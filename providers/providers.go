package providers // import "cirello.io/gochatbot/providers"

import (
	"log"

	"cirello.io/gochatbot/messages"
)

// Provider explains the interface for pluggable message providers (CLI, Slack,
// IRC etc.)
type Provider interface {
	IncomingChannel() chan messages.Message
	OutgoingChannel() chan messages.Message
	Error() error
}

var availableProviders []func(func(string) string) (Provider, bool)

// Detect try all available providers, and return the one which manages to
// configure itself by inspecting the environment. If all fail, them CLI is
// returned.
func Detect(getenv func(string) string) Provider {
	for _, ap := range availableProviders {
		if ret, ok := ap(getenv); ok {
			if ret.Error() != nil {
				log.Printf("providers: %T %v", ret, ret.Error())
				continue
			}
			return ret
		}
	}
	log.Println("providers: no message provider found.")
	log.Println("providers: falling back to CLI.")
	return CLI()
}
