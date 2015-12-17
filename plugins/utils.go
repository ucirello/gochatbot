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
	req, err := http.NewRequest("GET", fmt.Sprint("http://", p.rpcAddr, "/pop"), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request (pop): %v", err)
	}

	req.Close = true
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error talking to gochatbot (pop): %v", err)
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

	req, err := http.NewRequest("POST", fmt.Sprint("http://", p.rpcAddr, "/send"), &buf)
	if err != nil {
		return fmt.Errorf("error creating request (send): %v", err)
	}
	req.Close = true
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error talking to gochatbot (send): %v", err)
	}
	defer resp.Body.Close()

	return nil
}

func (p Comm) MemoryRead(ns, key string) ([]byte, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("http://%s/memoryRead?namespace=%s&key=%s", p.rpcAddr, url.QueryEscape(ns), url.QueryEscape(key)),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating request (memory read): %v", err)
	}

	req.Close = true
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return []byte{}, fmt.Errorf("error talking to gochatbot (memory read): %v", err)
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
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("http://%s/memorySave?namespace=%s&key=%s", p.rpcAddr, url.QueryEscape(ns), url.QueryEscape(key)),
		buf,
	)
	if err != nil {
		return fmt.Errorf("error creating request (memory save): %v", err)
	}

	req.Close = true
	resp, err := http.DefaultClient.Do(req)
	defer resp.Body.Close()

	if err != nil {
		return fmt.Errorf("error talking to gochatbot (memory save): %v", err)
	}

	return nil
}
