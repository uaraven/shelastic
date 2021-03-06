package utils

import (
	"encoding/json"
	"strconv"

	"gopkg.in/yaml.v2"
)

// MapToYaml converts JSON represented as a map[string]interface{} to yaml string
func MapToYaml(inp interface{}) (string, error) {
	text, err := yaml.Marshal(inp)

	if err != nil {
		return "", err
	}

	return string(text), nil
}

// MapToJSON converts JSON represented as a map[string]interface{} to JSON string
func MapToJSON(inp interface{}) (string, error) {
	text, err := json.Marshal(inp)

	if err != nil {
		return "", err
	}

	return string(text), nil
}

// YamlStrToJSON converts stringin YAML format to JSON
func YamlStrToJSON(yamls string) (string, error) {
	var holder map[string]interface{}

	err := yaml.Unmarshal([]byte(yamls), holder)

	if err != nil {
		return "", err
	}

	jsonb, err := json.Marshal(holder)
	if err != nil {
		return "", err
	}

	return string(jsonb), nil
}

// GetAsInt reads value from map "inp" by key "name" and tries to convert it to int
// If conversion fails and "orElse" is passed, then orElse[0] is returned, otherwise 0 is returned
func GetAsInt(inp map[string]interface{}, name string, orElse ...int) int {
	value, err := strconv.Atoi(inp[name].(string))
	if err != nil {
		if len(orElse) > 0 {
			return orElse[0]
		}
		return 0
	}
	return value
}

// GetAsBool reads value from map "inp" by key "name" and tries to convert it to bool
// If conversion fails and "orElse" is passed, then orElse[0] is returned, otherwise false is returned
func GetAsBool(inp map[string]interface{}, name string, orElse ...bool) bool {
	value, err := strconv.ParseBool(inp[name].(string))
	if err != nil {
		if len(orElse) > 0 {
			return orElse[0]
		}
		return false
	}
	return value
}

// DictToAny converts map[string]interface{} to any other interface
// This is naive implemenation which uses yaml as a convertation medium
func DictToAny(inp map[string]interface{}, receiver interface{}) error {
	data, err := MapToYaml(inp)
	if err != nil {
		return err
	}
	return yaml.Unmarshal([]byte(data), receiver)
}

// DictToAnyJ converts map[string]interface{} to any other interface
// This is naive implemenation which uses JSON as a convertation medium
func DictToAnyJ(inp map[string]interface{}, receiver interface{}) error {
	data, err := MapToJSON(inp)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), receiver)
}
