// -*- tab-width: 4; -*-

package main

import (
	"log"
	"os"
	"os/user"
	"sort"
	"time"

	"github.com/fatih/color"
)

var homedir string
var conf Config

func main() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	homedir = usr.HomeDir

	// color even on non-tty (less!)
	color.NoColor = false

	configpath := conf.Read()
	cache := Loadcache(configpath)

	alltweets := get_tweets(cache)
	sort.Sort(alltweets)
	for _, tweet := range alltweets {
		print_tweet(tweet, time.Now().Round(time.Second))
	}

	cache.Store(configpath)

	os.Exit(0)

	return
}
