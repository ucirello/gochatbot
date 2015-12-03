package cron

import "cirello.io/gochatbot/messages"

type cronRule struct {
	When   string
	Action func() []messages.Message
}

var cronRules = map[string]cronRule{
	"message of the day": {
		"10:00",
		func() []messages.Message {
			return []messages.Message{
				{Message: "Good morning!"},
			}
		},
	},
}
