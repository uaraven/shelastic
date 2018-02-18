package es

import (
	"fmt"
	"shelastic/utils"
)

//DocumentProperty is a container for simple property information, it includes Name and Type
type DocumentProperty struct {
	Name string
	Type string
}

//SearchResult contains results for a simple search query
type SearchResult struct {
	Total int
	Hits  []string
}

// ListDocuments lists names of the documents in the index
func (e Es) ListDocuments(index string) ([]string, error) {
	body, err := e.getJSON(fmt.Sprintf("/%s/_mapping", index))

	if err != nil {
		return nil, err
	}

	err = checkError(body)
	if err != nil {
		return nil, fmt.Errorf("Index %s failed: %s", index, err.Error())
	}

	index = e.resolveAlias(index)

	body = body[index].(map[string]interface{})

	if mapping, ok := body["mappings"]; ok {
		body = mapping.(map[string]interface{})
	}

	result := make([]string, len(body))
	i := 0
	for doc := range body {
		result[i] = doc
		i++
	}
	return result, nil
}

//ListProperties lists properties of a given document of a given index
func (e Es) ListProperties(index string, doc string) ([]DocumentProperty, error) {
	body, err := e.getJSON(fmt.Sprintf("/%s/_mapping/%s", index, doc))

	if err != nil {
		return nil, err
	}

	err = checkError(body)
	if err != nil {
		return nil, fmt.Errorf("Index %s failed: %s", index, err.Error())
	}

	index = e.resolveAlias(index)

	body = body[index].(map[string]interface{})

	if mapping, ok := body["mappings"]; ok {
		body = mapping.(map[string]interface{})
	}

	if document, ok := body[doc]; ok {
		body = document.(map[string]interface{})["properties"].(map[string]interface{})
	} else {
		return nil, fmt.Errorf("No '%s' document in mapping", doc)
	}

	result := make([]DocumentProperty, len(body))
	i := 0
	for field := range body {
		v := DocumentProperty{
			Name: field,
			Type: body[field].(map[string]interface{})["type"].(string),
		}
		result[i] = v
		i++
	}
	return result, nil
}

// GetDocument reads document by id and returns string with YAML-formatted document
func (e Es) GetDocument(index string, docType string, id string) (string, error) {
	body, err := e.getJSON(fmt.Sprintf("/%s/%s/%s", index, docType, id))

	if err != nil {
		return "", err
	}

	err = checkError(body)
	if err != nil {
		return "", fmt.Errorf("Index %s failed: %s", index, err.Error())
	}

	result, err := utils.MapToYaml(body)

	if err != nil {
		return "", err
	}
	return result, nil
}

// DeleteDocument deletes document by id
func (e Es) DeleteDocument(index string, docType string, id string) error {
	body, err := e.delete(fmt.Sprintf("/%s/%s/%s", index, docType, id))

	if err != nil {
		return err
	}

	err = checkError(body)
	if err != nil {
		return fmt.Errorf("Index %s failed: %s", index, err.Error())
	}

	result, ok := body["result"].(string)

	if ok && result == "deleted" {
		return nil
	}

	if ok {
		return fmt.Errorf("Failed to delete document: " + result)
	}
	return fmt.Errorf("Failed to parse response from server")
}

//PutDocument stores JSON document in index/doc with provided id
func (e Es) PutDocument(index string, doc string, id string, reqBody string) (string, error) {
	if id == "-" {
		id = ""
	}
	body, err := e.putJSON(fmt.Sprintf("/%s/%s/%s", index, doc, id), reqBody)
	if err != nil {
		return "failed", err
	}
	err = checkError(body)
	if err != nil {
		return "failed", err
	}
	result, ok := body["result"].(string)
	if !ok {
		return "failed", fmt.Errorf("Failed to parse response")
	}
	return result, nil
}

//Search function implements ES URL search
func (e Es) Search(index string, doc string, query string) (*SearchResult, error) {
	if doc != "" {
		doc = "/" + doc
	}
	body, err := e.getJSON(fmt.Sprintf("/%s%s/_search?q=%s", index, doc, query))

	if err != nil {
		return nil, err
	}

	err = checkError(body)
	if err != nil {
		return nil, err
	}

	allhits, ok := body["hits"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Failed to retrieve hits from response")
	}

	total := int(allhits["total"].(float64))
	hits := allhits["hits"].([]interface{})

	result := make([]string, len(hits))
	i := 0
	for _, hit := range hits {
		record, err := utils.MapToYaml(hit)
		if err != nil {
			return nil, err
		}
		result[i] = record
		i++
	}

	return &SearchResult{
		Total: total,
		Hits:  result,
	}, nil
}
