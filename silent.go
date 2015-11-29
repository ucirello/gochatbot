// +build silent

package main

import (
	"io/ioutil"
	"log"
	"os"
)

func init() {
	verbose := os.Getenv("GOCHATBOT_VERBOSE")
	if verbose != "1" {
		log.SetOutput(ioutil.Discard)
	}
}
