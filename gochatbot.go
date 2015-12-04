package main // import "cirello.io/gochatbot"

import (
	"log"
	"os"

	"cirello.io/gochatbot/bot"
	"cirello.io/gochatbot/brain"
	"cirello.io/gochatbot/providers"
	"cirello.io/gochatbot/rules/cron"
	"cirello.io/gochatbot/rules/regex"
)

func main() {
	name := os.Getenv("GOCHATBOT_NAME")
	if name == "" {
		name = "gochatbot"
	}
	provider := providers.Detect(os.Getenv)
	if err := provider.Error(); err != nil {
		log.SetOutput(os.Stderr)
		log.Fatalln("error in message provider:", err)
	}

	memory := brain.Detect(os.Getenv)
	if err := memory.Error(); err != nil {
		log.SetOutput(os.Stderr)
		log.Fatalln("error in brain memory:", err)
	}

	bot.New(
		name,
		memory,
		bot.MessageProvider(provider),
		bot.RegisterRuleset(regex.New()),
		bot.RegisterRuleset(cron.New()),
	).Process()
}
