package main // import "cirello.io/gochatbot/plugins/gochatbot-plugin-ops"

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"cirello.io/gochatbot/messages"
	"cirello.io/gochatbot/plugins"
	"cirello.io/gochatbot/plugins/gochatbot-plugin-ops/ssh"
)

type sshConf struct {
	Username   string
	SSHKeyFile string
}

type OpsPlugin struct {
	comm    *plugins.Comm
	botName string

	cmds map[string]string

	mu             sync.Mutex
	hostGroups     map[string][]string
	hostGroupsConf map[string]sshConf
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

	r := &OpsPlugin{
		comm:    plugins.NewComm(rpcBind),
		botName: botName,

		cmds: allowedCmds,

		hostGroups:     make(map[string][]string),
		hostGroupsConf: make(map[string]sshConf),
	}

	log.Println("ops: reading from memory")
	memHG, err := r.comm.MemoryRead("ops", "hostGroups")
	if err != nil {
		log.Println("ops: error reading hostGroups from Bots memory")
	}
	if err := json.Unmarshal(memHG, &r.hostGroups); err == nil {
		log.Println("ops: hostGroups read")
	}

	memHGC, err := r.comm.MemoryRead("ops", "hostGroupsConf")
	if err != nil {
		log.Println("ops: error reading hostGroupsConf from Bots memory")
	}
	if err := json.Unmarshal(memHGC, &r.hostGroupsConf); err == nil {
		log.Println("ops: hostGroupsConf read")
	}

	for {
		in, err := r.comm.Pop()
		if err != nil {
			log.Println("ops: error popping message from gochatbot:", err)
			continue
		}
		if in.Message == "" {
			time.Sleep(1 * time.Second)
		}
		if err := r.parseMessage(in); err != nil {
			log.Println("ops: error parsing message:", err)
		}
	}
}

func (r OpsPlugin) helpMessage() string {
	msg := fmt.Sprintln(r.botName, "ops add host host-group - add host to host-group (the group is created at first host addition)")
	msg = fmt.Sprintln(msg, r.botName, "ops remove host host-group - remove host from host-group (the group is removed after last host deletion)")
	msg = fmt.Sprintln(msg, r.botName, "ops configure hostgroup username keyfile - configure ssh login credentials (don't provide keyfile to force the use of ssh-agent)")

	for cmd, desc := range r.cmds {
		msg = fmt.Sprintln(msg, r.botName, "ops", cmd, "-", desc)
	}
	return msg
}

func (r OpsPlugin) parseMessage(in *messages.Message) error {
	if in.Message == "help" || in.Message == "ops help" {
		return r.comm.Send(&messages.Message{
			Room:       in.Room,
			ToUserID:   in.FromUserID,
			ToUserName: in.FromUserName,
			Message:    r.helpMessage(),
		})
	}

	cmd := strings.TrimSpace(strings.TrimPrefix(in.Message, r.botName))
	parts := strings.Split(cmd, " ")

	if strings.HasPrefix(cmd, "ops add") {
		host, hostGroup := parts[2], parts[3]
		return r.comm.Send(&messages.Message{
			Room:         in.Room,
			FromUserID:   in.ToUserID,
			FromUserName: in.ToUserName,
			ToUserID:     in.FromUserID,
			ToUserName:   in.FromUserName,
			Message:      r.add(host, hostGroup),
		})
	} else if strings.HasPrefix(cmd, "ops remove") {
		host, hostGroup := parts[2], parts[3]
		return r.comm.Send(&messages.Message{
			Room:         in.Room,
			FromUserID:   in.ToUserID,
			FromUserName: in.ToUserName,
			ToUserID:     in.FromUserID,
			ToUserName:   in.FromUserName,
			Message:      r.remove(host, hostGroup),
		})
	} else if strings.HasPrefix(cmd, "ops configure") {
		var sshKeyFile string
		hostGroup, username := parts[2], parts[3]
		if len(parts) == 5 {
			sshKeyFile = parts[4]
		}
		return r.comm.Send(&messages.Message{
			Room:         in.Room,
			FromUserID:   in.ToUserID,
			FromUserName: in.ToUserName,
			ToUserID:     in.FromUserID,
			ToUserName:   in.FromUserName,
			Message:      r.configure(hostGroup, username, sshKeyFile),
		})
	} else {
		for allowedCmd := range r.cmds {
			if strings.HasPrefix(cmd, strings.TrimSpace(fmt.Sprintln("ops", allowedCmd))) {
				hostGroup := strings.TrimSpace(strings.TrimPrefix(cmd, strings.TrimSpace(fmt.Sprintln("ops", allowedCmd))))
				go r.run(in, hostGroup, allowedCmd)
				return r.comm.Send(&messages.Message{
					Room:         in.Room,
					FromUserID:   in.ToUserID,
					FromUserName: in.ToUserName,
					ToUserID:     in.FromUserID,
					ToUserName:   in.FromUserName,
					Message:      fmt.Sprintf("dispatched '%s' call to %s", allowedCmd, hostGroup),
				})
			}
		}
	}

	return nil
}

func (r *OpsPlugin) add(host, hostGroup string) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.hostGroups[hostGroup] = append(r.hostGroups[hostGroup], strings.TrimSpace(host))

	b, err := json.Marshal(r.hostGroups)
	if err != nil {
		return fmt.Sprintf("error adding host to host group. got:", err)
	}
	r.comm.MemorySave("ops", "hostGroups", b)

	return fmt.Sprintln(host, "added to host group", hostGroup)
}

func (r *OpsPlugin) remove(host, hostGroup string) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.hostGroups[hostGroup]; !ok {
		return "host group not found"
	}

	var newHG []string
	for _, h := range r.hostGroups[hostGroup] {
		if strings.TrimSpace(h) == strings.TrimSpace(host) {
			continue
		}
		newHG = append(newHG, host)
	}
	r.hostGroups[hostGroup] = newHG

	b, err := json.Marshal(r.hostGroups)
	if err != nil {
		return fmt.Sprintf("error removing host from host group. got: %v", err)
	}
	r.comm.MemorySave("ops", "hostGroups", b)

	return fmt.Sprintln(host, "removed from host group", hostGroup)
}

func (r *OpsPlugin) configure(hostGroup, username, sshKeyFile string) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.hostGroupsConf[hostGroup] = sshConf{
		Username:   username,
		SSHKeyFile: sshKeyFile,
	}

	b, err := json.Marshal(r.hostGroupsConf)
	if err != nil {
		return fmt.Sprintf("error configuring host group. got: %v", err)
	}
	r.comm.MemorySave("ops", "hostGroupsConf", b)

	return fmt.Sprintln(hostGroup, "configured")
}

func (r *OpsPlugin) run(in *messages.Message, hostGroup, cmd string) {
	if _, ok := r.hostGroups[hostGroup]; !ok {
		r.comm.Send(&messages.Message{
			Room:         in.Room,
			FromUserID:   in.ToUserID,
			FromUserName: in.ToUserName,
			ToUserID:     in.FromUserID,
			ToUserName:   in.FromUserName,
			Message:      fmt.Sprintln("could not find hostgroup", hostGroup),
		})
		return
	}
	if _, ok := r.hostGroupsConf[hostGroup]; !ok {
		r.comm.Send(&messages.Message{
			Room:         in.Room,
			FromUserID:   in.ToUserID,
			FromUserName: in.ToUserName,
			ToUserID:     in.FromUserID,
			ToUserName:   in.FromUserName,
			Message:      fmt.Sprintln("could not find configuration for hostgroup", hostGroup),
		})
		return
	}

	conf := r.hostGroupsConf[hostGroup]

	for _, hostname := range r.hostGroups[hostGroup] {
		go func(conf sshConf, hostname string) {
			host, port, err := net.SplitHostPort(hostname)
			if err != nil {
				r.comm.Send(&messages.Message{
					Room:         in.Room,
					FromUserID:   in.ToUserID,
					FromUserName: in.ToUserName,
					ToUserID:     in.FromUserID,
					ToUserName:   in.FromUserName,
					Message:      err.Error(),
				})
				return
			}
			authMethod := ssh.PublicKeyFile(conf.SSHKeyFile)
			if conf.SSHKeyFile == "" {
				authMethod = ssh.SshAgent(os.Getenv)
			}
			out, err := ssh.Run(
				cmd,
				[]string{},
				conf.Username,
				host,
				port,
				authMethod,
			)
			if err != nil {
				r.comm.Send(&messages.Message{
					Room:         in.Room,
					FromUserID:   in.ToUserID,
					FromUserName: in.ToUserName,
					ToUserID:     in.FromUserID,
					ToUserName:   in.FromUserName,
					Message:      err.Error(),
				})
				return
			}
			r.comm.Send(&messages.Message{
				Room:         in.Room,
				FromUserID:   in.ToUserID,
				FromUserName: in.ToUserName,
				ToUserID:     in.FromUserID,
				ToUserName:   in.FromUserName,
				Message:      fmt.Sprintf("%s: %s", hostname, out),
			})
		}(conf, hostname)
	}
}
