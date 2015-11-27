package providers

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"

	"cirello.io/gobot/messages"
)

func TestProviderCLI(t *testing.T) {
	const rawMsg = "hello world"
	stdin = strings.NewReader(rawMsg)
	inPrompt = ioutil.Discard
	var buf bytes.Buffer
	outPrompt = &buf

	cli := CLI()

	inMsg := <-cli.IncomingChannel()
	if inMsg.Message != rawMsg {
		t.Error("CLI provider not ingesting incoming messages")
	}

	outChan := cli.OutgoingChannel()
	outChan <- messages.Message{Room: "room", FromUserID: "uid", FromUserName: "name", Message: rawMsg}
	close(outChan)
	t.Log(buf.String())
}
