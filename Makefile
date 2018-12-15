.NOTPARALLEL:

export GO111MODULE = on
export GOARCH ?=

ifeq ($(GOARCH),)
BINNAME = wolssh
else
BINNAME = wolssh-$(GOARCH)
endif

all: build

build:
	go build -v -mod=vendor -o $(BINNAME)

clean:
	rm -f $(BINNAME)

goclean: clean
	go clean -cache

mod:
	go mod tidy -v
	go mod vendor -v

.PHONY: all build clean goclean mod
