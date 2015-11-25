package gobot // import "cirello.io/gobot"

type IncomingMessage struct {
	Room         string
	FromUserID   string
	FromUserName string
	Message      string
}

type OutgoingMessage struct {
	Room         string
	FromUserID   string
	FromUserName string
	Message      string
}

type Self struct {
	Name        string
	ProviderOut chan OutgoingMessage
	ProviderIn  chan IncomingMessage
}

type Provider interface {
	OutgoingChannel() chan OutgoingMessage
	IncomingChannel() chan IncomingMessage
}

type RuleParser interface {
	ParseMessage(*Rule, *IncomingMessage) []*OutgoingMessage
}

func main() {

}
