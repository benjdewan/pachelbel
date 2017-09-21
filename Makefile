# Constants
VERSION := $(shell git describe --tags || echo DEV-BUILD)
LDFLAGS := "-w -s -X github.com/benjdewan/pachelbel/cmd.version=$(VERSION)"
export CGO_ENABLED := 0
export GOARCH := amd64

# Targets/Source
TARGETS := pachelbel-linux pachelbel-windows.exe pachelbel-darwin
SOURCE := $(shell find . -type f -iname '*.go')

# Executables
DEP := $(GOPATH)/bin/dep
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
	go get -u github.com/alecthomas/gometalinter github.com/golang/dep/cmd/dep
	$(GOMETALINTER) --install
	$(DEP) ensure
.PHONY: setup

pachelbel-%.exe: $(SOURCE)
	GOOS=$* go build -ldflags $(LDFLAGS) -o "$@"

pachelbel-%: $(SOURCE)
	GOOS=$* go build -ldflags $(LDFLAGS) -o "$@"

lint:
	$(GOMETALINTER) --deadline=90s cmd/ connection/ config/ progress/ output/ main.go
.PHONY: lint

test:
	go test -v ./progress ./config ./output
.PHONY: test

clean:
	rm -rf $(TARGETS)
.PHONY: clean
