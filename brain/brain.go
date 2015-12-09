package brain // import "cirello.io/gochatbot/brain"

import "sync"

type brainMemory struct {
	mu    sync.Mutex        // serializes items access
	items map[string][]byte // rule name + key
}

// Brain constructs brainMemory
func Brain() *brainMemory {
	b := &brainMemory{
		items: make(map[string][]byte),
	}
	return b
}

// Save stores into Brain some arbritary value.
func (b *brainMemory) Save(ruleName, key string, value []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()

	k := fullKeyName(ruleName, key)
	b.items[k] = value
}

// Read reads from Brain some arbritary value.
func (b *brainMemory) Read(ruleName, key string) []byte {
	b.mu.Lock()
	defer b.mu.Unlock()

	k := fullKeyName(ruleName, key)
	v, ok := b.items[k]
	if !ok {
		return []byte{}
	}
	return v
}

func (b *brainMemory) Error() error {
	return nil
}

func fullKeyName(ruleName, key string) string {
	return "\x02" + ruleName + "\x03" + "\x02" + key + "\x03"
}
