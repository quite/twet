// -*- tab-width: 4; -*-

package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

func underline(s string) string {
	return fmt.Sprintf("\033[4m%s\033[0m", s)
}
func bold(s string) string {
	return fmt.Sprintf("\033[1m%s\033[0m", s)
}
func blue(s string) string {
	return fmt.Sprintf("\033[34m%s\033[0m", s)
}

func print_tweet(tweet Tweet, now time.Time) {
	text := shorten_mentions(tweet.Text)

	fmt.Printf("> %s (%s)\n  %s\n",
		underline(tweet.Tweeter.Nick),
		pretty_duration(now.Sub(tweet.Created)),
		text)
}

// Turns "@<nick URL>" into "@nick" if we're following URL (or it's us!). If
// we're following as another nick then "@nick(followednick)".
func shorten_mentions(text string) string {
	re := regexp.MustCompile(`@<([^ ]+) *([^>]+)>`)
	return re.ReplaceAllStringFunc(text, func(match string) string {
		parts := re.FindStringSubmatch(match)
		mentioned := Tweeter{
			Nick: parts[1],
			URL:  parts[2],
		}
		for followednick, followedurl := range conf.Following {
			if isSameURL(mentioned.URL, followedurl) {
				return format_mention(mentioned, followednick)
			}
		}
		if conf.Nick != "" && conf.Twturl != "" {
			// Maybe we got mentioned ourselves?
			if isSameURL(mentioned.URL, conf.Twturl) {
				return format_mention(mentioned, conf.Nick)
			}
		}
		// Not shortening if we're not following
		return match
	})
}

func isSameURL(a string, b string) bool {
	a = strings.TrimPrefix(a, "http://")
	a = strings.TrimPrefix(a, "https://")
	a = strings.TrimSuffix(a, "/")
	b = strings.TrimPrefix(b, "http://")
	b = strings.TrimPrefix(b, "https://")
	b = strings.TrimSuffix(b, "/")
	return a == b
}

// Takes followednick to be able to indicated when somebody (URL) was mentioned
// using a nick other than the one we follow the person as.
func format_mention(mentioned Tweeter, followednick string) string {
	str := "@" + mentioned.Nick
	if followednick != mentioned.Nick {
		str = str + fmt.Sprintf("(%s)", followednick)
	}
	if conf.Twturl != "" && isSameURL(mentioned.URL, conf.Twturl) {
		return blue(str)
	}
	return bold(str)
}

func pretty_duration(duration time.Duration) string {
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
