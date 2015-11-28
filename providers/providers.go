package providers // import "cirello.io/gochatbot/providers"

import "cirello.io/gochatbot/messages"

// Provider explains the interface for pluggable message providers (CLI, Slack,
// IRC etc.)
type Provider interface {
	IncomingChannel() chan messages.Message
	OutgoingChannel() chan messages.Message
	Error() error
}

var availableProviders []func(func(string) string) Provider

// Detect try all available providers, and return the one which manages to
// configure itself by inspecting the environment. If all fail, them CLI is
// returned.
func Detect(getenv func(string) string) Provider {
	for _, ap := range availableProviders {
		if ret := ap(getenv); ret != nil {
			return ret
		}
	}
	return CLI()
}
