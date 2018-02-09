package es

import (
	"bytes"
	"encoding/json"
	"fmt"
	"shelastic/utils"
	"strconv"
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
type IndexSettings struct {
	NumberOfShards   int
	NumberOfReplicas int
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

func (e Es) buildAliasCache() (map[string]string, error) {
	body, err := e.getJSON("/_alias")

	fmt.Println(body)

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
func (e Es) IndexViewMapping(indexName string, documentType string, propertyName string) (string, error) {
	body, err := e.getJSON(fmt.Sprintf("/%s/_mapping", indexName))

	if err != nil {
		return "", err
	}

	if doc, ok := body["error"]; ok {
		body = doc.(map[string]interface{})
		reason := body["reason"].(string)
		return "", fmt.Errorf("Index %s failed: %s", indexName, reason)
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
			return "", fmt.Errorf("No '%s' document in mapping", documentType)
		}
	}
	if propertyName != "" {
		if doc, ok := body[propertyName]; ok {
			body = doc.(map[string]interface{})
		} else {
			return "", fmt.Errorf("No '%s' property in document '%s'", propertyName, documentType)
		}
	}

	data, err := json.MarshalIndent(body, "", "  ")
	if err == nil {
		return string(data), nil
	}
	return "", err
}

// IndexViewSettings retrieves index settings
func (e Es) IndexViewSettings(indexName string) (*IndexSettings, error) {
	body, err := e.getJSON(fmt.Sprintf("/%s/_settings", indexName))

	if err != nil {
		return nil, err
	}
	if doc, ok := body["error"]; ok {
		body = doc.(map[string]interface{})
		reason := body["reason"].(string)
		return nil, fmt.Errorf("Index %s failed: %s", indexName, reason)
	}

	indexName = e.resolveAlias(indexName)

	settings := body[indexName].(map[string]interface{})["settings"].(map[string]interface{})["index"].(map[string]interface{})

	noReplicas, err := strconv.Atoi(settings["number_of_replicas"].(string))
	if err != nil {
		return nil, err
	}

	noShards, err := strconv.Atoi(settings["number_of_shards"].(string))
	if err != nil {
		return nil, err
	}

	return &IndexSettings{
		NumberOfReplicas: noReplicas,
		NumberOfShards:   noShards,
	}, nil
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

func (e Es) resolveAlias(indexName string) string {
	if idx, ok := e.aliases[indexName]; ok {
		indexName = idx
	}
	return indexName
}
