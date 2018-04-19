package cmd

import (
	"shelastic/es"
	"shelastic/utils"

	ishell "gopkg.in/abiosoft/ishell.v2"
)

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

func viewIndexSegments(c *ishell.Context) {
	if context == nil {
		errorMsg(c, errNotConnected)
	}

	selector, err := parseDocumentArgs(c.Args)
	if err != nil {
		errorMsg(c, err.Error())
		return
	}

	if selector.Index == "" {
		errorMsg(c, errIndexNotSelected)
		return
	}

	result, err := context.IndexShards(selector.Index)
	if err != nil {
		errorMsg(c, err.Error())
	} else {
		yamlstr, err := utils.MapToYaml(result)
		if err != nil {
			errorMsg(c, err.Error())
		}
		cprintln(c, yamlstr)
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
			if shard.Shard.Routing.Primary {
				prim = "Primary"
			} else {
				prim = "Replica"
			}
			cprintln(c, "   %d: %s, %s", shard.ID, shard.Shard.Routing.State, prim)
		}
	}
}

func printIndexShardsByShard(c *ishell.Context, indexShards es.IndexShards) {
	for _, indexShard := range indexShards {
		cprintln(c, "Shard %d:", indexShard.ID)
		for shardIdx, shardInfo := range indexShard.Shards {

			var prim string
			if shardInfo.Routing.Primary {
				prim = "Primary"
			} else {
				prim = "Replica"
			}

			cprintln(c, "   %d: %s - %s - %s", shardIdx, prim, shardInfo.Routing.State, shardInfo.Node.String())
		}
	}
}
