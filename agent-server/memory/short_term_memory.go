package memory

// ShortTermMemory 短期记忆实现
type ShortTermMemory struct {
	items []MemoryItem
}

// NewShortTermMemory 创建短期记忆
func NewShortTermMemory() *ShortTermMemory {
	return &ShortTermMemory{
		items: make([]MemoryItem, 0),
	}
}

// Add 添加记忆
func (m *ShortTermMemory) Add(item MemoryItem) {
	m.items = append(m.items, item)
}

// GetAll 获取所有记忆
func (m *ShortTermMemory) GetAll() []MemoryItem {
	return m.items
}