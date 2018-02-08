## shelastic 

Elastic search shell

### Commands

| Command                           | Description                                                                                    |
|:---------------------------------:|:-----------------------------------------------------------------------------------------------|
| `connect [host]`                  | Connects to ES cluster. If host name is omitted, tries to connect to localhost                 |
| `disconnect`                      | Disconnects from ES cluster. You need to disconnect before connecting to another cluster       |
| `list indices`                    | Lists indices on the cluster |
| `list nodes`                      | Lists nodes of the cluster. Each node is displayed in format `<name> @ <hostname> [<ip-address>]` |
| `index view mappings <index-name> [doc-name] [property-name]` | View mappings for index `<index-name>`. Optionally can display mappings only for specified document and/or property|


### Supported operations:

  - Basic operations:
    - [x] List index name
    - [x] List basic node information - name, ip, hostname
  - Index metadata operations:
    - [ ] View settings
    - [ ] Change settings
    - [ ] View mappings
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
   

  