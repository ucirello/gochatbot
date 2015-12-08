package rpc // import "cirello.io/gochatbot/rules/rpc"

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"cirello.io/gochatbot/bot"
	"cirello.io/gochatbot/messages"
)

type rpcRuleset struct {
	mux *http.ServeMux

	bindAddr string
	outCh    chan messages.Message

	mu    sync.Mutex
	inbox []messages.Message
}

// Name returns this rules name - meant for debugging.
func (r *rpcRuleset) Name() string {
	return "RPC Ruleset"
}

// Boot runs preparatory steps for ruleset execution
func (r *rpcRuleset) Boot(self *bot.Self) {
	r.outCh = self.MessageProviderOut()
	r.mux.HandleFunc("/pop", r.httpPop)
	r.mux.HandleFunc("/send", r.httpSend)
	log.Println("rpc: listening", r.bindAddr)
	go http.ListenAndServe(r.bindAddr, r.mux)
}

func (r rpcRuleset) HelpMessage(self bot.Self) string {
	return fmt.Sprintln("RPC listens to", r.bindAddr, "for RPC calls")
}

func (r *rpcRuleset) ParseMessage(self bot.Self, in messages.Message) []messages.Message {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.inbox = append(r.inbox, in)

	return []messages.Message{}
}

// New returns a RPC ruleset
func New(bindAddr string) *rpcRuleset {
	return &rpcRuleset{
		mux:      http.NewServeMux(),
		bindAddr: bindAddr,
	}
}
