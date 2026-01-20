package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

type SchemaDefinition struct {
	Structure struct {
		Root struct {
			Files map[string]struct {
				Required bool                   `json:"required"`
				Schema   map[string]interface{} `json:"schema"`
			} `json:"files"`
		} `json:"root"`
	} `json:"structure"`
}

func TestElementalExampleValidation(t *testing.T) {
	schemaData, err := os.ReadFile("elemental-schema.json")
	if err != nil {
		t.Fatalf("failed to read schema: %v", err)
	}

	var schemaDef SchemaDefinition
	if err := json.Unmarshal(schemaData, &schemaDef); err != nil {
		t.Fatalf("failed to unmarshal schema: %v", err)
	}

	examples := []string{"linux-only", "single-node", "multi-node"}
	for _, example := range examples {
		t.Run(example, func(t *testing.T) {
			dir := filepath.Join("elemental-example", example)
			for filename, fileSpec := range schemaDef.Structure.Root.Files {
				if fileSpec.Schema == nil {
					// Skip if no schema is defined (e.g. butane.yaml has schemaReference)
					continue
				}

				path := filepath.Join(dir, filename)
				if _, err := os.Stat(path); os.IsNotExist(err) {
					if fileSpec.Required {
						t.Errorf("required file %s is missing in %s", filename, example)
					}
					continue
				}

				yamlData, err := os.ReadFile(path)
				if err != nil {
					t.Errorf("failed to read %s: %v", path, err)
					continue
				}

				var yamlObj interface{}
				if err := yaml.Unmarshal(yamlData, &yamlObj); err != nil {
					t.Errorf("failed to unmarshal YAML %s: %v", path, err)
					continue
				}

				// Convert to JSON-compatible for validation (ensures all map keys are strings)
				jsonObj := convertToJSONCompatible(yamlObj)

				schemaLoader := gojsonschema.NewGoLoader(fileSpec.Schema)
				documentLoader := gojsonschema.NewGoLoader(jsonObj)

				result, err := gojsonschema.Validate(schemaLoader, documentLoader)
				if err != nil {
					t.Errorf("validation error for %s: %v", path, err)
					continue
				}

				if !result.Valid() {
					t.Errorf("file %s is not valid against schema:", path)
					for _, desc := range result.Errors() {
						t.Errorf("- %s", desc)
					}
				}
			}
		})
	}
}

// convertToJSONCompatible ensures that the object can be serialized to JSON and back,
// specifically converting map[interface{}]interface{} to map[string]interface{}.
func convertToJSONCompatible(i interface{}) interface{} {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[fmt.Sprintf("%v", k)] = convertToJSONCompatible(v)
		}
		return m2
	case map[string]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k] = convertToJSONCompatible(v)
		}
		return m2
	case []interface{}:
		res := make([]interface{}, len(x))
		for i, v := range x {
			res[i] = convertToJSONCompatible(v)
		}
		return res
	default:
		return i
	}
}
