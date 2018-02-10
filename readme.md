# shelastic

Elastic search shell aims to support all ElasticSearch versions from 1.7 to 6.x.

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

| Command                           | Description                                                                                    |
|:---------------------------------:|:-----------------------------------------------------------------------------------------------|
| `index clear-cache [<index-name>]`| Clears cache of given index. If no `<index-name>` is specified then cache for all indices is cleared|
| `index flush [<index-name>]` | Flushes index. If no `<index-name>` given, flushesh all indices. Supported options: `force` - forces flush even if it is not needed; `wait` - waits for other ongoing flush operation to complete |
| `index view mappings <index-name> [doc-name] [property-name]` | View mappings for index `<index-name>`. Optionally can display mappings only for specified document and/or property. Mappings are printed in YAML format for better readability|
| `index view settings <index-name>` | View index settings|
| `index view shards <index-name> [by-node|by-shard]` | View index shards|
| `index configure <yaml-config>` | Set index setting. See below for syntax |

## Supported operations

- Basic operations:
    - [x] List index name
    - [x] List basic node information - name, ip, hostname
- Index administration:
    - [ ] Clear cache
    - [x] Flush
    - [ ] Optimize
    - [ ] Refresh
- Index metadata operations:
    - [x] View settings
    - [ ] Change settings
    - [x] View mappings
    - [x] View routing - as part of settings
    - [x] View shards allocation
    - [ ] Change routing
    - [ ] View statistics
- Index operations:
    - [ ] Insert/Update document
    - [ ] Delete document
    - [ ] View document by id
- Query operations:
    - [ ] JSON requests
    - [ ] SQL like?
- Node operations:
    - [ ] JVM stats
        - [ ] JVM name and version
        - [ ] JVM arguments
    - [ ] OS stats
        - [ ] CPU, memory
- Cluster operations
    - Routing:
        - [ ] Decomission a node

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
    
