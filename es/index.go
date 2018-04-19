package es

import (
	"bytes"
	"encoding/json"
	"fmt"
	"shelastic/utils"
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
	if len(e.aliases) != 0 {
		if idx, ok := e.aliases[indexName]; ok {
			return idx, nil
		}
		for als := range e.aliases {
			if indexName == e.aliases[als] {
				return indexName, nil
			}
		}
	}

	indices, err := e.ListIndices()
	if err != nil {
		return "", err
	}
	for _, idx := range indices {
		if idx.Name == indexName {
			return indexName, nil
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

//CopyIndex creates a new index named 'newName' and copies data from 'indexName' to it
//On ES 5.0+ this uses reindex API, on older versions this copies data using bulk APIs
//Mappings and settings are copied from original index
func (e Es) CopyIndex(indexName string, newName string) error {
	indexName = e.resolveAlias(indexName)

	// Create new index with the same settings as the old one, but without aliases

	// retrieve index settings
	body, err := e.getJSON(fmt.Sprintf("/%s", indexName))
	if err != nil {
		return err
	}
	err = checkError(body)
	if err != nil {
		return err
	}

	// Create settings for new index
	indexSettingsJSON := body[indexName].(map[string]interface{})
	indexSettingsJSON["aliases"] = make(map[string]interface{})
	indexSettingsJSON["settings"] = make(map[string]interface{})

	indexSettingsStr, err := json.Marshal(indexSettingsJSON)
	if err != nil {
		return err
	}

	// Create new index
	response, err := e.putJSON(newName, string(indexSettingsStr))
	if err != nil {
		return err
	}
	err = checkError(response)
	if err != nil {
		return err
	}

	if e.Version[0] < 2 && e.Version[1] < 3 {
		return e.copyData(indexName, newName)
	}
	return e.reindex(indexName, newName)
}

func (e Es) reindex(oldIndex string, newIndex string) error {
	body := fmt.Sprintf("{\"source\":{\"index\":\"%s\"}, \"dest\":{\"index\":\"%s\"}}", oldIndex, newIndex)
	resp, err := e.postJSON("_reindex?wait_for_completion=true", body)
	if err != nil {
		return err
	}
	err = checkError(resp)
	return err
}

func (e Es) copyData(oldIndex string, newIndex string) error {
	docs, err := e.ListDocuments(oldIndex)
	if err != nil {
		return err
	}
	for _, doc := range docs {
		var recordsIn = make(chan *BulkRecord, 1)
		var recordsOut = make(chan *BulkRecord, 1)
		var controlIn = make(chan error)
		var controlOut = make(chan error)

		go e.BulkExport(oldIndex, doc, "{\"query\": {\"match_all\":{}}}", recordsIn, controlIn)
		go e.bulkSink(newIndex, doc, recordsOut, controlOut)

		done := false

		for !done {
			select {
			case record, recOk := <-recordsIn:
				if recOk {
					recordsOut <- record
				} else {
					done = true
				}
			case err = <-controlIn:
				controlOut <- err
				if e.Debug {
					fmt.Printf("Received from control In: %v\n", err)
				}
			case err = <-controlOut:
				if e.Debug {
					fmt.Printf("Received from control Out: %v\n", err)
				}
				done = true
			}

		}

		if e.Debug {
			fmt.Println("Closing channels")
		}
		close(recordsIn)
		close(recordsOut)
		close(controlIn)
		close(controlOut)
	}

	if e.Debug {
		fmt.Printf("Returning: %v\n", err)
	}

	return err
}

// DeleteIndex deletes index completely
func (e Es) DeleteIndex(indexName string) error {
	indexName = e.resolveAlias(indexName)
	_, err := e.delete("/" + indexName)
	return err
}

// AddIndexAlias adds alias to index
func (e Es) AddIndexAlias(indexName string, alias string) error {
	return e.aliasOperation("add", indexName, alias)
}

// DeleteIndexAlias deletes alias from index
func (e Es) DeleteIndexAlias(indexName string, alias string) error {
	return e.aliasOperation("remove", indexName, alias)
}

// OpenIndex opens previously closed index
func (e Es) OpenIndex(indexName string) error {
	response, err := e.postJSON(fmt.Sprintf("/%s/_open", indexName), "")
	if err != nil {
		return err
	}
	err = checkError(response)
	return err
}

// CloseIndex closes previously open index
func (e Es) CloseIndex(indexName string) error {
	response, err := e.postJSON(fmt.Sprintf("/%s/_close", indexName), "")
	if err != nil {
		return err
	}
	err = checkError(response)
	return err
}

func (e Es) aliasOperation(operation string, indexName string, alias string) error {
	url := fmt.Sprintf("/%s/_alias/%s", indexName, alias)
	var resp map[string]interface{}
	var err error
	if operation == "add" {
		resp, err = e.putJSON(url, "")
	} else if operation == "remove" {
		resp, err = e.delete(url)
	}
	if err != nil {
		return err
	}
	err = checkError(resp)
	if err == nil {
		aliases, err := e.buildAliasCache()
		if err == nil {
			e.aliases = aliases
		}
	}
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

func getErrorReason(errorBody interface{}) string {
	body, ok := errorBody.(map[string]interface{})
	var reason string
	if !ok {
		reason = errorBody.(string)
	} else {
		reason = body["reason"].(string)
		if causeBody, ok := body["caused_by"]; ok {
			cause := getErrorReason(causeBody)
			return fmt.Sprintf("%s, caused by '%s'", reason, cause)
		}
	}
	return reason
}

func checkError(body map[string]interface{}) error {
	if doc, ok := body["error"]; ok {
		return fmt.Errorf(getErrorReason(doc))
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
