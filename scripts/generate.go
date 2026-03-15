// generate.go generates OpenAPI specs and JSON Schema config files for all services.
//
// Usage: go run scripts/generate.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"

	apiInternal "github.com/trianalab/pacto-demo/services/api/service"
	inferenceInternal "github.com/trianalab/pacto-demo/services/inference/service"
	runtimeInternal "github.com/trianalab/pacto-demo/services/runtime/service"
)

type serviceSpec struct {
	name    string
	version string
	setup   func(api huma.API)
	config  any
}

func main() {
	services := []serviceSpec{
		{
			name:    "runtime",
			version: "1.0.0",
			setup: func(api huma.API) {
				huma.Post(api, "/predict", runtimeInternal.Predict)
				huma.Get(api, "/models", runtimeInternal.ListModels)
				huma.Get(api, "/health", runtimeInternal.HealthCheck)
			},
			config: runtimeInternal.Config{},
		},
		{
			name:    "inference",
			version: "1.0.0",
			setup: func(api huma.API) {
				huma.Post(api, "/infer", inferenceInternal.Infer)
				huma.Post(api, "/infer/batch", inferenceInternal.BatchInfer)
				huma.Get(api, "/health", inferenceInternal.HealthCheck)
			},
			config: inferenceInternal.Config{},
		},
		{
			name:    "api",
			version: "1.0.0",
			setup: func(api huma.API) {
				huma.Post(api, "/analyze", apiInternal.Analyze)
				huma.Get(api, "/status", apiInternal.Status)
				huma.Get(api, "/health", apiInternal.HealthCheck)
			},
			config: apiInternal.Config{},
		},
	}

	for _, svc := range services {
		generateOpenAPI(svc)
		generateConfigSchema(svc)
	}

	fmt.Println("All specs generated successfully.")
}

func generateOpenAPI(svc serviceSpec) {
	router := chi.NewMux()
	title := strings.ToUpper(svc.name[:1]) + svc.name[1:] + " Service"
	api := humachi.New(router, huma.DefaultConfig(title, svc.version))
	svc.setup(api)

	spec := api.OpenAPI()
	data, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal OpenAPI for %s: %v", svc.name, err)
	}

	outPath := filepath.Join("services", svc.name, "pacto", "interfaces", "openapi.json")
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		log.Fatalf("Failed to create dir for %s: %v", svc.name, err)
	}
	if err := os.WriteFile(outPath, data, 0o644); err != nil {
		log.Fatalf("Failed to write OpenAPI for %s: %v", svc.name, err)
	}
	fmt.Printf("  Generated %s\n", outPath)
}

func generateConfigSchema(svc serviceSpec) {
	schema := structToJSONSchema(svc.config)
	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal config schema for %s: %v", svc.name, err)
	}

	outPath := filepath.Join("services", svc.name, "pacto", "configuration", "schema.json")
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		log.Fatalf("Failed to create dir for %s: %v", svc.name, err)
	}
	if err := os.WriteFile(outPath, data, 0o644); err != nil {
		log.Fatalf("Failed to write config schema for %s: %v", svc.name, err)
	}
	fmt.Printf("  Generated %s\n", outPath)
}

type jsonSchema struct {
	Type        string                `json:"type"`
	Properties  map[string]jsonProp   `json:"properties,omitempty"`
	Required    []string              `json:"required,omitempty"`
}

type jsonProp struct {
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Default     any      `json:"default,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}

func structToJSONSchema(v any) jsonSchema {
	t := reflect.TypeOf(v)
	schema := jsonSchema{
		Type:       "object",
		Properties: make(map[string]jsonProp),
	}

	var required []string

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}
		name := strings.Split(jsonTag, ",")[0]

		prop := jsonProp{
			Description: field.Tag.Get("doc"),
		}

		switch field.Type.Kind() {
		case reflect.String:
			prop.Type = "string"
		case reflect.Int, reflect.Int64:
			prop.Type = "integer"
		case reflect.Bool:
			prop.Type = "boolean"
		case reflect.Float64:
			prop.Type = "number"
		default:
			prop.Type = "string"
		}

		if enumTag := field.Tag.Get("enum"); enumTag != "" {
			prop.Enum = strings.Split(enumTag, ",")
		}

		if defaultTag := field.Tag.Get("default"); defaultTag != "" {
			prop.Default = defaultTag
		}

		// Fields without a default are considered required
		if field.Tag.Get("default") == "" {
			required = append(required, name)
		}

		schema.Properties[name] = prop
	}

	schema.Required = required
	return schema
}
