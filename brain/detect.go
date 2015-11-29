package brain // import "cirello.io/gochatbot/brain"

var AvailableDrivers []func(func(string) string) (Memorizer, bool)

type Memorizer interface {
	Save(ruleName, key string, value interface{})
	Read(ruleName, key string) interface{}
	Error() error
}

func Detect(getenv func(string) string) Memorizer {
	for _, driver := range AvailableDrivers {
		memo, ok := driver(getenv)
		if ok {
			return memo
		}
	}
	return Brain()
}
