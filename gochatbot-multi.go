// +build all,multi multi

package main // import "cirello.io/gochatbot"

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"cirello.io/gochatbot/bot"
	"cirello.io/gochatbot/brain"
	"cirello.io/gochatbot/providers"
	"cirello.io/gochatbot/rules/cron"
	"cirello.io/gochatbot/rules/ops"
	"cirello.io/gochatbot/rules/reddit"
	"cirello.io/gochatbot/rules/regex"
	"cirello.io/gochatbot/rules/rpc"
)

func main() {

	var wg sync.WaitGroup
	var botCount int

	for {
		e := &envGet{botCount}
		if e.getenv("GOCHATBOT_NAME") == "" {
			break
		}
		wg.Add(1)
		go func(e *envGet) {
			name := e.getenv("GOCHATBOT_NAME")
			if name == "" {
				name = "gochatbot"
			}

			provider := providers.Detect(e.getenv)
			if err := provider.Error(); err != nil {
				log.SetOutput(os.Stderr)
				log.Fatalln("error in message provider:", err)
			}

			memory := brain.Detect(e.getenv)
			if err := memory.Error(); err != nil {
				log.SetOutput(os.Stderr)
				log.Fatalln("error in brain memory:", err)
			}

			options := []bot.Option{
				bot.MessageProvider(provider),
				bot.RegisterRuleset(regex.New(regexRules)),
				bot.RegisterRuleset(cron.New(cronRules)),
				bot.RegisterRuleset(reddit.New()),
				bot.RegisterRuleset(ops.New()),
			}

			rpcHostAddr := e.getenv("GOCHATBOT_RPC_BIND")
			if rpcHostAddr != "" {
				options = append(
					options,
					bot.RegisterRuleset(rpc.New(rpcHostAddr)),
				)
			}

			bot.New(
				name,
				memory,
				options...,
			).Process()
			wg.Done()
		}(e)
		botCount++
	}
	wg.Wait()
}

type envGet struct {
	idx int
}

func (e envGet) getenv(key string) string {
	newKey := strings.Replace(key, "GOCHATBOT_", fmt.Sprint("GOCHATBOT_", e.idx, "_"), 1)
	return os.Getenv(newKey)
}
