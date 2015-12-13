// +build all,!multi !multi

package main // import "cirello.io/gochatbot"

import (
	"log"
	"net"
	"os"
	"path/filepath"

	"cirello.io/gochatbot/bot"
	"cirello.io/gochatbot/brain"
	"cirello.io/gochatbot/providers"
	"cirello.io/gochatbot/rules/cron"
	"cirello.io/gochatbot/rules/ops"
	"cirello.io/gochatbot/rules/plugins"
	"cirello.io/gochatbot/rules/reddit"
	"cirello.io/gochatbot/rules/regex"
	"cirello.io/gochatbot/rules/rpc"
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

	wd, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatalln("error detecting working directory:", err)
	}

	options := []bot.Option{
		bot.MessageProvider(provider),
		bot.RegisterRuleset(regex.New(regexRules)),
		bot.RegisterRuleset(cron.New(cronRules)),
		bot.RegisterRuleset(reddit.New()),
		bot.RegisterRuleset(ops.New(opsCmds)),
		bot.RegisterRuleset(plugins.New(wd)),
	}

	rpcHostAddr := os.Getenv("GOCHATBOT_RPC_BIND")
	if rpcHostAddr != "" {
		l, err := net.Listen("tcp4", rpcHostAddr)
		if err != nil {
			log.Fatalf("rpc: cannot bind. err: %v", err)
		}
		options = append(
			options,
			bot.RegisterRuleset(rpc.New(l)),
		)
	}

	bot.New(
		name,
		memory,
		options...,
	).Process()
}
