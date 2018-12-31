# bin2go: program to create go byte arrays from files
#
# Copyright 2018 Allen Wild <allenwild93@gmail.com>
# SPDX-License-Identifier: MIT

.NOTPARALLEL:

BINNAME = bin2go
GOFLAGS ?= -v
LDFLAGS ?= -s -w

all: build

build:
	go build $(GOFLAGS) -ldflags='$(LDFLAGS)' -o $(BINNAME) ./cmd

test:
	go test -v ./...

clean:
	rm -f $(BINNAME)

goclean: clean
	go clean -cache

.PHONY: all build test clean goclean
