// -*- tab-width: 4; -*-

package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/goware/urlx"
)

func underline(s string) string {
	return fmt.Sprintf("\033[4m%s\033[0m", s)
}
func bold(s string) string {
	return fmt.Sprintf("\033[1m%s\033[0m", s)
}
func red(s string) string {
	return fmt.Sprintf("\033[31m%s\033[0m", s)
}
func green(s string) string {
	return fmt.Sprintf("\033[32m%s\033[0m", s)
}
func boldgreen(s string) string {
	return fmt.Sprintf("\033[32;1m%s\033[0m", s)
}
func blue(s string) string {
	return fmt.Sprintf("\033[34m%s\033[0m", s)
}

func PrintTweet(tweet Tweet, now time.Time) {
	text := ShortenMentions(tweet.Text)

	nick := green(tweet.Tweeter.Nick)
	if NormalizeURL(tweet.Tweeter.URL) == NormalizeURL(conf.Twturl) {
		nick = boldgreen(tweet.Tweeter.Nick)
	}
	fmt.Printf("> %s (%s)\n%s\n",
		nick,
		PrettyDuration(now.Sub(tweet.Created)),
		text)
}

func PrintTweetRaw(tweet Tweet) {
	fmt.Printf("%s\t%s\t%s",
		tweet.Tweeter.URL,
		tweet.Created.Format(time.RFC3339),
		tweet.Text)
}

// Turns "@<nick URL>" into "@nick" if we're following URL (or it's us!). If
// we're following as another nick then "@nick(followednick)".
func ShortenMentions(text string) string {
	re := regexp.MustCompile(`@<([^ ]+) *([^>]+)>`)
	return re.ReplaceAllStringFunc(text, func(match string) string {
		parts := re.FindStringSubmatch(match)
		nick, url := parts[1], parts[2]
		if fnick := conf.urlToNick(url); fnick != "" {
			return FormatMention(nick, url, fnick)
		}
		// Not shortening if we're not following
		return match
	})
}

func NormalizeURL(url string) string {
	if url == "" {
		return ""
	}
	u, err := urlx.Parse(url)
	if err != nil {
		return ""
	}
	if u.Scheme == "https" {
		u.Scheme = "http"
		u.Host = strings.TrimSuffix(u.Host, ":443")
	}
	u.User = nil
	u.Path = strings.TrimSuffix(u.Path, "/")
	norm, err := urlx.Normalize(u)
	if err != nil {
		return ""
	}
	return norm
}

// Takes followednick to be able to indicated when somebody (URL) was mentioned
// using a nick other than the one we follow the person as.
func FormatMention(nick string, url string, followednick string) string {
	str := "@" + nick
	if followednick != nick {
		str += fmt.Sprintf("(%s)", followednick)
	}
	if NormalizeURL(url) == NormalizeURL(conf.Twturl) {
		return red(str)
	}
	return blue(str)
}

func PrettyDuration(duration time.Duration) string {
	s := int(duration.Seconds())
	d := s / 86400
	s = s % 86400
	if d >= 365 {
		return fmt.Sprintf("%dy %dw ago", d/365, d%365/7)
	}
	if d >= 14 {
		return fmt.Sprintf("%dw ago", d/7)
	}
	h := s / 3600
	s = s % 3600
	if d > 0 {
		str := fmt.Sprintf("%dd", d)
		if h > 0 && d <= 6 {
			str += fmt.Sprintf(" %dh", h)
		}
		return str + " ago"
	}
	m := s / 60
	s = s % 60
	if h > 0 || m > 0 {
		str := ""
		hh := ""
		if h > 0 {
			str += fmt.Sprintf("%dh", h)
			hh = " "
		}
		if m > 0 {
			str += fmt.Sprintf("%s%dm", hh, m)
		}
		return str + " ago"
	}
	return fmt.Sprintf("%ds ago", s)
}
