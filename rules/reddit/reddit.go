package reddit // import "cirello.io/gochatbot/rules/reddit"

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/jeffail/gabs"

	"cirello.io/gochatbot/bot"
	"cirello.io/gochatbot/messages"
)

// Golang/.json
const baseURL = "https://www.reddit.com/r/"

type redditRuleset struct {
	outCh chan messages.Message

	mu         sync.Mutex
	subreddits map[string][]string
	recents    map[string]string

	memoryRead func(string, string) []byte
	memorySave func(string, string, []byte)
}

// Name returns this rules name - meant for debugging.
func (r *redditRuleset) Name() string {
	return "Reddit Ruleset"
}

// Boot runs preparatory steps for ruleset execution
func (r *redditRuleset) Boot(self *bot.Self) {
	r.outCh = self.MessageProviderOut()
	r.memoryRead = self.MemoryRead
	r.memorySave = self.MemorySave
	r.loadMemory()
}

func (r *redditRuleset) loadMemory() {
	log.Println("reddit: reading from memory")
	if err := json.Unmarshal(r.memoryRead("reddit", "follow"), &r.subreddits); err == nil {
		log.Println("reddit: memory (follow) read")
	}

	if err := json.Unmarshal(r.memoryRead("reddit", "recents"), &r.recents); err == nil {
		log.Println("reddit: memory (recent) read")
	}
	go r.start()
}

func (r redditRuleset) HelpMessage(self bot.Self) string {
	helpMsg := fmt.Sprintln("reddit follow <subreddit>- follow one subreddit in a room")
	helpMsg = fmt.Sprintln(helpMsg, "reddit unfollow <subreddit> - unfollow one subreddit in a room")
	helpMsg = fmt.Sprintln(helpMsg, "reddit list - list the followed subreddits in a room")
	helpMsg = fmt.Sprintln(helpMsg, "reddit help - this message")

	return helpMsg
}

func (r *redditRuleset) ParseMessage(self bot.Self, in messages.Message) []messages.Message {
	if strings.HasPrefix(in.Message, "reddit follow") {
		subreddit := strings.TrimSpace(strings.TrimPrefix(in.Message, "reddit follow"))
		ret := []messages.Message{
			{
				Room:       in.Room,
				ToUserID:   in.FromUserID,
				ToUserName: in.FromUserName,
				Message:    r.follow(subreddit, in.Room),
			},
		}
		return ret
	}

	if strings.HasPrefix(in.Message, "reddit unfollow") {
		subreddit := strings.TrimSpace(strings.TrimPrefix(in.Message, "reddit unfollow"))
		return []messages.Message{
			{
				Room:       in.Room,
				ToUserID:   in.FromUserID,
				ToUserName: in.FromUserName,
				Message:    r.unfollow(subreddit, in.Room),
			},
		}
	}

	if strings.HasPrefix(in.Message, "reddit list") {
		return []messages.Message{
			{
				Room:       in.Room,
				ToUserID:   in.FromUserID,
				ToUserName: in.FromUserName,
				Message:    r.list(in.Room),
			},
		}
	}
	return []messages.Message{}
}

func (r *redditRuleset) list(room string) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	return fmt.Sprintln("followed subreddits in this room:", r.subreddits[room])
}

func (r *redditRuleset) follow(subreddit, room string) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !strings.HasPrefix(subreddit, "/r/") {
		subreddit = fmt.Sprint("/r/", subreddit)
	}

	subredditURL, err := subredditURL(subreddit)
	if err != nil {
		return "could not follow " + subreddit
	}

	for _, sr := range r.subreddits[room] {
		if sr == subredditURL {
			return subreddit + " already followed in this room"
		}
	}

	r.subreddits[room] = append(r.subreddits[room], subredditURL)

	b, err := json.Marshal(r.subreddits)
	if err != nil {
		return "could not follow " + subreddit
	}
	r.memorySave("reddit", "follow", b)
	return subredditURL + " followed in this room"
}

func (r *redditRuleset) unfollow(subreddit, room string) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.subreddits[room]; !ok {
		return "room not found in reddit memory"
	}

	if !strings.HasPrefix(subreddit, "/r/") {
		subreddit = fmt.Sprint("/r/", subreddit)
	}

	url, err := subredditURL(subreddit)
	if err != nil {
		return subreddit + " cannot not be removed"
	}
	var newRoom []string
	for _, sr := range r.subreddits[room] {
		if sr == url {
			continue
		}
		newRoom = append(newRoom, url)
	}
	r.subreddits[room] = newRoom

	b, err := json.Marshal(r.subreddits)
	if err != nil {
		return "could not unfollow " + subreddit
	}
	r.memorySave("reddit", "follow", b)

	return subreddit + " not followed in this room anymore"
}

func (r *redditRuleset) start() {
	c := time.Tick(30 * time.Second)
	for range c {
		r.mu.Lock()
		for room, subreddits := range r.subreddits {
			for _, subreddit := range subreddits {
				r.readSubreddit(subreddit, room)
			}
		}
		r.mu.Unlock()
	}
}

func (r *redditRuleset) readSubreddit(subreddit, room string) {
	resp, err := http.Get(subreddit + ".json")
	if err != nil {
		log.Printf("redit: error loading subreddit %s. got: %v", subreddit, err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("redit: error reading subreddit %s. got: %v", subreddit, err)
		return
	}

	jsonParsed, err := gabs.ParseJSON(body)
	if err != nil {
		log.Printf("redit: error parsing subreddit json %s. got: %v", subreddit, err)
		return
	}

	children, err := jsonParsed.S("data", "children").Children()
	if err != nil {
		log.Printf("redit: error parsing subreddit json %s. got: %v", subreddit, err)
		return
	}

	var recent string
	subredditName := strings.TrimSuffix(strings.TrimPrefix(subreddit, baseURL), ".json")
	if _, ok := r.recents[subredditName]; !ok {
		r.recents[subredditName] = ""
	}

	for _, child := range children {
		title := child.Path("data.title").Data()
		url := child.Path("data.url").Data()

		if recent == "" {
			recent = fmt.Sprint(title, url)
		}

		if fmt.Sprint(title, url) == r.recents[subredditName] {
			break
		}

		r.outCh <- messages.Message{
			Room:    room,
			Message: fmt.Sprint("/r/", subredditName, ": ", title, " (", url, ")"),
		}

		if r.recents[subredditName] == "" {
			break
		}
	}

	r.recents[subredditName] = recent

	b, err := json.Marshal(r.recents)
	if err != nil {
		log.Printf("redit: error serializing subreddit json. got: %v", err)
	}
	r.memorySave("reddit", "recents", b)
}

// New returns a reddit rule set
func New() *redditRuleset {
	r := &redditRuleset{
		subreddits: make(map[string][]string),
		recents:    make(map[string]string),
	}
	return r
}

func subredditURL(subreddit string) (string, error) {
	u, err := url.Parse(subreddit)
	if err != nil {
		return "", err
	}
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	return base.ResolveReference(u).String(), nil
}
