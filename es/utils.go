package es

import (
	"gopkg.in/yaml.v2"
)

// MapToYaml converts JSON representeda as map[string]interface{} to yaml string
func MapToYaml(inp map[string]interface{}) (string, error) {
	text, err := yaml.Marshal(inp)

	if err != nil {
		return "", err
	}

	return string(text), nil
}
