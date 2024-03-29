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
	Cfg   *Configs
	Rag   *autog.Rag
	Query string
	LastHtml string
	LastHtmlContext string
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
		cxt, cancel := context.WithCancel(context.Background())

		sigChan := make(chan os.Signal, 1)
		done := make(chan bool)
	
		signal.Notify(sigChan, syscall.SIGINT)
	
		go func() {
			select {
			case <-sigChan:
				cancel()
			case <-cxt.Done():
			}
			done <- true
		}()
	
		defer func() {
			cancel()
			<-done
			signal.Stop(sigChan)
		}()

		msgs := chromeAgent.GetShortHistory()

		currentHtml := GetHtmlContext()

		if chromeAction.LastHtml != currentHtml {
			// HTML太大，不能完整的送给大模型，所以这里进行RAG增强检索，因为页面会刷新，所以每次都重新间索引
			ShowAgentLog(1, fmt.Sprintf("Indexing HTML...\n"))
			splitter := &rag.TextSplitter{
				ChunkSize: chromeAgent.Cfg.ChunkSize,
				Overlap: float64(chromeAgent.Cfg.ChunkOverlap)/float64(100.0),
				BreakStartChars: []rune { '<' },
				BreakEndChars:   []rune { '>' },
			}
	
			chromeAgent.Rag.EmbeddingCallback = func (stage autog.EmbeddingStage, texts []string, embeds []autog.Embedding, i, j int, finished, tried int, err error) bool {
				if stage != autog.EmbeddingStageIndexing {
					return tried < 1
				}
	
				if err != nil {
					ShowAgentLog(1, fmt.Sprintf("Embedding HTML (%d/%d) Retry...\n", len(texts), finished))
					return tried < 1
				}
				ShowAgentLog(1, fmt.Sprintf("Embedding HTML (%d/%d) Done!\n", len(texts), finished))
				return false
			}
	
			err := chromeAgent.Rag.Indexing(cxt, "/html", currentHtml, splitter, true)
			if err != nil {
				ShowAgentLog(-1, fmt.Sprintf("RAG Indexing ERROR: %s\n", err))
				return msgs
			}
		}

		ShowAgentLog(1, fmt.Sprintf("Retrieval HTML...\n"))
		var scoredss []autog.ScoredChunks
		scoredss, err  = chromeAgent.Rag.Retrieval(cxt, "/html", []string{query}, chromeAgent.Cfg.TopK)
		if err != nil {
			ShowAgentLog(-1, fmt.Sprintf("RAG Retrieval ERROR: %s\n", err))
			chromeAction.LastHtml = ""
			return msgs
		}

		content := "最新的HTML内容如下\nHTML:\n"
		for _, scoreds := range scoredss {
			for _, scored := range scoreds {
				content += fmt.Sprintf("...\n%s\n...\n", scored.Chunk.GetContent())
			}
		}

		chromeAgent.LastHtmlContext = content
		chromeAction.LastHtml = currentHtml

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
		return "问题：" + chromeAgent.Query
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

func CreateMemoryRag(embedmodel autog.EmbeddingModel, chunkBatch int, routines int) *autog.Rag {
	memDB, err := rag.NewMemDatabase()
	if err != nil {
		fmt.Printf("CreateMemoryRag ERROR: %s\n", err)
		os.Exit(0)
	}

	memRag := &autog.Rag{
		Database: memDB,
		EmbeddingModel: embedmodel,
		EmbeddingBatch: chunkBatch,
		EmbeddingRoutines: routines,
	}

	return memRag
}

func ShowAgentLog(level int, str string) {
	if chromeAgent == nil || chromeAgent.ShowLog == nil {
		return
	}
	chromeAgent.ShowLog(level, str)
}

func GetLastHtmlContext() string {
	if chromeAgent != nil {
		return chromeAgent.LastHtmlContext
	}
	return "Empty!\n"
}

func RunChromeAgent(cfg *Configs, llm autog.LLM, embedmodel autog.EmbeddingModel, query string) {
	cxt, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	done := make(chan bool)

	signal.Notify(sigChan, syscall.SIGINT)

    go func() {
        select {
        case <-sigChan:
            cancel()
        case <-cxt.Done():
        }
        done <- true
    }()

	defer func() {
		cancel()
		<-done
		signal.Stop(sigChan)
	}()

	if chromeAgent.Rag == nil {
		chromeAgent.Rag = CreateMemoryRag(embedmodel, cfg.ChunkBatch, cfg.ChunkRoutines)
	}
	chromeAgent.Cfg   = cfg
	chromeAgent.Query = query
	chromeAgent.Prompt(systemPrompt, longHistory, shortHistory).
    ReadQuestion(cxt, input, output).
    AskLLM(llm, true). // `true` means stream response
    WaitResponse(cxt).
    Action(doaction).
    Reflection(nil, 3). // `nil` == no reflection
    Summarize(cxt, summaryPrompt, summaryPrefix, false) // `false` == disable force summary
}
