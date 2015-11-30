package regex // import "cirello.io/gochatbot/rules/regex"

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"cirello.io/gochatbot/bot"
	"cirello.io/gochatbot/messages"
)

type regexRule struct {
	Regex        string
	HelpMessage  string
	ParseMessage func(bot.Self, string, []string) []string
}

type regexRuleset struct {
	regexes map[string]*template.Template
}

// Name returns this rules name - meant for debugging.
func (r regexRuleset) Name() string {
	return "Regex Ruleset"
}

func (r regexRuleset) ParseMessage(self bot.Self, in messages.Message) []messages.Message {
	localRegexRules := regexRules
	localRegexRules = append(localRegexRules,
		regexRule{
			`{{ .RobotName }} help`, `this help screen`,
			func(self bot.Self, msg string, _ []string) []string {
				botName := self.Name()
				ret := []string{fmt.Sprint("available commands:")}
				for _, rule := range localRegexRules {
					var finalRegex bytes.Buffer
					r.regexes[rule.Regex].Execute(&finalRegex, struct{ RobotName string }{botName})

					ret = append(ret, fmt.Sprintln(finalRegex.String(), "-", rule.HelpMessage))
				}
				return ret
			},
		},
	)

	for _, rule := range localRegexRules {
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
func New() *regexRuleset {
	r := &regexRuleset{
		regexes: make(map[string]*template.Template),
	}
	for _, rule := range regexRules {
		r.regexes[rule.Regex] = template.Must(template.New(rule.Regex).Parse(rule.Regex))
	}
	return r
}
