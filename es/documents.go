package es

import (
	"fmt"
)

//DocumentProperty is a container for simple property information, it includes Name and Type
type DocumentProperty struct {
	Name string
	Type string
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
