# Pachelbel
[![Build Status](https://travis-ci.org/benjdewan/pachelbel.svg?branch=master)](https://travis-ci.org/benjdewan/pachelbel)

pachelbel is designed to be an idempotent provisioning tool for [IBM Compose](compose.io) deployments. It's still under heavy development

## Usage
pachelbel is built atop cobra and has extensive information available using `--help`.

## Configuration schema
pachelbel is designed to read yaml configuration files. The YAML objects read in must adhere to the following schema:

```yaml
# The config is versioned, but only 1 version currently exists.
config_version: 1

# Compose offers many database types. Pachelbel only supports a subset
# currently since extracting credentials from connection strings is required
# and different for every database type
type: postgresql|redis|rabbitmq|etcd|elastic_search

# By default the latest stable version of the specified databse type is used.
# Use the 'version' field to specify an older version, or if you want to update
# an existing deployment to a newer version
version: 3.2.10

# pachelbel will attempt to map the cluster-name to an ID. If that cluster
# does not exist or is not visible when using the provided API token, pachelbel
# exits. There is currently no support for the creation/deleting of clusters.
cluster: my-softlayer-cluster

# The name of the deployment must be <64 characters, but is otherwise very
# flexible
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
```

Multiple YAML configuration objects can be combined into a single file (separated using a newline and the `---` string), or they can span multiple files passed in on the command line. Pachelbel can also read directories of configuration files, but does not do so recursively.
