// +build all redis

package brain // import "cirello.io/gochatbot/brain"

import (
	"log"
	"strconv"

	redis "gopkg.in/redis.v3"
)

func init() {
	availableDrivers = append(availableDrivers, func(getenv func(string) string) (Memorizer, bool) {
		log.Println("brain: trying registering redis")
		memo, ok := RedisFromEnv(getenv)
		if ok {
			log.Println("brain: registered redis")
		} else {
			log.Println("brain: if you want Redis enabled, please set a valid value for the environment variables", redisMemoryDatabase, redisMemoryHostPort, redisMemoryPassword)
		}
		return memo, ok
	})
}

type RedisMemory struct {
	brain *brainMemory
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
func (r *RedisMemory) Save(ruleName, key string, value []byte) {
	r.brain.Save(ruleName, key, value)

	if err := r.db.Set(r.calculateKey(ruleName, key), value, 0).Err(); err != nil {
		log.Println("redis err (set):", err)
		r.err = err
		return
	}
}

// Read reads from Brain some arbritary value.
func (r *RedisMemory) Read(ruleName, key string) []byte {
	v := r.brain.Read(ruleName, key)
	if len(v) > 0 {
		return v
	}

	found, err := r.db.Get(r.calculateKey(ruleName, key)).Result()
	if err != nil {
		log.Println("redis err (get):", err)
		r.err = err
		return nil
	}

	return []byte(found)
}
