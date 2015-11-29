package regex

import (
	"fmt"
	"time"

	"cirello.io/gochatbot/bot"
)

var regexRules = []regexRule{
	{
		`{{ .RobotName }} jump`, `tells the robot to jump`,
		func(bot bot.Self, msg string, matches []string) []string {
			var ret []string
			ret = append(ret, "{{ .User }}, How high?")
			lastJumpTS := bot.MemoryRead("jump", "lastJump")
			ret = append(ret, fmt.Sprint("{{ .User }} (last time I jumped:", lastJumpTS, ")"))
			bot.MemorySave("jump", "lastJump", fmt.Sprint(time.Now()))

			return ret
		},
	},
}
