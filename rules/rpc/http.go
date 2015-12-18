package rpc // import "cirello.io/gochatbot/rules/rpc"

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"cirello.io/gochatbot/messages"
)

func (r *rpcRuleset) httpPop(w http.ResponseWriter, req *http.Request) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer req.Body.Close()

	var msg messages.Message
	if len(r.inbox) > 1 {
		msg, r.inbox = r.inbox[0], r.inbox[1:]
	} else if len(r.inbox) == 1 {
		msg = r.inbox[0]
		r.inbox = []messages.Message{}
	} else if len(r.inbox) == 0 {
		fmt.Fprint(w, "{}")
		return
	}

	if err := json.NewEncoder(w).Encode(&msg); err != nil {
		log.Fatal(err)
	}
}

func (r *rpcRuleset) httpSend(w http.ResponseWriter, req *http.Request) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var msg messages.Message
	if err := json.NewDecoder(req.Body).Decode(&msg); err != nil {
		log.Fatal(err)
	}
	defer req.Body.Close()

	go func(m messages.Message) {
		r.outCh <- m
	}(msg)
	fmt.Fprintln(w, "OK")
}

func (r *rpcRuleset) httpMemoryRead(w http.ResponseWriter, req *http.Request) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer req.Body.Close()

	namespace := req.URL.Query().Get("namespace")
	key := req.URL.Query().Get("key")

	fmt.Fprintf(w, "%s", r.memoryRead(namespace, key))
}

func (r *rpcRuleset) httpMemorySave(w http.ResponseWriter, req *http.Request) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer req.Body.Close()

	namespace := req.URL.Query().Get("namespace")
	key := req.URL.Query().Get("key")

	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	r.memorySave(namespace, key, b)
	fmt.Fprintln(w, "OK")
}
