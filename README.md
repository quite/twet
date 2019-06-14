
# twet
[![Build Status](https://travis-ci.org/quite/twet.svg?branch=master)](https://travis-ci.org/quite/twet)

twet is a simple client in Go for
[`twtxt`](https://github.com/buckket/twtxt) -- *the decentralised, minimalist
microblogging service for hackers*.

Please see the [TODO](TODO.md).

## Configuration

twet looks for `config.yaml` in these directories:

```
  $XDG_BASE_DIR/config/twet
  $HOME/config/twet
  $HOME/Library/Application Support/twet
  $HOME/.twet
```

Or you can set a directory with `twet -dir /some/dir`.

The config looks like this:

```
  # define yourself! this is the author:
  nick: quite
  twturl: https://lublin.se/twtxt.txt

  # tweets appended here
  twtfile: ~/public_html/twtxt.txt

  following:
    twtxt: https://buckket.org/twtxt_news.txt
```

A cache file will be stored next to the config file.

If you want to read your own tweets, you should follow yourself. The `twturl`
above is used for highlighting mentions, and for revealing who you are in the
HTTP User-Agent when fetching feeds.
