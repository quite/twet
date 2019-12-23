// -*- tab-width: 4; -*-

package main

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Cached struct {
	Tweets       Tweets
	Lastmodified string
}

// key: url
type Cache map[string]Cached

func (cache Cache) Store(configpath string) {
	b := new(bytes.Buffer)
	enc := gob.NewEncoder(b)
	err := enc.Encode(cache)
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

const maxfetchers = 50

func (cache Cache) FetchTweets(sources map[string]string) {
	var mu sync.RWMutex

	// buffered to let goroutines write without blocking before the main thread
	// begins reading
	tweetsch := make(chan Tweets, len(sources))

	var wg sync.WaitGroup
	// max parallel http fetchers
	var fetchers = make(chan struct{}, maxfetchers)

	for nick, url := range sources {
		wg.Add(1)
		fetchers <- struct{}{}
		// anon func takes needed variables as arg, avoiding capture of iterator variables
		go func(nick string, url string) {
			defer func() {
				<-fetchers
				wg.Done()
			}()

			if strings.HasPrefix(url, "file://") {
				err := ReadLocalFile(url, nick, tweetsch, cache, &mu)
				if err != nil {
					if debug {
						log.Printf("%s: Failed to read and cache local file: %s", url, err)
					}
				}
				return
			}

			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				if debug {
					log.Printf("%s: http.NewRequest fail: %s", url, err)
				}
				tweetsch <- nil
				return
			}

			if conf.Nick != "" && conf.Twturl != "" {
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

			client := http.Client{
				Timeout: time.Second * 15,
			}
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

func ReadLocalFile(url, nick string, tweetsch chan<- Tweets, cache Cache, mu sync.Locker) error {
	path := url[6:]
	file, err := os.Stat(path)
	if err != nil {
		if debug {
			log.Printf("%s: Can't stat local file: %s", path, err)
		}
		return err
	}
	if cached, ok := (cache)[url]; ok {
		if cached.Lastmodified == file.ModTime().String() {
			tweets := (cache)[url].Tweets
			tweetsch <- tweets
			return nil
		}
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		if debug {
			log.Printf("%s: Can't read local file: %s", path, err)
		}
		tweetsch <- nil
		return err
	}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	tweets := ParseFile(scanner, Tweeter{Nick: nick, URL: url})
	lastmodified := file.ModTime().String()
	mu.Lock()
	cache[url] = Cached{Tweets: tweets, Lastmodified: lastmodified}
	mu.Unlock()
	tweetsch <- tweets
	return nil
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

func (cache Cache) GetByURL(url string) Tweets {
	if cached, ok := cache[url]; ok {
		return cached.Tweets
	}
	return Tweets{}
}
