// -*- tab-width: 4; -*-

package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/chzyer/readline"
)

func TimelineCommand(args []string) error {
	fs := flag.NewFlagSet("timeline", flag.ExitOnError)
	durationFlag := fs.Duration("d", 0, "only show tweets created at most `duration` back in time. Example: -d 12h")
	sourceFlag := fs.String("s", "", "only show timeline for given nick (URL, if dry-run)")
	dryFlag := fs.Bool("n", false, "dry-run, only locally cached tweets")
	rawFlag := fs.Bool("r", false, "output tweets in ..TODO")
	fs.Usage = func() {
		fmt.Printf("usage: %s timeline [arguments]\n\nDisplays the timeline.\n\n", progname)
		fs.PrintDefaults()
	}
	fs.Parse(args) // currently using flag.ExitOnError, so we won't get an error on -h
	if fs.NArg() > 0 {
		return errors.New("too many arguments given")
	}
	if *durationFlag < 0 {
		return errors.New("negative duration doesn't make sense")
	}

	cache := LoadCache(configpath)
	var tweets Tweets

	if !*dryFlag {
		var sources map[string]string = conf.Following
		if *sourceFlag != "" {
			url, ok := conf.Following[*sourceFlag]
			if !ok {
				return errors.New(fmt.Sprintf("no source with nick %q", *sourceFlag))
			}
			sources = make(map[string]string)
			sources[*sourceFlag] = url
		}

		tweets = GetTweets(cache, sources)
		cache.Store(configpath)
	} else {
		if debug {
			log.Print("dry run\n")
		}
		tweets = CachedTweets(cache, *sourceFlag)
	}

	sort.Sort(tweets)
	now := time.Now()
	for _, tweet := range tweets {
		if *durationFlag == 0 || (now.Sub(tweet.Created)) <= *durationFlag {
			if !*rawFlag {
				PrintTweet(tweet, now)
			} else {
				PrintTweetRaw(tweet)
			}
			fmt.Println()
		}
	}

	return nil
}

func TweetCommand(args []string) error {
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
		if text, err = GetLine(); err != nil {
			return fmt.Errorf("readline: %v", err)
		}
	} else {
		text = strings.Join(fs.Args(), " ")
	}
	text = strings.TrimSpace(text)
	if len(text) == 0 {
		return errors.New("cowardly refusing to tweet empty text, or only spaces")
	}
	text = fmt.Sprintf("%s\t%s\n", time.Now().Format(time.RFC3339), ExpandMentions(text))
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

func GetLine() (string, error) {
	rl, err := readline.NewEx(&readline.Config{
		Prompt:       "> ",
		AutoComplete: new(NicksCompleter),
	})
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

type NicksCompleter struct{ nicks []string }

func (n *NicksCompleter) Do(line []rune, pos int) (newLine [][]rune, offset int) {
	if len(n.nicks) < len(conf.Following) {
		for nick, _ := range conf.Following {
			n.nicks = append(n.nicks, nick)
		}
		sort.Strings(n.nicks)
	}

	linestr := string(line)
	i := strings.LastIndex(string(line), "@")
	if i == -1 {
		return
	}
	i++
	nickpart := linestr[i:pos]

	for _, nick := range n.nicks {
		if strings.HasPrefix(nick, nickpart) {
			newLine = append(newLine, []rune(strings.TrimPrefix(nick, nickpart)))
		}
	}

	offset = len(linestr) - i
	return
}

// Turns "@nick" into "@<nick URL>" if we're following nick.
func ExpandMentions(text string) string {
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
