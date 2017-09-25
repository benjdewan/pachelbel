# Examples

The Pachelbel input format is decently complicated, so it may be useful to
demonstrate some of its functionality using examples. These examples are broken
into 2 categories, those can be run using a 30-day free trial account from
Compose, and those that require access to an Enterprise account.

NOTE: All of the command-line snippets assume that
*   pachelbel is in your PATH as `pachelbel`
*   You have a valid compose API key stored in your terminal env as
    COMPOSE_API_KEY
*   You are running these examples from this examples directory

## Examples that can be run with a free trial account
You can sign up for a free 30-day trial with Compose.io
[here](https://www.compose.com/signup).
After which all of the following examples should work.

NOTE: This free trial _does_ require providing credit card information, and if you
do not remember to cancel the account you will be automatically charged after the
thirty days expire for any deployments you still have provisioned. Do not blame me
if this happens.

### 1. Provision a postgres database

*   [input.yml](free/01-postgres/input.yml)

Running
```console
$ pachelbel provision ./free/01-postgres/input.yml
```

will create a postgres database in the AWS US-East-1 datacenter.

You can deprovision this instance with this command:
```console
$ pachelbel deprovision pachelbel-postgres-01
```

### 2. Provision a postgress database with SSL enabled

*   [input.yml](free/02-postgres-ssl/input.yml)

Running
```console
$ pachelbel provision ./free/02-postgres-ssl/input.yml
```

will create a postgres databse in the AWS US-East-1 datacenter with SSL enabled.
Note that in the output `connection-info.yml` file there is a base64 encoded CA
Certificate you will need to use to talk to your database alongside the connection
information you saw in [example #1](#1. Provision a postgres database)

You can deprovision this deployment with:
```console
$ pachelbel deprovision pachelbel-postgres-02
```

### 3. Provisioning a larger databse

*   [input.yml](free/03-postgres-bigger/input.yml)

The (Compose API)[https://apidocs.compose.com/v1.0/reference#2016-07-post-deployments]
offers a single knob for changing the size of a deployment which is
referred to as the deployment's `scaling`.

Running
```console
$ pachelbel provision ./free/03-postgres-bigger/input.yml
```

will provision a postgres database of scaling `5` which maps to 5GB of storage.
The default scaling for a new dpeloyment is `1`, and for postgres that translates to
1GB of storage

You can deprovision this deployment with:
```console
$ pachelbel deprovision pachelbel-postgres-03
```

### 4. Provisioning multiple deployments at once

*   [input.yml](free/04-multiple-deployments/input.yml)

A single input configuration file can define multiple deployment objects that
pachelbel will provision in parallel.

Running
```console
$ pachelbel provision ./free/04-multiple-deployments
```

will deploy two postgres databases, a redis instance as well as rabbitmq. Note that
in the configuration the datacenter in which a deployment is created is tied just
to that deployment, not the `provision` run as a whole, so that a single `provision`
call can not only create multiple deployments, but do so across datacenters, with
different security settings and different sizes

Also note, for Redis, there is a special configuration field, `cache_mode` that
only applies to redis deployments and cannot be set for any other deployment type
(this is also true for `wired_tiger` and MongoDB deployments).


You can deprovision these deployments with:
```console
$ pachelbel deprovision pachelbel-postgres-04 pachelbel-postgres-ssl pachelbel-redis pachelbel-rabbitmq
````

### 5. Connecting to an exisiting deployment

*   [input.yml](free/05-exisiting-deployment/input.yml)

The pachelbel input format is declarative, and the provision command is idempotent.
This means if you define a deployment that does not exist pachelbel will create
it, but if you define a deployment that already exists pachelbel will ensure it
fits the declaration (size, version, &c.) in your input, and return connection
information to that existing deployment when it has finished. If there is no
modification to be done (e.g. you re-run a `provision` with the same input file)
pachelbel will simply return the exact same output.

Running
```console
$ pachelbel provision ./free/05-existing-deployment/input.yml
```

Will create a new postgres deployment. If you then run:
```console
$ pachelbel provision --output re-run.yml ./free/05-existing-deployment/input.yml
  # Omitting pachelbel output for brevity
$ diff ./connection-info.yml ./re-run.yml
```

you will see that the two outputs are the same. Pachelbel does not have different
syntaxes for creating new deployment.

You can deprovision this deployment with:
```console
$ pachelbel deprovision pachelbel-postgres-05
```

### 6. Resizing a deployment

*   [input.yml](free/06-resizing-deployments/input.yml)
*   [upgrade.yml](free/06-resizing-deployments/upgrade.yml)

If a deployment already exists, and your input to pachelbel's provision call
includes scaling or version fields pachelbel will attempt to ensure the size
and/or version of the deployment matches your input.

This is not always possible (for example, the Compose API does not allow you
to upgrade from Elastic Search v2 to v5), and pachelbel will error out in this case.

NOTE: Version upgrades can be _extremely_ slow, so you may want to use the `timeout`
field to increase the amount of time pachelbel waits for the operation to finish.

Running
```console
$ pachelbel provision ./free/06-resizing-deployments/input.yml
```

will create a new postgres database with the default scaling of 1. Now running
``console
$ pachelbel provision ./free/06-resizing-deployments/upgrade.yml
```

will resize that deployment to scaling 4 (for postgres this means supporting 4gb
of storage).

You can deprovision this deployment using:
```console
$ pachelbel deprovision pachelbel-postgres-06
```

### 7. Deployment uniqueness

*   [input.yml](free/07-deployment-uniqueness/input.yml)
*   [input-advanced.yml](free/07-deployment-uniqueness/input-advanced.yml)

When specifying multiple deployment objects the names must be unique. Deployment
names in Compose must be unique, and Pachelbel lacks the ability to merge
configuration objects on the `name` field, so specifying the same deployment
multiple times will cause pachelbel to fail.

You can see this error:
```
$ pachelbel provision ./free/07-deployment-uniqueness/input.yml
```

NOTE: This restriction also applies to `deployment_client` objects:
```
$ pachelbel provision ./free/07-deployment-uniqueness/input-advanced.yml
```

### 8. Deployment Clients

*   [owner.yml](free/08-deployment-clients/owner.yml)
*   [client.yml](free/08-deployment-clients/client.yml)

Some Compose deployments will be shared across development teams where one team
owns/administrates it while another subscribes to it (for example, using a
redis queue to push jobs from a front-end API to back-end worker services). In
this case the downstream/consumers of the queue should not have the ability to
modify or control the deployment even though they do need access credentials.

The `deployment_client` object handles this use case: if a deployment exists
the connection information is returned, and if not an error is thrown because
the client cannot create or modify a deployment.

So if you run:
```console
$ pachelbel provision ./free/08-deployment-clients/client.yml
```

an error will be thrown, but if you run:
```console
$ pachelbel provision ./free/08-deployment-clients/owner.yml
$ pachelbel provision --output client.yml ./free/08-deployment-clients/client.yml
```

You will see that the output from both runs is the same.

You can deprovision this deployment by running:
```console
$ pachelbel deprovision pachelbel-postgres-08
```

NOTE: The name uniqueness requirement explained in
[example 7](#7. Deployment uniqueness) applies across deployment and
deployment_client objects; pachelbel will not attempt to resolve whether or not
it should treat the deployment as read-only or if it had the authority to
create/munge it and instead errors out. You can easily see this:
```console
$ pachelbel provision ./free/08-deployment-clients/owner.yml ./free/08-deployment-clients/client.yml
```

### 9. Dry run

*   [input.yml](free/09-dry-run/input.yml)

Pachelbel has a `--dry-run` flag for both provisioning and deprovisioning that can
be used for testing without making any changes to your Compose account. It is
important to note that a dry-run is **not** a complete simulation. API requests
are still made to Compose, so an API Key is still required. In general the behavior
is the following:

For deployments:
*   If the deployment exists a dry-run will return the real connection strings for
    that object, but will ignore any resizing or version changes it would perform
    during a regular run
*   If the deployment does not exist fake connection information will be returned.

For deployment_clients:
*   If the deployment exists a dry-run will return the real connection strings
    for that object.
*   If the deployment does not exist fake conneciton information will be returned.

To see an example of this you can run
```console
$ pachelbel provision --dry-run ./free/09-dry-run/input.yml
```

If you then log-in to your account in a web browser you will see no deployments
were actually created, but if you are using pachelbel as part of a larger
deployment pipeline you still generate an output suitable for testing downstream
functionality without having to worry about incurring charged from Compose.

## Examples that must be run with an enterprise account

I am not aware of a free trial of Compose' enterprise account functionality,
and because of the flexibility of Compose enterprise clusters these examples
are not guaranteed to work, but they should provide decent models for
developing against your own enterprise account (if you have access to one).

### 1. Deploying to clusters

*   [input.yml](enterprise/01-clusters/input.yml)

In enterprise accounts Compose offers a datacenter abstraction layer,
clusters, in which account administrators can define private networking,
hardware and more. It's expected that deployments will be into these
clusters, and not directly to datacenters (as one does with a free/standard
plan). The syntax for Pachelbel, however, is almost identical.

### 2. VPNs and Routing

*   [input.yml](enterprise/02-vpns-and-routing/input.yml)

By default Compose deployments to datacenters return publicly addressable
endpoint URLs (that is, the database endpoints are on the public internet).
A Compose enterprise customer can set-up their cluster to put deployments inside
a VPN, and have all endpoint URLs for those deployments only work when inside
the VPN even though the Compose API endpoint itself is publicly available.
Alternatively a cluster can be configured so that deployments have endpoints both
within a VPN and externally. The problem with this is that the Compose API in this
case will not return both sets of connection strings, just the one chosen by the
account administrators.

If, for example, querying the Compose API returns the public-facing endpoint URLs,
but you want to use the VPN endpoints, Pachelbel supports user-defined endpoint
mapping to perform that translation for you during the `provision` call.

NOTE: This feature certainly works with free accounts, but the functionality
is for an obscure enough use case that doing so is almost never requried
