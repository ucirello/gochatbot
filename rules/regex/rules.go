package regex

import "cirello.io/gochatbot/bot"

var regexRules = []regexRule{
	{
		`{{ .RobotName }} jump`, func(bot bot.Self, msg string) []string {
			return []string{"{{ .User }}, How high?"}
		},
	},
}
