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

### Index commads

| Command                           | Description                                                                                    |
|:---------------------------------:|:-----------------------------------------------------------------------------------------------|
| `index clear-cache [<index-name>]`| Clears cache of given index. If no `<index-name>` is specified then cache for all indices is cleared|
| `index flush [<index-name>]` | Flushes index. If no `<index-name>` given, flushesh all indices. Supported options: `force` - forces flush even if it is not needed; `wait` - waits for other ongoing flush operation to complete |
| `index view mappings <index-name> [doc-name] [property-name]` | View mappings for index `<index-name>`. Optionally can display mappings only for specified document and/or property|
| `index view settings <index-name>` | View index settings|

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
  - [ ] View settings
  - [ ] Change settings
  - [x] View mappings
  - [ ] View routing
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
