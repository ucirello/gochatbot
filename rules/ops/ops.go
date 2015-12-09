package ops // import "cirello.io/gochatbot/rules/ops"

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"

	"cirello.io/gochatbot/bot"
	"cirello.io/gochatbot/messages"
	"cirello.io/gochatbot/rules/ops/ssh"
)

type sshConf struct {
	Username   string
	SSHKeyFile string
}

type opsRuleset struct {
	outCh chan messages.Message

	mu             sync.Mutex
	hostGroups     map[string][]string
	hostGroupsConf map[string]sshConf
}

// Name returns this rules name - meant for debugging.
func (r opsRuleset) Name() string {
	return "Remote Ops Ruleset"
}

// Boot runs preparatory steps for ruleset execution
func (r *opsRuleset) Boot(self *bot.Self) {
	r.outCh = self.MessageProviderOut()

	log.Println("ops: reading from memory")
	if err := json.Unmarshal(self.MemoryRead("ops", "hostGroups"), &r.hostGroups); err == nil {
		log.Println("ops: hostGroups read")
	}

	if err := json.Unmarshal(self.MemoryRead("ops", "hostGroupsConf"), &r.hostGroupsConf); err == nil {
		log.Println("ops: hostGroupsConf read")
	}
}

func (r opsRuleset) HelpMessage(self bot.Self) string {
	botName := self.Name()
	msg := fmt.Sprintln(botName, "ops uptime host-group - get uptime of all hosts of a host-group")
	msg = fmt.Sprintln(msg, botName, "ops df host-group - get 'df -h' of all hosts of a host-group")
	msg = fmt.Sprintln(msg, botName, "ops memfree host-group - get 'free -m' of all hosts of a host-group")
	msg = fmt.Sprintln(msg, botName, "ops add host host-group - add host to host-group (the group is created at first host addition)")
	msg = fmt.Sprintln(msg, botName, "ops remove host host-group - remove host from host-group (the group is removed after last host deletion)")
	msg = fmt.Sprintln(msg, botName, "ops configure hostgroup username keyfile - configure ssh login credentials (don't provide keyfile to force the use of ssh-agent)")
	return msg
}

func (r opsRuleset) ParseMessage(self bot.Self, in messages.Message) []messages.Message {
	cmd := strings.TrimSpace(strings.TrimPrefix(in.Message, self.Name()))
	parts := strings.Split(cmd, " ")

	var msg messages.Message
	if strings.HasPrefix(cmd, "ops add") {
		host, hostGroup := parts[2], parts[3]
		msg = messages.Message{
			Room:         in.Room,
			FromUserID:   in.ToUserID,
			FromUserName: in.ToUserName,
			ToUserID:     in.FromUserID,
			ToUserName:   in.FromUserName,
			Message:      r.add(self, host, hostGroup),
		}
	} else if strings.HasPrefix(cmd, "ops remove") {
		host, hostGroup := parts[2], parts[3]
		msg = messages.Message{
			Room:         in.Room,
			FromUserID:   in.ToUserID,
			FromUserName: in.ToUserName,
			ToUserID:     in.FromUserID,
			ToUserName:   in.FromUserName,
			Message:      r.remove(self, host, hostGroup),
		}
	} else if strings.HasPrefix(cmd, "ops configure") {
		var sshKeyFile string
		hostGroup, username := parts[2], parts[3]
		if len(parts) == 5 {
			sshKeyFile = parts[4]
		}
		msg = messages.Message{
			Room:         in.Room,
			FromUserID:   in.ToUserID,
			FromUserName: in.ToUserName,
			ToUserID:     in.FromUserID,
			ToUserName:   in.FromUserName,
			Message:      r.configure(self, hostGroup, username, sshKeyFile),
		}
	} else if strings.HasPrefix(cmd, "ops uptime") {
		hostGroup := parts[2]
		go r.uptime(in, hostGroup)
		msg = messages.Message{
			Room:         in.Room,
			FromUserID:   in.ToUserID,
			FromUserName: in.ToUserName,
			ToUserID:     in.FromUserID,
			ToUserName:   in.FromUserName,
			Message:      fmt.Sprintln("dispatched uptime call to", hostGroup),
		}
	} else if strings.HasPrefix(cmd, "ops df") {
		hostGroup := parts[2]
		go r.dfH(in, hostGroup)
		msg = messages.Message{
			Room:         in.Room,
			FromUserID:   in.ToUserID,
			FromUserName: in.ToUserName,
			ToUserID:     in.FromUserID,
			ToUserName:   in.FromUserName,
			Message:      fmt.Sprintln("dispatched 'df -h' call to", hostGroup),
		}
	} else if strings.HasPrefix(cmd, "ops memfree") {
		hostGroup := parts[2]
		go r.freeM(in, hostGroup)
		msg = messages.Message{
			Room:         in.Room,
			FromUserID:   in.ToUserID,
			FromUserName: in.ToUserName,
			ToUserID:     in.FromUserID,
			ToUserName:   in.FromUserName,
			Message:      fmt.Sprintln("dispatched 'free -m' call to", hostGroup),
		}
	}

	return []messages.Message{msg}
}

func (r *opsRuleset) add(self bot.Self, host, hostGroup string) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.hostGroups[hostGroup] = append(r.hostGroups[hostGroup], strings.TrimSpace(host))

	b, err := json.Marshal(r.hostGroups)
	if err != nil {
		return fmt.Sprintf("error adding host to host group. got:", err)
	}
	self.MemorySave("ops", "hostGroups", b)

	return fmt.Sprintln(host, "added to host group", hostGroup)
}

func (r *opsRuleset) remove(self bot.Self, host, hostGroup string) string {
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
	self.MemorySave("ops", "hostGroups", b)

	return fmt.Sprintln(host, "removed from host group", hostGroup)
}

func (r *opsRuleset) configure(self bot.Self, hostGroup, username, sshKeyFile string) string {
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
	self.MemorySave("ops", "hostGroupsConf", b)

	return fmt.Sprintln(hostGroup, "configured")
}

func (r *opsRuleset) uptime(in messages.Message, hostGroup string) {
	r.run(in, hostGroup, "/usr/bin/uptime")
}

func (r *opsRuleset) dfH(in messages.Message, hostGroup string) {
	r.run(in, hostGroup, "/bin/df -h")
}

func (r *opsRuleset) freeM(in messages.Message, hostGroup string) {
	r.run(in, hostGroup, "free -m")
}

func (r *opsRuleset) run(in messages.Message, hostGroup, cmd string) {
	if _, ok := r.hostGroups[hostGroup]; !ok {
		r.outCh <- messages.Message{
			Room:         in.Room,
			FromUserID:   in.ToUserID,
			FromUserName: in.ToUserName,
			ToUserID:     in.FromUserID,
			ToUserName:   in.FromUserName,
			Message:      fmt.Sprintln("could not find hostgroup", hostGroup),
		}
		return
	}
	if _, ok := r.hostGroupsConf[hostGroup]; !ok {
		r.outCh <- messages.Message{
			Room:         in.Room,
			FromUserID:   in.ToUserID,
			FromUserName: in.ToUserName,
			ToUserID:     in.FromUserID,
			ToUserName:   in.FromUserName,
			Message:      fmt.Sprintln("could not find configuration for hostgroup", hostGroup),
		}
		return
	}

	conf := r.hostGroupsConf[hostGroup]

	for _, hostname := range r.hostGroups[hostGroup] {
		go func(conf sshConf, hostname string) {
			host, port, err := net.SplitHostPort(hostname)
			if err != nil {
				r.outCh <- messages.Message{
					Room:         in.Room,
					FromUserID:   in.ToUserID,
					FromUserName: in.ToUserName,
					ToUserID:     in.FromUserID,
					ToUserName:   in.FromUserName,
					Message:      err.Error(),
				}
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
				r.outCh <- messages.Message{
					Room:         in.Room,
					FromUserID:   in.ToUserID,
					FromUserName: in.ToUserName,
					ToUserID:     in.FromUserID,
					ToUserName:   in.FromUserName,
					Message:      err.Error(),
				}
				return
			}
			r.outCh <- messages.Message{
				Room:         in.Room,
				FromUserID:   in.ToUserID,
				FromUserName: in.ToUserName,
				ToUserID:     in.FromUserID,
				ToUserName:   in.FromUserName,
				Message:      fmt.Sprintf("%s: %s", hostname, out),
			}
		}(conf, hostname)
	}
}

// New returns a ops ruleset
func New() *opsRuleset {
	return &opsRuleset{
		hostGroups:     make(map[string][]string),
		hostGroupsConf: make(map[string]sshConf),
	}
}
