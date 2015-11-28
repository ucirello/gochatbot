package providers

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"
	"time"

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
