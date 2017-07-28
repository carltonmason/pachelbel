# Pachelbel

pachelbel is designed to be an idempotent provisioning tool for [IBM Compose](compose.io) deployments. It's still under heavy development

## Usage
pachelbel is built atop cobra and has extensive information available using `--help`.

## Configuration schema
pachelbel is designed to read yaml configuration files. The YAML objects read in must adhere to the following schema:

```yaml
# The config is versioned, but only 1 version currently exists.
version: 1

# Compose offers many database types. Pachelbel only supports a subset
# currently since extracting credentials from connection strings is required
# and different for every database type
type: postgresql|redis|rabbitmq|etcd|elastic_search

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

# For databases that support ssl, use this line to ensure it is set.
ssl: true

# If you want to make this deployment visible to anyone other than the user that
# created it, you should create a team via the web interface, grab the team ID, and
# then specify it here along with the roles that team should have:
teams:
    - { id: "123456789", role: "admin" }
    - { id: "123456789", role: "developer" }
```

Multiple YAML configuration objects can be combined into a single file (separated using a newline and the `---` string), or they can span multiple files passed in on the command line. Pachelbel can also read directories of configuration files, but does not do so recursively.
