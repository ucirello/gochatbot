package bot

import (
	"log"

	"cirello.io/gochatbot/messages"
	"cirello.io/gochatbot/providers"
)

// MessageProvider is the self-referencing option that plugs Message Providers
// into the robot.
func MessageProvider(provider providers.Provider) Option {
	return func(s *Self) {
		log.Printf("bot: changing message provider %T\n", provider)
		s.providerIn = provider.IncomingChannel()
		s.providerOut = provider.OutgoingChannel()
	}
}

// RuleParser explains the interface needed for a certain type to be considered
// a valid message parsing rule.
type RuleParser interface {
	Name() string
	Boot(*Self)
	HelpMessage(Self) string
	ParseMessage(Self, messages.Message) []messages.Message
}

// RegisterRuleset is the self-referencing option that plugs Rules into the robot.
func RegisterRuleset(rule RuleParser) Option {
	return func(s *Self) {
		log.Printf("bot: registering ruleset %T", rule)
		rule.Boot(s)
		s.rules = append(s.rules, rule)
	}
}
