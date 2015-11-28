package regex // import "cirello.io/gochatbot/rules/regex"

import (
	"bytes"
	"regexp"
	"strings"
	"text/template"

	"cirello.io/gochatbot/bot"
	"cirello.io/gochatbot/messages"
)

type regexRule struct {
	Regex        string
	ParseMessage func(bot.Self, string) []string
}

type regexRuleset struct {
	regexes map[string]*template.Template
}

// Name returns this rules name - meant for debugging.
func (r regexRuleset) Name() string {
	return "Regex Ruleset"
}

func (r regexRuleset) ParseMessage(bot bot.Self, in messages.Message) []messages.Message {
	for _, rule := range regexRules {
		botName := bot.Name()
		if in.Direct {
			botName = ""
		}

		var finalRegex bytes.Buffer
		r.regexes[rule.Regex].Execute(&finalRegex, struct{ RobotName string }{botName})
		matched, err := regexp.MatchString(strings.TrimSpace(finalRegex.String()), in.Message)
		if err != nil || !matched {
			continue
		}

		if ret := rule.ParseMessage(bot, in.Message); len(ret) > 0 {
			var retMsgs []messages.Message
			for _, m := range ret {
				var finalMsg bytes.Buffer
				template.Must(template.New("tmpl").Parse(m)).Execute(&finalMsg, struct{ User string }{"<@" + in.UserID + ">"})
				retMsgs = append(
					retMsgs,
					messages.Message{
						Room:     in.Room,
						UserID:   in.UserID,
						UserName: in.UserName,
						Message:  finalMsg.String(),
					},
				)
			}
			return retMsgs
		}
	}

	return []messages.Message{}
}

// New returns a regex rule set
func New() *regexRuleset {
	r := &regexRuleset{
		regexes: make(map[string]*template.Template),
	}
	for _, rule := range regexRules {
		r.regexes[rule.Regex] = template.Must(template.New(rule.Regex).Parse(rule.Regex))
	}
	return r
}
