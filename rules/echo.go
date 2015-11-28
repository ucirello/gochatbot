package rules // import "cirello.io/gochatbot/rules"

import (
	"strings"

	"cirello.io/gochatbot/bot"
	"cirello.io/gochatbot/messages"
)

// echoRule is a demonstration rule to prove that the basic bot functionality is
// correctly implemented.
type echoRule struct{}

// Name returns this rules name - meant for debugging.
func (e *echoRule) Name() string {
	return "Echo"
}

// ParseMessage for Echo reads the current message and echoes back both the last
// heard message and the newly gotten one.
func (e *echoRule) ParseMessage(bot bot.Self, in messages.Message) []messages.Message {
	if strings.TrimSpace(in.Message) != "" {
		var out []messages.Message
		lastMsg := bot.MemoryRead(e.Name(), "lastMsg")
		if lastMsg != nil {
			out = append(out, lastMsg.(messages.Message))
		}
		out = append(out, in)
		bot.MemorySave(e.Name(), "lastMsg", in)
		return out
	}

	return []messages.Message{}
}

// Echo returns an echo rule set
func Echo() *echoRule {
	return new(echoRule)
}
