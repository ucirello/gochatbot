package providers

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"text/template"

	"cirello.io/gochatbot/messages"

	ircevent "github.com/thoj/go-ircevent"
)

// IRC message provider configuration environment variables.
const (
	IrcUserEnvVarName     = "GOCHATBOT_IRC_USER"
	IrcNickEnvVarName     = "GOCHATBOT_IRC_NICK"
	IrcServerEnvVarName   = "GOCHATBOT_IRC_SERVER"
	IrcChannelsEnvVarName = "GOCHATBOT_IRC_CHANNELS"
	IrcPasswordEnvVarName = "GOCHATBOT_IRC_PASSWORD"
	IrcTLSEnvVarName      = "GOCHATBOT_IRC_TLS"
)

func init() {
	availableProviders = append(availableProviders, func(getenv func(string) string) (Provider, bool) {
		user := getenv(IrcUserEnvVarName)
		nick := getenv(IrcNickEnvVarName)
		server := getenv(IrcServerEnvVarName)
		channels := getenv(IrcChannelsEnvVarName)
		password := getenv(IrcPasswordEnvVarName)
		tls := getenv(IrcTLSEnvVarName)

		if server == "" || channels == "" || user == "" || nick == "" {
			log.Println("providers: if you want IRC enabled, please set a valid value for the environment variables", IrcUserEnvVarName, IrcNickEnvVarName, IrcServerEnvVarName, IrcChannelsEnvVarName)
			return nil, false
		}
		return IRC(user, nick, server, channels, password, tls), true
	})
}

type providerIRC struct {
	channels []string
	ircConn  *ircevent.Connection
	in       chan messages.Message
	out      chan messages.Message
	err      error
}

// IRC returns the IRC message provider
func IRC(user, nick, server, channels, password, useTLS string) *providerIRC {
	pi := &providerIRC{
		channels: strings.Split(channels, ","),
		in:       make(chan messages.Message),
		out:      make(chan messages.Message),
	}

	ircConn := ircevent.IRC(user, nick)
	ircConn.Password = password

	ircConn.UseTLS = false
	if useTLS != "" {
		serverName, _, err := net.SplitHostPort(server)
		if err != nil {
			pi.err = fmt.Errorf("error spliting host port: %v", err)
			return pi
		}

		useTLSBool, err := strconv.ParseBool(useTLS)
		if err != nil {
			pi.err = fmt.Errorf("error parsing bool (useTLS): %v", err)
			return pi
		}
		ircConn.UseTLS = useTLSBool
		if useTLSBool {
			ircConn.TLSConfig = &tls.Config{
				ServerName: serverName,
			}
		}
	}

	ircConn.AddCallback(
		"001",
		func(e *ircevent.Event) {
			for _, channel := range pi.channels {
				ircConn.Join(channel)
			}
		},
	)

	ircConn.AddCallback("PRIVMSG", func(e *ircevent.Event) {
		msg := messages.Message{
			Room:         e.Arguments[0],
			FromUserName: e.Nick,
			Message:      e.Message(),
			// Direct:       strings.HasPrefix(data.Channel, "D"),
		}
		pi.in <- msg
	})

	if err := ircConn.Connect(server); err != nil {
		pi.err = fmt.Errorf("error connecting to server (%s): %v", server, err)
		return pi
	}
	pi.ircConn = ircConn
	go pi.ircConn.Loop()
	go pi.loop()
	return pi
}

func (p *providerIRC) loop() {
	for msg := range p.out {
		channel := msg.Room
		if p.ircConn.GetNick() == msg.Room {
			channel = msg.FromUserName
		}

		var finalMsg bytes.Buffer
		template.Must(template.New("tmpl").Parse(msg.Message)).Execute(&finalMsg, struct{ User string }{msg.ToUserName})

		p.ircConn.Privmsg(channel, finalMsg.String())
	}
}

func (p *providerIRC) IncomingChannel() chan messages.Message {
	return p.in
}

func (p *providerIRC) OutgoingChannel() chan messages.Message {
	return p.out
}

func (p *providerIRC) Error() error {
	return p.err
}
