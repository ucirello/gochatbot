// +build all bolt

package brain // import "cirello.io/gochatbot/brain"

import (
	"log"

	"github.com/boltdb/bolt"
)

func init() {
	availableDrivers = append(availableDrivers, func(getenv func(string) string) (Memorizer, bool) {
		log.Println("brain: trying registering bolt")
		memo, ok := BoltFromEnv(getenv)
		if ok {
			log.Println("brain: registered bolt")
		} else {
			log.Println("brain: if you want BoltDB enabled, please set a valid value for the environment variable", boltMemoryFilename)
		}
		return memo, ok
	})
}

type BoltMemory struct {
	brain *brainMemory
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
func (b *BoltMemory) Save(ruleName, key string, value []byte) {
	b.brain.Save(ruleName, key, value)

	b.bolt.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(ruleName))
		if err != nil {
			log.Println("bolt: error saving:", err)
			return err
		}
		return b.Put([]byte(key), value)
	})
}

// Read reads from Brain some arbritary value.
func (b *BoltMemory) Read(ruleName, key string) []byte {
	v := b.brain.Read(ruleName, key)
	if len(v) > 0 {
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

	return found
}
