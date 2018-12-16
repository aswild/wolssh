.NOTPARALLEL:

export GO111MODULE = on
export GOARCH ?=

GO ?= go

LDFLAGS ?= -s -w

ifeq ($(GOARCH),)
BINNAME = wolssh
else
BINNAME = wolssh-$(GOARCH)
endif

all: build

build:
	$(GO) build -v -mod=vendor -ldflags="$(LDFLAGS)" -o $(BINNAME) $(BUILDFLAGS)

clean:
	rm -f $(BINNAME)

goclean: clean
	$(GO) clean -cache

mod:
	$(GO) mod tidy -v
	$(GO) mod vendor -v

.PHONY: all build clean goclean mod
