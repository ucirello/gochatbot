package main // import "cirello.io/gochatbot/plugins/gochatbot-plugin-reddit"

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/jeffail/gabs"

	"cirello.io/gochatbot/messages"
	"cirello.io/gochatbot/plugins"
)

const baseURL = "https://www.reddit.com/r/"

type RedditPlugin struct {
	comm       *plugins.Comm
	mu         sync.Mutex
	subreddits map[string][]string
	recents    map[string]map[string]string
}

func main() {
	rpcBind := os.Getenv("GOCHATBOT_RPC_BIND")
	if rpcBind == "" {
		log.Fatal("GOCHATBOT_RPC_BIND empty or not set. Cannot start plugin.")
	}
	r := &RedditPlugin{
		comm:       plugins.NewComm(rpcBind),
		subreddits: make(map[string][]string),
		recents:    make(map[string]map[string]string),
	}
	r.loadMemory()
	go r.watch()
	for {
		in, err := r.comm.Pop()
		if err != nil {
			log.Println("reddit: error popping message from gochatbot:", err)
			continue
		}
		if in.Message == "" {
			time.Sleep(1 * time.Second)
		}
		if err := r.parseMessage(in); err != nil {
			log.Println("reddit: error parsing message:", err)
		}
	}
}

func (r *RedditPlugin) loadMemory() {
	log.Println("reddit: reading from memory")
	followMem, err := r.comm.MemoryRead("reddit", "follow")
	if err != nil {
		log.Println("reddit: error memory (follow) read:", err)
		return
	}
	if err := json.Unmarshal(followMem, &r.subreddits); err == nil {
		log.Println("reddit: memory (follow) read")
	}

	recentsMem, err := r.comm.MemoryRead("reddit", "recents")
	if err != nil {
		log.Println("reddit: error memory (recent) read:", err)
		return
	}
	if err := json.Unmarshal(recentsMem, &r.recents); err == nil {
		log.Println("reddit: memory (recent) read")
	}
}

func (r RedditPlugin) helpMessage() string {
	helpMsg := fmt.Sprintln("reddit follow <subreddit>- follow one subreddit in a room")
	helpMsg = fmt.Sprintln(helpMsg, "reddit unfollow <subreddit> - unfollow one subreddit in a room")
	helpMsg = fmt.Sprintln(helpMsg, "reddit list - list the followed subreddits in a room")
	helpMsg = fmt.Sprintln(helpMsg, "reddit help - this message")

	return helpMsg
}

func (r *RedditPlugin) parseMessage(in *messages.Message) error {
	if in.Message == "help" || in.Message == "reddit help" {
		return r.comm.Send(&messages.Message{
			Room:       in.Room,
			ToUserID:   in.FromUserID,
			ToUserName: in.FromUserName,
			Message:    r.helpMessage(),
		})
	}
	if strings.HasPrefix(in.Message, "reddit follow") {
		subreddit := strings.TrimSpace(strings.TrimPrefix(in.Message, "reddit follow"))
		return r.comm.Send(&messages.Message{
			Room:       in.Room,
			ToUserID:   in.FromUserID,
			ToUserName: in.FromUserName,
			Message:    r.follow(subreddit, in.Room),
		})
	}

	if strings.HasPrefix(in.Message, "reddit unfollow") {
		subreddit := strings.TrimSpace(strings.TrimPrefix(in.Message, "reddit unfollow"))
		return r.comm.Send(&messages.Message{
			Room:       in.Room,
			ToUserID:   in.FromUserID,
			ToUserName: in.FromUserName,
			Message:    r.unfollow(subreddit, in.Room),
		})
	}

	if strings.HasPrefix(in.Message, "reddit list") {
		return r.comm.Send(&messages.Message{
			Room:       in.Room,
			ToUserID:   in.FromUserID,
			ToUserName: in.FromUserName,
			Message:    r.list(in.Room),
		})
	}

	return nil
}

func (r *RedditPlugin) list(room string) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	return fmt.Sprintln("followed subreddits in this room:", r.subreddits[room])
}

func (r *RedditPlugin) follow(subreddit, room string) string {
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
	r.comm.MemorySave("reddit", "follow", b)
	return subredditURL + " followed in this room"
}

func (r *RedditPlugin) unfollow(subreddit, room string) string {
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
		newRoom = append(newRoom, sr)
	}
	r.subreddits[room] = newRoom

	b, err := json.Marshal(r.subreddits)
	if err != nil {
		return "could not unfollow " + subreddit
	}
	r.comm.MemorySave("reddit", "follow", b)

	return subreddit + " not followed in this room anymore"
}

func (r *RedditPlugin) watch() {
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

func (r *RedditPlugin) readSubreddit(subreddit, room string) {
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
	if _, ok := r.recents[room]; !ok {
		r.recents[room] = make(map[string]string)
	}
	if _, ok := r.recents[room][subredditName]; !ok {
		r.recents[room][subredditName] = ""
	}

	for _, child := range children {
		title := child.Path("data.title").Data()
		url := child.Path("data.url").Data()

		if recent == "" {
			recent = fmt.Sprint(title, url)
		}

		if fmt.Sprint(title, url) == r.recents[room][subredditName] {
			break
		}

		r.comm.Send(&messages.Message{
			Room:    room,
			Message: fmt.Sprint("/r/", subredditName, ": ", title, " (", url, ")"),
		})

		if r.recents[room][subredditName] == "" {
			break
		}
	}

	r.recents[room][subredditName] = recent

	b, err := json.Marshal(r.recents)
	if err != nil {
		log.Printf("redit: error serializing subreddit json. got: %v", err)
	}
	r.comm.MemorySave("reddit", "recents", b)
}

func subredditURL(subreddit string) (string, error) {
	u, err := url.Parse(strings.ToLower(subreddit))
	if err != nil {
		return "", err
	}
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	return base.ResolveReference(u).String(), nil
}
