// -*- tab-width: 4; -*-

package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
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

func GetTweets(cache Cache, sources map[string]string) Tweets {
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
				if debug {
					log.Printf("%s: http.NewRequest fail: %s", url, err)
				}
				tweetsch <- nil
				return
			}

			if conf.Nick != "" && conf.Twturl != "" {
				// TODO: version goes here
				req.Header.Set("User-Agent",
					fmt.Sprintf("%s/%s (+%s; @%s)", progname, progversion,
						conf.Twturl, conf.Nick))
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
				if debug {
					log.Printf("%s: client.Do fail: %s", url, err)
				}
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
				tweets = ParseFile(scanner, Tweeter{Nick: nick, URL: url})
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

	if debug {
		log.Print("fetching:\n")
	}
	var alltweets Tweets
	var n = 0
	// loop until channel closed
	for tweets := range tweetsch {
		n++
		if debug {
			log.Printf("%d ", len(sources)+1-n)
		}
		alltweets = append(alltweets, tweets...)
		if debug && len(tweets) > 0 {
			log.Printf("%s\n", tweets[0].Tweeter.URL)
		}
	}
	if debug {
		log.Print("\n")
	}

	return alltweets
}

func ParseFile(scanner *bufio.Scanner, tweeter Tweeter) Tweets {
	var tweets Tweets
	re := regexp.MustCompile(`^(.+?)(\s+)(.+)$`) // .+? is ungreedy
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}
		if strings.HasPrefix(line, "#") {
			if debug {
				log.Printf("skipped #-line: '%s' (source:%s)\n", line, tweeter.URL)
			}
			continue
		}
		parts := re.FindStringSubmatch(line)
		if len(parts) != 4 {
			if debug {
				log.Printf("could not parse: '%s' (source:%s)\n", line, tweeter.URL)
			}
			continue
		}
		tweets = append(tweets,
			Tweet{
				Tweeter: tweeter,
				Created: ParseTime(parts[1]),
				Text:    parts[3],
			})
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	return tweets
}

func ParseTime(timestr string) time.Time {
	var tm time.Time
	var err error
	// Twtxt clients generally uses basically time.RFC3339Nano, but sometimes
	// there's a colon in the timezone, or no timezone at all.
	for _, layout := range []string{
		"2006-01-02T15:04:05.999999999Z07:00",
		"2006-01-02T15:04:05.999999999Z0700",
		"2006-01-02T15:04:05.999999999",
		"2006-01-02T15:04.999999999Z07:00",
		"2006-01-02T15:04.999999999Z0700",
		"2006-01-02T15:04.999999999",
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
	return tm
}
