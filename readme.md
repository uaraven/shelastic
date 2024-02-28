# shelastic

**This project is abandoned!**

Version 0.3

Shelastic is an interactive shell for Elastic search which provides commands for most common administration tasks and aims to support all ElasticSearch versions from 1.7 to 6.x.

This project started as study in Go language and was not intented as seriuous administration tool. Despite that I found that it is sometimes useful.

This project is not affiliated with Elastic in any way.

## Commands

### General commands

    connect [host]

Connects to ES cluster. If host name is omitted, tries to connect to localhost

    disconnect

Disconnects from ES cluster. You need to disconnect before connecting to another cluster

    list indices

Lists indices in the cluster. Displays number of documents in index, size of index in bytes and index aliases

    list nodes

Lists nodes of the cluster. Each node is displayed in format `<name> @ <hostname> [<ip-address>]`

    use <index-name>

Select index to be used for document manipulations. If no `<index-name>` is specified it will display index that is currently in use. To stop using index pass `--` argument to `use`, as in `use --`

    _debug

Toggle debug output (mostly HTTP traces). Use for bug reporting purposes

### Index commads

All index commands can accept index name as argument to `--index` option. By using `use index-name` command one can "open" an index and it will be implicitly used in all document commands.

Even when an index is in use, explicit index name may be supplied to any document command. Index specified with `--index` option will take precedence.

Some of the commands (like `refresh` or `flush`) will be applied to all indices if no index name is specified. If an index is selected with `use` command it first have to be deselected with `use --` for such commands to be applied to all indices in the cluster.

For some commands you can use wildcards in index name, for example following command will set replica count to 0 for all indices:

        es-cluster $> index configure --index *
        Enter configuration parameters, one per line. Finish with ;
        >>> number_of_replicas: 0;
        Ok

    index clear-cache [--index <index-name>]
Clears cache of given index. If no index is in use then cache for all indices is cleared

    index flush  [--index <index-name>] [--force] [--wait]
Flushes index. If no index is in use then flushes all indices. Supported options: `--force` - forces flush even if it is not needed; `--wait` - waits for other ongoing flush operation to complete

    index refresh [--index <index-name>]
Refreshes index, making all operations performed since last refresh available for search. If no index is in use then all indices are refreshed

    index force-merge [--index <index-name>]
Forces merging of one or more indices through an API. For ES version 1.x and 2.x this calls _Optimize_ API. If no index is in use then all indices are forced to merge

    index view mappings [--index <index-name>] [--doc <doc-name>] [property-name]
View mappings for index `<index-name>`. Optionally can display mappings only for specified document and/or property. Mappings are printed in YAML format for better readability

    index view settings [--index <index-name>]
View index settings

    index view shards [--index <index-name>] [--mode by-node | by-shard]
View index shards. If `--mode` option is not specified, `by-shard` is used

    index restrict [--index <index-name> selector [<target>]
Moves all shards to given node by selector. Selector can be one of `name`, `host` or `ip`. If `<target>` is not specified, then restriction is removed

    index delete [--index] [<index-name>]

Deletes index.

    index truncate [--index] [<index-name>]

Deletes all the data in the index, leaving settings, aliases and mappings intact.

    index configure [--index <index-name>]
Sets index setting.

At prompt enter index configuration line by line. Each line of configuration consists of configuration key and value separated by colon. Semicolon indicates end of entry

        > index configure index-name
        Enter configuration parameters, one per line. Finish with ;
        >>> index.routing.allocation.require._name: "host1"
        >>> index.routing.allocation.require._ip:  "host2"
        >>>;


    index add-alias [--index <index-name>] <alias-name>
Creates a new alias `<alias-name>` for index `<index-name>`

    index delete-alias [--index <index-name>] <alias-name>
Delete alias `<alias-name`> from index `<index-name>`

    index close [--index <index-name>]
Closes index. Closed index is not available for read or write

    index open [--index <index-name>]
Opens previously closed index. 

    index copy [--index <index-name>] --target <target-index-name>
Copies mappings and documents from `<index-name>` to `<target-index-name>`. Target index should not exist. No index settings or
aliases are copied. For ES version 2.4 and above this will use `_reindex` API. For older Elasticsearch versions all the documents will
be copied using bulk APIs. There is no progress indication, so be patient. 

### Snapshot commands

    snapshot repo list
Lists all configured snapshot repositories with their settings

    snapshot repo register <name> <type> <settings>
Registers new repository of type `<type>` and named `<name>`. Repository settings can be passed as key-value pairs on command line. Each key and value must be separated by space

    snapshot repo verify <name>
Verifies repository

    snapshot create <repo> <name>
Creates snapshot named `<name>` in repository `<repo>`

    snapshot info <repo> [<name>]
Retrieves snapshot information from repository `<repo>`. If snapshot `<name>` is specified then its information is retrieved, otherwise information for all snapshots in the repository is printed

    snapshot restore <repo> <name>
Restores snapshot named `<name>` from repository `<repo>`

    snapshot delete <repo> <name>
Deletes snapshot named `<name>` from repository `<repo>`

### Node commands

    node stats [node-name]
Displays node statistics. If `node-name` is specified then only this node stats are displayed otherwise statistics for all nodes is retrieved

    node environment [node-name]
Displays node environment: OS name version and JVM name and version

    node shards [<node-name>]
Displays indices and shards located on node. If node name is not specified, information is printed for all nodes

    node decomission [--selector ip|host|node --clear|<list-of-selectors>]
Disables allocation for given node. Nodes can be disabled by ip address, host name or node name. Multiple nodes can be disabled by listing selectors separated with space. If `--clear` parameter is provided instead of list of selectors, then restrictions will be removed for selector chosen by `--selector`. If all parameters are omitted, then current restrictions will be printed. This command modifies cluster _transient_ settings, so its effects will last till cluster restart.


### Document commands

All document commands can accept index name as argument to `--index` option. By using 'use index-name' command one can "open" an index and it will be implicitly used in all document commands.

Even when an index is in use, explicit index name may be supplied to any document command. Index specified with `--index` option will take precedence.

    document list [--index <index-name>]
Lists all documents in index

    document properties [--index <index-name>] --doc <doc-name>
Lists properties of `<doc-name>` document. This does not display full metadata, just properies names and types

    document get [--index <index-name>]--doc  <doc-name> <id>
Retrieves document by id

    document delete [--index <index-name>] --doc <doc-name> <id>
Deletes document by id

    document search [--index <index-name>] [--doc <doc-names>] <query>
Search for query in `<doc-names>`. Document name can be omitted. Number of records returned by query is limited to 20.

    document query [--index <index-name>] [--doc <doc-name>]
Search using Query DSL. Query must be entered as JSON at the prompt. Empty query (single `;` character) will be interpreted as `{"query":{"match_all":{}}}`

Number of records returned by query is limited to 20. If more document is needed use `bulk export` command.

    document put [--index <index-name>] --doc <doc-name> id
Upserts document into `index.doc-name` with id == id. This command will start multi-line editor to enter JSON of the document. Complete document with ";". Number of documents returned with this query will be limited to 20. If you need more results use `export` command

### Bulk export/import commands

All bulk commands can accept index name as argument to `--index` option. By using 'use index-name' command one can "open" an index and it will be implicitly used in all document commands.

Even when an index is in use, explicit index name may be supplied to any document command. Index specified with `--index` option will take precedence.

    bulk export [--index <index-name>] [--doc <doc-type>] [--format ndjson|array] [--source] <filename>
Exports all records from a search into a file. Each line in file will contain JSON with one search result.

Query for search is entered as JSON at the prompt. Empty query (single `;` character) will be interpreted as `{"query":{"match_all":{}}}`. If `--source` parameter is specified only `_source` field of records will be exported.
If `--format ndjson` option is specified then data will be written in Elasticsearch NDJSON format, with action and metadata (see [ES bulk API](https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-bulk.html) for details If `--format array` is used, then file will contain JSON array with all the records. If `--format` option is omitted then `array` format is assumed by default.

    bulk import [--format ndjson|array]|[--index <index-name>] [--doc <doc-type>] [--id-field field] <filename>
Imports records from the file into Elasticsearch. Import supports two file formats, just like export.

If `--format ndjson` option is specified then file will be treated like Elasticsearch NDJSON file, with action and metadata (see [ES bulk API](https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-bulk.html) for details). `--index` and `--doc` options are ignored when used with `--format array`. If `--format` option is omitted then `array` format is assumed by default.

If `--ndjson` is not specified then shelastic expects the file to contain json array of recordsIndex and document names should be specified on command line and optional `--idfield <id-field-name>` parameter can be used to pick record id from its `<id-field-name>` field.


## Release history

### 0.3.1

Changelog:
  - Bulk import and export
  - Index data copying
  - Index alias control
  - Bugfixes
  
  Binaries can be downloaded [here](https://github.com/uaraven/shelastic/releases/tag/v0.3.1).

### 0.3

Latest version of shelastic is 0.3, binaries compiled for `amd64` architecture can be downloaded [here](https://github.com/uaraven/shelastic/releases/tag/v0.3). It includes pre-built binaries for Linux, MacOS and Windows. Note that Windows binary is untested. 
