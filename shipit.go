// +build all shipit

package main

import (
	"math/rand"

	"cirello.io/gochatbot/bot"
	"cirello.io/gochatbot/rules/regex"
)

// Ported from github.com/hubot-scripts/hubot-shipit
func init() {
	regexRules = append(regexRules, regex.Rule{
		`\bship(ping|z|s|ped)?\s*it\b`, `display a motivation squirrel`,
		func(bot bot.Self, msg string, matches []string) []string {
			squirrels := []string{
				"http://1.bp.blogspot.com/_v0neUj-VDa4/TFBEbqFQcII/AAAAAAAAFBU/E8kPNmF1h1E/s640/squirrelbacca-thumb.jpg",
				"http://28.media.tumblr.com/tumblr_lybw63nzPp1r5bvcto1_500.jpg",
				"http://d2f8dzk2mhcqts.cloudfront.net/0772_PEW_Roundup/09_Squirrel.jpg",
				"http://i.imgur.com/DPVM1.png",
				"http://images.cheezburger.com/completestore/2011/11/2/46e81db3-bead-4e2e-a157-8edd0339192f.jpg",
				"http://images.cheezburger.com/completestore/2011/11/2/aa83c0c4-2123-4bd3-8097-966c9461b30c.jpg",
				"http://img70.imageshack.us/img70/4853/cutesquirrels27rn9.jpg",
				"http://img70.imageshack.us/img70/9615/cutesquirrels15ac7.jpg",
				"http://shipitsquirrel.github.io/images/ship%20it%20squirrel.png",
				"http://www.cybersalt.org/images/funnypictures/s/supersquirrel.jpg",
				"http://www.zmescience.com/wp-content/uploads/2010/09/squirrel.jpg",
				"https://dl.dropboxusercontent.com/u/602885/github/sniper-squirrel.jpg",
				"https://dl.dropboxusercontent.com/u/602885/github/soldier-squirrel.jpg",
				"https://dl.dropboxusercontent.com/u/602885/github/squirrelmobster.jpeg",
			}

			chosenSquirrels := squirrels[rand.Intn(len(squirrels)-1)]

			return []string{chosenSquirrels}
		},
	})
}
