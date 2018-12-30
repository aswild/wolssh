.NOTPARALLEL:

export VERSION ?= $(shell scripts/version.sh)
export GO111MODULE = on
export GOARCH ?=

GO ?= go

GOFLAGS ?= -v
LDFLAGS ?= -s -w

NAME := wolssh
BINNAME := $(patsubst %-,%,$(NAME)-$(GOARCH))

all: build

build:
	$(GO) build -mod=vendor $(GOFLAGS) -ldflags='-X main.version=$(VERSION) $(LDFLAGS)' -o $(BINNAME)

clean:
	rm -f $(BINNAME) *.deb

goclean: clean
	$(GO) clean -cache

mod:
	$(GO) mod tidy -v
	$(GO) mod vendor -v

deb: build
	scripts/make-deb.sh

.PHONY: all build clean goclean mod deb
