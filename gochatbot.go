package main // import "cirello.io/gochatbot"

import (
	"log"
	"os"

	"cirello.io/gochatbot/bot"
	"cirello.io/gochatbot/providers"
	"cirello.io/gochatbot/rules"
)

func main() {
	slackToken := os.Getenv("GOCHATBOT_SLACK_TOKEN")
	var provider bot.Provider
	if slackToken != "" {
		provider = providers.Slack(slackToken)
	} else {
		provider = providers.CLI()
	}
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
