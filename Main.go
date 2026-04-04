package main

import (
	"fmt"

	"nanoagent/config"
	"nanoagent/llm"
)

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		fmt.Println("load config error:", err)
		return
	}

	client := llm.NewClient(cfg)

	messages := []llm.ChatMessage{
		{Role: "user", Content: "作为一名营销专家，请为我的产品创作一个吸引人的口号"},
		{Role: "assistant", Content: "当然，要创作一个吸引人的口号，请告诉我一些关于您产品的信息"},
		{Role: "user", Content: "智谱AI 开放平台"},
	}

	fmt.Println("=== response begin ===")
	err = client.InvokeMessage(messages, func(content string) {
		fmt.Print(content)
	})
	if err != nil {
		fmt.Println("invoke error:", err)
		return
	}
	fmt.Println("\n=== response end ===")
}