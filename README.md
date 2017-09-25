# Pachelbel
[![Build Status](https://travis-ci.org/benjdewan/pachelbel.svg?branch=master)](https://travis-ci.org/benjdewan/pachelbel)
[![Go Report Card](https://goreportcard.com/badge/github.com/benjdewan/pachelbel)](https://goreportcard.com/report/github.com/benjdewan/pachelbel)
[![codebeat badge](https://codebeat.co/badges/12ff5e7c-e6f5-4791-a706-2ee1a8fe7653)](https://codebeat.co/projects/github-com-benjdewan-pachelbel-master)

pachelbel is designed to be an idempotent provisioning tool for [IBM Compose](compose.io) deployments.

## Usage
`pachelbel` is built atop the [cobra](https://github.com/spf13/cobra) CLI framework,
so every one of these commands as help information built-in:
```console
$ pachelbel --help
```
```console
$ pachelbel provision --help
```
```console
$ pachelbel deprovision --help
```
```console
$ pachelbel version --help
```

### `pachelbel provision`
The most complicated and powerful command, `provision` is used to create and/or
modify existing Compose deployments. `provision` takes a list of yaml configuration
files and/or directories as input and returns a single yaml configuration file
containing connection information to those deployments as output.

The valuable aspect of this provision step is that it's idempotent. Re-running
`provision` with the same inputs will always produce the same output.

#### The `provision` input schema
`pachelbel provision` is designed to read yaml configuration files. The YAML objects read in must adhere to the following schemas:
* [v1 schema](schema/v1.md)
* [v2 schema](schema/v2.md)

Each object is versioned, and the v2 schema does not yet support the functionality of the v1 schema, so the mixing of objects from different schemas is both supported and expected.

Checkout the [the examples](examples/README.md) to see runnable input files as well as the commands to use them.


#### The `provision` output schema

Pachelbel's output schema is also a yaml file to be consumed by other tools in a configuration/deployment workflow. The schema is not currently strictly versioned.
* [output schema](schema/output.md)

### `pachelbel deprovision`
This command deprovisions existing Compose deployments. It takes a mixed list of
deployment names and deployment IDs as input parameters, resolves them to
deployment objects (or drops them if no matching deployment exists), and then
triggers deprovision API commands.

By default this command does not wait for the deprovision recipes to finish, but
you can use the `--wait` flag to force it to.

This command has no output on success.

### `pachelbel version`
This command prints the version.
