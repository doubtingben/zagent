SHELL := /bin/bash

VERSION := `git describe --always`
GITCOMMIT := `git rev-parse HEAD`
BRANCH := `git rev-parse --abbrev-ref HEAD`
BUILDDATE := `date +%Y-%m-%d`
BUILDUSER := `whoami`

LDFLAGSSTRING :=-X github.com/doubtingben/zagent/cmd.Version=$(VERSION)
LDFLAGSSTRING +=-X github.com/doubtingben/zagent/cmd.GitCommit=$(GITCOMMIT)
LDFLAGSSTRING +=-X github.com/doubtingben/zagent/cmdBranch=$(BRANCH)
LDFLAGSSTRING +=-X github.com/doubtingben/zagent/cmd.BuildDate=$(BUILDDATE)
LDFLAGSSTRING +=-X github.com/doubtingben/zagent/cmd.BuildUser=$(BUILDUSER)

LDFLAGS :=-ldflags "$(LDFLAGSSTRING)"

.PHONY: all build

all: build

# Build binary
build:
	go build $(LDFLAGS) 

test:
	go test -v ./...