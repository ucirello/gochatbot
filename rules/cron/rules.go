package cron

import "cirello.io/gochatbot/messages"

var cronRules = []cronRule{
	{
		"03:20",
		"message of the day",
		func(out chan messages.Message) {
			out <- messages.Message{
				Message: "Good morning!",
			}
		},
	},
}
