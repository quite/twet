// -*- tab-width: 4; -*-

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"strings"
)

const progname = "twet"

var homedir string
var conf Config
var configpath string

var usage = fmt.Sprintf(`%s is a client for twtxt -- https://twtxt.readthedocs.org/en/stable/

Usage:
	%s command [arguments]

Commands:
	timeline
	tweet or twet

Use "%s help [command]" for more information about a command.
`, progname, progname, progname)

func main() {
	setversion()
	log.SetPrefix(fmt.Sprintf("%s: ", progname))
	log.SetFlags(0)

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	homedir = usr.HomeDir

	configpath = conf.Read()

	flag.Usage = func() {
		fmt.Printf(usage)
	}
	flag.Parse()
	switch flag.Arg(0) {
	case "timeline":
		if err := timeline_command(flag.Args()[1:]); err != nil {
			log.Fatal(err)
		}
	case "tweet", "twet":
		if err := tweet_command(flag.Args()[1:]); err != nil {
			log.Fatal(err)
		}
	case "help":
		switch flag.Arg(1) {
		case "timeline":
			timeline_command([]string{"-h"})
		case "tweet", "twet":
			tweet_command([]string{"-h"})
		case "":
			flag.Usage()
			os.Exit(2)
		default:
			fmt.Fprintf(os.Stderr, "Unknown help topic %q.\n", flag.Arg(1))
			os.Exit(2)
		}
	case "version":
		fmt.Printf("%s %s\nbuilt: %s\n", progname, progversion, buildtimestamp)
	case "":
		flag.Usage()
		os.Exit(2)
	default:
		log.Fatal(fmt.Sprintf("%q is not a valid command.\n", flag.Arg(0)))
	}
}

func setversion() {
	progversion = strings.TrimPrefix(gitontag, "v")
	if progversion == "" {
		progversion = strings.TrimPrefix(gitlasttag, "v")
		if gitcommit != "" {
			progversion += "+" + gitcommit
		}
	}
}

var (
	progversion    string = "v0.0.1" // setversion.sh
	buildtimestamp string
	gitontag       string
	gitlasttag     string
	gitcommit      string
)
