# Contributing
Pull requests and issues are always welcome. Bad attitudes are not. Collaborators have the ability to close PRs or issues that are seen as harassing or otherwise offensive to them with little or no explanation.

## Creating a Pull Request (PR)
1. Create a [GitHub Account](https://github.com/signup/free)
2. [Fork the pachelbel repository](https://help.github.com/articles/fork-a-repo/)
3. Set-up your local environment
    ```bash
    $ mkdir go ; cd go
    $ export GOPATH=$PWD
    $ mkdir -p src/github.com/<your_namespace>
    ```

4.  Clone your fork
    ```bash
    $ cd src/github.com/<your_namespace>
    $ git clone git@github.com:<your_namespace>/pachelbel
    $ cd pachelbel
    $ ls
    CONTRIBUTING.md LICENSE         Makefile        README.md       cmd             config          connection      main.go         vendor
    ```
5. Set up upstream repo tracking
    ```bash
    $ git remote add upstream git@github.com/benjdewan/pachelbel
    ```
6. Make your changes and ensure the build still passes by running make:
    ```bash
    $ make
    go get -u github.com/alecthomas/gometalinter github.com/kardianos/govendor
    /Users/bdewan/watson/golang/bin/gometalinter --install
      ... #omitting some output for brevity
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go install -ldflags "-X github.com/benjdewan/pachelbel/cmd.version=v0.3.0-3-g8d10621" github.com/benjdewan/pachelbel
    CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go install -ldflags "-X github.com/benjdewan/pachelbel/cmd.version=v0.3.0-3-g8d10621" github.com/benjdewan/pachelbel
    CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go install -ldflags "-X github.com/benjdewan/pachelbel/cmd.version=v0.3.0-3-g8d10621" github.com/benjdewan/pachelbel
    $ $GOPATH/bin/pachelbel --help
    pachelbel provisions and deprovisions deployments of compose resources
    in an idempotent fashion.

    Usage:
      pachelbel [command]
      ... #omitting some more output for brevity
    ```
7. Push your change to GitHub and [open a pull request](https://developer.github.com/v3/pulls/)

I will try and review new pull requests as quickly as I am able, but life can get in the way.
