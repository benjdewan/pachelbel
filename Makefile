LDFLAGS="-X github.com/benjdewan/pachelbel/cmd.version=$(shell git describe --tags || echo DEV-BUILD)"

.PHONY: all clean

all: post-build

pre-build:
ifndef GOPATH
    $(error GOPATH must be defined to build this project)
endif
ifdef GOBIN
    $(error To do cross-compilation GOBIN cannot be set)
endif
	go get -u github.com/alecthomas/gometalinter github.com/kardianos/govendor
	$(GOPATH)/bin/gometalinter --install
	$(GOPATH)/bin/govendor sync

post-build: linux-build macos-build windows-build

linux-build: test-build
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go install -ldflags $(LDFLAGS) github.com/benjdewan/pachelbel

macos-build: test-build
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go install -ldflags $(LDFLAGS) github.com/benjdewan/pachelbel

windows-build: test-build
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go install -ldflags $(LDFLAGS) github.com/benjdewan/pachelbel

test-build: pre-build
	$(GOPATH)/bin/gometalinter cmd/ connection/ main.go

clean:
	rm -rf $(GOPATH)/bin
