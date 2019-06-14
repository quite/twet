// -*- tab-width: 4; -*-

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/go-yaml/yaml"
)

type Config struct {
	Nick      string
	Twturl    string
	Twtfile   string
	Following map[string]string // nick -> url
	nicks     map[string]string // normalizeURL(url) -> nick
}

func (conf *Config) Parse(data []byte) error {
	return yaml.Unmarshal(data, conf)
}

func (conf *Config) Read(confdir string) string {
	var paths []string
	if confdir != "" {
		paths = append(paths, confdir)
	} else {
		if xdg := os.Getenv("XDG_BASE_DIR"); xdg != "" {
			paths = append(paths, fmt.Sprintf("%s/config/twet", xdg))
		}
		paths = append(paths, fmt.Sprintf("%s/config/twet", homedir))
		paths = append(paths, fmt.Sprintf("%s/Library/Application Support/twet", homedir))
		paths = append(paths, fmt.Sprintf("%s/.twet", homedir))
	}

	filename := "config.yaml"

	foundpath := ""
	for _, path := range paths {
		configfile := fmt.Sprintf("%s/%s", path, filename)
		data, err := ioutil.ReadFile(configfile)
		if err != nil {
			// try next path
			continue
		}
		if err := conf.Parse(data); err != nil {
			log.Fatal(fmt.Sprintf("error parsing config file: %s: %s", filename, err))
		}
		foundpath = path
		break
	}
	if foundpath == "" {
		log.Fatal(fmt.Sprintf("config file %q not found; looked in: %q", filename, paths))
	}
	return foundpath
}

func (conf *Config) urlToNick(url string) string {
	if conf.nicks == nil {
		conf.nicks = make(map[string]string)
		for n, u := range conf.Following {
			if u = NormalizeURL(u); u == "" {
				continue
			}
			conf.nicks[u] = n
		}
		if conf.Nick != "" && conf.Twturl != "" {
			conf.nicks[NormalizeURL(conf.Twturl)] = conf.Nick
		}
	}
	return conf.nicks[NormalizeURL(url)]
}
