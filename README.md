# Pachelbel
[![Build Status](https://travis-ci.org/benjdewan/pachelbel.svg?branch=master)](https://travis-ci.org/benjdewan/pachelbel)
[![Go Report Card](https://goreportcard.com/badge/github.com/benjdewan/pachelbel)](https://goreportcard.com/report/github.com/benjdewan/pachelbel)
[![codebeat badge](https://codebeat.co/badges/12ff5e7c-e6f5-4791-a706-2ee1a8fe7653)](https://codebeat.co/projects/github-com-benjdewan-pachelbel-master)

pachelbel is designed to be an idempotent provisioning tool for [IBM Compose](compose.io) deployments. It's still under heavy development

## Usage
pachelbel is built atop cobra and has extensive information available using `--help`.

## Configuration schema
pachelbel is designed to read yaml configuration files. The YAML objects read in must adhere to the following schema:

```yaml
# The config is versioned. Objects of different config_versions can be used in
# a single pachelbel provision run.
config_version: 1

# Compose offers many database types.
type: mongodb|rethinkdb|postgresql|redis|rabbitmq|etcd|elastic_search|mysql|janusgraph

# By default the latest stable version of the specified databse type is used.
# Use the 'version' field to specify an older version, or if you want to update
# an existing deployment to a newer version
version: 3.2.10

# pachelbel supports deployments to datacenters *or* clusters *or* tags. You
# cannot specify more than one of these fields for any single deployment.
#
# pachelbel will attempt to map the cluster-name to an ID. If that cluster
# does not exist or is not visible when using the provided API token, pachelbel
# will throw an error. There is currently no support for the creation/deleting
# of clusters.
cluster: my-softlayer-cluster

# pachelbel supports deployments to datacenters *or* clusters *or* tags. You
# cannot specify more than one of these fields for any single deployment.
#
# the value provided here should be a datacenter slug as returned by the compose
# API
datacenter: aws:us-east-1

# pachelbel supports deployments to datacenters *or* clusters *or* tags. You
# cannot specify more than one of these fields for any single deployment.
#
# pachelbel does not currently validate that the provided tags exist, but
# provisioning a deployment will fail if they do not.
tags:
  - dev
  - benjdewan

# The name of the deployment must be <64 characters, but is otherwise very
# flexible.
name: postgres-benjdewan-01

# notes can include additional metadata about this deployment
notes: |
    This is a test of pachelbel

# Optionally specify the scaling size of the deployment. The default is '1'
scaling: 2

# For databases that support ssl, use this line to ensure it is set.
ssl: true

# The timeout period, in seconds, to wait for provisioning recipes (creating
# new deployments, scaling existing deployments &c.) to complete. If no recipes
# are triggered, no waiting will occur.
#
# If this field is not set a default timeout of 300 seconds (5 minutes) is used.
timeout: 900

# If you want to make this deployment visible to anyone other than the user that
# created it, you should create a team via the web interface, grab the team ID, and
# then specify it here along with the roles that team should have:
teams:
  - id: "123456789"
    role: "admin"
  - id: "123456789"
    role: "developer"

# WiredTiger is a storage engine option for MongoDB. Setting this field for
# any other type of deployment will throw an error
wired_tiger: true

# Multiple YAML configuration objects can be combined into a single file separated
# using the standard yaml separator: `\n---\n`.
---
config_version: 2

# Objects of config_version == 2 have a new top-level field, object type. There is
# currently only one v2 object, an endpoint_map. These objects define conversions
# to be run against the deployment endpoints returned by Compose.io.
#
# endpoint_map objects are very basic structures, mapping one URL hostname to
# another.
object_type: endpoint_map
  cluster-benjdewan-01.compose.direct: haproxy-v1.legacy.com
  haproxy-v1.legacy.com: endpoint.example.net

# endpoint_maps are interpolated recursively, so given these mappings if my
# postgres deployment, 'postgres-benjdewan-01' returns the connection string:
#    postgres://admin:PASSWORD@cluster-benjdewan-01.compose.direct:45678/compose
#
# the output of pachelbel will be
#    name: postgres-benjdewan-01
#    type: postgresql
#    connections:
#      - scheme: postgres
#        host: endpoint.example.net
#        username: admin
#        password: PASSWORD
#        port: 45678
#        path: /compose
```

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
