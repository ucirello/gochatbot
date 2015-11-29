package regex

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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
	{
		`{{ .RobotName }} godoc (.*)`, `search godoc.org and return the first result`,
		func(bot bot.Self, msg string, matches []string) []string {
			if len(matches) < 2 {
				return []string{}
			}

			resp, err := http.Get(fmt.Sprintf("http://api.godoc.org/search?q=%s", url.QueryEscape(matches[1])))
			if err != nil {
				return []string{err.Error()}
			}
			defer resp.Body.Close()

			var data struct {
				Results []struct {
					Path     string `json:"path"`
					Synopsis string `json:"synopsis"`
				} `json:"results"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
				return []string{err.Error()}
			}

			if len(data.Results) == 0 {
				return []string{"package not found"}
			}

			return []string{fmt.Sprintf("%s %s/%s", data.Results[0].Synopsis, "http://godoc.org", data.Results[0].Path)}

		},
	},
}
