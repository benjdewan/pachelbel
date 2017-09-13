# Pachelbel
[![Build Status](https://travis-ci.org/benjdewan/pachelbel.svg?branch=master)](https://travis-ci.org/benjdewan/pachelbel)
[![Go Report Card](https://goreportcard.com/badge/github.com/benjdewan/pachelbel)](https://goreportcard.com/report/github.com/benjdewan/pachelbel)
[![codebeat badge](https://codebeat.co/badges/12ff5e7c-e6f5-4791-a706-2ee1a8fe7653)](https://codebeat.co/projects/github-com-benjdewan-pachelbel-master)

pachelbel is designed to be an idempotent provisioning tool for [IBM Compose](compose.io) deployments. It's still under heavy development

## Usage
pachelbel is built atop cobra and has extensive information available using `--help`.

## Input schema
pachelbel is designed to read yaml configuration files. The YAML objects read in must adhere to the following schemas:
* [v1 schema](schema/v1.md)
* [v2 schema](schema/v2.md)

Each object is versioned, and the v2 schema does not yet support the functionality of the v1 schema, so the mixing of objects from different schemas is both supported and expected.

## Output Schema

Pachelbel will write deployment connection information to disk if it successfully provisions all the requests deployments. The default location for this information is `./connection-info.yml`, but that can be overridden using the `--output` flag to specify a different file. The schema of this file is the following:

```yaml
<deployment_name>:
  cacert: {{.base64_encoded_ca_bundle}}
  connections:
  - host:     {{.host}}
    password: {{.password}}
    path:     {{.path}}
    port:     {{.port}}
    scheme: http|https|mongodb|mysql|amqp|amqp|postgres
    username: {{.username}}
  type: elasticsearch|etcd|janusgraph|mongodb|mysql|postgresql|rabbitmq|redis|scylladb
```

For example:
```bash
$ cat ./cluster-config.yml
config_version: 1
cluster: benjdewan-cluster-01
type: postgresql
name: postgres-benjdewan-02
notes: This is a pachelbel test provisioning
ssl: true
---
config_version: 1
cluster: benjdewan-cluster-02
type: redis
name: redis-benjdewan-01
notes: This is a pachelbel test provisioning
ssl: true

$ ./pachelbel-darwin provision --dry-run ./cluster-config.yml
Pretending to create 'postgres-benjdewan-02'
Pretending to create 'redis-benjdewan-01'
         postgres-benjdewan-02                    redis-benjdewan-01
░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░  ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
                  DONE                                    DONE

Writing connection strings to './connection-info.yml'

$ cat ./connection-info.yml
redis-benjdewan-01:
  cacert: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS...
  connections:
  - host: pachelbel-dry-run.compose.direct
    password: Ez57510qVFnK7obJYKr3
    port: 4801
    scheme: redis
    username: alice
  type: redis

postgres-benjdewan-02:
  cacert: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS...
  connections:
  - host: pachelbel-dry-run.compose.direct
    password: h3lm3t5t4ck5
    path: /compose
    port: 1774
    scheme: postgres
    username: bjblazkowicz
  type: postgresql
```

NOTES:
* In this case I used a `--dry-run` on deployments that do not exist, so most of the returned information is bogus for testing purpose. If the deployments existed the information in `cluster-info.yml` would have been accurate and usable even though this was a dry-run
* I truncated the certificates for clarity. In actual use if a certificate bundle exists for a deployment you'll get the whole thing.
