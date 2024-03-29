package main

import (
	"flag"
	"fmt"
	"os"
	_ "embed"
)

//go:embed VERSION
var Version string

type Configs struct {
	Version            bool
	ApiVendor          string     `json:"api-vendor"`
	ApiBase            string     `json:"api-base"`
	ApiKey             string     `json:"api-key"`
	Model              string     `json:"model"`
	ModelEmbed         string     `json:"model-embed"`
	ChunkSize          int        `json:"chunk-size"`
	ChunkOverlap       int        `json:"chunk-overlap"`
	ChunkBatch         int        `json:"chunk-batch"`
	ChunkRoutines      int        `json:"chunk-routines"`
	TopK               int        `json:"topk"`
	URL                string     `json:"url"`
}

var cfgInited bool
var cfg Configs

func GetConfigs() *Configs{
	if !cfgInited {
		cfgInited = true
		ParseConfigs()
	}
	return &cfg
}

func ClearConfigs(cfgs *Configs) {
	*cfgs = Configs{}
}

func getenvOrDefault(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func ParseConfigs() *Configs {
    ClearConfigs(&cfg)

	flag.BoolVar(&cfg.Version, "version", false, "Show the version number")

	flag.StringVar(&cfg.ApiVendor, "api-vendor", getenvOrDefault("API_VENDOR", VendorOpenAI), "Specify the vendor decide which API type to use (openai or ollama)")
	flag.StringVar(&cfg.ApiBase, "api-base", getenvOrDefault("API_BASE", ""), "Specify the api base url")
	flag.StringVar(&cfg.Model, "model", getenvOrDefault("MODEL", ""), "Specify the main model to use")
	flag.StringVar(&cfg.ModelEmbed, "model-embed", getenvOrDefault("MODEL_EMBED", ""), "Specify the embedding model to use")
	flag.StringVar(&cfg.ApiKey, "api-key", getenvOrDefault("API_KEY", ""), "Specify the api key")

	flag.IntVar(&cfg.ChunkSize, "chunk-size", 1024, "Chunk size for split text")
	flag.IntVar(&cfg.ChunkOverlap, "chunk-overlap", 1, "Chunk overlap for split text (percent)")
	flag.IntVar(&cfg.ChunkBatch, "chunk-batch", 2, "Chunk batch for split text")
	flag.IntVar(&cfg.ChunkRoutines, "chunk-routines", 5, "Chunk routines for split text")

	flag.IntVar(&cfg.TopK, "topk", 10, "TopK for RAG")
	flag.StringVar(&cfg.URL, "url", "", "URL to open")

    flag.Parse()

    if cfg.Version {
        fmt.Println(Version)
        os.Exit(0)
    }

	if cfg.ChunkSize < 512 {
		cfg.ChunkSize = 512
	}
	if cfg.ChunkOverlap < 1 {
		cfg.ChunkOverlap = 1
	}
	if cfg.ChunkOverlap > 99 {
		cfg.ChunkOverlap = 99
	}
	if cfg.TopK < 1 {
		cfg.TopK = 1
	}

    return &cfg
}