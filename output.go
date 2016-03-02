// -*- tab-width: 4; -*-

package main

import (
	"fmt"
	"regexp"
	"time"

	"github.com/fatih/color"
)

func print_tweet(tweet Tweet, now time.Time) {
	text := shorten_mentions(tweet.Text)

	underline := color.New(color.Underline).SprintFunc()
	fmt.Printf("> %s (%s)\n  %s\n",
		underline(tweet.Tweeter.Nick),
		pretty_duration(now.Sub(tweet.Created)),
		text)
}

// Turns "@<nick URL>" into "@nick" if we're following URL (or it's us!). If
// we're following as another nick then "@nick(followednick)".
func shorten_mentions(text string) string {
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
		if conf.Nick != "" && conf.Twturl != "" {
			// Maybe we got mentioned ourselves?
			if mentioned.URL == conf.Twturl {
				return format_mention(mentioned, conf.Nick)
			}
		}
		// Not shortening if we're not following
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
	if conf.Twturl != "" {
		if mentioned.URL == conf.Twturl {
			coloring = color.New(color.FgBlue).SprintFunc()
		}
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
