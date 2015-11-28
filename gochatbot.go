package main // import "cirello.io/gochatbot"

import (
	"log"
	"os"

	"cirello.io/gochatbot/bot"
	"cirello.io/gochatbot/providers"
	"cirello.io/gochatbot/rules"
)

func main() {
	provider := providers.Detect(os.Getenv)
	robot := bot.New(
		"gobot",
		bot.MessageProvider(provider),
		bot.RegisterRule(rules.Echo()),
	)
	if err := provider.Error(); err != nil {
		log.Fatalln("error in message provider:", err)
	}
	robot.Process()
}
