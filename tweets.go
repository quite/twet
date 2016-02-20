// -*- tab-width: 4; -*-

package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

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

func get_tweets(cache Cache) Tweets {
	var alltweets Tweets

	client := http.Client{}

	for nick, url := range conf.Following {
		// TODO for test...
		var cacheonly bool = true
		if cacheonly == false {
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				// TODO handle in different way; when can this happen?
				fmt.Fprintf(os.Stderr, "http.NewRequest(..%s..) failed with: %s", url, err)
				continue
			}

			if cached, ok := cache[url]; ok {
				if cached.Lastmodified != "" {
					req.Header.Set("If-Modified-Since", cached.Lastmodified)
				}
			}

			resp, err := client.Do(req)
			if err != nil {
				// TODO handle in different way; when can this happen?
				fmt.Fprintf(os.Stderr, "client.Do failed with: %s", url, err)
				continue
			}

			actualurl := resp.Request.URL.String()
			if actualurl != url {
				url = actualurl
			}

			var tweets Tweets

			switch resp.StatusCode {
			case 200:
				scanner := bufio.NewScanner(resp.Body)
				tweets = parse_file(scanner, Tweeter{Nick: nick, URL: url})
				lastmodified := resp.Header.Get("Last-Modified")
				cache[url] = Cached{Tweets: tweets, Lastmodified: lastmodified}
			case 304:
				tweets = cache[url].Tweets
			}

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

	return alltweets
}

func parse_file(scanner *bufio.Scanner, tweeter Tweeter) Tweets {
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
