// +build cron

package cron // import "cirello.io/gochatbot/rules/cron"

import (
	"log"
	"time"

	"cirello.io/gochatbot/bot"
	"cirello.io/gochatbot/messages"
)

type cronRule struct {
	When        string
	Room        string
	Description string
	Action      func(out chan messages.Message)
}

type cronRuleset struct {
	outCh    chan messages.Message
	stopChan []chan struct{}
}

// Name returns this rules name - meant for debugging.
func (r *cronRuleset) Name() string {
	return "Cron Ruleset"
}

func (r *cronRuleset) ParseMessage(self bot.Self, in messages.Message) []messages.Message {
	if in.Message == "cron help" {
		return []messages.Message{
			{
				Room:       in.Room,
				ToUserID:   in.FromUserID,
				ToUserName: in.FromUserName,
				Message:    "cron list - list all crons",
			},
			{
				Room:       in.Room,
				ToUserID:   in.FromUserID,
				ToUserName: in.FromUserName,
				Message:    "cron start - start all crons",
			},
			{
				Room:       in.Room,
				ToUserID:   in.FromUserID,
				ToUserName: in.FromUserName,
				Message:    "cron stop - stop all crons",
			},
			{
				Room:       in.Room,
				ToUserID:   in.FromUserID,
				ToUserName: in.FromUserName,
				Message:    "cron help - this message",
			},
		}
	}
	if in.Message == "cron list" {
		var ret []messages.Message
		for _, rule := range cronRules {
			ret = append(ret, messages.Message{
				Room:       in.Room,
				ToUserID:   in.FromUserID,
				ToUserName: in.FromUserName,
				Message:    "@" + rule.When + " " + rule.Description,
			})
		}
		return ret
	}
	if in.Message == "cron start" {
		r.start()
		return []messages.Message{
			{
				Room:       in.Room,
				ToUserID:   in.FromUserID,
				ToUserName: in.FromUserName,
				Message:    "all cron jobs started",
			},
		}
	}
	if in.Message == "cron stop" {
		r.stop()
		return []messages.Message{
			{
				Room:       in.Room,
				ToUserID:   in.FromUserID,
				ToUserName: in.FromUserName,
				Message:    "all cron jobs stopped",
			},
		}
	}

	return []messages.Message{}
}

func (r *cronRuleset) start() {
	for _, rule := range cronRules {
		c := make(chan struct{})
		r.stopChan = append(r.stopChan, c)
		go func(r cronRule, stop chan struct{}, outCh chan messages.Message) {
			lastExecuted := ""
			for {
				select {
				case <-c:
					return
				default:
					now := time.Now().Format("15:04")
					if now == r.When && lastExecuted != now {
						lastExecuted = now
						log.Println("executed", r.Description)
						r.Action(outCh)
					}
					time.Sleep(1 * time.Second)
				}
			}
		}(rule, c, r.outCh)
	}
}

func (r *cronRuleset) stop() {
	for _, c := range r.stopChan {
		c <- struct{}{}
	}
}

// New returns a regex rule set
func New() *cronRuleset {
	r := &cronRuleset{}
	r.start()
	return r
}

func (r *cronRuleset) SetOutgoingChannel(outCh chan messages.Message) {
	r.outCh = outCh
}
