package cmd

import (
	"fmt"
	"shelastic/es"
	"shelastic/utils"
	"strings"

	ishell "gopkg.in/abiosoft/ishell.v2"
)

// Index wraps all index functions
func Index() *ishell.Cmd {
	index := &ishell.Cmd{
		Name: "index",
		Help: "Index operations",
	}

	view := &ishell.Cmd{
		Name: "view",
		Help: "View index data",
	}

	view.AddCmd(&ishell.Cmd{
		Name: "mapping",
		Help: "View index mapping. Usage: index view mapping <index-name> [<doc> [<property>]]",
		Func: viewIndexMapping,
	})

	view.AddCmd(&ishell.Cmd{
		Name: "settings",
		Help: "View index settings. Usage: index view settings <index-name>",
		Func: viewIndexSettings,
	})

	view.AddCmd(&ishell.Cmd{
		Name: "shards",
		Help: "View index shards. Usage: index view shards <index-name>",
		Func: viewIndexShards,
	})

	index.AddCmd(view)
	index.AddCmd(&ishell.Cmd{
		Name: "flush",
		Help: "Flushes index data to storage and clears transaction log",
		Func: flush,
	})

	index.AddCmd(view)
	index.AddCmd(&ishell.Cmd{
		Name: "clear-cache",
		Help: "Clears index cache",
		Func: clearCache,
	})

	index.AddCmd(view)
	index.AddCmd(&ishell.Cmd{
		Name: "refresh",
		Help: "Refreshes index, making all operations performed since last refresh available for search",
		Func: refresh,
	})

	index.AddCmd(view)
	index.AddCmd(&ishell.Cmd{
		Name: "force-merge",
		Help: "Allows to force merging of one or more indices through an API.",
		Func: forceMerge,
	})

	index.AddCmd(&ishell.Cmd{
		Name: "configure",
		Help: "Set index' setting",
		Func: configureIndex,
	})

	return index
}

func flush(c *ishell.Context) {
	if context != nil {
		var index string
		if len(c.Args) == 0 {
			index = ""
		} else {
			index = c.Args[0]
		}
		err := context.Flush(index, false, false)
		if err != nil {
			errorMsg(c, err.Error())
		} else {
			cprintln(c, "Ok")
		}
	} else {
		errorMsg(c, errNotConnected)
	}
}

func clearCache(c *ishell.Context) {
	if context != nil {
		var index string
		if len(c.Args) == 0 {
			index = ""
		} else {
			index = c.Args[0]
		}
		err := context.ClearCache(index)
		if err != nil {
			errorMsg(c, err.Error())
		} else {
			cprintln(c, "Ok")
		}
	} else {
		errorMsg(c, errNotConnected)
	}
}

func refresh(c *ishell.Context) {
	if context != nil {
		var index string
		if len(c.Args) == 0 {
			index = ""
		} else {
			index = c.Args[0]
		}
		err := context.Refresh(index)
		if err != nil {
			errorMsg(c, err.Error())
		} else {
			cprintln(c, "Ok")
		}
	} else {
		errorMsg(c, errNotConnected)
	}
}

func forceMerge(c *ishell.Context) {
	if context != nil {
		var index string
		if len(c.Args) == 0 {
			index = ""
		} else {
			index = c.Args[0]
		}
		err := context.ForceMerge(index)
		if err != nil {
			errorMsg(c, err.Error())
		} else {
			cprintln(c, "Ok")
		}
	} else {
		errorMsg(c, errNotConnected)
	}
}

func viewIndexMapping(c *ishell.Context) {
	if context != nil {
		indexName := ""
		docType := ""
		property := ""

		if len(c.Args) >= 1 {
			indexName = c.Args[0]
		}
		if len(c.Args) >= 2 {
			docType = c.Args[1]
		}
		if len(c.Args) >= 3 {
			property = c.Args[2]
		}
		if indexName == "" {
			errorMsg(c, "Index name not specified")
			return
		}

		result, err := context.IndexViewMapping(indexName, docType, property)
		if err != nil {
			errorMsg(c, err.Error())
			return
		}
		mappings, err := utils.MapToYaml(result)
		if err != nil {
			errorMsg(c, err.Error())
		} else {
			cprintln(c, mappings)
		}

	} else {
		errorMsg(c, errNotConnected)
	}
}

func viewIndexSettings(c *ishell.Context) {
	if context != nil {

		if len(c.Args) < 1 {
			errorMsg(c, "Index name not specified")
			return
		}
		indexName := c.Args[0]

		result, err := context.IndexViewSettings(indexName)
		if err != nil {
			errorMsg(c, err.Error())
		} else {

			text, err := utils.MapToYaml(result)
			if err != nil {
				errorMsg(c, err.Error())
			} else {
				cprintln(c, text)
			}
		}

	} else {
		errorMsg(c, errNotConnected)
	}
}

func viewIndexShards(c *ishell.Context) {
	if context != nil {

		if len(c.Args) < 1 {
			errorMsg(c, "Index name not specified")
			return
		}
		indexName := c.Args[0]

		var mode string
		if len(c.Args) > 1 && strings.ToLower(c.Args[1]) == "by-shard" {
			mode = "shard"
		} else {
			mode = "node"
		}

		result, err := context.IndexShards(indexName)
		if err != nil {
			errorMsg(c, err.Error())
		} else {
			if mode == "node" {
				printIndexShardsByNode(c, result)
			} else {
				printIndexShardsByShard(c, result)
			}
		}

	} else {
		errorMsg(c, errNotConnected)
	}
}

func printIndexShardsByNode(c *ishell.Context, indexShards es.IndexShards) {

	type ShardByNode struct {
		ID    int
		Shard *es.ShardInfo
	}

	nodes := make(map[string][]*ShardByNode)
	for _, indexShard := range indexShards {
		for _, shardInfo := range indexShard.Shards {
			nodeID := shardInfo.Node.String()
			if shardList, ok := nodes[nodeID]; ok {
				shardList = append(shardList, &ShardByNode{ID: indexShard.ID, Shard: shardInfo})
				nodes[nodeID] = shardList
			} else {
				shardList := make([]*ShardByNode, 1)
				shardList[0] = &ShardByNode{ID: indexShard.ID, Shard: shardInfo}
				nodes[nodeID] = shardList
			}
		}
	}
	for nodeID := range nodes {
		cprintln(c, "%s:", nodeID)
		for _, shard := range nodes[nodeID] {
			var prim string
			if shard.Shard.Primary {
				prim = "Primary"
			} else {
				prim = "Replica"
			}
			cprintln(c, "   %d: %s, %s", shard.ID, shard.Shard.State, prim)
		}
	}
}

func printIndexShardsByShard(c *ishell.Context, indexShards es.IndexShards) {
	for _, indexShard := range indexShards {
		cprintln(c, "Shard %d:", indexShard.ID)
		for shardIdx, shardInfo := range indexShard.Shards {

			var prim string
			if shardInfo.Primary {
				prim = "Primary"
			} else {
				prim = "Replica"
			}

			cprintln(c, "   %d: %s - %s - %s", shardIdx, prim, shardInfo.State, shardInfo.Node.String())
		}
	}
}

func configureIndex(c *ishell.Context) {
	if context != nil {
		if len(c.Args) < 1 {
			errorMsg(c, "Index name not specified")
			return
		}
		indexName := c.Args[0]

		var payload map[string]string
		if len(c.Args) < 3 {
			payload = make(map[string]string)
			cprintln(c, "Enter configuration parameters, one per line finish with empty line")
			c.SetPrompt(">>> ")
			lines := strings.Split(c.ReadMultiLinesFunc(func(ln string) bool {
				return len(ln) != 0
			}), "\n")
			for _, ln := range lines {
				ln = strings.TrimSpace(ln)
				kv := strings.Split(ln, ":")
				fmt.Println("pair:", kv)
				if len(kv) == 2 {
					payload[strings.TrimSpace(kv[0])] = kv[1]
				}
			}
			restorePrompt(context, c)
		} else {
			v := c.Args[2]
			if v[0] == '(' && v[len(v)-1] == ')' {
				v = "\"" + v[1:len(v)-1] + "\""
			}
			payload = map[string]string{c.Args[1]: v}
		}
		err := context.IndexConfigure(indexName, payload)
		if err != nil {
			errorMsg(c, err.Error())
		} else {
			cprintln(c, "Ok")
		}

	} else {
		errorMsg(c, errNotConnected)
	}
}
