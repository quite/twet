// -*- tab-width: 4; -*-

package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v1"
)

type Config struct {
	Nick      string
	Twturl    string
	Following map[string]string
}

func (c *Config) Parse(data []byte) error {
	if err := yaml.Unmarshal(data, c); err != nil {
		return err
	}
	if len(c.Following) < 1 {
		return errors.New("not following anyone!")
	}
	if c.Nick == "" || c.Twturl == "" {
		return errors.New("both nick and twtwurl must be set!")
	}
	return nil
}

func (c *Config) Read() string {
	var paths []string
	if xdg := os.Getenv("XDG_BASE_DIR"); xdg != "" {
		paths = append(paths, fmt.Sprintf("%s/config/twet", xdg))
	}
	paths = append(paths, fmt.Sprintf("%s/config/twet", os.Getenv("HOME")))
	paths = append(paths, fmt.Sprintf("%s/.twet", os.Getenv("HOME")))

	filename := "config.yaml"

	for _, path := range paths {
		data, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", path, filename))
		if err != nil {
			// try next path
			continue
		}
		if err := c.Parse(data); err != nil {
			fmt.Println("config error: ", err)
			os.Exit(1)
		}
		return path
	}
	fmt.Println("could not find config file")
	os.Exit(1)
	return ""
}
