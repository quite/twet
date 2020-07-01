// -*- tab-width: 4; -*-

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

const progname = "twet"

var homedir string
var conf Config
var configpath string

var debug bool
var dir string
var usage = fmt.Sprintf(`%s is a client for twtxt -- https://twtxt.readthedocs.org/en/stable/

Usage:
	%s [flags] command [arguments]

Commands:
	timeline
	tweet or twet

Use "%s help [command]" for more information about a command.

Flags:
`, progname, progname, progname)

func main() {
	log.SetPrefix(fmt.Sprintf("%s: ", progname))
	log.SetFlags(0)

	if homedir = os.Getenv("HOME"); homedir == "" {
		log.Fatal("HOME env variable empty?! can't proceed")
	}

	flag.CommandLine.SetOutput(os.Stdout)
	flag.BoolVar(&debug, "debug", false, "output debug info")
	flag.StringVar(&dir, "dir", "", "set config directory")
	flag.Usage = func() {
		fmt.Print(usage)
		flag.PrintDefaults()
	}
	flag.Parse()
	configpath = conf.Read(dir)

	switch flag.Arg(0) {
	case "timeline":
		if err := TimelineCommand(flag.Args()[1:]); err != nil {
			log.Fatal(err)
		}
	case "tweet", "twet":
		if conf.Hooks.Pre != "" {
			if _, err := execShell(homedir, conf.Hooks.Pre); err != nil {
				log.Fatalf("error executing pre tweet hook: %s", err)
			}
		}

		if err := TweetCommand(flag.Args()[1:]); err != nil {
			log.Fatal(err)
		}

		if conf.Hooks.Post != "" {
			if _, err := execShell(homedir, conf.Hooks.Post); err != nil {
				log.Fatalf("error executing post tweet hook: %s", err)
			}
		}
	case "help":
		switch flag.Arg(1) {
		case "timeline":
			_ = TimelineCommand([]string{"-h"})
		case "tweet", "twet":
			_ = TweetCommand([]string{"-h"})
		case "":
			flag.Usage()
			os.Exit(2)
		default:
			log.Printf("Unknown help topic %q.\n", flag.Arg(1))
			os.Exit(2)
		}
	case "version":
		fmt.Printf("%s %s\n", progname, GetVersion())
	case "":
		flag.Usage()
		os.Exit(2)
	default:
		log.Fatal(fmt.Sprintf("%q is not a valid command.\n", flag.Arg(0)))
	}
}
