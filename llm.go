package main

import (
	"github.com/autogorg/autog"
	"github.com/autogorg/autog/llm"
)

const (
	VendorOpenAI     = "openai"
	VendorOllama     = "ollama"

	OllamaApiBase    = "http://localhost:11434"
	OllamaModel      = "gemma:2b"
	OllamaModelEmbed = "nomic-embed-text"

	OpenAIApiBase    = "https://api.chatpp.org/v1"
	OpenAIModel      = "gpt-4-turbo-preview"
	OpenAIModelEmbed = "text-embedding-3-large"

	OpenAIApiKey     = "sk-***"
)

var llmInited
var llm autog.LLM

func GetLLM(cfg *Configs) autog.LLM {
	if !llmInited {
		if cfg.ApiVendor == VendorOpenAI {
			if len(cfg.ApiBase) <= 0 {
				cfg.ApiBase = OpenAIApiBase
			}
			if len(cfg.Model) <= 0 {
				cfg.Model = OpenAIModel
			}
			if len(cfg.Model) <= 0 {
				cfg.ModelEmbed = OpenAIModelEmbed
			}
			llm = &llm.OpenAi{ 
				ApiBase: cfg.ApiBase, 
				Model: cfg.Model,
				ModeWeak: cfg.Model,
				ModelEmbed: cfg.ModelEmbed,
				ApiKey: Ocfg.ApiKey,
			}
		} else if cfg.ApiVendor == VendorOllama {
			if len(cfg.ApiBase) <= 0 {
				cfg.ApiBase = OpenAIApiBase
			}
			if len(cfg.Model) <= 0 {
				cfg.Model = OpenAIModel
			}
			if len(cfg.Model) <= 0 {
				cfg.ModelEmbed = OpenAIModelEmbed
			}
			llm = &llm.Ollama{ 
				ApiBase: cfg.ApiBase, 
				Model: cfg.Model,
				ModeWeak: cfg.Model,
				ModelEmbed: cfg.ModelEmbed,
			}
		} else {
			fmt.Println("ApiVendor not supported!")
			os.Exit(0)
		}
		err := llm.InitLLM()
		if err != nil {
			fmt.Printf("LLM init ERROR: %s\n", err)
			os.Exit(0)
		}
		llmInited = true
	}
	
	return llm
}