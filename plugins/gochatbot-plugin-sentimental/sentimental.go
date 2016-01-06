package main // import "cirello.io/gochatbot/plugins/gochatbot-plugin-sentimental"

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"cirello.io/HumorChecker"
	"cirello.io/gochatbot/messages"
	"cirello.io/gochatbot/plugins"
)

const baseURL = "https://www.reddit.com/r/"

type SentimentalPlugin struct {
	comm    *plugins.Comm
	botName string
	quiet   bool
}

type userScore struct {
	Messages int
	Score    float64
	Average  float64
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
	quiet, err := strconv.ParseBool(os.Getenv("GOCHATBOT_SENTIMENTAL_QUIET"))
	if err != nil {
		log.Println("error reading variable GOCHATBOT_SENTIMENTAL_QUIET, defaulting to false")
		quiet = false
	}

	r := &SentimentalPlugin{comm: plugins.NewComm(rpcBind), botName: botName, quiet: quiet}
	for {
		in, err := r.comm.Pop()
		if err != nil {
			log.Println("sentimental: error popping message from gochatbot:", err)
			continue
		}
		if in.Message == "" {
			time.Sleep(1 * time.Second)
		}
		if err := r.parseMessage(in); err != nil {
			log.Println("sentimental: error parsing message:", err)
		}
	}
}

func (r SentimentalPlugin) helpMessage() string {
	helpMsg := fmt.Sprintln(r.botName, "check <user>")
	helpMsg = fmt.Sprintln(helpMsg, r.botName, "check everyone")

	return helpMsg
}

func (r *SentimentalPlugin) parseMessage(in *messages.Message) error {
	msg := strings.TrimSpace(in.Message)
	checkPrefix := r.botName + " check on"
	if msg == "" || (strings.HasPrefix(msg, r.botName) && !strings.HasPrefix(msg, checkPrefix)) {
		return nil
	}

	if strings.HasPrefix(msg, checkPrefix) {
		username := strings.TrimPrefix(strings.TrimSpace(msg), checkPrefix)
		log.Println("sentimental: state for ", username)
		return nil
		// return r.comm.Send(&messages.Message{
		// 	Room:       in.Room,
		// 	ToUserID:   in.FromUserID,
		// 	ToUserName: in.FromUserName,
		// 	Message:    r.helpMessage(),
		// })
	}

	if in.FromUserName == "" {
		log.Println("sentimental: got empty username")
		return nil
	}

	scorecard, err := r.readUsersSentiment()
	if err != nil {
		return err
	}

	if _, ok := scorecard[in.FromUserName]; !ok {
		scorecard[in.FromUserName] = userScore{}
	}

	score := HumorChecker.Analyze(in.Message)
	sc := r.updateScore(scorecard[in.FromUserName], score)
	scorecard[in.FromUserName] = sc

	if err := r.storeUsersSentiment(scorecard); err != nil {
		return err
	}

	log.Printf("sentimental: %s now has %v / %v", in.FromUserName, sc.Score, sc.Average)

	if score.Score < -2 && !r.quiet {
		return r.comm.Send(&messages.Message{
			Room:       in.Room,
			ToUserID:   in.FromUserID,
			ToUserName: in.FromUserName,
			Message:    fmt.Sprintln("stay positive", in.FromUserName),
		})
	}

	return nil
}

func (r *SentimentalPlugin) readUsersSentiment() (map[string]userScore, error) {
	scorecard := make(map[string]userScore)
	b, err := r.comm.MemoryRead("sentimental", "userScoreCard")
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &scorecard); len(b) > 0 && err != nil {
		return nil, err
	}
	return scorecard, nil
}

func (r *SentimentalPlugin) updateScore(sc userScore, score HumorChecker.FullScore) userScore {
	sc.Messages++
	sc.Score += score.Score
	sc.Average = sc.Score / float64(sc.Messages)
	return sc
}

func (r *SentimentalPlugin) storeUsersSentiment(scorecard map[string]userScore) error {
	record, err := json.Marshal(scorecard)
	if err != nil {
		return err
	}
	return r.comm.MemorySave("sentimental", "userScoreCard", record)
}
