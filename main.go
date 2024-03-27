package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"autochrome/readline"
	"github.com/autogorg/autog"
)

type MultilineState int

const (
	MultilineNone MultilineState = iota
	MultilinePrompt
)

func main() {
	cfg := GetConfigs()
	llm := GetLLM(cfg)
	embedModel := GetEmbeddModel(cfg)

	if len(cfg.URL) <= 0 {
		fmt.Println("URL is empty! -h for help!")
		return
	}

	
	action := GetChromeAction()
	action.ShowLog = func (level int, content string) {
		if level < 0 {
			fmt.Printf("%s", Red(content))
		} else if level > 0 {
			fmt.Printf("%s", BrightBlack(content))
		} else {
			fmt.Printf("%s", content)
		}
	}

	agent := GetChromeAgent()
	agent.ShowLog = action.ShowLog
	

	openerr := ChromeActionOpenUrl(cfg.URL)
	if openerr != nil {
		fmt.Printf("Open url error : %s\n", openerr)
		return
	}

	usage := func() {
		fmt.Fprintln(os.Stderr, "Available Commands:")
		fmt.Fprintln(os.Stderr, "  /bye            Exit")
		fmt.Fprintln(os.Stderr, "  /?, /help       Help for a command")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Use \"\"\" to begin a multi-line message.")
		fmt.Fprintln(os.Stderr, "")
	}

	scanner, err := readline.New(readline.Prompt{
		Prompt:         ">>> ",
		AltPrompt:      "... ",
		Placeholder:    "Send a message (/? for help)",
		AltPlaceholder: `Use """ to end multi-line input`,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	fmt.Print(readline.StartBracketedPaste)
	defer fmt.Printf(readline.EndBracketedPaste)

	var sb strings.Builder
	var multiline MultilineState

	output.WriteStreamStart = func() *strings.Builder {
		if output.AgentStage == autog.AsWaitResponse {
			fmt.Printf(Yellow("## --- AI ---\n"))
		}
		return &strings.Builder{}
	}
	output.WriteStreamError = func(contentbuf *strings.Builder, status autog.LLMStatus, errstr string) {
		fmt.Printf("\n%s\n", Red(errstr))
	}
	output.WriteStreamEnd = func(contentbuf *strings.Builder) {
		if output.AgentStage == autog.AsWaitResponse {
			fmt.Println()
		}
	}
	output.WriteStreamDelta = func(contentbuf *strings.Builder, delta string) {
		if output.AgentStage == autog.AsWaitResponse {
			fmt.Print(Cyan(delta))
		}
	}


	for {
		line, err := scanner.Readline()
		switch {
		case errors.Is(err, io.EOF):
			fmt.Println()
			return
		case errors.Is(err, readline.ErrInterrupt):
			if line == "" {
				fmt.Println("\nUse Ctrl + d or /bye to exit.")
			}

			scanner.Prompt.UseAlt = false
			sb.Reset()

			continue
		case err != nil:
			fmt.Fprintln(os.Stderr, err)
			return
		}
		switch {
		case multiline != MultilineNone:
			// check if there's a multiline terminating string
			before, ok := strings.CutSuffix(line, `"""`)
			sb.WriteString(before)
			if !ok {
				fmt.Fprintln(&sb)
				continue
			}

			multiline = MultilineNone
			scanner.Prompt.UseAlt = false
		case strings.HasPrefix(line, `"""`):
			line := strings.TrimPrefix(line, `"""`)
			line, ok := strings.CutSuffix(line, `"""`)
			sb.WriteString(line)
			if !ok {
				// no multiline terminating string; need more input
				fmt.Fprintln(&sb)
				multiline = MultilinePrompt
				scanner.Prompt.UseAlt = true
			}
		case scanner.Pasting:
			fmt.Fprintln(&sb, line)
			continue
		case strings.HasPrefix(line, "/help"), strings.HasPrefix(line, "/?"):
			usage()
		case strings.HasPrefix(line, "/exit"), strings.HasPrefix(line, "/bye"):
			return
		case strings.HasPrefix(line, "/"):
			sb.WriteString(line)
		default:
			sb.WriteString(line)
		}

		if sb.Len() > 0 && multiline == MultilineNone {
			fmt.Printf(Green("## ---USER---\n"))
			fmt.Printf("%s\n", BrightWhite(sb.String()))
			RunChromeAgent(cfg, llm, embedModel, sb.String())
			sb.Reset()
		}
	}
}