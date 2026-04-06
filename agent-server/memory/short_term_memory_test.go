package memory

import (
	"fmt"
	"testing"

	"agent-server/client"
	"agent-server/config"
)

func TestShortTermMemory(t *testing.T) {
	// 加载配置
	_, err := config.LoadConfig("../resource/config.yaml")
	if err != nil {
		t.Fatalf("Load config failed: %v", err)
	}

	// 创建短期记忆
	memory := NewShortTermMemory()
	fmt.Println("=== ShortTermMemory 测试 ===")

	// 测试 Add
	fmt.Println("\n--- 测试 Add ---")

	items := []struct {
		content    string
		role       string
		importance float32
	}{
		{"我喜欢吃苹果和香蕉", "user", 0.8},
		{"明天下午3点有个会议", "user", 0.9},
		{"北京今天天气晴朗", "assistant", 0.3},
		{"用户偏好使用中文交流", "user", 0.95},
		{"最近在学习Go语言开发", "user", 0.7},
	}

	for _, item := range items {
		err := memory.Add(client.ChatMessage{
			Content: item.content, Role: item.role,
		}, item.importance)
		if err != nil {
			t.Fatalf("Add failed for '%s': %v", item.content, err)
		}
		fmt.Printf("Added: [%s] %s (importance: %.2f)\n", item.role, item.content, item.importance)
	}

	// 验证存储数量
	all := memory.GetAll()
	if len(all) != len(items) {
		t.Fatalf("Expected %d items, got %d", len(items), len(all))
	}
	fmt.Printf("\nTotal items in memory: %d\n", len(all))

	// 测试 Recall
	fmt.Println("\n--- 测试 Recall ---")

	queries := []struct {
		query string
		topK  int
	}{
		{"水果相关的话题", 3},
		{"日程安排", 2},
		{"编程学习", 5},
	}

	for _, q := range queries {
		fmt.Printf("\nQuery: '%s' (topK: %d)\n", q.query, q.topK)

		results, err := memory.Recall(q.query, q.topK)
		if err != nil {
			t.Fatalf("Recall failed for '%s': %v", q.query, err)
		}

		if len(results) == 0 {
			fmt.Println("  No results found")
			continue
		}

		for i, item := range results {
			fmt.Printf("  %d. [%s] %s (importance: %.2f)\n", i+1, item.Role, item.Content, item.Importance)
		}

		// 验证返回数量不超过 topK
		if len(results) > q.topK {
			t.Errorf("Expected at most %d results, got %d", q.topK, len(results))
		}
	}

	// 测试使用默认 topK (从配置读取)
	fmt.Println("\n--- 测试默认 topK ---")
	results, err := memory.Recall("用户信息", 0) // 0 表示使用默认值
	if err != nil {
		t.Fatalf("Recall with default topK failed: %v", err)
	}
	fmt.Printf("Default topK recall returned %d results (config: %d)\n", len(results), config.GetConfig().Memory.RecallTopK)

	// 测试 Clear
	fmt.Println("\n--- 测试 Clear ---")
	memory.Clear()
	if len(memory.GetAll()) != 0 {
		t.Fatal("Clear failed, memory not empty")
	}
	fmt.Println("Memory cleared successfully")

	fmt.Println("\n=== 所有测试通过 ===")
}

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		a        []float32
		b        []float32
		expected float32
	}{
		{
			name:     "相同向量",
			a:        []float32{1.0, 0.0, 0.0},
			b:        []float32{1.0, 0.0, 0.0},
			expected: 1.0,
		},
		{
			name:     "正交向量",
			a:        []float32{1.0, 0.0, 0.0},
			b:        []float32{0.0, 1.0, 0.0},
			expected: 0.0,
		},
		{
			name:     "相反向量",
			a:        []float32{1.0, 0.0, 0.0},
			b:        []float32{-1.0, 0.0, 0.0},
			expected: -1.0,
		},
		{
			name:     "相似向量",
			a:        []float32{1.0, 1.0, 1.0},
			b:        []float32{1.0, 1.0, 0.5},
			expected: 0.96, // 约 0.96
		},
		{
			name:     "长度不等",
			a:        []float32{1.0, 0.0},
			b:        []float32{1.0, 0.0, 0.0},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cosineSimilarity(tt.a, tt.b)
			// 允许一定误差
			if tt.name == "相似向量" {
				if result < 0.95 || result > 0.97 {
					t.Errorf("cosineSimilarity() = %v, expected ~%v", result, tt.expected)
				}
			} else {
				if abs32(result-tt.expected) > 0.0001 {
					t.Errorf("cosineSimilarity() = %v, expected %v", result, tt.expected)
				}
			}
			fmt.Printf("%s: %.4f\n", tt.name, result)
		})
	}
}

func abs32(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}
