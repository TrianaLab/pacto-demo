// remove-endpoint.go removes a path from an OpenAPI JSON file.
//
// Usage: go run scripts/remove-endpoint.go <openapi.json> <path-to-remove>
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("Usage: %s <openapi.json> <path-to-remove>", os.Args[0])
	}

	file := os.Args[1]
	pathToRemove := os.Args[2]

	data, err := os.ReadFile(file)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	var spec map[string]any
	if err := json.Unmarshal(data, &spec); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	paths, ok := spec["paths"].(map[string]any)
	if !ok {
		log.Fatal("No paths found in OpenAPI spec")
	}

	if _, exists := paths[pathToRemove]; !exists {
		log.Fatalf("Path %s not found in spec", pathToRemove)
	}

	delete(paths, pathToRemove)
	fmt.Printf("Removed path: %s\n", pathToRemove)

	out, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal: %v", err)
	}

	if err := os.WriteFile(file, out, 0o644); err != nil {
		log.Fatalf("Failed to write: %v", err)
	}
}
