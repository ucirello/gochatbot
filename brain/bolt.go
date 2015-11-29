// +build all bolt

package brain // import "cirello.io/gochatbot/brain"

import (
	"encoding/json"

	"github.com/boltdb/bolt"
)

func init() {
	AvailableDrivers = append(AvailableDrivers, func(getenv func(string) string) (Memorizer, bool) {
		return BoltFromEnv(getenv)
	})
}

type BoltMemory struct {
	brain *BrainMemory
	bolt  *bolt.DB

	err error
}

const (
	boltMemoryFilename = "GOCHATBOT_BOLT_FILENAME"
)

func BoltFromEnv(getenv func(string) string) (*BoltMemory, bool) {
	fn := getenv(boltMemoryFilename)
	if fn == "" {
		return nil, false
	}
	return Bolt(fn), true
}

func Bolt(dbFn string) *BoltMemory {
	b := &BoltMemory{
		brain: Brain(),
	}

	db, err := bolt.Open(dbFn, 0600, nil)
	if err != nil {
		b.err = err
		return nil
	}

	b.bolt = db
	return b
}

func (b *BoltMemory) Error() error {
	return b.err
}

// Save stores into Brain some arbritary value.
func (b *BoltMemory) Save(ruleName, key string, value interface{}) {
	b.brain.Save(ruleName, key, value)

	b.bolt.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(ruleName))
		if err != nil {
			return err
		}
		output, err := json.Marshal(value)
		if err != nil {
			return err
		}
		return b.Put([]byte(key), output)
	})
}

// Read reads from Brain some arbritary value.
func (b *BoltMemory) Read(ruleName, key string) interface{} {
	v := b.brain.Read(ruleName, key)
	if v != nil {
		return v
	}

	var found []byte
	b.bolt.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ruleName))
		if b == nil {
			return nil
		}
		found = b.Get([]byte(key))

		return nil
	})

	var ret interface{}
	if err := json.Unmarshal(found, &ret); err != nil {
		return nil
	}
	return ret
}
