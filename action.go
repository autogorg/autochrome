package main

import (
	"os"
	"fmt"
	"regexp"
	"autochrome/executor"
	"autochrome/executor/chrome"
	"github.com/autogorg/autog"
)

type ChromeAction struct {
	autog.Action
	Executor *executor.Executor
	Chrome   *chrome.Chrome
	ShowLog  func (level int, content string)
}

var chromeActionInited bool
var chromeAction *ChromeAction

func GetChromeAction() *ChromeAction {
	if !chromeActionInited {
		chromeAction = &ChromeAction{}
		exec, aerr := executor.NewExecutor()
		if aerr != nil {
			fmt.Printf("Executor create ERROR: %s\n", aerr)
			os.Exit(0)
		}
		chro, berr := exec.ChromeNew()
		if berr != nil {
			fmt.Printf("Chrome create ERROR: %s\n", berr)
			os.Exit(0)
		}
		chromeAction.Executor = exec
		chromeAction.Chrome   = chro
		chromeAction.NeedRun  = ChromeActionNeedRun
		chromeAction.Check    = ChromeActionCheck
		chromeAction.Run      = ChromeActionRun
		chromeActionInited = true
	}
	return chromeAction
}

func GetHtmlContext() string {
	if chromeAction.Executor != nil {
		html, err := chromeAction.Executor.ChromeGetHtml()
		if err == nil {
			return fmt.Sprintf("HTML:\n%s\n", html)
		}
	}
	return "HTML:\nEmpty!\n"
}

func ChromeActionNeedRun(content string) bool {
	codeBlockPattern := regexp.MustCompile(`(?s)\x60\x60\x60go\n(.*?)\n\x60\x60\x60`)
	match := codeBlockPattern.FindStringSubmatch(content)
	return match != nil && len(match) > 1
}

func ChromeActionCheck(content string) (ok bool, err string, payload interface{}) {
	codeBlockPattern := regexp.MustCompile(`(?s)\x60\x60\x60go\n(.*?)\n\x60\x60\x60`)
	match := codeBlockPattern.FindStringSubmatch(content)
	if match != nil && len(match) > 1 {
		codeBlock := match[1]
		return true, "", codeBlock
	}
	return false, "", ""
}

func ShowActionLog(level int, str string) {
	if chromeAction == nil || chromeAction.ShowLog == nil {
		return
	}
	chromeAction.ShowLog(level, str)
}

func ChromeActionRun(content string, payload interface{}) (ok bool, err string) {
	if codeBlock, ok := payload.(string); ok && len(codeBlock) > 0 {
		ShowActionLog(1, fmt.Sprintf("ACTION: Processing...\n"))
		err := chromeAction.Executor.ChromeRunTasks(codeBlock)
		if err != nil {
			ShowActionLog(-1, fmt.Sprintf("ACTION: ERROR -- %s\n", err))
		} else {
			ShowActionLog(1, fmt.Sprintf("ACTION: Success!\n"))
		}
	}
	return true, ""
}

func ChromeActionOpenUrl(url string) error {
	err := chromeAction.Executor.ChromeSetUrl(url)
	if err != nil {
		return err
	}
	err = chromeAction.Executor.ChromeNewTab()
	if err != nil {
		return err
	}
	err = chromeAction.Executor.ChromeNavigateAndWaitReady()
	if err != nil {
		return err
	}
	return err
}