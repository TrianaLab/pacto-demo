SERVICES := runtime inference api
REGISTRY := ghcr.io/trianalab/pacto-demo

.PHONY: generate build validate explain graph doc pack push diff breaking-change clean

## -- Code Generation -------------------------------------------

# Generate JSON Schema and OpenAPI specs using Pacto plugins
generate:
	@for svc in $(SERVICES); do \
		echo "==> Generating schemas for $$svc..."; \
		pacto generate schema-infer services/$$svc/pacto --option file=config.yaml -o services/$$svc/pacto; \
		ln -sf pacto/pacto.yaml services/$$svc/pacto.yaml; \
		pacto generate openapi-infer services/$$svc --option framework=huma --option output=interfaces/openapi.json -o services/$$svc/pacto; \
		rm -f services/$$svc/pacto.yaml; \
	done

# Build all service binaries
build:
	@for svc in $(SERVICES); do \
		go build -o bin/$$svc ./services/$$svc/cmd/...; \
	done

## -- Pacto Commands --------------------------------------------

# Validate every contract against the Pacto specification
validate:
	@for svc in $(SERVICES); do \
		pacto validate services/$$svc/pacto; \
	done

# Print a human-readable summary of each contract
explain:
	@for svc in $(SERVICES); do \
		pacto explain services/$$svc/pacto; \
		echo ""; \
	done

# Resolve and display the full dependency graph (from the top-level service)
graph:
	@pacto graph services/api/pacto

# Generate rich Markdown documentation for every contract
doc:
	@mkdir -p dist/docs
	@for svc in $(SERVICES); do \
		pacto doc services/$$svc/pacto -o dist/docs/$$svc; \
		echo "Generated docs for $$svc -> dist/docs/$$svc/"; \
	done

# Package every contract into an OCI-ready tar.gz bundle
pack:
	@mkdir -p dist
	@for svc in $(SERVICES); do \
		pacto pack services/$$svc/pacto -o dist/$$svc.tar.gz; \
	done

# Push all contract bundles to an OCI registry (requires authentication)
push:
	@for svc in $(SERVICES); do \
		pacto push oci://$(REGISTRY)/$$svc -p services/$$svc/pacto; \
	done

# Compare two contract versions (usage: make diff OLD=<path-or-oci> NEW=<path-or-oci>)
diff:
	@pacto diff $(OLD) $(NEW)

## -- Demo ------------------------------------------------------

# Detect breaking changes using overrides (no file edits needed)
breaking-change:
	@pacto diff oci://ghcr.io/trianalab/pacto-demo/runtime services/runtime/pacto \
		--new-set service.version=2.0.0 \
		--new-set 'service.image.ref=ghcr.io/trianalab/pacto-demo/runtime:2.0.0' \
		--new-set 'interfaces[0].port=9090' \
		--output-format markdown || true

## -- Housekeeping ----------------------------------------------

clean:
	@rm -rf bin/ dist/
