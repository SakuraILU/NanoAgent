package memory

import (
	"math"
	"sort"
	"time"

	"agent-server/client"
	"agent-server/config"
)

type ShortTermMemory struct {
	items         []MemoryItem
	embedClient   *client.QueryAnalyzeClient
	nextID        int
	recallTopK    int
}

func NewShortTermMemory() *ShortTermMemory {
	cfg := config.GetConfig()
	return &ShortTermMemory{
		items:       make([]MemoryItem, 0),
		embedClient: client.NewQueryAnalyzeClient(),
		nextID:      1,
		recallTopK:  cfg.Memory.RecallTopK,
	}
}

func (m *ShortTermMemory) AddAll(msgs []client.ChatMessage, importance float32) error {
	for _, msg := range msgs {
		err := m.Add(msg, importance)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *ShortTermMemory) Add(msg client.ChatMessage, importance float32) error {
	embeddings, err := m.embedClient.Embed([]string{msg.Content}, true)
	if err != nil {
		return err
	}

	var vector []float32
	if len(embeddings) > 0 && len(embeddings[0]) > 0 {
		vector = make([]float32, len(embeddings[0]))
		for i, v := range embeddings[0] {
			vector[i] = float32(v)
		}
	}

	item := MemoryItem{
		ID:         m.nextID,
		Content:    msg.Content,
		Role:       msg.Role,
		Vector:     vector,
		Timestamp:  time.Now().Unix(),
		Importance: importance,
	}
	m.nextID++
	m.items = append(m.items, item)
	return nil
}

// Recall 向量召回，返回相关性排序的前 num 个记忆
func (m *ShortTermMemory) Recall(query string, num int) ([]MemoryItem, error) {
	if num <= 0 {
		num = m.recallTopK
	}
	if len(m.items) == 0 {
		return []MemoryItem{}, nil
	}

	embeddings, err := m.embedClient.Embed([]string{query}, true)
	if err != nil {
		return nil, err
	}
	var queryVector []float32
	if len(embeddings) > 0 && len(embeddings[0]) > 0 {
		queryVector = make([]float32, len(embeddings[0]))
		for i, v := range embeddings[0] {
			queryVector[i] = float32(v)
		}
	}

	type scoredItem struct {
		item  MemoryItem
		score float32
	}
	scored := make([]scoredItem, 0, len(m.items))
	for _, item := range m.items {
		if len(item.Vector) > 0 && len(queryVector) > 0 {
			score := cosineSimilarity(queryVector, item.Vector)
			scored = append(scored, scoredItem{item: item, score: score})
		}
	}

	// 按相似度降序排序
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// 取前 num 个
	result := make([]MemoryItem, 0, num)
	for i := 0; i < len(scored) && i < num; i++ {
		result = append(result, scored[i].item)
	}

	return result, nil
}

func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}
	var dotProduct, normA, normB float32
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dotProduct / (sqrt32(normA) * sqrt32(normB))
}

func sqrt32(x float32) float32 {
	return float32(math.Sqrt(float64(x)))
}

func (m *ShortTermMemory) GetAll() []MemoryItem {
	return m.items
}

func (m *ShortTermMemory) Clear() {
	m.items = make([]MemoryItem, 0)
	m.nextID = 1
}