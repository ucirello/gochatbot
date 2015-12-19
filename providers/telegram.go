// +build all telegram

package providers

import (
	"bytes"
	"log"
	"strconv"
	"strings"
	"text/template"
	"time"

	"cirello.io/gochatbot/messages"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
)

const (
	telegramEnvVarName = "GOCHATBOT_TELEGRAM_TOKEN"
)

func init() {
	availableProviders = append(availableProviders, func(getenv func(string) string) (Provider, bool) {
		token := getenv(telegramEnvVarName)
		if token == "" {
			log.Println("providers: skipping Telegram. if you want Telegram enabled, please set a valid value for the environment variables", telegramEnvVarName)
			return nil, false
		}
		return Telegram(token), true
	})
}

type providerTelegram struct {
	token string
	tg    *tgbotapi.BotAPI

	in  chan messages.Message
	out chan messages.Message
	err error
}

// Telegram is the message provider meant to be used in development of rule sets.
func Telegram(token string) *providerTelegram {
	telegram := &providerTelegram{
		token: token,
		in:    make(chan messages.Message),
		out:   make(chan messages.Message),
	}

	tg, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		telegram.err = err
		return telegram
	}

	log.Println("telegram: logged as", tg.Self.UserName)
	telegram.tg = tg

	go telegram.intakeLoop()
	go telegram.dispatchLoop()
	return telegram
}

func (p *providerTelegram) IncomingChannel() chan messages.Message {
	return p.in
}

func (p *providerTelegram) OutgoingChannel() chan messages.Message {
	return p.out
}

func (p *providerTelegram) Error() error {
	return p.err
}

func (p *providerTelegram) intakeLoop() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := p.tg.GetUpdatesChan(u)
	if err != nil {
		p.err = err
		return
	}

	go func(upds <-chan tgbotapi.Update) {
		log.Println("telegram: started message intake loop")
		for {
			for update := range upds {
				senderID := strconv.Itoa(update.Message.From.ID)
				senderName := update.Message.From.FirstName
				targetID := strconv.Itoa(update.Message.Chat.ID)
				targetName := update.Message.Chat.FirstName
				msg := messages.Message{
					Room:         "",
					FromUserID:   senderID,
					FromUserName: senderName,
					ToUserID:     targetID,
					ToUserName:   targetName,
					Message:      update.Message.Text,
					// todo(carlos): do DM detection
					// Direct:     strings.HasPrefix(data.Channel, "D"),
				}
				p.in <- msg
			}
		}
	}(updates)
}

func (p *providerTelegram) dispatchLoop() {
	log.Println("telegram: started message dispatch loop")
	for msg := range p.out {
		id, err := strconv.Atoi(msg.ToUserID)
		if err != nil {
			continue
		}
		var finalMsg bytes.Buffer
		template.Must(template.New("tmpl").Parse(msg.Message)).Execute(&finalMsg, struct{ User string }{"@" + msg.ToUserName})

		if strings.TrimSpace(finalMsg.String()) == "" {
			continue
		}

		p.tg.Send(tgbotapi.NewMessage(id, finalMsg.String()))
		time.Sleep(1 * time.Second)
	}
}
