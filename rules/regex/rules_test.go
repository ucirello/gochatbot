package regex

import (
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"cirello.io/gochatbot/bot"
	"cirello.io/gochatbot/brain"
	"cirello.io/gochatbot/messages"
)

func TestRegex(t *testing.T) {
	mockBot := bot.New("gochatbot", brain.Brain())
	reSet := New()
	tests := []struct {
		name    string
		in      string
		out     []string
		httpGet func(u string) (io.ReadCloser, error)
	}{
		{
			"jump",
			"gochatbot jump",
			[]string{
				"{{ .User }}, How high?",
				"{{ .User }} (last time I jumped:<nil>)",
			},
			nil,
		},
		{
			"godoc",
			"gochatbot godoc mock",
			[]string{
				"http://api.godoc.org/search?q=mock http://godoc.org/path",
			},
			func(u string) (io.ReadCloser, error) {
				return ioutil.NopCloser(strings.NewReader(`{"results":[{"path":"path","synopsis":"` + u + `"}]}`)), nil
			},
		},
	}

	for _, test := range tests {
		httpGet = test.httpGet
		out := reSet.ParseMessage(*mockBot, messages.Message{
			Message: test.in,
		})

		outLen := len(out)
		for i := 0; i < outLen; i++ {
			if out[i].Message != test.out[i] {
				t.Error("missed:", test.name, test.out[i], out[i])
			}
		}
	}
}
