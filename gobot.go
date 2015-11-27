package main // import "cirello.io/gobot"

import (
	"cirello.io/gobot/bot"
	"cirello.io/gobot/providers"
	"cirello.io/gobot/rules"
)

func main() {
	robot := bot.New(
		"gobot",
		bot.MessageProvider(providers.CLI()),
		bot.RegisterRule(rules.Echo()),
	)
	robot.Process()
}
