package providers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/net/websocket"

	"cirello.io/gochatbot/messages"
)

const SlackEnvVarName = "GOCHATBOT_SLACK_TOKEN"

func init() {
	availableProviders = append(availableProviders, func(getenv func(string) string) Provider {
		token := getenv(SlackEnvVarName)
		if token == "" {
			return nil
		}
		return Slack(token)
	})
}

const URLSlackAPI = "https://slack.com/api/"

type providerSlack struct {
	token  string
	wsURL  string
	wsConn *websocket.Conn
	selfID string

	in  chan messages.Message
	out chan messages.Message
	err error
}

// Slack is the message provider meant to be used in development of rule sets.
func Slack(token string) *providerSlack {
	slack := &providerSlack{
		token: token,
		in:    make(chan messages.Message),
		out:   make(chan messages.Message),
	}
	slack.handshake()
	slack.dial()
	if slack.err == nil {
		go slack.loop()
	}
	return slack
}

func (p *providerSlack) IncomingChannel() chan messages.Message {
	return p.in
}

func (p *providerSlack) OutgoingChannel() chan messages.Message {
	return p.out
}

func (p *providerSlack) Error() error {
	return p.err
}

func (p *providerSlack) handshake() {
	resp, err := http.Get(fmt.Sprint(URLSlackAPI, "rtm.start?no_unreads&simple_latest&token=", p.token))
	if err != nil {
		p.err = err
		return
	}
	defer resp.Body.Close()
	var data struct {
		OK   interface{} `json:"ok"`
		URL  string      `json:"url"`
		Self struct {
			ID string `json:"id"`
		} `json:"self"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		p.err = err
		return
	}

	switch v := data.OK.(type) {
	case bool:
		if !v {
			p.err = err
			return

		}
	default:
		p.err = err
		return
	}
	p.wsURL = data.URL
	p.selfID = data.Self.ID
}

func (p *providerSlack) dial() {
	ws, err := websocket.Dial(p.wsURL, "", URLSlackAPI)
	if err != nil {
		p.err = err
		return
	}
	p.wsConn = ws
}

func (p *providerSlack) loop() {
	go func() {
		for {
			var data struct {
				Type    string `json:"type"`
				Channel string `json:"channel"`
				UserID  string `json:"user"`
				Text    string `json:"text"`
			}
			if err := json.NewDecoder(p.wsConn).Decode(&data); err != nil {
				continue
			}
			if data.Type != "message" {
				continue
			}

			msg := messages.Message{
				Room:    data.Channel,
				UserID:  data.UserID,
				Message: data.Text,
				Direct:  strings.HasPrefix(data.Channel, "D"),
			}
			p.in <- msg
		}
	}()

	go func() {
		for msg := range p.out {
			data := struct {
				Type    string `json:"type"`
				User    string `json:"user"`
				Channel string `json:"channel"`
				Text    string `json:"text"`
			}{"message", p.selfID, msg.Room, msg.Message}

			// TODO(carlos): look for an idiomatic way of doing limited writers
			b, err := json.Marshal(data)
			if err != nil {
				continue
			}

			wsMsg := string(b)
			if len(wsMsg) > 16*1024 {
				continue
			}
			fmt.Fprint(p.wsConn, wsMsg)
		}
	}()
}
