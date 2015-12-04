gochatbot
=========

This is a chatops bot implemented in Go.

It features:
- Support for Slack, Telegram and IRC
- Non-durable and durable memory with BoltDB and Redis
- Two ready to use rulesets: regex parsed messages and cron events
- Container ready to use and easy to deploy

Requirements:
[glide](https://github.com/Masterminds/glide) and [Go 1.5 or newer](http://golang.org).

### Environmental Variables

Basic
 * `GOCHATBOT_NAME` - defines the name of the bot. It will honor this name to
 answer messages in chatrooms.

Durable Memory (BoltDB)
 * `GOCHATBOT_BOLT_FILENAME` - local file name of BoltDB file. Set it if you
 want BoltDB support activated.

Durable Memory (Redis)
 * `GOCHATBOT_REDIS_DATABASE` - Database index
 * `GOCHATBOT_REDIS_HOST` - Redis Hostname (with port)
 * `GOCHATBOT_REDIS_PASSWORD` - Redis database password

IRC message provider
 * `GOCHATBOT_IRC_USER` - IRC user to connect to the server
 * `GOCHATBOT_IRC_NICK` - IRC nickname for this account (will not handle nick
 renaming)
 * `GOCHATBOT_IRC_SERVER` - IRC server
 * `GOCHATBOT_IRC_CHANNELS` - comma-separated list of channels to connect to.
 * `GOCHATBOT_IRC_PASSWORD` - IRC password for user
 * `GOCHATBOT_IRC_TLS` - whether TLS connection must be used or not (0/1)

Slack message provider
 * `GOCHATBOT_SLACK_TOKEN` - Slack user token for the chatbot

Telegram message provider
 * `GOCHATBOT_TELEGRAM_TOKEN` - Telegram user token for the chatbot

### Quick start (Docker version - Slack - Non-durable memory)

```ShellSession
# docker run -e "GOCHATBOT_SLACK_TOKEN=xxxx-xxxx" -d ccirello/gochatbot
```

### Quick start (Local Docker version - CLI - Non-durable memory)

```ShellSession
# brew install glide
# make docker
# docker run -ti ccirello/gochatbot
```
(Type `gochatbot jump` and press enter. Do it twice. `docker stop` to exit.)

### Quick start (Compiled version - Telegram - BoltDB memory)

```ShellSession
# brew install glide
# make all
# GOCHATBOT_TELEGRAM_TOKEN=#####:... GOCHATBOT_BOLT_FILENAME=gochatbot.db ./gochatbot
```

### Extending

A more thorough manual of how to extend the bot is yet to be written. This
section will cover just adding, removing and modifying rules into the two
shipped rulesets: regex and cron.

#### Extending regex ruleset

Edit the file `rules.go` in package root directory. Look for the line starting
with: `var regexRules = []regex.Rule{`. This is the beginning of the data
structure that comprises all regex rules.

Why regex rule? Because it uses regex to extract from the body of the message
the parameters for later message parsing.

All regex rules have the following structure:
```Go
{
	`regex pattern (.*)`, `explanation of the rule`,
	func(bot bot.Self, msg string, args []string) []string {
		var ret []string

		// your logic comes here

		return ret
	},
},
```

`regex pattern (.*)` - is the regex match pattern to be used against the
incoming message. Note that you can tell apart the messages that are sent in the
room from those sent to the chatbot, by append a Go `text/template` variable.

```Go
`{{ .RobotName }} jump`, `tells the robot to jump`, // Only messages starting with bot's name will be parsed
// vs
`jump`, `tells the robot to jump`, // All messages whose content matches jump will be matched
```

`explanation of the rule` - is a human readable explanation of the rule, meant
to be displayed with `{{ .RobotName }} help` command (if you haven't changed the
bot's name, it will be `gochatbot help`).

`func(bot bot.Self, msg string, args []string) []string {` - is the function
which parses and reacts on the incoming message. Few important as aspects:
 * Each incoming message is its own goroutine. It means you can execute blocking
 calls and the bot will keep working as usual.
 * Look at [bot package](https://godoc.org/cirello.io/gochatbot/bot) documentation
 for what you can do with the bot. In the following example, we'll see how we
 can use the bot's brain to store state across messages.
 * `msg` is the raw version of the message.
 * `args` is the slice of strings that matched against the `regex pattern (.*)`.
 * It must return a slice of strings, even if it is empty.

So, this is a practical and annotated version of the `jump` regex rule.
```Go
var regexRules = []regex.Rule{
	{
		`{{ .RobotName }} jump`, // Regex rule. No matching, so args will be empty
		`tells the robot to jump`, // Just a simple explanation of the rule
		func(bot bot.Self, msg string, args []string) []string {
			var ret []string

			ret = append(ret, "{{ .User }}, How high?") // In the messages, the text/template variable "{{ .User }}" is replaced with username.

			lastJumpTS := bot.MemoryRead("jump", "lastJump") // Reads from the bot's brain the last time this command was executed.
			ret = append(ret, fmt.Sprint("{{ .User }} (last time I jumped:", lastJumpTS, ")")) // Append this information to the outgoing messages slice.

			bot.MemorySave("jump", "lastJump", fmt.Sprint(time.Now())) // Saves into the bot's brain that it has just jumped.

			return ret
		},
	},
}
```

#### Extending cron ruleset

Edit the file `rules.go` in package root directory. Look for the line starting
with: `cronRules = map[string]cron.Rule{`. This is the beginning of the data
structure that comprises all cron rules.

Why cron rule? Because it uses crontab format to periodically execute tasks,
that may, or may not yield messages onto a chat room.

All croon rules have the following structure:
```Go
"job name": {
	"crontab format",
	func() []messages.Message {
		return []messages.Message{}
	},
},
```

`"job name"` - the name of the cron task to be executed. Each cron task must be
attached to a chatroom, otherwise they don't get executed. This attachment is
done through the command "cron attach _task-name_" in the desired chatroom.

`"crontab format"` - the crontab-like set that tells how often this rule is
executed. Refer to the
[cronexpr](https://github.com/gorhill/cronexpr#implementation) for the
implementation.

`func() []messages.Message {` - the niladic function that's executed
periodically. It returns a slice of `messages.Message`. But the only required
value is the `messages.Message.Message` string field. It will overwrite the
other values with the context information for correct delivery.

So, this is a practical and annotated version of the `good morning` cron rule.
```Go
var cronRules = map[string]cron.Rule{
	// name of the cron rule
	"message of the day": {
		"0 10 * * *", // every day at 10:00
		func() []messages.Message {
			return []messages.Message{
				{Message: "Good morning!"}, // Returns "Good Morning"
			}
		},
	},
}
```

### Guarantees

I guarantee that I will maintain this chatops bot for the next 2 years, provide
it with updates, Github Issues support and issue updates to keep it compatible
with newer Go versions.

Also, I will work to my best to ensure a vibrant community around this bot, so
even in the case I step down, I hope by the end of the guaranteed period, to
have a project bigger than one man effort.

The last day of guaranteed action on this bot is `2017-12-05`.

