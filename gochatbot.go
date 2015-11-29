package main // import "cirello.io/gochatbot"

import (
	"log"
	"os"

	"cirello.io/gochatbot/bot"
	"cirello.io/gochatbot/brain"
	"cirello.io/gochatbot/providers"
	"cirello.io/gochatbot/rules/regex"
)

func main() {
	provider := providers.Detect(os.Getenv)
	memory := brain.Detect(os.Getenv)
	robot := bot.New(
		"gochatbot",
		bot.MessageProvider(provider),
		bot.RegisterRuleset(regex.New()),
		bot.RegisterMemorizer(memory),
	)
	if err := provider.Error(); err != nil {
		log.SetOutput(os.Stderr)
		log.Fatalln("error in message provider:", err)
	}
	if err := memory.Error(); err != nil {
		log.SetOutput(os.Stderr)
		log.Fatalln("error in brain memory:", err)
	}
	robot.Process()
}
