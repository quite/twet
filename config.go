// -*- tab-width: 4; -*-

package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-yaml/yaml"
)

type Hooks struct {
	Pre  string
	Post string
}

type Config struct {
	Nick             string
	Twturl           string
	Twtfile          string
	Following        map[string]string // nick -> url
	DiscloseIdentity bool
	Timeline         string
	Hooks            Hooks
	nicks            map[string]string // normalizeURL(url) -> nick
	path             string            // location of loaded config
}

func (conf *Config) Write() error {
	if conf.path == "" {
		return errors.New("error: no config file path found")
	}

	data, err := yaml.Marshal(conf)
	if err != nil {
		return fmt.Errorf("error marshalling config: %s", err)
	}

	return ioutil.WriteFile(conf.path, data, 0666)
}

func (conf *Config) Parse(data []byte) error {
	return yaml.Unmarshal(data, conf)
}

func (conf *Config) Read(confdir string) string {
	var paths []string
	if confdir != "" {
		paths = append(paths, confdir)
	} else {
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			paths = append(paths, fmt.Sprintf("%s/twet", xdg))
		}
		paths = append(paths,
			fmt.Sprintf("%s/.config/twet", homedir),
			fmt.Sprintf("%s/Library/Application Support/twet", homedir),
			fmt.Sprintf("%s/.twet", homedir))
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

	if conf.Timeline == "" {
		conf.Timeline = "full"
	}
	conf.Timeline = strings.ToLower(conf.Timeline)
	if conf.Timeline != "new" && conf.Timeline != "full" {
		log.Fatal(fmt.Sprintf("unexpected config timeline: %s", conf.Timeline))
	}

	conf.path = filepath.Join(foundpath, filename)
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
