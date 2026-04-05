package client

import (
	"fmt"
	"testing"
	"time"

	Config "agent-server/config"
)

func TestQueryAnalyzeClient(t *testing.T) {
	// 加载配置
	_, err := Config.LoadConfig("../resource/config.yaml")
	if err != nil {
		t.Fatalf("Load config failed: %v", err)
	}

	// 创建客户端（从配置读取）
	client := NewQueryAnalyzeClient()

	// 测试健康检查
	fmt.Println("=== 测试健康检查 ===")
	health, err := client.Health()
	if err != nil {
		t.Fatalf("Health check failed: %v", err)
	}
	fmt.Printf("Health: %s\n", health)

	// 测试分词
	fmt.Println("\n=== 测试分词 ===")
	segmentResults, err := client.Segment([]string{"今天我想去北京旅游，但是不知道北京有什么好玩的地方"}, 5)
	if err != nil {
		t.Fatalf("Segment failed: %v", err)
	}
	for _, r := range segmentResults {
		fmt.Printf("Text: %s\n", r.GetText())
		for _, w := range r.GetWords() {
			fmt.Printf("  - %s: %d\n", w.GetWord(), w.GetCount())
		}
	}

	// 测试 NER
	fmt.Println("\n=== 测试 NER ===")
	nerResults, err := client.NER([]string{"马云在北京创立了阿里巴巴公司"})
	if err != nil {
		t.Fatalf("NER failed: %v", err)
	}
	for _, r := range nerResults {
		fmt.Printf("Text: %s\n", r.GetText())
		for _, e := range r.GetEntities() {
			fmt.Printf("  - %s (%s): [%d, %d]\n", e.GetText(), e.GetType(), e.GetStartPos(), e.GetEndPos())
		}
	}

	// 测试向量化
	fmt.Println("\n=== 测试向量化 ===")
	embeddings, err := client.Embed([]string{"你好世界", "今天天气不错"}, true)
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}
	fmt.Printf("Embedding count: %d, dimension: %d\n", len(embeddings), len(embeddings[0]))

	// 测试相似度
	fmt.Println("\n=== 测试相似度 ===")
	sim, err := client.Similarity("北京天气", "首都天气")
	if err != nil {
		t.Fatalf("Similarity failed: %v", err)
	}
	fmt.Printf("Similarity: %.4f\n", sim)

	// 测试综合分析
	fmt.Println("\n=== 测试综合分析 ===")
	analysisResults, err := client.Analyze(
		[]string{"张三想去上海出差，查询一下机票价格"},
		true,  // keywords
		true,  // ner
		false, // embedding
		5,     // topK
	)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	for _, r := range analysisResults {
		fmt.Printf("Text: %s\n", r.GetText())
		if r.Keywords != nil {
			fmt.Println("Keywords:")
			for _, w := range r.Keywords {
				fmt.Printf("  - %s: %d\n", w.GetWord(), w.GetCount())
			}
		}
		if r.Entities != nil {
			fmt.Println("Entities:")
			for _, e := range r.Entities {
				fmt.Printf("  - %s (%s)\n", e.GetText(), e.GetType())
			}
		}
	}
}

func TestQueryAnalyzeClientWithAddr(t *testing.T) {
	// 测试指定地址创建客户端
	client := NewQueryAnalyzeClientWithAddr("localhost:9090", 30*time.Second)

	health, err := client.Health()
	if err != nil {
		t.Fatalf("Health check failed: %v", err)
	}
	fmt.Printf("Health: %s\n", health)
}