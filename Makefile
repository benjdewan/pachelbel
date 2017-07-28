CC=gcc

.PHONY: all clean

all: post-build

pre-build:
ifndef GOPATH
    $(error GOPATH must be defined to build this project)
endif
ifdef GOBIN
    $(error To do cross-compilation GOBIN cannot be set)
endif
	go get -v github.com/alecthomas/gometalinter
	$(GOPATH)/bin/gometalinter --install

post-build: linux-build macos-build windows-build

linux-build: test-build
	GOOS=linux GOARCH=amd64 go install github.com/benjdewan/pachelbel

macos-build: test-build
	GOOS=darwin GOARCH=amd64 go install github.com/benjdewan/pachelbel

windows-build: test-build
	GOOS=windows GOARCH=amd64 go install github.com/benjdewan/pachelbel

test-build: pre-build
	$(GOPATH)/bin/gometalinter cmd/ connection/ main.go

clean:
	rm -rf $(GOPATH)/bin
