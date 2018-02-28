package cmd

import (
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
		Help: "View index mapping. Usage: index mapping [--index <index-name>] [--doc <doc> [<property>]]",
		Func: viewIndexMapping,
	})

	view.AddCmd(&ishell.Cmd{
		Name: "settings",
		Help: "View index settings. Usage: view settings [--index <index-name>]",
		Func: viewIndexSettings,
	})

	view.AddCmd(&ishell.Cmd{
		Name: "shards",
		Help: "View index shards. Usage: view shards [--index <index-name>] [--mode by-node|by-shard]",
		Func: viewIndexShards,
	})

	index.AddCmd(view)
	index.AddCmd(&ishell.Cmd{
		Name: "flush",
		Help: "Flushes index data to storage and clears transaction log. Usage: flush [--index <index-name>]",
		Func: flush,
	})

	index.AddCmd(view)
	index.AddCmd(&ishell.Cmd{
		Name: "clear-cache",
		Help: "Clears index cache. Usage: clear-cache [--index <index-name>]",
		Func: clearCache,
	})

	index.AddCmd(view)
	index.AddCmd(&ishell.Cmd{
		Name: "refresh",
		Help: "Refreshes index, making all operations performed since last refresh available for search. Usage: refresh [--index <index-name>]",
		Func: refresh,
	})

	index.AddCmd(view)
	index.AddCmd(&ishell.Cmd{
		Name: "force-merge",
		Help: "Allows to force merging of one or more indices through an API. Usage: force-merge [--index <index-name>]",
		Func: forceMerge,
	})

	index.AddCmd(&ishell.Cmd{
		Name: "configure",
		Help: "Set index' setting",
		Func: configureIndex,
	})

	index.AddCmd(&ishell.Cmd{
		Name: "restrict",
		Help: "Move index shards to one node. " + restrictUsage,
		Func: restrictIndex,
	})

	return index
}

func flush(c *ishell.Context) {
	if context != nil {
		type flushArgs struct {
			documentSelectorData
			Wait  bool `long:"wait" description:"Wait for flush to complete"`
			Force bool `long:"force" description:"Force flush even if it is not required"`
		}

		slct, err := parseDocumentArgsCustom(c.Args, &flushArgs{})
		if err != nil {
			errorMsg(c, err.Error())
			return
		}
		selector := slct.(*flushArgs)
		if selector.Index == "" {
			cprintln(c, "Flusing all indices")
		} else {
			cprintlist(c, "Flushing ", cy(selector.Index))
		}
		err = context.Flush(selector.Index, selector.Force, selector.Wait)
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
		selector, err := parseDocumentArgs(c.Args)
		if err != nil {
			errorMsg(c, err.Error())
			return
		}
		if selector.Index == "" {
			cprintln(c, "Clearing cache for all indices")
		} else {
			cprintlist(c, "Clearing cache for ", cy(selector.Index))
		}
		err = context.ClearCache(selector.Index)
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
		selector, err := parseDocumentArgs(c.Args)
		if err != nil {
			errorMsg(c, err.Error())
			return
		}
		if selector.Index == "" {
			cprintln(c, "Refreshing all indices")
		} else {
			cprintlist(c, "Refreshing ", cy(selector.Index))
		}
		err = context.Refresh(selector.Index)
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
		selector, err := parseDocumentArgs(c.Args)
		if err != nil {
			errorMsg(c, err.Error())
			return
		}
		if selector.Index == "" {
			cprintln(c, "Merging all indices")
		} else {
			cprintlist(c, "Merging ", cy(selector.Index))
		}
		err = context.ForceMerge(selector.Index)
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

		selector, err := parseDocumentArgs(c.Args)
		if err != nil {
			errorMsg(c, err.Error())
			return
		}
		if selector.Index == "" {
			errorMsg(c, errIndexNotSelected)
			return
		}
		var property string
		if len(selector.Args) > 0 {
			property = selector.Args[0]
		}

		result, err := context.IndexViewMapping(selector.Index, selector.Document, property)
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

		index := selectIndex(c)
		if index == "" {
			return
		}

		result, err := context.IndexViewSettings(index)
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

		type shardsArgs struct {
			documentSelectorData
			Mode string `long:"mode" choice:"by-shard" choice:"by-node" default:"by-node" description:"Display result grouped either by node or by shard"`
		}

		sel, err := parseDocumentArgsCustom(c.Args, &shardsArgs{})
		if err != nil {
			errorMsg(c, err.Error())
			return
		}
		selector := sel.(*shardsArgs)

		if selector.Index == "" {
			errorMsg(c, errIndexNotSelected)
			return
		}

		result, err := context.IndexShards(selector.Index)
		if err != nil {
			errorMsg(c, err.Error())
		} else {
			if selector.Mode == "by-node" {
				printIndexShardsByNode(c, result)
			} else {
				printIndexShardsByShard(c, result)
			}
		}

	} else {
		errorMsg(c, errNotConnected)
	}
}

func selectIndex(c *ishell.Context) string {
	selector, err := parseDocumentArgs(c.Args)
	if err != nil {
		errorMsg(c, err.Error())
		return ""
	}
	var index string
	if selector.Index == "" && len(selector.Args) > 0 {
		index = selector.Args[0]
	} else {
		index = selector.Index
	}

	if index == "" {
		errorMsg(c, errIndexNotSelected)
		return ""
	}
	return index
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

		selector, err := parseDocumentArgs(c.Args)
		if err != nil {
			errorMsg(c, err.Error())
			return
		}
		if selector.Index == "" {
			errorMsg(c, errIndexNotSelected)
			return
		}

		var payload map[string]string

		payload = make(map[string]string)
		cprintlist(c, "Enter configuration parameters, one per line. Finish with ", cy(";"))
		c.SetPrompt(">>> ")
		lines := strings.Split(c.ReadMultiLines(";"), "\n")
		for _, ln := range lines {
			ln = strings.TrimSpace(ln)
			kv := strings.Split(ln, ":")
			if len(kv) == 2 {
				payload[strings.TrimSpace(kv[0])] = kv[1]
			}
		}
		restorePrompt(context, c)

		err = context.IndexConfigure(selector.Index, payload)
		if err != nil {
			errorMsg(c, err.Error())
		} else {
			cprintln(c, "Ok")
		}

	} else {
		errorMsg(c, errNotConnected)
	}
}

const restrictUsage = "Usage: restrict [--index <index-name>] name|ip|host <target>"

func restrictIndex(c *ishell.Context) {
	if context == nil {
		errorMsg(c, errIndexNotSelected)
		return
	}
	selector, err := parseDocumentArgs(c.Args)

	if len(selector.Args) < 1 {
		errorMsg(c, "Not enough parameters."+restrictUsage)
		return
	}
	mode := selector.Args[0]
	if mode != "name" && mode != "ip" && mode != "host" {
		errorMsg(c, "Restriction can be done by node name, host name or by ip address."+restrictUsage)
		return
	}
	var route string
	if len(selector.Args) == 2 {
		route = selector.Args[1]
	} else {
		route = ""
	}

	err = context.MoveAllShardsToNode(selector.Index, "_"+mode, route)

	if err != nil {
		errorMsg(c, err.Error())
	} else {
		if route == "" {
			cprintln(c, "Restrictions removed")
		} else {
			cprintln(c, "Ok")
		}
	}
}
