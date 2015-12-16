package regex // import "cirello.io/gochatbot/rules/regex"

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/template"
	"unicode"

	"cirello.io/gochatbot/bot"
	"cirello.io/gochatbot/messages"
)

type Rule struct {
	Regex        string
	HelpMessage  string
	ParseMessage func(bot.Self, string, []string) []string
}

type regexRuleset struct {
	regexes map[string]*template.Template
	rules   []Rule
}

// Name returns this rules name - meant for debugging.
func (r regexRuleset) Name() string {
	return "Regex Ruleset"
}

// Boot runs preparatory steps for ruleset execution
func (r regexRuleset) Boot(_ *bot.Self) {
}

func (r regexRuleset) HelpMessage(self bot.Self, _ string) string {
	botName := self.Name()
	var helpMsg string
	for _, rule := range r.rules {
		var finalRegex bytes.Buffer
		r.regexes[rule.Regex].Execute(&finalRegex, struct{ RobotName string }{botName})

		helpMsg = fmt.Sprintln(helpMsg, finalRegex.String(), "-", rule.HelpMessage)
	}
	return strings.TrimLeftFunc(helpMsg, unicode.IsSpace)
}

func (r regexRuleset) ParseMessage(self bot.Self, in messages.Message) []messages.Message {
	for _, rule := range r.rules {
		botName := self.Name()
		if in.Direct {
			botName = ""
		}

		var finalRegex bytes.Buffer
		if _, ok := r.regexes[rule.Regex]; !ok {
			r.regexes[rule.Regex] = template.Must(template.New(rule.Regex).Parse(rule.Regex))
		}
		r.regexes[rule.Regex].Execute(&finalRegex, struct{ RobotName string }{botName})
		sanitizedRegex := strings.TrimSpace(finalRegex.String())
		re := regexp.MustCompile(sanitizedRegex)
		matched := re.MatchString(in.Message)
		if !matched {
			continue
		}

		args := re.FindStringSubmatch(in.Message)
		if ret := rule.ParseMessage(self, in.Message, args); len(ret) > 0 {
			var retMsgs []messages.Message
			for _, m := range ret {
				retMsgs = append(
					retMsgs,
					messages.Message{
						Room:       in.Room,
						ToUserID:   in.FromUserID,
						ToUserName: in.FromUserName,
						Message:    m,
					},
				)
			}
			return retMsgs
		}
	}

	return []messages.Message{}
}

// New returns a regex rule set
func New(rules []Rule) *regexRuleset {
	r := &regexRuleset{
		regexes: make(map[string]*template.Template),
		rules:   rules,
	}
	for _, rule := range rules {
		r.regexes[rule.Regex] = template.Must(template.New(rule.Regex).Parse(rule.Regex))
	}
	return r
}
