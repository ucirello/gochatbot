package rpc // import "cirello.io/gochatbot/rules/rpc"

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"

	"cirello.io/gochatbot/bot"
	"cirello.io/gochatbot/messages"
)

type rpcRuleset struct {
	mux *http.ServeMux

	memoryRead func(ruleName, key string) []byte
	memorySave func(ruleName, key string, value []byte)

	listener net.Listener
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
	r.memoryRead = self.MemoryRead
	r.memorySave = self.MemorySave
	r.outCh = self.MessageProviderOut()
	r.mux.HandleFunc("/pop", r.httpPop)
	r.mux.HandleFunc("/send", r.httpSend)
	r.mux.HandleFunc("/memoryRead", r.httpMemoryRead)
	r.mux.HandleFunc("/memorySave", r.httpMemorySave)
	log.Println("rpc: listening", r.listener.Addr())
	go http.Serve(r.listener, r.mux)
}

func (r rpcRuleset) HelpMessage(self bot.Self, _ string) string {
	return fmt.Sprintln("RPC listens to", r.listener.Addr(), "for RPC calls")
}

func (r *rpcRuleset) ParseMessage(self bot.Self, in messages.Message) []messages.Message {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.inbox = append(r.inbox, in)

	return []messages.Message{}
}

// New returns a RPC ruleset
func New(listener net.Listener) *rpcRuleset {
	return &rpcRuleset{
		mux:      http.NewServeMux(),
		listener: listener,
	}
}
