# Constants
VERSION := $(shell git describe --tags || echo DEV-BUILD)
LDFLAGS := "-w -X github.com/benjdewan/pachelbel/cmd.version=$(VERSION)"
export CGO_ENABLED := 0
export GOARCH := amd64

# Targets/Source
TARGETS := pachelbel-linux pachelbel-windows.exe pachelbel-darwin
SOURCE := $(shell find . -type f -iname '*.go')

# Executables
GOVENDOR := $(GOPATH)/bin/govendor
GOMETALINTER := $(GOPATH)/bin/gometalinter

# Sanity Check
ifndef GOPATH
    $(error GOPATH must be defined to build this project)
endif

default: setup lint test $(TARGETS)
.PHONY: all

all: $(TARGETS)
.PHONY: all

setup:
	go get -u github.com/alecthomas/gometalinter github.com/kardianos/govendor
	$(GOMETALINTER) --install
	$(GOVENDOR) sync
.PHONY: setup

pachelbel-%.exe: $(SOURCE)
	GOOS=$* $(GOVENDOR) build -ldflags $(LDFLAGS) -o "$@"

pachelbel-%: $(SOURCE)
	GOOS=$* $(GOVENDOR) build -ldflags $(LDFLAGS) -o "$@"

lint:
	$(GOMETALINTER) --deadline=30s cmd/ connection/ main.go
.PHONY: lint

test:
	$(GOVENDOR) test -v +local
.PHONY: test

clean:
	rm -rf $(TARGETS)
.PHONY: clean
