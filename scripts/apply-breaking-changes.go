// apply-breaking-changes.go applies multiple breaking changes to a contract directory.
//
// Usage: go run scripts/apply-breaking-changes.go <contract-dir>
//
// Changes applied:
//   1. Bumps service version from 1.0.0 to 2.0.0
//   2. Changes the HTTP port from 8081 to 9090
//   3. Removes the POST /predict endpoint from the OpenAPI spec
//   4. Removes the model_path property from the config schema
package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: %s <contract-dir>", os.Args[0])
	}
	dir := os.Args[1]

	// 1. Modify pacto.yaml: bump version, change port
	pactoFile := filepath.Join(dir, "pacto.yaml")
	pactoData, err := os.ReadFile(pactoFile)
	if err != nil {
		log.Fatalf("Failed to read pacto.yaml: %v", err)
	}
	content := string(pactoData)
	content = strings.Replace(content, "version: 1.0.0", "version: 2.0.0", 1)
	content = strings.Replace(content, "port: 8081", "port: 9090", 1)
	content = strings.Replace(content, "ref: ghcr.io/trianalab/pacto-demo/runtime:1.0.0", "ref: ghcr.io/trianalab/pacto-demo/runtime:2.0.0", 1)
	if err := os.WriteFile(pactoFile, []byte(content), 0o644); err != nil {
		log.Fatalf("Failed to write pacto.yaml: %v", err)
	}

	// 2. Remove /predict from OpenAPI spec
	openapiFile := filepath.Join(dir, "interfaces", "openapi.json")
	openapiData, err := os.ReadFile(openapiFile)
	if err != nil {
		log.Fatalf("Failed to read openapi.json: %v", err)
	}
	var spec map[string]any
	if err := json.Unmarshal(openapiData, &spec); err != nil {
		log.Fatalf("Failed to parse openapi.json: %v", err)
	}
	if paths, ok := spec["paths"].(map[string]any); ok {
		delete(paths, "/predict")
	}
	openapiOut, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal openapi.json: %v", err)
	}
	if err := os.WriteFile(openapiFile, openapiOut, 0o644); err != nil {
		log.Fatalf("Failed to write openapi.json: %v", err)
	}

	// 3. Remove model_path from config schema
	schemaFile := filepath.Join(dir, "configuration", "schema.json")
	schemaData, err := os.ReadFile(schemaFile)
	if err != nil {
		log.Fatalf("Failed to read schema.json: %v", err)
	}
	var schema map[string]any
	if err := json.Unmarshal(schemaData, &schema); err != nil {
		log.Fatalf("Failed to parse schema.json: %v", err)
	}
	if props, ok := schema["properties"].(map[string]any); ok {
		delete(props, "model_path")
	}
	if req, ok := schema["required"].([]any); ok {
		var filtered []any
		for _, r := range req {
			if r != "model_path" {
				filtered = append(filtered, r)
			}
		}
		if len(filtered) == 0 {
			delete(schema, "required")
		} else {
			schema["required"] = filtered
		}
	}
	schemaOut, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal schema.json: %v", err)
	}
	if err := os.WriteFile(schemaFile, schemaOut, 0o644); err != nil {
		log.Fatalf("Failed to write schema.json: %v", err)
	}
}
