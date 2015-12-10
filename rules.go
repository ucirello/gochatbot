package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"cirello.io/gochatbot/bot"
	"cirello.io/gochatbot/messages"
	"cirello.io/gochatbot/rules/cron"
	"cirello.io/gochatbot/rules/regex"
)

var opsCmds = map[string]string{
	"uptime":  "get 'uptime' of all hosts of a host-group",
	"df -h":   "get 'df -h' of all hosts of a host-group",
	"free -m": "get 'free -m' of all hosts of a host-group",
}

var cronRules = map[string]cron.Rule{
	"message of the day": {
		"0 10 * * *",
		func() []messages.Message {
			return []messages.Message{
				{Message: "Good morning!"},
			}
		},
	},
}

var regexRules = []regex.Rule{
	{
		`{{ .RobotName }} jump`, `tells the robot to jump`,
		func(bot bot.Self, msg string, matches []string) []string {
			var ret []string
			ret = append(ret, "{{ .User }}, How high?")
			lastJumpTS := bot.MemoryRead("jump", "lastJump")
			ret = append(ret, fmt.Sprintf("{{ .User }} (last time I jumped: %s)", lastJumpTS))
			bot.MemorySave("jump", "lastJump", []byte(time.Now().String()))

			return ret
		},
	},
	{
		`{{ .RobotName }} qr code (.*)`, `turn a URL into a QR Code`,
		func(bot bot.Self, msg string, matches []string) []string {
			const qrUrl = "https://chart.googleapis.com/chart?chs=178x178&cht=qr&chl=%s"
			return []string{
				fmt.Sprintf(qrUrl, url.QueryEscape(matches[1])),
			}
		},
	},
	{
		`{{ .RobotName }} http status (.*)`, `return the description of the given HTTP status`,
		func(bot bot.Self, msg string, matches []string) []string {
			httpCode, err := strconv.Atoi(matches[1])
			if err != nil {
				return []string{fmt.Sprintln("I could not convert", matches[1], "into HTTP code.")}
			}
			return []string{
				fmt.Sprintln(
					"{{ .User }},", matches[1], "is",
					http.StatusText(httpCode),
				),
			}
		},
	},
	{
		`{{ .RobotName }} explainshell (.*)`, `links to explainshell.com on given command`,
		func(bot bot.Self, msg string, matches []string) []string {
			const explainShellUrl = "http://explainshell.com/explain?cmd=%s"
			return []string{
				strings.Replace(
					fmt.Sprintf(explainShellUrl, url.QueryEscape(matches[1])),
					"%20",
					"+",
					-1,
				),
			}
		},
	},
	{
		`{{ .RobotName }} godoc (.*)`, `search godoc.org and return the first result`,
		func(bot bot.Self, msg string, matches []string) []string {
			if len(matches) < 2 {
				return []string{}
			}

			respBody, err := httpGet(fmt.Sprintf("http://api.godoc.org/search?q=%s", url.QueryEscape(matches[1])))
			if err != nil {
				return []string{err.Error()}
			}
			defer respBody.Close()

			var data struct {
				Results []struct {
					Path     string `json:"path"`
					Synopsis string `json:"synopsis"`
				} `json:"results"`
			}

			if err := json.NewDecoder(respBody).Decode(&data); err != nil {
				return []string{err.Error()}
			}

			if len(data.Results) == 0 {
				return []string{"package not found"}
			}

			return []string{fmt.Sprintf("%s %s/%s", data.Results[0].Synopsis, "http://godoc.org", data.Results[0].Path)}

		},
	},
}

func httpGet(u string) (io.ReadCloser, error) {
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
