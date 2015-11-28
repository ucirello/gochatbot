package main // import "cirello.io/gochatbot"

import (
	"cirello.io/gochatbot/bot"
	"cirello.io/gochatbot/providers"
	"cirello.io/gochatbot/rules"
)

func main() {
	robot := bot.New(
		"gobot",
		bot.MessageProvider(providers.CLI()),
		bot.RegisterRule(rules.Echo()),
	)
	robot.Process()
}
