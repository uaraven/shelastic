package es

import (
	"bytes"
	"encoding/json"
	"fmt"
	"shelastic/utils"
	"sort"
	"strconv"
	"strings"
)

// ShortIndexInfo contains basic index information
type ShortIndexInfo struct {
	Name          string
	DocumentCount int
	DeletedCount  int
	Size          int
	Aliases       []*ShortAliasInfo
}

// ShortAliasInfo contains basic alias information
type ShortAliasInfo struct {
	Name     string
	Filtered bool
	Filter   string
}

// IndexSettings contains index settings (surprise!)
type IndexSettings map[string]interface{}

// IndexMappings contains index mappings
type IndexMappings map[string]interface{}

// ShardInfo contains information about the shard
type ShardInfo struct {
	State            string
	Primary          bool
	Node             *ShortNodeInfo
	CommitedSegments int
	SearchSegments   int
}

// IndexShard is information about index shard.
// ID is number of shard
// Shards contains information on actual shards allocated to nodes
type IndexShard struct {
	ID     int
	Shards []*ShardInfo
}

// ListIndices returns slice of *ShortIndexInfo containing names of indices
func (e Es) ListIndices() ([]*ShortIndexInfo, error) {
	body, err := e.getJSON("/_stats")

	if err != nil {
		return nil, err
	}

	body = body["indices"].(map[string]interface{})

	result := make([]*ShortIndexInfo, len(body))
	i := 0
	for index := range body {
		idx := body[index].(map[string]interface{})
		prim := idx["primaries"].(map[string]interface{})
		docs := prim["docs"].(map[string]interface{})
		store := prim["store"].(map[string]interface{})

		aliases, err := e.GetAliases(index)
		if err != nil {
			aliases = make([]*ShortAliasInfo, 0)
		}

		sii := &ShortIndexInfo{
			Name:          index,
			DocumentCount: int(docs["count"].(float64)),
			DeletedCount:  int(docs["deleted"].(float64)),
			Size:          int(store["size_in_bytes"].(float64)),
			Aliases:       aliases,
		}

		result[i] = sii
		i++
	}
	return result, nil
}

// GetAliases retrieves aliases for a given index
// Each alias contains name and filter in yaml format, if alias is filtered
func (e Es) GetAliases(indexName string) ([]*ShortAliasInfo, error) {
	body, err := e.getJSON(fmt.Sprintf("/%s/_alias/*", indexName))

	if err != nil {
		return nil, err
	}

	if len(body) == 0 {
		return []*ShortAliasInfo{}, nil
	}

	body = body[indexName].(map[string]interface{})["aliases"].(map[string]interface{})

	result := make([]*ShortAliasInfo, len(body))
	i := 0
	for alias := range body {
		filter := body[alias].(map[string]interface{})
		filterYaml, err := utils.MapToYaml(filter)

		if err != nil {
			return nil, err
		}

		sai := &ShortAliasInfo{
			Name:     alias,
			Filtered: len(filter) > 0,
			Filter:   filterYaml,
		}

		result[i] = sai
		i++
	}
	return result, nil
}

func getAnyKey(m map[string]interface{}) string {
	for k := range m {
		return k
	}
	return ""
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

// Flush flushes ES index
func (e Es) Flush(indexName string, force bool, wait bool) error {
	var path string
	if indexName != "" {
		path = fmt.Sprintf("/%s/_flush", indexName)
	} else {
		path = "/_flush"
	}
	_, err := e.post(path, "")
	return err
}

// ClearCache clears index's cache
func (e Es) ClearCache(indexName string) error {
	var path string
	if indexName != "" {
		path = fmt.Sprintf("/%s/_cache/clear", indexName)
	} else {
		path = "/_cache/clear"
	}
	_, err := e.post(path, "")
	return err
}

// Refresh refreshes index, making all operations performed since last refresh available for search
func (e Es) Refresh(indexName string) error {
	var path string
	if indexName != "" {
		path = fmt.Sprintf("/%s/_refresh", indexName)
	} else {
		path = "/_refresh"
	}
	_, err := e.post(path, "")
	return err
}

// ForceMerge allows to force merging of one or more indices through an API.
// For ES version 1.x and 2.x this calls _optimize API
func (e Es) ForceMerge(indexName string) error {
	var apiName string
	if e.Version[0] < 5 {
		apiName = "_optimize"
	} else {
		apiName = "_forcemerge"
	}
	var path string
	if indexName != "" {
		path = fmt.Sprintf("/%s/%s", indexName, apiName)
	} else {
		path = "/" + apiName
	}
	_, err := e.post(path, "")
	return err
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
			routing := subShard["routing"].(map[string]interface{})

			state := routing["state"].(string)
			primary := routing["primary"].(bool)
			nodeID := routing["node"].(string)

			node, ok := e.Nodes[nodeID]
			if !ok {
				return nil, fmt.Errorf("Failed to parse response: Unknown node %s", node)
			}

			commitedSegments := int(subShard["num_committed_segments"].(float64))
			searchSegments := int(subShard["num_search_segments"].(float64))

			shardInfo := &ShardInfo{
				State:            state,
				Primary:          primary,
				Node:             node,
				CommitedSegments: commitedSegments,
				SearchSegments:   searchSegments,
			}
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

// IndexConfigure updates configuration for a given index.
func (e Es) IndexConfigure(indexName string, params map[string]string) error {
	if len(params) == 0 {
		return fmt.Errorf("No settings to update")
	}

	var kv = make([]string, len(params))
	var idx = 0
	for key := range params {
		kv[idx] = "\"" + key + "\": " + params[key]
		idx++
	}
	var payload bytes.Buffer
	payload.WriteString("{\n")
	payload.WriteString(strings.Join(kv, ",\n"))
	payload.WriteString("\n}")

	resp, err := e.putJSON(fmt.Sprintf("/%s/_settings", indexName), payload.String())

	err = checkError(resp)
	if err != nil {
		return err
	}
	return nil
}

//ResolveAndValidateIndex checks if indexName parameter is a valid index and in case it is an alias it resolves it to actual index name
func (e Es) ResolveAndValidateIndex(indexName string) (string, error) {
	if len(e.aliases) == 0 {
		indices, err := e.ListIndices()
		if err != nil {
			return "", err
		}
		for _, idx := range indices {
			if idx.Name == indexName {
				return indexName, nil
			}
		}
	} else {
		if idx, ok := e.aliases[indexName]; ok {
			return idx, nil
		}
		for als := range e.aliases {
			if indexName == e.aliases[als] {
				return indexName, nil
			}
		}
	}
	return "", fmt.Errorf("Unknown index: %s", indexName)
}

// MoveAllShardsToNode changes index routing allocation to require given node
func (e Es) MoveAllShardsToNode(index string, selector string, node string) error {
	if node != "" {
		node = "\"" + node + "\""
	} else {
		node = "null"
	}
	postBody := fmt.Sprintf("{\"index.routing.allocation.require.%s\": %s}", selector, node)

	resp, err := e.putJSON(fmt.Sprintf("/%s/_settings", index), postBody)

	if err != nil {
		return err
	}

	err = checkError(resp)

	return err
}

// TruncateIndex deletes all records in index keeping the index itself intact (with all settings and mappings)
// It does so by backing up settings and mappings, deleting and recreating index and configuring index as before
func (e Es) TruncateIndex(indexName string) error {
	indexName = e.resolveAlias(indexName)
	indexData, err := e.getJSON("/" + indexName)
	if err != nil {
		return err
	}
	indexData, ok := indexData[indexName].(map[string]interface{})
	if !ok {
		return fmt.Errorf("Cannot read index settings")
	}
	databody, err := json.Marshal(indexData)
	if err != nil {
		return err
	}
	e.delete("/" + indexName)

	_, err = e.postJSON("/"+indexName, string(databody))
	return err
}

// DeleteIndex deletes index completely
func (e Es) DeleteIndex(indexName string) error {
	indexName = e.resolveAlias(indexName)
	_, err := e.delete("/" + indexName)
	return err
}

func (sii ShortIndexInfo) String() string {
	return fmt.Sprintf("%s [docs: %d, bytes: %d, aliases:%v]", sii.Name, sii.DocumentCount, sii.Size, sii.Aliases)
}

func (sai ShortAliasInfo) String() string {
	var buffer bytes.Buffer
	buffer.WriteString(sai.Name)
	if sai.Filtered {
		buffer.WriteString("*")
	}
	return buffer.String()
}

func checkError(body map[string]interface{}) error {
	if doc, ok := body["error"]; ok {
		body, ok := doc.(map[string]interface{})
		var reason string
		if !ok {
			reason = doc.(string)
		} else {
			reason = body["reason"].(string)
		}
		return fmt.Errorf(reason)
	}
	return nil
}

func (e Es) buildAliasCache() (map[string]string, error) {
	body, err := e.getJSON("/_alias")

	if err != nil {
		return nil, err
	}
	var result = make(map[string]string)
	for index := range body {
		aliases := body[index].(map[string]interface{})["aliases"].(map[string]interface{})
		if len(aliases) > 0 {
			for a := range aliases {
				result[a] = index
			}
		}
	}
	return result, nil
}

func (e Es) resolveAlias(indexName string) string {
	if idx, ok := e.aliases[indexName]; ok {
		indexName = idx
	}
	return indexName
}
