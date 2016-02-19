// -*- tab-width: 4; -*-

package main

import (
	//	"strings"
	//	"log"
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

type HttpResponse struct {
	url      string
	response *http.Response
	err      error
}

/////////////
/////////////
/////////////
func asyncHttpGets(urls []string) []*HttpResponse {
	ch := make(chan *HttpResponse)
	responses := []*HttpResponse{}
	client := http.Client{}
	for _, url := range urls {
		go func(url string) {
			fmt.Printf("Fetching %s \n", url)
			resp, err := client.Get(url)
			ch <- &HttpResponse{url, resp, err}
			if err != nil && resp != nil && resp.StatusCode == http.StatusOK {
				resp.Body.Close()
			}
		}(url)
	}

	for {
		select {
		case r := <-ch:
			fmt.Printf("%s was fetched\n", r.url)
			if r.err != nil {
				fmt.Println("with an error", r.err)
			}
			responses = append(responses, r)
			if len(responses) == len(urls) {
				return responses
			}
		case <-time.After(50 * time.Millisecond):
			fmt.Printf(".")
		}
	}
	///////return responses
}

/////////////
/////////////
/////////////

// for _, httpresp := range asyncHttpGets(urls) {
// 	if httpresp != nil && httpresp.response != nil {
// 		fmt.Printf("%s status: %s\n", httpresp.url,
// 			httpresp.response.Status)
// 		fmt.Printf("\n'%s'\n", httpresp.response.Body)
// 	}
// }

/////////////
/////////////
/////////////

type Tweeter struct {
	Nick string
	URL  string
}

type Tweet struct {
	Tweeter Tweeter
	Created time.Time
	Text    string
}

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

func formatmention(mentioned Tweeter, followednick string) string {
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

func printtweet(tweet Tweet, now time.Time) {
	// twtxt regexen:
	// mention_re = re.compile(r'@<(?:(?P<name>\S+?)\s+)?(?P<url>\S+?://.*?)>')
	// short_mention_re = re.compile(r'@(?P<name>\w+)')
	// TODO: optional <name> in re?
	re := regexp.MustCompile(`@<([^ ]+) ([^>]+)>`)
	text := re.ReplaceAllStringFunc(tweet.Text,
		func(match string) string {
			parts := re.FindStringSubmatch(match)
			mentioned := Tweeter{Nick: parts[1], URL: parts[2]}

			for followednick, followedurl := range conf.Following {
				if mentioned.URL == followedurl {
					return formatmention(mentioned, followednick)
				}
			}
			if mentioned.URL == conf.Twturl {
				return formatmention(mentioned, conf.Nick)
			}
			return fmt.Sprintf("@<%s %s>", mentioned.Nick, mentioned.URL)
		})

	// color even on non-tty
	color.NoColor = false

	underline := color.New(color.Underline).SprintFunc()
	fmt.Printf("> %s (%s)\n  %s\n",
		underline(tweet.Tweeter.Nick),
		pretty_duration(now.Sub(tweet.Created)),
		text)
}

func parsetime(timestr string) time.Time {
	var tm time.Time
	var err error
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

func gettweets(scanner *bufio.Scanner, tweeter Tweeter) Tweets {
	var tweets Tweets
	i := 0
	for scanner.Scan() {
		i += 1
		a := strings.SplitN(scanner.Text(), "\t", 2)
		if len(a) != 2 {
			fmt.Fprintf(os.Stderr, color.RedString("could not parse: ", scanner.Text()))
		} else {
			tweets = append(tweets, Tweet{tweeter, parsetime(a[0]), a[1]})
		}
	}
	return tweets
}

// TODO... FIXME
var cacheonly bool = true

func main() {
	configpath := conf.Read()

	// TODO let cache have func too? or have conf not have...eh
	cache := Load(configpath)

	var alltweets Tweets

	client := http.Client{}

	for nick, url := range conf.Following {
		fmt.Fprintf(os.Stderr, "* fetching %s %s", nick, url)

		if cacheonly == false {
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				panic(err)
			}

			if cached, ok := cache[url]; ok {
				// makes sense yeah? some gave us 200 before, but no last-modifed...
				if cached.Lastmodified != "" {
					req.Header.Set("If-Modified-Since", cached.Lastmodified)
					fmt.Fprintf(os.Stderr, " [I-M-S:'%s']", cached.Lastmodified)
				}
			}

			/////////////
			// dump, _ := httputil.DumpRequestOut(req, true)
			// fmt.Fprintf(os.Stderr, "\n%q\n", dump)

			resp, err := client.Do(req)
			if err != nil {
				panic(err)
			}

			/////////////
			// dump, _ = httputil.DumpResponse(resp, true)
			// fmt.Fprintf(os.Stderr, "%q\n", dump)

			// TODO handle redirect?
			actualurl := resp.Request.URL.String()
			if actualurl != url {
				fmt.Fprintf(os.Stderr, "\n * %s->%s\n", url, actualurl)
				url = actualurl
			}

			lastmodified := resp.Header.Get("Last-Modified")
			fmt.Fprintf(os.Stderr, " _%d_ | L-M:'%s'", resp.StatusCode, lastmodified)

			fmt.Fprintf(os.Stderr, "\n")

			var tweets Tweets

			// TODO, stuff and 401, 301? 404? ...
			switch resp.StatusCode {
			case 200:
				scanner := bufio.NewScanner(resp.Body)
				tweets = gettweets(scanner, Tweeter{Nick: nick, URL: url})
				cache[url] = Cached{Tweets: tweets, Lastmodified: lastmodified}
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
		printtweet(tweet, time.Now().Round(time.Second))
	}

	cache.Store(configpath)

	os.Exit(0)

	return
}
