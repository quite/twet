// -*- tab-width: 4; -*-

package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
)

type Cached struct {
	Tweets       Tweets
	Lastmodified string
}

// key: url
type Cache map[string]Cached

func (c Cache) Store(configpath string) {
	b := new(bytes.Buffer)
	enc := gob.NewEncoder(b)
	err := enc.Encode(c)
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

func Loadcache(configpath string) Cache {
	cache := make(Cache)

	f, err := os.Open(fmt.Sprintf("%s/cache", configpath))
	if err != nil {
		if os.IsNotExist(err) {
			return cache
		}
		panic(err)
	}
	defer f.Close()

	dec := gob.NewDecoder(f)
	err = dec.Decode(&cache)
	if err != nil {
		panic(err)
	}
	return cache
}
