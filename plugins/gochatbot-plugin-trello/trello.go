package main // import "cirello.io/gochatbot/plugins/gochatbot-plugin-trello"

import (
	"fmt"
	"log"
	"os"
	"time"

	"cirello.io/gochatbot/messages"
	"cirello.io/gochatbot/plugins"
	"github.com/VojtechVitek/go-trello"
)

type TrelloPlugin struct {
	comm    *plugins.Comm
	botName string

	client *trello.Client
	board  string
}

func main() {
	rpcBind := os.Getenv("GOCHATBOT_RPC_BIND")
	if rpcBind == "" {
		log.Fatal("GOCHATBOT_RPC_BIND empty or not set. Cannot start plugin.")
	}
	botName := os.Getenv("GOCHATBOT_NAME")
	if botName == "" {
		log.Fatal("GOCHATBOT_NAME empty or not set. Cannot start plugin.")
	}

	trelloKey := os.Getenv("GOCHABOT_TRELLO_KEY")
	if trelloKey == "" {
		log.Fatal("GOCHABOT_TRELLO_KEY empty or not set. Cannot start plugin.")
	}
	trelloToken := os.Getenv("GOCHABOT_TRELLO_TOKEN")
	if trelloToken == "" {
		log.Fatal("GOCHABOT_TRELLO_TOKEN empty or not set. Cannot start plugin.")
	}
	trelloBoard := os.Getenv("GOCHABOT_TRELLO_BOARD")
	if trelloBoard == "" {
		log.Fatal("GOCHABOT_TRELLO_BOARD empty or not set. Cannot start plugin.")
	}

	tc, err := trello.NewAuthClient(trelloKey, &trelloToken)
	if err != nil {
		log.Fatal("could not connect to Trello")
	}

	r := &TrelloPlugin{
		comm:    plugins.NewComm(rpcBind),
		botName: botName,
		client:  tc,
		board:   trelloBoard,
	}
	for {
		in, err := r.comm.Pop()
		if err != nil {
			log.Println("trello error popping message from gochatbot:", err)
			continue
		}
		if in.Message == "" {
			time.Sleep(1 * time.Second)
		}
		if err := r.parseMessage(in); err != nil {
			log.Println("trello error parsing message:", err)
		}
	}
}

func (r TrelloPlugin) helpMessage() string {
	helpMsg := fmt.Sprintln(r.botName, "new <list> <name>")
	helpMsg = fmt.Sprintln(helpMsg, r.botName, "list <list>")
	helpMsg = fmt.Sprintln(helpMsg, r.botName, "move <shortLink> <list>")

	return helpMsg
}

func (r *TrelloPlugin) parseMessage(in *messages.Message) error {
	// return r.comm.Send(&messages.Message{
	// 	Room:       in.Room,
	// 	ToUserID:   in.FromUserID,
	// 	ToUserName: in.FromUserName,
	// 	Message:    fmt.Sprintln("stay positive", in.FromUserName),
	// })

	return nil
}

func (r *TrelloPlugin) createCard(in *messages.Message, list, card string) *messages.Message {
	return nil
}

func (r *TrelloPlugin) showCards(in *messages.Message, list string) *messages.Message {
	return nil
}

func (r *TrelloPlugin) moveCard(in *messages.Message, cardId, list string) *messages.Message {
	return nil
}
