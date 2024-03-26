package main

import (
	"os"
	"fmt"
	"os/signal"
	"syscall"
	"context"
	_ "embed"
	"github.com/autogorg/autog"
	"github.com/autogorg/autog/rag"
)

//go:embed CHROME_PROMPT.md
var systemStr string


type ChromeAgent struct {
	autog.Agent
	Rag   *autog.Rag
	Query string
	ShowLog  func (level int, content string)
}

var chromeAgent *ChromeAgent = &ChromeAgent{}

func GetChromeAgent() *ChromeAgent {
	return chromeAgent
}

var systemPrompt *autog.PromptItem = &autog.PromptItem{
	GetPrompt : func (query string) (role string, prompt string) {
		return autog.ROLE_SYSTEM, systemStr
	},
}

var longHistory *autog.PromptItem =  &autog.PromptItem{
	GetMessages : func (query string) []autog.ChatMessage {
		return chromeAgent.GetLongHistory()
	},
}

var shortHistory *autog.PromptItem =  &autog.PromptItem{
	GetMessages : func (query string) []autog.ChatMessage {
		msgs := chromeAgent.GetShortHistory()

		cxt, cancel := context.WithCancel(context.Background())
		defer cancel()
	
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT)
	
		go func() {
			<-sigChan
			cancel()
		}()

		// HTML太大，不能完整的送给大模型，所以这里进行RAG增强检索，因为页面会刷新，所以每次都重新间索引
		ShowAgentLog(1, fmt.Sprintf("Indexing HTML...\n"))

		splitter := &rag.TextSplitter{
			ChunkSize : 8192,
		}
		err := chromeAgent.Rag.Indexing(cxt, "/html", GetHtmlContext(), splitter, true)
		if err != nil {
			ShowAgentLog(-1, fmt.Sprintf("RAG Indexing ERROR: %s\n", err))
			return msgs
		}

		ShowAgentLog(1, fmt.Sprintf("Retrieval HTML...\n"))
		var scoredss []autog.ScoredChunks
		scoredss, err  = chromeAgent.Rag.Retrieval(cxt, "/html", []string{query}, 3)
		if err != nil {
			ShowAgentLog(-1, fmt.Sprintf("RAG Retrieval ERROR: %s\n", err))
			return msgs
		}

		content := "请基于以下内容进行回答，你的所有操作来源仅限于如下HTML内容和我的提问内容，不允许进行任何假设!!!我的问题将在下一条消息中发给你!\nHTML:\n"
		for _, scoreds := range scoredss {
			for _, scored := range scoreds {
				content += fmt.Sprintf("...\n%s\n...\n", scored.Chunk.GetContent())
			}
		}

		ShowAgentLog(1, fmt.Sprintf("Sending...\n"))

		msgs = append(msgs, autog.ChatMessage{Role:autog.ROLE_USER, Content: content})
		msgs = append(msgs, autog.ChatMessage{Role:autog.ROLE_ASSISTANT, Content: "OK"})
		return msgs
	},
}

var summaryPrompt *autog.PromptItem =  &autog.PromptItem{
	GetPrompt : func (query string) (role string, prompt string) {
		return "", "用500字以内总计一下我们的历史对话！"
	},
}

var summaryPrefix *autog.PromptItem =  &autog.PromptItem{
	GetPrompt : func (query string) (role string, prompt string) {
		return "", "我们的历史对话总结如下："
	},
}

var input *autog.Input = &autog.Input{
	ReadContent: func() string {
		return chromeAgent.Query
	},
}

var output *autog.Output = &autog.Output{}


var doaction *autog.DoAction = &autog.DoAction {
	Do: func (content string) (ok bool, reflection string) {
		action := GetChromeAction()
		if action.NeedRun(content) {
			cok, _, payload := action.Check(content)
			if cok {
				rok, rerr := action.Run(content, payload)
				return rok, rerr
			}
		}
		return true, ""
	},
}

func CreateMemoryRag(embedmodel autog.EmbeddingModel) *autog.Rag {
	memDB, err := rag.NewMemDatabase()
	if err != nil {
		fmt.Printf("CreateMemoryRag ERROR: %s\n", err)
		os.Exit(0)
	}

	memRag := &autog.Rag{
		Database: memDB,
		EmbeddingModel: embedmodel,
		EmbeddingBatch: 5,
	}

	return memRag
}

func ShowAgentLog(level int, str string) {
	if chromeAgent == nil || chromeAgent.ShowLog == nil {
		return
	}
	chromeAgent.ShowLog(level, str)
}

func RunChromeAgent(llm autog.LLM, embedmodel autog.EmbeddingModel, query string) {
	cxt, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	go func() {
		<-sigChan
		cancel()
	}()

	if chromeAgent.Rag == nil {
		chromeAgent.Rag = CreateMemoryRag(embedmodel)
	}

	chromeAgent.Query = query
	chromeAgent.Prompt(systemPrompt, longHistory, shortHistory).
    ReadQuestion(cxt, input, output).
    AskLLM(llm, true). // `true` means stream response
    WaitResponse(cxt).
    Action(doaction).
    Reflection(nil, 3). // `nil` == no reflection
    Summarize(cxt, summaryPrompt, summaryPrefix, false) // `false` == disable force summary
}
