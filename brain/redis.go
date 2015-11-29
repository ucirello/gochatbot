// +build all redis

package brain // import "cirello.io/gochatbot/brain"

import (
	"strconv"

	redis "gopkg.in/redis.v3"
)

func init() {
	AvailableDrivers = append(AvailableDrivers, func(getenv func(string) string) (Memorizer, bool) {
		return RedisFromEnv(getenv)
	})
}

type RedisMemory struct {
	brain *BrainMemory
	db    *redis.Client

	err error
}

const (
	redisMemoryDatabase = "GOCHATBOT_REDIS_DATABASE"
	redisMemoryHostPort = "GOCHATBOT_REDIS_HOST"
	redisMemoryPassword = "GOCHATBOT_REDIS_PASSWORD"
)

func RedisFromEnv(getenv func(string) string) (*RedisMemory, bool) {
	var dbID int64 = 0
	rawDbID := getenv(redisMemoryDatabase)
	if rawDbID != "" {
		var err error
		dbID, err = strconv.ParseInt(rawDbID, 10, 0)
		if err != nil {
			return nil, false
		}
	}
	hostPort, password := getenv(redisMemoryHostPort), getenv(redisMemoryPassword)
	if hostPort == "" {
		return nil, false
	}
	r := Redis(hostPort, password, dbID)
	if r.err != nil {
		return nil, false
	}
	return r, true
}

func Redis(hostPort, password string, dbID int64) *RedisMemory {
	r := &RedisMemory{
		brain: Brain(),
	}

	r.db = redis.NewClient(&redis.Options{
		Addr:     hostPort,
		Password: password, // no password set
		DB:       dbID,     // use default DB
	})

	_, err := r.db.Ping().Result()
	if err != nil {
		r.err = err
	}

	return r
}

func (r *RedisMemory) Error() error {
	return r.err
}

func (r *RedisMemory) calculateKey(ruleName, key string) string {
	return "\x02" + ruleName + "\x03" + "\x02" + key + "\x03"
}

// Save stores into Brain some arbritary value.
func (r *RedisMemory) Save(ruleName, key string, value interface{}) {
	r.brain.Save(ruleName, key, value)

	err := r.db.Set(r.calculateKey(ruleName, key), value, 0).Err()
	if err != nil {
		r.err = err
		return
	}
}

// Read reads from Brain some arbritary value.
func (r *RedisMemory) Read(ruleName, key string) interface{} {
	v := r.brain.Read(ruleName, key)
	if v != nil {
		return v
	}

	found, err := r.db.Get(r.calculateKey(ruleName, key)).Result()
	if err != nil {
		r.err = err
		return nil
	}

	return found
}
