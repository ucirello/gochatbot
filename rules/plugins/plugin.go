package plugins // import "cirello.io/gochatbot/rules/plugins"

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path"
	"strings"

	"cirello.io/gochatbot/bot"
	"cirello.io/gochatbot/messages"
	"cirello.io/gochatbot/rules/rpc"
)

type pluginRuleset struct {
	pluginBins []string
	plugins    []bot.RuleParser
}

// Name returns this rules name - meant for debugging.
func (r *pluginRuleset) Name() string {
	return "Plugins Ruleset"
}

// Boot runs preparatory steps for ruleset execution
func (r *pluginRuleset) Boot(self *bot.Self) {
	for _, pluginBin := range r.pluginBins {
		l, err := net.Listen("tcp4", "0.0.0.0:0")
		if err != nil {
			log.Println("could not start plugin %s. error while setting listener: %v", pluginBin, err)
			continue
		}
		log.Printf("plugin: starting %s", pluginBin)
		rs := rpc.New(l)
		rs.Boot(self)
		cmd := exec.Command(pluginBin)
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, fmt.Sprintf("GOCHATBOT_RPC_BIND=%s", l.Addr()))
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			log.Printf("plugin: %s - error: %v", pluginBin, err)
			log.Printf("plugin: %s closing listener.", pluginBin)
			l.Close()
			continue
		}
		r.plugins = append(r.plugins, rs)
	}
}

func (r pluginRuleset) HelpMessage(self bot.Self, room string) string {
	if len(r.plugins) > 0 {
		var names []string
		for _, pluginBin := range r.pluginBins {
			names = append(names, path.Base(pluginBin))
		}
		msg := fmt.Sprintln("Loaded plugins:", names)
		for _, plugin := range r.plugins {
			go plugin.ParseMessage(self, messages.Message{Room: room, Message: "help"})
		}
		return msg
	}
	return "no external plugins loaded"
}

func (r *pluginRuleset) ParseMessage(self bot.Self, in messages.Message) []messages.Message {
	var ret []messages.Message
	for _, plugin := range r.plugins {
		msgs := plugin.ParseMessage(self, in)
		ret = append(ret, msgs...)
	}
	return ret
}

func (p *pluginRuleset) detectPluginBinaries(workdir string) error {
	files, err := ioutil.ReadDir(workdir)
	if err != nil {
		return err
	}

	for _, file := range files {
		fn := file.Name()
		if strings.HasPrefix(fn, "gochatbot-plugin-") && file.Mode()&0111 != 0 {
			p.pluginBins = append(p.pluginBins, path.Join(workdir, fn))
		}
	}

	return nil
}

// New returns plugin ruleset
func New(workdir string) *pluginRuleset {
	p := &pluginRuleset{}
	if err := p.detectPluginBinaries(workdir); err != nil {
		log.Fatal("error loading plugins: %v", err)
	}
	return p
}
