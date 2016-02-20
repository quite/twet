// -*- tab-width: 4; -*-

package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"net/http"

	"github.com/fatih/color"
)

// TODO when add/removing following in config, what happens?
// TODO require url (and nick...)  unique in following?
// TODO panics etc

var conf Config

type Tweeter struct {
	Nick string
	URL  string
}
type Tweet struct {
	Tweeter Tweeter
	Created time.Time
	Text    string
}

// typedef to be able to attach sort methods
type Tweets []Tweet

func (tweets Tweets) Len() int {
	return len(tweets)
}
func (tweets Tweets) Less(i, j int) bool {
	return tweets[i].Created.Before(tweets[j].Created)
}
func (tweets Tweets) Swap(i, j int) {
	tweets[i], tweets[j] = tweets[j], tweets[i]
}

func get_tweets(scanner *bufio.Scanner, tweeter Tweeter) Tweets {
	var tweets Tweets
	i := 0
	for scanner.Scan() {
		i += 1
		a := strings.SplitN(scanner.Text(), "\t", 2)
		if len(a) != 2 {
			fmt.Fprintf(os.Stderr,
				color.RedString("could not parse: ", scanner.Text()))
		} else {
			tweet := Tweet{
				Tweeter: tweeter,
				Created: parsetime(a[0]),
				Text:    a[1],
			}
			tweets = append(tweets, tweet)
		}
	}
	return tweets
}

func parsetime(timestr string) time.Time {
	var tm time.Time
	var err error
	// Twtxt clients generally uses basically time.RFC3339Nano, but sometimes
	// there's a colon in the timezone, or no timezone at all.
	for _, layout := range []string{
		"2006-01-02T15:04:05.999999999Z07:00",
		"2006-01-02T15:04:05.999999999Z0700",
		"2006-01-02T15:04:05.999999999",
	} {
		tm, err = time.Parse(layout, timestr)
		if err != nil {
			continue
		} else {
			break
		}
	}
	if err != nil {
		return time.Unix(0, 0)
	}
	return tm.Round(time.Second)
}

func print_tweet(tweet Tweet, now time.Time) {
	text := shorten_mentions(tweet.Text)

	underline := color.New(color.Underline).SprintFunc()
	fmt.Printf("> %s (%s)\n  %s\n",
		underline(tweet.Tweeter.Nick),
		pretty_duration(now.Sub(tweet.Created)),
		text)
}

func shorten_mentions(text string) string {
	// a mention: @<somenick url>
	re := regexp.MustCompile(`@<([^ ]+) ([^>]+)>`)
	return re.ReplaceAllStringFunc(text, func(match string) string {
		parts := re.FindStringSubmatch(match)
		mentioned := Tweeter{
			Nick: parts[1],
			URL:  parts[2],
		}

		for followednick, followedurl := range conf.Following {
			if mentioned.URL == followedurl {
				return format_mention(mentioned, followednick)
			}
		}
		// Maybe we got mentioned ourselves?
		if mentioned.URL == conf.Twturl {
			return format_mention(mentioned, conf.Nick)
		}
		// Couldn't
		return match
	})
}

// Takes followednick to be able to indicated when somebody (URL) was mentioned
// using a nick other than the one we follow the person as.
func format_mention(mentioned Tweeter, followednick string) string {
	str := "@" + mentioned.Nick
	if followednick != mentioned.Nick {
		str = str + fmt.Sprintf("(%s)", followednick)
	}
	coloring := color.New(color.Bold).SprintFunc()
	if mentioned.URL == conf.Twturl {
		coloring = color.New(color.FgBlue).SprintFunc()
	}
	return coloring(str)
}

func pretty_duration(duration time.Duration) string {
	if duration.Hours() > 24 {
		return fmt.Sprintf("%d days ago", int(duration.Hours())/24)
	}
	if duration.Minutes() > 60 {
		return fmt.Sprintf("%d hours ago", int(duration.Hours()))
	}
	if duration.Seconds() > 60 {
		return fmt.Sprintf("%d minutes ago", int(duration.Minutes()))
	}
	return fmt.Sprintf("%d seconds ago", int(duration.Seconds()))
}

// TODO... FIXME
var cacheonly bool = true

func main() {
	// color even on non-tty (less!)
	color.NoColor = false

	configpath := conf.Read()

	// TODO let cache have func too? or have conf not have...eh
	cache := Load(configpath)

	var alltweets Tweets

	client := http.Client{}

	// what the httpgetter needs: url, lastmod
	for nick, url := range conf.Following {
		if cacheonly == false {
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				panic(err)
			}

			if cached, ok := cache[url]; ok {
				// makes sense yeah? some gave us 200 before, but no last-modifed...
				if cached.Lastmodified != "" {
					req.Header.Set("If-Modified-Since", cached.Lastmodified)
				}
			}

			resp, err := client.Do(req)
			if err != nil {
				panic(err)
			}

			// TODO handle redirect?
			actualurl := resp.Request.URL.String()
			if actualurl != url {
				url = actualurl
			}

			lastmodified := resp.Header.Get("Last-Modified")

			var tweets Tweets

			// TODO, stuff and 401, 301? 404? ...
			switch resp.StatusCode {
			case 200:
				scanner := bufio.NewScanner(resp.Body)
				tweets = get_tweets(scanner, Tweeter{
					Nick: nick,
					URL:  url,
				})
				cache[url] = Cached{
					Tweets:       tweets,
					Lastmodified: lastmodified,
				}
			case 304:
				tweets = cache[url].Tweets
			}

			//TODO defer?
			resp.Body.Close()

			alltweets = append(alltweets, tweets...)
		} else {
			var tweets Tweets
			if cached, ok := cache[url]; ok {
				tweets = cached.Tweets
			}
			alltweets = append(alltweets, tweets...)
		}
	}

	sort.Sort(alltweets)

	// TODO: limit from commandline
	for _, tweet := range alltweets {
		print_tweet(tweet, time.Now().Round(time.Second))
	}

	cache.Store(configpath)

	os.Exit(0)

	return
}
