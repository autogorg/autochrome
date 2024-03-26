package main


func main() {
	cfg := GetConfigs()
	llm := GetLLM(cfg)
	RunChromeAgent(llm, cfg.Question)
}