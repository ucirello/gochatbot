package providers

import (
	"fmt"
	"testing"
)

func TestDetect(t *testing.T) {
	providersDetect := []struct {
		Getenv   func(string) string
		Expected string
	}{
		{
			func(name string) string {
				if name == SlackEnvVarName {
					return "some token"
				}
				return ""
			},
			"*providers.providerSlack",
		},
	}

	for _, p := range providersDetect {
		gotProvider := Detect(p.Getenv)
		if p.Expected != fmt.Sprintf("%T", gotProvider) {
			t.Errorf("expected: %v. got: %v\n", p.Expected, fmt.Sprintf("%T", gotProvider))
		}
	}
}
