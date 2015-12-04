package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"cirello.io/gochatbot/bot"
	"cirello.io/gochatbot/messages"
	"cirello.io/gochatbot/rules/cron"
	"cirello.io/gochatbot/rules/regex"
)

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
			ret = append(ret, fmt.Sprint("{{ .User }} (last time I jumped:", lastJumpTS, ")"))
			bot.MemorySave("jump", "lastJump", fmt.Sprint(time.Now()))

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
