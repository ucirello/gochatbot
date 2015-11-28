package bot

import (
	"cirello.io/gochatbot/messages"
	"cirello.io/gochatbot/providers"
)

// MessageProvider is the self-referencing option that plugs Message Providers
// into the robot.
func MessageProvider(provider providers.Provider) Option {
	return func(s *Self) {
		s.providerIn = provider.IncomingChannel()
		s.providerOut = provider.OutgoingChannel()
	}
}

// RuleParser explains the interface needed for a certain type to be considered
// a valid message parsing rule.
type RuleParser interface {
	Name() string
	ParseMessage(Self, messages.Message) []messages.Message
}

// RegisterRule is the self-referencing option that plugs Rules into the robot.
func RegisterRule(rule RuleParser) Option {
	return func(s *Self) {
		s.rules = append(s.rules, rule)
	}
}
