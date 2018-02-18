# shelastic

Shelastic is an interactive shell for Elastic search which provides commands for most common administration tasks and aims to support all ElasticSearch versions from 1.7 to 6.x.

This project is not affiliated with Elastic in any way.

## Commands

### General commands

| Command                           | Description                                                                                    |
|:---------------------------------:|:-----------------------------------------------------------------------------------------------|
| `connect [host]`                  | Connects to ES cluster. If host name is omitted, tries to connect to localhost                 |
| `disconnect`                      | Disconnects from ES cluster. You need to disconnect before connecting to another cluster       |
| `list indices`                    | Lists indices on the cluster |
| `list nodes`                      | Lists nodes of the cluster. Each node is displayed in format `<name> @ <hostname> [<ip-address>]` |
| `_debug`                          | Toggle HTTP output. Use for bug reporting purposes |

### Index commads

| Command                            | Description                                                                                    |
|:-----------------------------------|:-----------------------------------------------------------------------------------------------|
| `index clear-cache [<index-name>]` | Clears cache of given index. If no `<index-name>` is specified then cache for all indices is cleared|
| `index flush [<index-name>]`        | Flushes index. If no `<index-name>` given, flushesh all indices. Supported options: `force` - forces flush even if it is not needed; `wait` - waits for other ongoing flush operation to complete |
| `index clear-cache [<index-name>]` | Clear index cache. If `<index-name>` is omitted, cache is cleared for all indices |
| `index refresh [<index-name>]` | Refreshes index, making all operations performed since last refresh available for search.|
| `index force-merge [<index-name>]` | Forces merging of one or more indices through an API. For ES version 1.x and 2.x this calls _Optimize_ API |
| `index view mappings <index-name> [doc-name] [property-name]` | View mappings for index `<index-name>`. Optionally can display mappings only for specified document and/or property. Mappings are printed in YAML format for better readability|
| `index view settings <index-name>` | View index settings|
| `index view shards <index-name> [by-node | by-shard]` | View index shards|
| `index configure <index-name> <config-item>`        | Set index setting. See below for syntax |
| `index restrict <index-name> selector [<target>]`    | moves all shards to given node by selector. Selector can be one of `name`, `host` or `ip`. If `<target>` is not specified, then restriction is removed |


### Snapshot commands

| Command                           | Description                                                                                    |
|:----------------------------------|:-----------------------------------------------------------------------------------------------|
| `snapshot repo list`              | Lists all configured snapshot repositories with their settings|
| `snapshot repo register <name> <type> <settings>` | Registers new repository of type `<type>` and named `<name>`. Repository settings can be passed as key-value pairs on command line. Each key and value must be separated by space|
| `snapshot repo verify <name>`     | Verifies repository |
| `snapshot create <repo> <name>`   | Creates snapshot named `<name>` in repository `<repo>` |
| `snapshot info <repo> [<name>]`   | Retrieves snapshot information from repository `<repo>`. If snapshot `<name>` is specified then its information is retrieved, otherwise information for all snapshots in the repository is printed|
| `snapshot restore <repo> <name>`  | Restores snapshot named `<name>` from repository `<repo>` |
| `snapshot delete <repo> <name>`   | Deletes snapshot named `<name>` from repository `<repo>` |

### Node commands

| Command                           | Description                                                                                    |
|:----------------------------------|:-----------------------------------------------------------------------------------------------|
| `node stats [node-name]`          | Displays node statistics. If `node-name` is specified then only this node stats are displayed otherwise statistics for all nodes is retrieved |
| `node environment [node-name]`    | Displays node environment: OS name version and JVM name and version|
| `node shards [<node-name>]`         | Displays indices and shards located on node. If node name is not specified, information is printed for all nodes|
| `node decomission <node-name>`    | Disables allocation for given node|

### Document commands

All document commands accept index name as first parameter. By using 'use index-name' command one can "open" and it will be implicitly used
in all document commands.

Even when an index is in use, explicit index name may be supplied to any document command. Explicitly specified index will take precedence.

| Command                           | Description                                                                                    |
|:----------------------------------|:-----------------------------------------------------------------------------------------------|
| `use <index-name>`                | Select index to be used for document manipulations. If no `<index-name>` is specified it will display index that is currently in use |
| `document list [index]`                   | Lists all documents in index |
| `document properties [index] <doc-name>`  | Lists properties of `<doc-name>` document. This does not display full metadata, just properies names
and types |
| `document get [index] <doc-name> <id>`    | Retrieves document by id |
| `document delete [index] <doc-name> <id>` | Deletes document by id| 
| `document search [index] [<doc-names>] <query>` | Search for query in `<doc-names>`. Document name can be omitted.|
| `document put [index] <doc-name> id`      | Upserts document into `index.doc-name` with id == id. This command will start multi-line editor to enter JSON
of the document. Complete document with ";" |

## Supported operations

- Basic operations:
    - [x] List index name
    - [x] List basic node information - name, ip, hostname
- Index administration:
    - [x] Clear cache
    - [x] Flush
    - [x] Refresh
    - [x] Force merge/Optimize
- Index metadata operations:
    - [x] View settings
    - [x] Change settings
    - [x] View mappings
    - [x] View routing - as part of settings
    - [x] View shards allocation
    - [ ] Change routing
    - [x] View shards
- Document operations:
    - [x] List documents in index
    - [x] List properties for document
    - [x] Insert/Update document
    - [x] Delete document
    - [x] View document by id
- Snapshots
    - [x] Create repository
    - [x] List repositories
    - [x] Create snapshot
    - [x] View snapshot information
    - [x] Delete snapshot
    - [x] Restore snapshot
- Search operations:
    - [x] Simple URL search
    - [ ] JSON requests
- Node operations:
    - [x] Node stats
        - [x] OS and JVM name and version
        - [x] Memory
        - [x] File system
    - [ ] View indices and shards allocated to this node
- Cluster operations
    - Routing:
        - [x] Ensure node has all shards
        - [ ] Decomission a node (ensure node has no shards)

## Changing index settings

To change index setting one can use yaml syntax.

Let's take a look at changing  allocation routing for the index.

Using REST APIs it can be done with following request:

        PUT http://localhost:9200/index-name/_settings

        {
        "settings": {
                "index": {
                    "routing": {
                        "allocation" : {
                            "require._name": "host1"
                        }
                    }
                }
            }
        }

To change the same setting using shelastic there are several syntax options. They are all implemented
using `index configure` command.

1. Interactive input. When no configuration key is specified on command line then `index configure` command will switch to multiline editor. Enter index configuration line by line, finish with semicolon. Each line of configuration consists of configuration key and value separated by colon

        > index configure index-name
        Enter configuration parameters, one per line, finish with ;
        index.routing.allocation.require._name: "host1"
        index.routing.allocation.require._ip:  "host2";

2. Enter configuration on a command line. Everything after index name will be interpreted as configuration in YAML syntax.

        > index configure index-name index.routing.allocation.require._name (host1)

   _Warning_: May change in future version.

    As commands are interpreted using shell rules, quotes and double quotes will be used to enclose multi-word parameters. Use parenthesis to pass string parameters. Parenthesis will be replaced with quotes in REST call.

### Specific use cases

Case: *Disable all hosts but one*

Steps to execute

- change number of replicas for index to 0

- Move all shards to one node

        PUT /<index>/_settings
        {
        "settings": {
            "index.routing.allocation.require._name": "enabled_node",
            }
        }