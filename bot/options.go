package bot

import (
	"log"

	"cirello.io/gochatbot/brain"
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
	ParseMessage(Self, messages.Message) []messages.Message
}

// RegisterRuleset is the self-referencing option that plugs Rules into the robot.
func RegisterRuleset(rule RuleParser) Option {
	return func(s *Self) {
		log.Printf("bot: registering ruleset %T\n", rule)
		s.rules = append(s.rules, rule)
	}
}

// RegisterMemorizer plugs a durable memory to the robot's brain.
func RegisterMemorizer(memo brain.Memorizer) Option {
	return func(s *Self) {
		log.Printf("bot: registering memorizer %T\n", memo)
		s.brain = memo
	}
}
