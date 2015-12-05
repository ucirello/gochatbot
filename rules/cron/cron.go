package cron // import "cirello.io/gochatbot/rules/cron"

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"cirello.io/gochatbot/bot"
	"cirello.io/gochatbot/messages"
	"github.com/gorhill/cronexpr"
)

type Rule struct {
	When   string
	Action func() []messages.Message
}

type cronRuleset struct {
	outCh     chan messages.Message
	stopChan  []chan struct{}
	cronRules map[string]Rule

	loadOnce sync.Once

	mu            sync.Mutex
	attachedCrons map[string][]string
}

// Name returns this rules name - meant for debugging.
func (r *cronRuleset) Name() string {
	return "Cron Ruleset"
}

// Boot runs preparatory steps for ruleset execution
func (r *cronRuleset) Boot(self *bot.Self) {
	r.loadMemory(self)
}

func (r *cronRuleset) loadMemory(self *bot.Self) {
	log.Println("cron: reading from memory")
	v := self.MemoryRead("cron", "attached")
	if v != nil {
		for room, irules := range v.(map[string]interface{}) {
			rules := irules.([]interface{})
			for _, rule := range rules {
				r.attachedCrons[room] = append(r.attachedCrons[room], fmt.Sprint(rule))
			}
		}
		log.Println("cron: memory read")
		r.start()
	}
}

func (r cronRuleset) HelpMessage(self bot.Self) string {
	helpMsg := fmt.Sprintln("cron attach <job name>- attach one cron job to a room")
	helpMsg = fmt.Sprintln(helpMsg, "cron detach <job name> - detach one cron job from a room")
	helpMsg = fmt.Sprintln(helpMsg, "cron list - list all available crons")
	helpMsg = fmt.Sprintln(helpMsg, "cron start - start all crons")
	helpMsg = fmt.Sprintln(helpMsg, "cron stop - stop all crons")
	helpMsg = fmt.Sprintln(helpMsg, "cron help - this message")

	return helpMsg
}

func (r *cronRuleset) ParseMessage(self bot.Self, in messages.Message) []messages.Message {
	if strings.HasPrefix(in.Message, "cron attach") {
		ruleName := strings.TrimSpace(strings.TrimPrefix(in.Message, "cron attach"))
		ret := []messages.Message{
			{
				Room:       in.Room,
				ToUserID:   in.FromUserID,
				ToUserName: in.FromUserName,
				Message:    r.attach(self, ruleName, in.Room),
			},
		}
		r.start()
		return ret
	}

	if strings.HasPrefix(in.Message, "cron detach") {
		ruleName := strings.TrimSpace(strings.TrimPrefix(in.Message, "cron detach"))
		return []messages.Message{
			{
				Room:       in.Room,
				ToUserID:   in.FromUserID,
				ToUserName: in.FromUserName,
				Message:    r.detach(self, ruleName, in.Room),
			},
		}
	}

	if in.Message == "cron list" {
		var ret []messages.Message
		for ruleName, rule := range r.cronRules {
			ret = append(ret, messages.Message{
				Room:       in.Room,
				ToUserID:   in.FromUserID,
				ToUserName: in.FromUserName,
				Message:    "@" + rule.When + " " + ruleName,
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

func (r *cronRuleset) attach(self bot.Self, ruleName, room string) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.cronRules[ruleName]; !ok {
		return ruleName + " not found"
	}

	r.attachedCrons[room] = append(r.attachedCrons[room], ruleName)
	self.MemorySave("cron", "attached", r.attachedCrons)
	return ruleName + " attached to this room"
}

func (r *cronRuleset) detach(self bot.Self, ruleName, room string) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.attachedCrons[room]; !ok {
		return room + " not found"
	}

	var newRoom []string
	for _, rn := range r.attachedCrons[room] {
		if rn == ruleName {
			continue
		}
		newRoom = append(newRoom, rn)
	}
	r.attachedCrons[room] = newRoom
	self.MemorySave("cron", "attached", r.attachedCrons)
	return ruleName + " detached to this room"
}

func (r *cronRuleset) start() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for room, rules := range r.attachedCrons {
		for _, rule := range rules {
			c := make(chan struct{})
			r.stopChan = append(r.stopChan, c)
			go processCronRule(r.cronRules[rule], c, r.outCh, room)
		}
	}
}

func processCronRule(rule Rule, stop chan struct{}, outCh chan messages.Message, cronRoom string) {
	nextTime := cronexpr.MustParse(rule.When).Next(time.Now())
	for {
		select {
		case <-stop:
			return
		default:
			if nextTime.Format("2006-01-02 15:04") == time.Now().Format("2006-01-02 15:04") {
				msgs := rule.Action()
				for _, msg := range msgs {
					msg.Room = cronRoom
					outCh <- msg
				}
			}
			nextTime = cronexpr.MustParse(rule.When).Next(time.Now())
			time.Sleep(2 * time.Second)
		}
	}
}

func (r *cronRuleset) stop() {
	for _, c := range r.stopChan {
		c <- struct{}{}
	}
	r.stopChan = []chan struct{}{}
}

// New returns a cron rule set
func New(rules map[string]Rule) *cronRuleset {
	r := &cronRuleset{
		attachedCrons: make(map[string][]string),
		cronRules:     rules,
	}
	return r
}

func (r *cronRuleset) SetOutgoingChannel(outCh chan messages.Message) {
	r.outCh = outCh
}
