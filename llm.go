package main

import (
	"fmt"
	"os"
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

var aLLMInited bool
var aLLM autog.LLM
var openaiLLM *llm.OpenAi
var ollamaLLM *llm.Ollama

func GetEmbeddModel(cfg *Configs) autog.EmbeddingModel {
	GetLLM(cfg)
	if cfg.ApiVendor == VendorOpenAI {
		return openaiLLM
	} else if cfg.ApiVendor == VendorOllama {
		return ollamaLLM
	} else {
		fmt.Println("ApiVendor not supported!")
		os.Exit(0)
	}
	return nil
}

func GetLLM(cfg *Configs) autog.LLM {
	var err error
	if !aLLMInited {
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
			openaiLLM = &llm.OpenAi{ 
				ApiBase: cfg.ApiBase, 
				Model: cfg.Model,
				ModelWeak: cfg.Model,
				ModelEmbedding: cfg.ModelEmbed,
				ApiKey: cfg.ApiKey,
			}
			err = openaiLLM.InitLLM()
			aLLM = openaiLLM
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
			ollamaLLM = &llm.Ollama{ 
				ApiBase: cfg.ApiBase, 
				Model: cfg.Model,
				ModelWeak: cfg.Model,
				ModelEmbedding: cfg.ModelEmbed,
			}
			err = ollamaLLM.InitLLM()
			aLLM = ollamaLLM
		} else {
			fmt.Println("ApiVendor not supported!")
			os.Exit(0)
		}
		if err != nil {
			fmt.Printf("LLM init ERROR: %s\n", err)
			os.Exit(0)
		}
		aLLMInited = true
	}
	
	return aLLM
}