package brain // import "cirello.io/gochatbot/brain"

var availableDrivers []func(func(string) string) (Memorizer, bool)

// Memorizer interface describes functions to cross-messages memory.
type Memorizer interface {
	Save(ruleName, key string, value interface{})
	Read(ruleName, key string) interface{}
	Error() error
}

// Detect finds a suitable durable memory driver by analyzing the environment.
func Detect(getenv func(string) string) Memorizer {
	for _, driver := range availableDrivers {
		memo, ok := driver(getenv)
		if ok {
			return memo
		}
	}
	return Brain()
}
