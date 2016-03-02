
# twet

twet is a simple client in Go for
[```twtxt```](https://github.com/buckket/twtxt) -- *the decentralised, minimalist
microblogging service for hackers*.

Please see the [TODO](TODO.md).

## Configuration

twet looks for ```config.yaml``` in these directories:

```
$XDG_BASE_DIR/config/twet
$HOME/config/twet
$HOME/Library/Application Support/twet
$HOME/.twet
```

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

The ```cache``` file will be stored where the config is found.

