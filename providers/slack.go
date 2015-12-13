// +build all slack

package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/websocket"

	"cirello.io/gochatbot/messages"
)

const (
	slackEnvVarName = "GOCHATBOT_SLACK_TOKEN"
	urlSlackAPI     = "https://slack.com/api/"
)

func init() {
	availableProviders = append(availableProviders, func(getenv func(string) string) (Provider, bool) {
		token := getenv(slackEnvVarName)
		if token == "" {
			log.Println("providers: skipping Slack. if you want Slack enabled, please set a valid value for the environment variables", slackEnvVarName)
			return nil, false
		}
		return Slack(token), true
	})
}

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
		go slack.intakeLoop()
		go slack.dispatchLoop()
	}
	go slack.reconnect()
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
	log.Println("slack: connecting to HTTP API handshake interface")
	resp, err := http.Get(fmt.Sprint(urlSlackAPI, "rtm.start?no_unreads&simple_latest&token=", p.token))
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
	log.Println("slack: dialing to HTTP WS rtm interface")
	if p.wsURL == "" {
		p.err = fmt.Errorf("could not connnect to Slack HTTP WS rtm. please, check your connection and your token (%s). error: %v", slackEnvVarName, p.err)
		return
	}
	ws, err := websocket.Dial(p.wsURL, "", urlSlackAPI)
	if err != nil {
		p.err = err
		return
	}
	p.wsConn = ws
}

func (p *providerSlack) intakeLoop() {
	log.Println("slack: started message intake loop")
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
			Room:       data.Channel,
			FromUserID: data.UserID,
			Message:    data.Text,
			Direct:     strings.HasPrefix(data.Channel, "D"),
		}
		p.in <- msg
	}
}

func (p *providerSlack) dispatchLoop() {
	log.Println("slack: started message dispatch loop")
	for msg := range p.out {
		// TODO(ccf): find a way in that text/template does not escape username DMs.
		var finalMsg bytes.Buffer
		template.Must(template.New("tmpl").Parse(msg.Message)).Execute(&finalMsg, struct{ User string }{"<@" + msg.ToUserID + ">"})

		data := struct {
			Type    string `json:"type"`
			User    string `json:"user"`
			Channel string `json:"channel"`
			Text    string `json:"text"`
		}{"message", p.selfID, msg.Room, html.UnescapeString(finalMsg.String())}

		// TODO(ccf): look for an idiomatic way of doing limited writers
		b, err := json.Marshal(data)
		if err != nil {
			continue
		}

		wsMsg := string(b)
		if len(wsMsg) > 16*1024 {
			continue
		}
		fmt.Fprint(p.wsConn, wsMsg)
		time.Sleep(1 * time.Second) // https://api.slack.com/docs/rate-limits
	}
}

func (p *providerSlack) reconnect() {
	for {
		time.Sleep(1 * time.Second)
		if p.wsConn == nil {
			log.Println("slack: cannot reconnect")
			break
		}
		_, err := p.wsConn.Write([]byte(`{"type":"hello"}`))
		if err != nil {
			log.Println("slack: reconnecting")
			p.handshake()
			p.dial()
		}
	}
}
