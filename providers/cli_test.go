package providers

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"cirello.io/gochatbot/messages"
)

func TestProviderCLI(t *testing.T) {
	const rawMsg = "hello world"
	stdin = strings.NewReader(rawMsg)
	var buf bytes.Buffer
	outPrompt = &buf

	cli := CLI()

	inMsg := <-cli.IncomingChannel()
	if inMsg.Message != rawMsg {
		t.Error("CLI provider not ingesting incoming messages")
	}

	outChan := cli.OutgoingChannel()
	outChan <- messages.Message{Room: "room", ToUserID: "uid", ToUserName: "name", Message: rawMsg}
	close(outChan)

	to := time.After(5 * time.Second)
	for buf.Len() == 0 {
		select {
		case <-to:
			t.Fatal("could not read output buffer")
		default:
		}
	}

	const expectedOutPrompt = "\nout:> room uid name : hello world\n"
	gotOut := buf.String()
	if expectedOutPrompt != gotOut {
		t.Errorf("wrong output prompt. Expected output:%v. Got: %v", expectedOutPrompt, gotOut)
	}
}
