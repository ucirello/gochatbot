package regex // import "cirello.io/gochatbot/rules/regex"

import (
	"bytes"
	"regexp"
	"text/template"

	"cirello.io/gochatbot/bot"
	"cirello.io/gochatbot/messages"
)

type regexRule struct {
	Regex        string
	ParseMessage func(bot.Self, messages.Message) []messages.Message
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
		var finalRegex bytes.Buffer
		r.regexes[rule.Regex].Execute(&finalRegex, struct{ RobotName string }{bot.Name()})
		matched, err := regexp.MatchString(finalRegex.String(), in.Message)
		if err != nil || !matched {
			continue
		}

		if ret := rule.ParseMessage(bot, in); len(ret) > 0 {
			return ret
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
