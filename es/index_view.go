package es

import (
	"fmt"
	"shelastic/utils"
	"sort"
	"strconv"
)

// ShardRouting contains shard routing information
type ShardRouting struct {
	State   string `json:"state" yaml:"State"`
	Primary bool   `json:"primary" yaml:"Primary"`
	Node    string `json:"node" yaml:"Node"`
}

// ShardInfo contains information about the shard
type ShardInfo struct {
	Routing          *ShardRouting `json:"routing" yaml:"Routing"`
	Node             *ShortNodeInfo
	CommitedSegments int                     `json:"num_committed_segments" yaml:"Committed Segments"`
	SearchSegments   int                     `json:"num_search_segments" yaml:"Search Segments"`
	Segments         map[string]*SegmentInfo `json:"segments" yaml:"Segments"`
}

// SegmentInfo contains information on search segment
type SegmentInfo struct {
	NumDocuments     int    `json:"num_docs" yaml:"Documents"`
	DeletedDocuments int    `json:"deleted_docs" yaml:"Deleted docs"`
	SizeBytes        int    `json:"size_in_bytes" yaml:"Size"`
	MemoryBytes      int    `json:"memory_in_bytes" yaml:"Memory"`
	Committed        bool   `json:"committed" yaml:"Committed"`
	Search           bool   `json:"search" yaml:"Search"`
	Version          string `json:"version" yaml:"Version"`
	Compound         bool   `json:"compound" yaml:"Compound"`
}

// IndexShard is information about index shard.
// ID is number of shard
// Shards contains information on actual shards allocated to nodes
type IndexShard struct {
	ID     int
	Shards []*ShardInfo
}

// IndexViewMapping returns string containing JSON of mapping information
func (e Es) IndexViewMapping(indexName string, documentType string, propertyName string) (*IndexMappings, error) {
	body, err := e.getJSON(fmt.Sprintf("/%s/_mapping", indexName))

	if err != nil {
		return nil, err
	}

	err = checkError(body)
	if err != nil {
		return nil, fmt.Errorf("Index %s failed: %s", indexName, err.Error())
	}

	indexName = e.resolveAlias(indexName)

	body = body[indexName].(map[string]interface{})

	if doc, ok := body["mappings"]; ok {
		body = doc.(map[string]interface{})
	}

	if documentType != "" {
		if doc, ok := body[documentType]; ok {
			body = doc.(map[string]interface{})["properties"].(map[string]interface{})
		} else {
			return nil, fmt.Errorf("No '%s' document in mapping", documentType)
		}
	}
	if propertyName != "" {
		if doc, ok := body[propertyName]; ok {
			body = doc.(map[string]interface{})
		} else {
			return nil, fmt.Errorf("No '%s' property in document '%s'", propertyName, documentType)
		}
	}
	mappings := &IndexMappings{}
	err = utils.DictToAny(body, mappings)
	if err != nil {
		return nil, err
	}
	return mappings, nil
}

// IndexViewSettings retrieves index settings
func (e Es) IndexViewSettings(indexName string) (*IndexSettings, error) {
	body, err := e.getJSON(fmt.Sprintf("/%s/_settings", indexName))

	if err != nil {
		return nil, err
	}

	// return settings, nil
	err = checkError(body)
	if err != nil {
		return nil, fmt.Errorf("Index %s failed: %s", indexName, err.Error())
	}

	indexName = e.resolveAlias(indexName)

	settings := &IndexSettings{}
	err = utils.DictToAny(body[indexName].(map[string]interface{})["settings"].(map[string]interface{})["index"].(map[string]interface{}), settings)

	return settings, nil
}

// IndexShards is just a slice of IndexShard
type IndexShards []IndexShard

//IndexShards returns list of shards allocated for a given index
func (e Es) IndexShards(indexName string) (IndexShards, error) {
	body, err := e.getJSON(fmt.Sprintf("/%s/_segments", indexName))

	if err != nil {
		return nil, err
	}

	err = checkError(body)
	if err != nil {
		return nil, fmt.Errorf("Index %s failed: %s", indexName, err.Error())
	}

	indexName = e.resolveAlias(indexName)

	shards := body["indices"].(map[string]interface{})[indexName].(map[string]interface{})["shards"].(map[string]interface{})

	var result []IndexShard
	for shardIdx := range shards {
		shard := shards[shardIdx].([]interface{})
		var shardInfos = make([]*ShardInfo, len(shard))
		for idx, subShardM := range shard {
			subShard := subShardM.(map[string]interface{})

			shardInfo := &ShardInfo{}
			err = utils.DictToAnyJ(subShard, shardInfo)
			if err != nil {
				return nil, err
			}
			nodeID := shardInfo.Routing.Node

			node, ok := e.Nodes[nodeID]
			if !ok {
				return nil, fmt.Errorf("Failed to parse response: Unknown node %s", node)
			}

			shardInfo.Node = node

			shardInfos[idx] = shardInfo
		}
		id, _ := strconv.Atoi(shardIdx)
		indexShard := IndexShard{
			ID:     id,
			Shards: shardInfos,
		}
		result = append(result, indexShard)
	}
	res := IndexShards(result)
	sort.Sort(res)
	return res, nil
}

func (is IndexShards) Len() int {
	return len(is)
}

func (is IndexShards) Less(i, j int) bool {
	return is[i].ID < is[j].ID
}

func (is IndexShards) Swap(i, j int) {
	is[i], is[j] = is[j], is[i]
}
