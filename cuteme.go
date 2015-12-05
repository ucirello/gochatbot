// +build all cuteme

package main

import (
	"encoding/json"
	"math/rand"

	"cirello.io/gochatbot/bot"
	"cirello.io/gochatbot/rules/regex"
)

// Ported from github.com/hubot-scripts/hubot-cute-me
func init() {
	cuteMe := func(bot bot.Self, msg string, matches []string) []string {
		r, err := httpGet("http://www.reddit.com/r/aww/.json")
		if err != nil {
			return []string{err.Error()}
		}
		defer r.Close()

		var data struct {
			Data struct {
				Children []struct {
					Data struct {
						URL string `json:"url"`
					} `json:"data"`
				} `json:"children"`
			} `json:"data"`
		}
		if err := json.NewDecoder(r).Decode(&data); err != nil {
			return []string{err.Error()}
		}
		options := data.Data.Children
		if len(options) == 0 {
			return []string{"could not find a cute thing for you."}
		}
		return []string{options[rand.Intn(len(options)-1)].Data.URL}
	}

	regexRules = append(regexRules, regex.Rule{
		`unicorn chaser `, `Receive a cute thing`,
		cuteMe,
	})

	regexRules = append(regexRules, regex.Rule{
		`{{ .RobotName }} cute me`, `Receive a cute thing`,
		cuteMe,
	})
}
