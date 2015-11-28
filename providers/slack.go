package providers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/net/websocket"

	"cirello.io/gochatbot/messages"
)

const URLSlackAPI = "https://slack.com/api/"

type providerSlack struct {
	token string
	wsURL string

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
	slack.connect()
	slack.talk()
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

func (p *providerSlack) connect() {
	resp, err := http.Get(fmt.Sprint(URLSlackAPI, "rtm.start?no_unreads&simple_latest&token=", p.token))
	if err != nil {
		p.err = err
		return
	}
	defer resp.Body.Close()
	var data struct {
		OK  interface{} `json:"ok"`
		URL string      `json:"url"`
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
}

func (p *providerSlack) talk() {
	ws, err := websocket.Dial(p.wsURL, "", URLSlackAPI)
	if err != nil {
		p.err = err
		return
	}

	var msg = make([]byte, 512)
	var n int
	if n, err = ws.Read(msg); err != nil {
		p.err = err
		return
	}
	log.Printf("Received: %s.\n", msg[:n])
}
