package ssh // import "cirello.io/gochatbot/rules/ops/ssh"

import (
	"fmt"
	"io/ioutil"
	"net"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type sshCommand struct {
	command string
	env     []string
}

type clientSSH struct {
	config *ssh.ClientConfig
	host   string
	port   string
}

func (client *clientSSH) runCommand(cmd *sshCommand) ([]byte, error) {
	var session *ssh.Session
	var err error

	if session, err = client.newSession(); err != nil {
		return []byte{}, err
	}
	defer session.Close()

	if err = client.prepareCommand(session, cmd); err != nil {
		return []byte{}, err
	}

	return session.CombinedOutput(cmd.command)
}

func (client *clientSSH) prepareCommand(session *ssh.Session, cmd *sshCommand) error {
	for _, env := range cmd.env {
		variable := strings.Split(env, "=")
		if len(variable) != 2 {
			continue
		}

		if err := session.Setenv(variable[0], variable[1]); err != nil {
			return err
		}
	}

	return nil
}

func (client *clientSSH) newSession() (*ssh.Session, error) {
	connection, err := ssh.Dial("tcp", net.JoinHostPort(client.host, client.port), client.config)
	if err != nil {
		return nil, fmt.Errorf("cannot dial: %s", err)
	}

	session, err := connection.NewSession()
	if err != nil {
		return nil, fmt.Errorf("cannot create session: %s", err)
	}

	modes := ssh.TerminalModes{
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		session.Close()
		return nil, fmt.Errorf("cannot request pseudo terminal: %s", err)
	}

	return session, nil
}

func PublicKeyFile(file string) ssh.AuthMethod {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}

	key, err := ssh.ParsePrivateKey(b)
	if err != nil {
		return nil
	}

	return ssh.PublicKeys(key)
}

func SshAgent(getenv func(string) string) ssh.AuthMethod {
	if sshAgent, err := net.Dial("unix", getenv("SSH_AUTH_SOCK")); err == nil {
		return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)
	}

	return nil
}

func Run(command string, env []string, username, host, port string, authMethod ssh.AuthMethod) ([]byte, error) {
	sshConfig := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{authMethod},
	}

	client := &clientSSH{
		config: sshConfig,
		host:   host,
		port:   port,
	}
	if client.port == "" {
		client.port = "22"
	}

	cmd := &sshCommand{
		command: command,
		env:     env,
	}

	return client.runCommand(cmd)
}
