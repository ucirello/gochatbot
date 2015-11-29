// +build !all,!bolt,!redis

package brain

import "log"

func init() {
	availableDrivers = append(availableDrivers, func(getenv func(string) string) (Memorizer, bool) {
		log.Println("You have not built gochatbot with any durable database.")
		log.Println("It will not persist state across service restarts.")
		log.Println("Consider rebuilding it choosing one or more of these drivers:")
		log.Println("  $ go build -tags bolt # Bolt")
		log.Println("  $ go build -tags redis # Redis")
		log.Println("Should you want to have all drivers available, run:")
		log.Println("  $ go build -tags all")
		return nil, false
	})
}
