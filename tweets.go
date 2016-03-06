// -*- tab-width: 4; -*-

package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
	"unicode"
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

const maxfetchers = 50

func get_tweets(cache Cache, sources map[string]string) Tweets {
	var mu sync.RWMutex

	tweetsch := make(chan Tweets)
	var wg sync.WaitGroup
	// max parallel http fetchers
	var fetchers = make(chan struct{}, maxfetchers)

	for nick, url := range sources {
		wg.Add(1)

		fetchers <- struct{}{}

		// anon func takes needed variables as arg, avoiding capture of iterator variables
		go func(nick string, url string) {
			defer wg.Done()

			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: http.NewRequest fail: %s", url, err)
				tweetsch <- nil
				return
			}

			if conf.Nick != "" && conf.Twturl != "" {
				// TODO: version goes here
				req.Header.Set("User-Agent",
					fmt.Sprintf("%s/0.1 (+%s; @%s)", progname, conf.Twturl, conf.Nick))
			}

			mu.RLock()
			if cached, ok := cache[url]; ok {
				if cached.Lastmodified != "" {
					req.Header.Set("If-Modified-Since", cached.Lastmodified)
				}
			}
			mu.RUnlock()

			client := http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: client.Do fail: %s", url, err)
				tweetsch <- nil
				return
			}
			defer resp.Body.Close()

			actualurl := resp.Request.URL.String()
			if actualurl != url {
				url = actualurl
			}

			var tweets Tweets

			switch resp.StatusCode {
			case http.StatusOK: // 200
				scanner := bufio.NewScanner(resp.Body)
				tweets = parse_file(scanner, Tweeter{Nick: nick, URL: url})
				lastmodified := resp.Header.Get("Last-Modified")
				mu.Lock()
				cache[url] = Cached{Tweets: tweets, Lastmodified: lastmodified}
				mu.Unlock()
			case http.StatusNotModified: // 304
				mu.RLock()
				tweets = cache[url].Tweets
				mu.RUnlock()
			}

			tweetsch <- tweets

		}(nick, url)

		<-fetchers
	}

	// close tweets channel when all goroutines are done
	go func() {
		wg.Wait()
		close(tweetsch)
	}()

	fmt.Fprintf(os.Stderr, "fetching: ")
	var alltweets Tweets
	var n = 0
	// loop until channel closed
	for tweets := range tweetsch {
		n++
		fmt.Fprintf(os.Stderr, "%d ", len(sources)+1-n)
		alltweets = append(alltweets, tweets...)
	}
	fmt.Fprintf(os.Stderr, "\n")

	return alltweets
}

func parse_file(scanner *bufio.Scanner, tweeter Tweeter) Tweets {
	var tweets Tweets
	for scanner.Scan() {
		line := strings.TrimRightFunc(scanner.Text(),
			func(r rune) bool {
				return unicode.IsSpace(r)
			})
		if len(line) == 0 {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			fmt.Fprintf(os.Stderr, "could not parse: '%s' (source:%s)\n", line, tweeter.URL)
			continue
		}
		tweets = append(tweets,
			Tweet{
				Tweeter: tweeter,
				Created: parsetime(parts[0]),
				Text:    parts[1],
			})
	}
	if err := scanner.Err(); err != nil {
		panic(err)
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
