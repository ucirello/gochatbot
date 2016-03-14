package main // import "cirello.io/gochatbot/plugins/gochatbot-plugin-goapp-release"

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"cirello.io/gochatbot/messages"
	"cirello.io/gochatbot/plugins"
)

type GoAppReleasePlugin struct {
	comm    *plugins.Comm
	botName string
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

	r := &GoAppReleasePlugin{comm: plugins.NewComm(rpcBind), botName: botName}
	for {
		in, err := r.comm.Pop()
		if err != nil {
			log.Println("goapp-release: error popping message from gochatbot:", err)
			continue
		}
		if in.Message == "" {
			time.Sleep(1 * time.Second)
		}
		if err := r.parseMessage(in); err != nil {
			log.Println("goapp-release: error parsing message:", err)
		}
	}
}

func (r GoAppReleasePlugin) helpMessage() string {
	helpMsg := fmt.Sprintln(r.botName, "release <project> - releases a google app engine project")
	return helpMsg
}

func (r *GoAppReleasePlugin) parseMessage(in *messages.Message) error {
	msg := strings.TrimSpace(in.Message)
	checkPrefix := r.botName + " release "
	if msg == "" || (strings.HasPrefix(msg, r.botName) && !strings.HasPrefix(msg, checkPrefix)) {
		return nil
	}

	project := filepath.Clean(strings.TrimSpace(strings.TrimPrefix(msg, checkPrefix)))

	r.comm.Send(&messages.Message{
		Room:       in.Room,
		ToUserID:   in.FromUserID,
		ToUserName: in.FromUserName,
		Message:    fmt.Sprintf("git-pull project %s ...", project),
	})

	outGit, err := exec.Command("git", "-C", project, "pull").CombinedOutput()
	if err != nil {
		return r.comm.Send(&messages.Message{
			Room:       in.Room,
			ToUserID:   in.FromUserID,
			ToUserName: in.FromUserName,
			Message:    fmt.Sprintf("could not do git-pull for %s: %s", project, outGit),
		})
	}

	r.comm.Send(&messages.Message{
		Room:       in.Room,
		ToUserID:   in.FromUserID,
		ToUserName: in.FromUserName,
		Message:    fmt.Sprintf("running appcfg.py on project %s ...", project),
	})

	outAppcfg, err := exec.Command("appcfg.py", "update", project).CombinedOutput()
	if err != nil {
		return r.comm.Send(&messages.Message{
			Room:       in.Room,
			ToUserID:   in.FromUserID,
			ToUserName: in.FromUserName,
			Message:    fmt.Sprintf("could not do appcfg.py for %s: %s", project, outAppcfg),
		})
	}

	return r.comm.Send(&messages.Message{
		Room:       in.Room,
		ToUserID:   in.FromUserID,
		ToUserName: in.FromUserName,
		Message:    fmt.Sprintf("Release log:\n```\n%s\n```", outAppcfg),
	})
}
