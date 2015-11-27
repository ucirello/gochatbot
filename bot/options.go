package bot

import "cirello.io/gobot/messages"

// Provider explains the interface for pluggable message providers (CLI, Slack,
// IRC etc.)
type Provider interface {
	IncomingChannel() chan messages.Message
	OutgoingChannel() chan messages.Message
}

// MessageProvider is the self-referencing option that plugs Message Providers
// into the robot.
func MessageProvider(provider Provider) Option {
	return func(s *Self) {
		s.ProviderIn = provider.IncomingChannel()
		s.ProviderOut = provider.OutgoingChannel()
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
		s.Rules = append(s.Rules, rule)
	}
}
