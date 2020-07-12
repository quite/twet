// -*- tab-width: 4; -*-

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/peterh/liner"
)

func FollowingCommand(args []string) error {
	fs := flag.NewFlagSet("following", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	rawFlag := fs.Bool("r", false, "output following users in machine parsable format")

	fs.Usage = func() {
		fmt.Printf("usage: %s following [arguments]\n\nDisplays a list of users being followed.\n\n", progname)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return fmt.Errorf("error parsing arguments")
	}

	if fs.NArg() > 0 {
		return fmt.Errorf("too many arguments given")
	}

	for nick, url := range conf.Following {
		if *rawFlag {
			PrintFolloweeRaw(nick, url)
		} else {
			PrintFollowee(nick, url)
		}
		fmt.Println()
	}

	return nil
}

func FollowCommand(args []string) error {
	fs := flag.NewFlagSet("follow", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)

	fs.Usage = func() {
		fmt.Printf("usage: %s follow <nick> <twturl>\n\nStart following @<nick url>.\n\n", progname)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return fmt.Errorf("error parsing arguments")
	}

	if fs.NArg() < 2 {
		return fmt.Errorf("too few arguments given")
	}

	nick := fs.Args()[0]
	url := fs.Args()[1]

	conf.Following[nick] = url
	if err := conf.Write(); err != nil {
		return fmt.Errorf("error: writing config failed with  %s", err)
	}

	fmt.Printf("%s successfully started following %s @ %s", yellow("✓"), blue(nick), url)

	return nil
}

func UnfollowCommand(args []string) error {
	fs := flag.NewFlagSet("unfollow", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)

	fs.Usage = func() {
		fmt.Printf("usage: %s unfollow <nick>\n\nStop following @nick.\n\n", progname)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return fmt.Errorf("error parsing arguments")
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("too few arguments given")
	}

	nick := fs.Args()[0]
	delete(conf.Following, nick)
	if err := conf.Write(); err != nil {
		return fmt.Errorf("error: writing config failed with  %s", err)
	}

	fmt.Printf("%s successfully stopped following %s", yellow("✓"), blue(nick))

	return nil
}

func TimelineCommand(args []string) error {
	fs := flag.NewFlagSet("timeline", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	durationFlag := fs.Duration("d", 0, "only show tweets created at most `duration` back in time. Example: -d 12h")
	sourceFlag := fs.String("s", "", "only show timeline for given nick (URL, if dry-run)")
	fullFlag := fs.Bool("f", false, "display full timeline (overrides timeline config)")
	dryFlag := fs.Bool("n", false, "dry-run, only locally cached tweets")
	rawFlag := fs.Bool("r", false, "output tweets in URL-prefixed twtxt format")
	reversedFlag := fs.Bool("desc", false, "tweets shown in descending order (newer tweets at top)")

	fs.Usage = func() {
		fmt.Printf("usage: %s timeline [arguments]\n\nDisplays the timeline.\n\n", progname)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return fmt.Errorf("error parsing arguments")
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("too many arguments given")
	}
	if *durationFlag < 0 {
		return fmt.Errorf("negative duration doesn't make sense")
	}
	if *fullFlag {
		if *durationFlag > 0 {
			return fmt.Errorf("full timeline with duration makes no sense")
		}
		conf.Timeline = "full"
	}

	cache := LoadCache(configpath)
	cacheLastModified, err := CacheLastModified(configpath)
	if err != nil {
		return fmt.Errorf("error calculating last modified cache time: %s", err)
	}

	var sourceURL string

	if !*dryFlag {
		var sources = conf.Following
		if *sourceFlag != "" {
			url, ok := conf.Following[*sourceFlag]
			if !ok {
				return fmt.Errorf("no source with nick %q", *sourceFlag)
			}
			sources = make(map[string]string)
			sources[*sourceFlag] = url
			sourceURL = url
		}

		cache.FetchTweets(sources)
		cache.Store(configpath)

		// Did the url for *sourceFlag change?
		if sources[*sourceFlag] != conf.Following[*sourceFlag] {
			sources[*sourceFlag] = conf.Following[*sourceFlag]
			sourceURL = conf.Following[*sourceFlag]
		}
	}

	if debug && *dryFlag {
		log.Print("dry run\n")
	}

	var tweets Tweets
	if *sourceFlag != "" {
		tweets = cache.GetByURL(sourceURL)
	} else {
		for _, url := range conf.Following {
			tweets = append(tweets, cache.GetByURL(url)...)
		}
	}
	if *reversedFlag {
		sort.Sort(sort.Reverse(tweets))
	} else {
		sort.Sort(tweets)
	}

	now := time.Now()
	for _, tweet := range tweets {
		if (*durationFlag > 0 && now.Sub(tweet.Created) <= *durationFlag) ||
			(conf.Timeline == "full" && *durationFlag == 0) ||
			(conf.Timeline == "new" && tweet.Created.Sub(cacheLastModified) >= 0) {
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
	fs := flag.NewFlagSet("tweet", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Printf(`usage: %s tweet [words]
   or: %s twet [words]

Adds a new tweet to your twtfile. Words are joined together with a single
space. If no words are given, user will be prompted to input the text
interactively.
`, progname, progname)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return fmt.Errorf("error parsing arguments")
	}

	twtfile := conf.Twtfile
	if twtfile == "" {
		return fmt.Errorf("cannot tweet without twtfile set in config")
	}
	// We don't support shell style ~user/foo.txt :P
	if strings.HasPrefix(twtfile, "~/") {
		twtfile = strings.Replace(twtfile, "~", homedir, 1)
	}

	var text string
	if fs.NArg() == 0 {
		var err error
		if text, err = getLine(); err != nil {
			return fmt.Errorf("readline: %v", err)
		}
	} else {
		text = strings.Join(fs.Args(), " ")
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return fmt.Errorf("cowardly refusing to tweet empty text, or only spaces")
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

func getLine() (string, error) {
	l := liner.NewLiner()
	defer l.Close()
	l.SetCtrlCAborts(true)
	l.SetMultiLineMode(true)
	l.SetTabCompletionStyle(liner.TabCircular)
	l.SetBeep(false)

	var tags, nicks []string
	for tag := range LoadCache(configpath).GetAll().Tags() {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	for nick := range conf.Following {
		nicks = append(nicks, nick)
	}
	sort.Strings(nicks)

	l.SetCompleter(func(line string) (candidates []string) {
		i := strings.LastIndexAny(line, "@#")
		if i == -1 {
			return
		}

		vocab := nicks
		if line[i] == '#' {
			vocab = tags
		}
		i++

		for _, item := range vocab {
			if strings.HasPrefix(strings.ToLower(item), strings.ToLower(line[i:])) {
				candidates = append(candidates, line[:i]+item)
			}
		}

		return
	})

	return l.Prompt("> ")
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
