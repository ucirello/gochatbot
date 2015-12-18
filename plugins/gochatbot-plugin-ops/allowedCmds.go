package main

var allowedCmds = map[string]string{
	"uptime":  "get 'uptime' of all hosts of a host-group",
	"df -h":   "get 'df -h' of all hosts of a host-group",
	"free -m": "get 'free -m' of all hosts of a host-group",
}
