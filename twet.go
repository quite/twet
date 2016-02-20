// -*- tab-width: 4; -*-

package main

import (
	"os"
	"sort"
	"time"

	"github.com/fatih/color"
)

var conf Config

func main() {
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
