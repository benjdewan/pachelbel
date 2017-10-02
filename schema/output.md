# Output schema

Pachelbel will write deployment connection information to disk if it
successfully provisions all the deployments specified in the input configuration
file(s). The default location for this output is `./connection-info.yml`, but
that can be overridden using the `--output` flag to specify a different location
on disk.

## Format
The output schema is a single yaml map of deployment name to deployment connection information with this format:
```yaml
<deployment_name>:
  cacert: {{.base64_encoded_ca_bundle}}
  connections:
  - host:     {{.host}}
    password: {{.password}}
    path:     {{.path}}
    port:     {{.port}}
    scheme:   http|https|mongodb|mysql|amqp|amqp|postgres
    username: {{.username}}
  type: elasticsearch|etcd|disque|janusgraph|mongodb|mysql|postgresql|rabbitmq|redis|rethink|scylla
  version: {{.version}}
```

## Example

Here is an example run of pachelbel that shows the input schema for 2 deployments,
the output of pachelbel running and then the resulting output file:
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
  version: 4.0.2

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
  version: 9.6.5
```

