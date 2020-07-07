
# twet
[![Build Status](https://travis-ci.org/quite/twet.svg?branch=master)](https://travis-ci.org/quite/twet)

twet is a simple client in Go for
[`twtxt`](https://github.com/buckket/twtxt) -- *the decentralised, minimalist
microblogging service for hackers*.

Please see the [TODO](TODO.md).

## Configuration

twet looks for `config.yaml` in the following directories. Example
configuration in [`config.yaml.example`](config.yaml.example).

```
  $XDG_BASE_DIR/config/twet
  $HOME/config/twet
  $HOME/Library/Application Support/twet
  $HOME/.twet
```

Or you can set a directory with `twet -dir /some/dir`.

A cache file will be stored next to the config file.

If you want to read your own tweets, you should follow yourself. The `twturl`
above is used for highlighting mentions, and for revealing who you are in the
HTTP User-Agent when fetching feeds.

## TODO?

* http: think about redirect, and handling of 401, 301, 404?
* cli/http: a "follow" command should probably resolve 301s (cache-control or not?)
* cache: behaviour when adding/removing following
* following: require unique URL?
* ...
