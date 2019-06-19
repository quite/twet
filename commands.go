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

	"github.com/chzyer/readline"
)

func TimelineCommand(args []string) error {
	fs := flag.NewFlagSet("timeline", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	durationFlag := fs.Duration("d", 0, "only show tweets created at most `duration` back in time. Example: -d 12h")
	sourceFlag := fs.String("s", "", "only show timeline for given nick (URL, if dry-run)")
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

	cache := LoadCache(configpath)

	if !*dryFlag {
		var sources = conf.Following
		if *sourceFlag != "" {
			url, ok := conf.Following[*sourceFlag]
			if !ok {
				return fmt.Errorf("no source with nick %q", *sourceFlag)
			}
			sources = make(map[string]string)
			sources[*sourceFlag] = url
			*sourceFlag = url
		}

		cache.FetchTweets(sources)
		cache.Store(configpath)
	}

	if debug && *dryFlag {
		log.Print("dry run\n")
	}

	var tweets Tweets
	if *sourceFlag != "" {
		tweets = cache.GetByURL(*sourceFlag)
	} else {
		tweets = cache.GetAll()
	}
	if *reversedFlag {
		sort.Sort(sort.Reverse(tweets))
	} else {
		sort.Sort(tweets)
	}

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
		c := newCompleter(LoadCache(configpath).GetAll().Tags())
		if text, err = GetLine(c); err != nil {
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

func GetLine(completer readline.AutoCompleter) (string, error) {
	rl, err := readline.NewEx(&readline.Config{
		Prompt:       "> ",
		AutoComplete: completer,
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

type Completer struct {
	nicks []string
	tags  []string
}

func newCompleter(tags map[string]int) *Completer {
	c := new(Completer)

	for nick := range conf.Following {
		c.nicks = append(c.nicks, nick)
	}
	sort.Strings(c.nicks)

	for tag := range tags {
		c.tags = append(c.tags, tag)
	}
	sort.Strings(c.tags)

	return c
}

func (n *Completer) Do(line []rune, pos int) (newLine [][]rune, offset int) {
	linestr := string(line)
	i := strings.LastIndexAny(linestr[:pos], "@#")
	if i == -1 {
		return
	}

	words := n.nicks
	if linestr[i] == '#' {
		words = n.tags
	}
	i++
	wordpart := linestr[i:pos]

	for _, word := range words {
		if strings.HasPrefix(word, wordpart) {
			newLine = append(newLine, []rune(strings.TrimPrefix(word, wordpart)))
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
