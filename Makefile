SERVICES := runtime inference api
REGISTRY := ghcr.io/trianalab/pacto-demo

.PHONY: generate build validate explain graph doc pack push diff breaking-change clean

## ── Code Generation ─────────────────────────────────────

# Generate OpenAPI specs and JSON Schema config files from Go source code
generate:
	@go run scripts/generate.go

# Build all service binaries
build:
	@for svc in $(SERVICES); do \
		go build -o bin/$$svc ./services/$$svc/cmd/...; \
	done

## ── Pacto Commands ──────────────────────────────────────

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

## ── Demo ────────────────────────────────────────────────

# Simulate a breaking change and run pacto diff to detect it
breaking-change:
	@bash scripts/breaking-change.sh

## ── Housekeeping ────────────────────────────────────────

clean:
	@rm -rf bin/ dist/
