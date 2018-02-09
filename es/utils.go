package es

import (
	"encoding/json"

	"gopkg.in/yaml.v2"
)

func JsonToYaml(inp string) (string, error) {
	bodyBytes := []byte(inp)

	var body map[string]interface{}

	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		return "", err
	}

	text, err := yaml.Marshal(body)

	if err != nil {
		return "", err
	}

	return string(text), nil
}

func MapToYaml(inp map[string]interface{}) (string, error) {
	text, err := yaml.Marshal(inp)

	if err != nil {
		return "", err
	}

	return string(text), nil
}
