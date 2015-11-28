package regex

import (
	"cirello.io/gochatbot/bot"
	"cirello.io/gochatbot/messages"
)

var regexRules = []regexRule{
	{
		`{{ .RobotName }} jump`, func(bot bot.Self, in messages.Message) []messages.Message {
			return []messages.Message{{Message: "How high?"}}
		},
	},
}
