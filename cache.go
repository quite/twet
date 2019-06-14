// -*- tab-width: 4; -*-

package main

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
)

type Cached struct {
	Tweets       Tweets
	Lastmodified string
}

// key: url
type Cache map[string]Cached

func (c Cache) Store(configpath string) {
	b := new(bytes.Buffer)
	enc := gob.NewEncoder(b)
	err := enc.Encode(c)
	if err != nil {
		panic(err)
	}

	f, err := os.OpenFile(fmt.Sprintf("%s/cache", configpath),
		os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	if _, err = f.Write(b.Bytes()); err != nil {
		panic(err)
	}
}

func LoadCache(configpath string) Cache {
	cache := make(Cache)

	f, err := os.Open(fmt.Sprintf("%s/cache", configpath))
	if err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
		return cache
	}
	defer f.Close()

	dec := gob.NewDecoder(f)
	err = dec.Decode(&cache)
	if err != nil {
		panic(err)
	}
	return cache
}

func (cache Cache) FetchTweets(sources map[string]string) {
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
					fmt.Sprintf("%s/%s (+%s; @%s)", progname, GetVersion(),
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

	var n = 0
	// loop until channel closed
	for tweets := range tweetsch {
		n++
		if debug {
			log.Printf("%d ", len(sources)+1-n)
		}
		if debug && len(tweets) > 0 {
			log.Printf("%s\n", tweets[0].Tweeter.URL)
		}
	}
	if debug {
		log.Print("\n")
	}
}

func (cache Cache) GetAll() Tweets {
	var alltweets Tweets
	for url, cached := range cache {
		alltweets = append(alltweets, cached.Tweets...)
		if debug {
			log.Printf("%s\n", url)
		}
	}
	return alltweets
}

func (cache Cache) GetByURL(URL string) Tweets {
	if cached, ok := cache[URL]; ok {
		return cached.Tweets
	}
	return Tweets{}
}
