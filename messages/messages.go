package messages // import "cirello.io/gochatbot/messages"

// Message holds all the metadata for each sent/received message by the bot.
type Message struct {
	Room     string
	UserID   string
	UserName string
	Message  string
	Direct   bool
}
