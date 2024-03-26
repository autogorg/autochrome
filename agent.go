package main

import (
	"fmt"
	"strings"
	"context"
	_ "embed"
	"github.com/autogorg/autog"
)

//go:embed prompt/system.md
var systemStr string


type ChromeAgent struct {
	autog.Agent
	Query string
}

var chromeAgent *ChromeAgent = &ChromeAgent{}

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
		return chromeAgent.GetShortHistory()
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

var output *autog.Output = &autog.Output{
	WriteStreamStart: func() *strings.Builder {
		fmt.Println()
		return &strings.Builder{}
	},
	WriteStreamError: func(contentbuf *strings.Builder, status autog.LLMStatus, errstr string) {
		fmt.Printf("\n%s\n", errstr)
	},
	WriteStreamEnd: func(contentbuf *strings.Builder) {
		fmt.Println()
	},
}



func RunChromeAgent(llm autog.LLM, query string) {

	output.WriteStreamDelta = func(contentbuf *strings.Builder, delta string) {
		if output.AgentStage == autog.AsWaitResponse {
			fmt.Print(Cyan(delta))
		}
	}

	cxt := context.Background()
	chromeAgent.Query = query
	chromeAgent.
	Prompt(systemPrompt, longHistory, shortHistory).
    ReadQuestion(cxt, input, output).
    AskLLM(llm, true). // `true` means stream response
    WaitResponse(cxt).
    Action(nil).
    Reflection(nil, 3). // `nil` == no reflection
    Summarize(cxt, summaryPrompt, summaryPrefix, false) // `false` == disable force summary
}
