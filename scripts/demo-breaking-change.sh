#!/usr/bin/env bash
#
# Simulates multiple breaking changes in the runtime service and detects them with pacto diff.
#
# Changes applied:
#   1. Bumps service version from 1.0.0 to 2.0.0
#   2. Changes the HTTP port from 8081 to 9090
#   3. Removes the POST /predict endpoint from the OpenAPI spec
#   4. Removes the model_path config property from the JSON Schema
#
# The diff compares the published OCI artifact against the modified contract.
#
set -euo pipefail

DEMO_DIR="$(cd "$(dirname "$0")/.." && pwd)"
MODIFIED=$(mktemp -d)

trap 'rm -rf "$MODIFIED"' EXIT

# Copy the current contract to a working directory
cp -r "$DEMO_DIR/services/runtime/pacto"/* "$MODIFIED/"

# Apply all breaking changes to the copy
go run "$DEMO_DIR/scripts/apply-breaking-changes.go" "$MODIFIED" 2>&1

# Diff the published OCI artifact against the modified contract
pacto diff oci://ghcr.io/trianalab/pacto-demo/runtime "$MODIFIED" --output-format markdown || true
