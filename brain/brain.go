package brain // import "cirello.io/gochatbot/brain"

import "sync"

// BrainMemory is the basic memory facility for the gobot.
type BrainMemory struct {
	mu    sync.Mutex             // serializes items access
	items map[string]interface{} // rule name + key
}

// Brain constructs BrainMemory
func Brain() *BrainMemory {
	b := &BrainMemory{
		items: make(map[string]interface{}),
	}
	return b
}

// Save stores into Brain some arbritary value.
func (b *BrainMemory) Save(ruleName, key string, value interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()

	k := fullKeyName(ruleName, key)
	b.items[k] = value
}

// Read reads from Brain some arbritary value.
func (b *BrainMemory) Read(ruleName, key string) interface{} {
	b.mu.Lock()
	defer b.mu.Unlock()

	k := fullKeyName(ruleName, key)
	v, ok := b.items[k]
	if !ok {
		return nil
	}
	return v
}

func (b *BrainMemory) Error() error {
	return nil
}

func fullKeyName(ruleName, key string) string {
	return "\x02" + ruleName + "\x03" + "\x02" + key + "\x03"
}
