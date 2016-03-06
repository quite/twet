// -*- tab-width: 4; -*-

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"gopkg.in/readline.v1"
)

func timeline_command(args []string) {
	fs := flag.NewFlagSet("timeline", flag.ExitOnError)
	durationFlag := fs.Duration("d", 0, "only show tweets created at most `duration` back in time. Example: -d 12h")
	fs.Usage = func() {
		fmt.Printf("usage: %s timeline [arguments]\n\nDisplays the timeline.\n\n", progname)
		fs.PrintDefaults()
	}
	fs.Parse(args) // currently using flag.ExitOnError, so we won't get an error on -h
	if fs.NArg() > 0 {
		fmt.Printf("Too many arguments given.\n")
		os.Exit(2)
	}

	if *durationFlag < 0 {
		fmt.Printf("Negative duration doesn't make sense.\n")
		os.Exit(2)
	}

	if len(conf.Following) == 0 {
		fmt.Printf("You're not following anyone.\n")
		os.Exit(0)
	}

	cache := Loadcache(configpath)

	now := time.Now().Round(time.Second)

	alltweets := get_tweets(cache)
	sort.Sort(alltweets)
	for _, tweet := range alltweets {
		if *durationFlag == 0 || (now.Sub(tweet.Created)) <= *durationFlag {
			print_tweet(tweet, now)
			fmt.Println()
		}
	}

	cache.Store(configpath)
}

func getline() (string, error) {
	rl, err := readline.New("> ")
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	line, err := rl.Readline()
	if err != nil { // io.EOF, readline.ErrInterrupt
		return "", err
	}
	return line, nil
}

func tweet_command(args []string) error {
	fs := flag.NewFlagSet("tweet", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Printf(`usage: %s tweet [words]
   or: %s twet [words]

Adds a new tweet to your twtfile. Words are joined together with a single
space. If no words are given, user will be prompted to input the text
interactively.
`, progname, progname)
		fs.PrintDefaults()
	}
	fs.Parse(args) // currently using flag.ExitOnError, so we won't get an error on -h

	twtfile := conf.Twtfile
	if len(twtfile) == 0 {
		return errors.New("cannot tweet without twtfile set in config")
	}
	// We don't support shell style ~user/foo.txt :P
	if strings.HasPrefix(twtfile, "~/") {
		twtfile = strings.Replace(twtfile, "~", homedir, 1)
	}

	var text string
	if fs.NArg() == 0 {
		var err error
		if text, err = getline(); err != nil {
			return fmt.Errorf("readline: %v", err)
		}
	} else {
		text = strings.Join(fs.Args(), " ")
	}
	text = strings.TrimSpace(text)
	if len(text) == 0 {
		return errors.New("cowardly refusing to tweet empty text, or only spaces")
	}
	text = fmt.Sprintf("%s\t%s\n", time.Now().Format(time.RFC3339), expand_mentions(text))
	f, err := os.OpenFile(twtfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	var n int
	if n, err = f.WriteString(text); err != nil {
		return err
	}
	fmt.Printf("appended %d bytes to %s:\n%s", n, conf.Twtfile, text)

	return nil
}

// Turns "@nick" into "@<nick URL>" if we're following nick.
func expand_mentions(text string) string {
	re := regexp.MustCompile(`@([_a-zA-Z0-9]+)`)
	return re.ReplaceAllStringFunc(text, func(match string) string {
		parts := re.FindStringSubmatch(match)
		mentionednick := parts[1]

		for followednick, followedurl := range conf.Following {
			if mentionednick == followednick {
				return fmt.Sprintf("@<%s %s>", followednick, followedurl)
			}
		}
		// Not expanding if we're not following
		return match
	})
}
