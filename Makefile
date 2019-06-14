
BUILDDIR ?= build

GITVERSION ?= $(shell git describe --tags --long --dirty)
BUILDTIMESTAMP := $(shell date +%s)

LDFLAGS="\
-X main.gitVersion=$(GITVERSION) \
-X main.buildTimestamp=$(BUILDTIMESTAMP)"

BINARY := twet
SRCS := $(wildcard *.go)

$(BUILDDIR)/$(BINARY): $(SRCS)
	go build -ldflags $(LDFLAGS) -o $@

$(BUILDDIR):
	mkdir -p $(BUILDDIR)
