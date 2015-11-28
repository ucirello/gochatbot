package providers // import "cirello.io/gochatbot/providers"

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"

	"cirello.io/gochatbot/messages"
)

var (
	stdin     io.Reader = os.Stdin
	outPrompt io.Writer = os.Stdout
)

type providerCLI struct {
	in  chan messages.Message
	out chan messages.Message
}

// CLI is the message provider meant to be used in development of rule sets.
func CLI() *providerCLI {
	cli := &providerCLI{
		in:  make(chan messages.Message),
		out: make(chan messages.Message),
	}
	go cli.loop()
	return cli
}

func (c *providerCLI) IncomingChannel() chan messages.Message {
	return c.in
}

func (c *providerCLI) OutgoingChannel() chan messages.Message {
	return c.out
}

func (c *providerCLI) Error() error {
	return nil
}

func (c *providerCLI) loop() {
	go func() {
		scanner := bufio.NewScanner(stdin)
		for scanner.Scan() {
			c.in <- messages.Message{
				Room:         "CLI",
				FromUserID:   "CLI",
				FromUserName: "CLI",
				Message:      scanner.Text(),
			}
		forLoop:
			for {
				select {
				case msg := <-c.out:
					fmt.Fprint(outPrompt, processOutMessage(msg))
				default:
					break forLoop
				}
			}
		}
	}()
	go func() {
		for msg := range c.out {
			fmt.Fprint(outPrompt, processOutMessage(msg))
		}
	}()
}

func processOutMessage(msg messages.Message) string {
	var finalMsg bytes.Buffer
	template.Must(template.New("tmpl").Parse(msg.Message)).Execute(&finalMsg, struct{ User string }{msg.ToUserID})

	return fmt.Sprintln("\nout:>", msg.Room, msg.ToUserID, msg.ToUserName, ":", finalMsg.String())
}
