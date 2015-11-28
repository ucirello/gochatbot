package main // import "cirello.io/gochatbot"

import (
	"log"
	"os"

	"cirello.io/gochatbot/bot"
	"cirello.io/gochatbot/providers"
	"cirello.io/gochatbot/rules/regex"
)

func main() {
	provider := providers.Detect(os.Getenv)
	robot := bot.New(
		"gochatbot",
		bot.MessageProvider(provider),
		bot.RegisterRule(regex.New()),
	)
	if err := provider.Error(); err != nil {
		log.Fatalln("error in message provider:", err)
	}
	robot.Process()
}
