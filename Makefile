REGISTRY := ghcr.io/trianalab/pacto-demo

# Single-version bundles
SINGLE_BUNDLES := \
	platform-http-policy \
	platform-app-config \
	platform-worker-config \
	postgresql \
	redis \
	stripe-api \
	email-provider \
	auth-service \
	fraud-service \
	notification-worker \
	frontend \
	pacto-demo

# Multi-version bundles (version dirs under the bundle)
MULTI_BUNDLES := \
	api-gateway \
	orders-service \
	payments-service

.PHONY: validate explain graph doc pack push push-all diff breaking-change dashboard clean

## -- Pacto Commands --------------------------------------------

# Validate every contract bundle
validate:
	@echo "==> Validating single-version bundles..."
	@for svc in $(SINGLE_BUNDLES); do \
		echo "  $$svc"; \
		pacto validate bundles/$$svc; \
	done
	@echo "==> Validating multi-version bundles..."
	@for svc in $(MULTI_BUNDLES); do \
		for ver in bundles/$$svc/v*/; do \
			echo "  $$svc/$$(basename $$ver)"; \
			pacto validate $$ver; \
		done; \
	done

# Print a human-readable summary of each contract
explain:
	@for svc in $(SINGLE_BUNDLES); do \
		pacto explain bundles/$$svc; \
		echo ""; \
	done
	@for svc in $(MULTI_BUNDLES); do \
		for ver in bundles/$$svc/v*/; do \
			pacto explain $$ver; \
			echo ""; \
		done; \
	done

# Resolve and display the full dependency graph from the root
graph:
	@pacto graph oci://$(REGISTRY)/pacto-demo

# Generate documentation for every contract
doc:
	@mkdir -p dist/docs
	@for svc in $(SINGLE_BUNDLES); do \
		pacto doc bundles/$$svc -o dist/docs/$$svc; \
	done
	@for svc in $(MULTI_BUNDLES); do \
		for ver in bundles/$$svc/v*/; do \
			v=$$(basename $$ver); \
			pacto doc $$ver -o dist/docs/$$svc-$$v; \
		done; \
	done

# Package every contract into an OCI-ready bundle
pack:
	@mkdir -p dist
	@for svc in $(SINGLE_BUNDLES); do \
		pacto pack bundles/$$svc -o dist/$$svc.tar.gz; \
	done
	@for svc in $(MULTI_BUNDLES); do \
		for ver in bundles/$$svc/v*/; do \
			v=$$(basename $$ver | sed 's/^v//'); \
			pacto pack $$ver -o dist/$$svc-$$v.tar.gz; \
		done; \
	done

# Push all contract bundles to the OCI registry
push:
	@echo "==> Pushing single-version bundles..."
	@for svc in $(SINGLE_BUNDLES); do \
		echo "  $(REGISTRY)/$$svc"; \
		pacto push oci://$(REGISTRY)/$$svc -p bundles/$$svc; \
	done
	@echo "==> Pushing multi-version bundles..."
	@for svc in $(MULTI_BUNDLES); do \
		for ver in bundles/$$svc/v*/; do \
			v=$$(basename $$ver | sed 's/^v//'); \
			echo "  $(REGISTRY)/$$svc:$$v"; \
			pacto push oci://$(REGISTRY)/$$svc:$$v -p $$ver; \
		done; \
	done

# Compare two contract versions (usage: make diff OLD=<path-or-oci> NEW=<path-or-oci>)
diff:
	@pacto diff $(OLD) $(NEW)

## -- Demo Scenarios --------------------------------------------

# Show the payments-service breaking change (v1.2.0 -> v2.0.0)
breaking-change:
	@echo "==> Breaking change: payments-service v1.2.0 -> v2.0.0"
	@pacto diff \
		bundles/payments-service/v1.2.0 \
		bundles/payments-service/v2.0.0 \
		--output-format markdown || true

# Show a non-breaking evolution (payments-service v1.0.0 -> v1.1.0)
evolution:
	@echo "==> Non-breaking evolution: payments-service v1.0.0 -> v1.1.0"
	@pacto diff \
		bundles/payments-service/v1.0.0 \
		bundles/payments-service/v1.1.0 \
		--output-format markdown || true

# Launch the interactive dashboard from the root entry point
dashboard:
	@pacto dashboard oci://$(REGISTRY)/pacto-demo

## -- Housekeeping ----------------------------------------------

clean:
	@rm -rf dist/
