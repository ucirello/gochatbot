package plugins // import "cirello.io/gochatbot/plugins"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"cirello.io/gochatbot/messages"
)

type Comm struct {
	rpcAddr string
}

func NewComm(rpcAddr string) *Comm {
	return &Comm{rpcAddr}
}

func (p Comm) Pop() (*messages.Message, error) {
	resp, err := http.Get(fmt.Sprint("http://", p.rpcAddr, "/pop"))
	if err != nil {
		return nil, fmt.Errorf("error talking to gochatbot: %v", err)
	}
	defer resp.Body.Close()

	var msg messages.Message
	if err := json.NewDecoder(resp.Body).Decode(&msg); err != nil {
		return nil, fmt.Errorf("error parsing message from gochatbot: %v", err)
	}

	return &msg, nil
}

func (p Comm) Send(msg *messages.Message) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&msg); err != nil {
		return fmt.Errorf("error serializing message to gochatbot: %v", err)
	}

	if _, err := http.Post(fmt.Sprint("http://", p.rpcAddr, "/send"), "application/json", &buf); err != nil {
		return fmt.Errorf("error talking to gochatbot: %v", err)
	}

	return nil
}

func (p Comm) MemoryRead(ns, key string) ([]byte, error) {
	resp, err := http.Get(
		fmt.Sprintf("http://%s/memoryRead?namespace=%s&key=%s", p.rpcAddr, url.QueryEscape(ns), url.QueryEscape(key)),
	)
	if err != nil {
		return []byte{}, fmt.Errorf("error talking to gochatbot %v", err)
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("error reading memory from gochatbot: %v", err)
	}

	return b, nil
}

func (p Comm) MemorySave(ns, key string, content []byte) error {
	buf := bytes.NewReader(content)
	if _, err := http.Post(
		fmt.Sprintf("http://%s/memorySave?namespace=%s&key=%s", p.rpcAddr, url.QueryEscape(ns), url.QueryEscape(key)),
		"application/octet-stream",
		buf,
	); err != nil {
		return fmt.Errorf("error talking to gochatbot: %v", err)
	}

	return nil
}
