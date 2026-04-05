package memory

type MemoryItem struct {
	ID         int
	Content    string
	Vector     []float32
	Timestamp  int64
	Importance float32
	Role       string
}