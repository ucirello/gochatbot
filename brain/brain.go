package brain // import "cirello.io/gobot/brain"

import "sync"

// Brain is the basic memory facility for the gobot.
type Brain struct {
	mu    sync.Mutex             // serializes items access
	items map[string]interface{} // rule name + key
}

// New constructs Brain
func New() *Brain {
	b := &Brain{
		items: make(map[string]interface{}),
	}
	return b
}

// Save stores into Brain some arbritary value.
func (b *Brain) Save(ruleName, key string, value interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()

	k := fullKeyName(ruleName, key)
	b.items[k] = value
}

// Read reads from Brain some arbritary value.
func (b *Brain) Read(ruleName, key string) interface{} {
	b.mu.Lock()
	defer b.mu.Unlock()

	k := fullKeyName(ruleName, key)
	v, ok := b.items[k]
	if !ok {
		return nil
	}
	return v
}

func fullKeyName(ruleName, key string) string {
	return "\x02" + ruleName + "\x03" + "\x02" + key + "\x03"
}
