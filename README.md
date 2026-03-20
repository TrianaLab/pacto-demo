# Pacto Demo

Production failures rarely happen because the code is wrong. They happen because a port changed, a required config key was missing, or a dependency upgraded and broke an assumption no one documented.

**Pacto** is a runtime contract system. It captures what a service needs to run correctly — interfaces, configuration, dependencies, and runtime assumptions — in a single, machine-readable contract. That contract is versioned, published to an OCI registry, validated in CI, and diffed on every pull request.

> If a service passes Pacto validation, it is safe to run in production.

This repository demonstrates Pacto across three Go services connected through a dependency chain.

> **[github.com/trianalab/pacto](https://github.com/trianalab/pacto)** — Install the CLI, read the docs, and learn more.

---

## Breaking Change Detection

A port changes from `8081` to `9090`. In most systems, you find out when the deployment fails. With Pacto, you find out before code is merged.

Pacto's override system lets you simulate contract changes inline — no need to copy or edit files. Use `--new-set` on `pacto diff` to apply hypothetical changes to the new contract and see how they would be classified:

```
$ pacto diff oci://ghcr.io/trianalab/pacto-demo/runtime services/runtime/pacto \
    --new-set service.version=2.0.0 \
    --new-set 'service.image.ref=ghcr.io/trianalab/pacto-demo/runtime:2.0.0' \
    --new-set 'interfaces[0].port=9090'
```

**Classification:** `BREAKING`

| Classification | Path | Type | Reason | Old | New |
|---|---|---|---|---|---|
| NON_BREAKING | `service.version` | modified | service.version modified | `1.0.0` | `2.0.0` |
| NON_BREAKING | `service.image` | modified | service.image modified | `ghcr.io/trianalab/pacto-demo/runtime:1.0.0` | `ghcr.io/trianalab/pacto-demo/runtime:2.0.0` |
| BREAKING | `interfaces.port` | modified | interfaces.port modified | `8081` | `9090` |

Port changes are caught as breaking. Version and image updates are non-breaking. All detected without modifying a single file.

You can also use `--new-values` to load overrides from a file:

```yaml
# overrides/breaking-changes.yaml
service:
  version: "2.0.0"
  image:
    ref: ghcr.io/trianalab/pacto-demo/runtime:2.0.0
interfaces:
  - name: http
    type: http
    port: 9090
    visibility: internal
    contract: interfaces/openapi.json
```

```
$ pacto diff oci://ghcr.io/trianalab/pacto-demo/runtime services/runtime/pacto \
    --new-values overrides/breaking-changes.yaml
```

Both approaches produce the same result — choose whichever fits your workflow.

---

## The Contract

A single `pacto.yaml` describes everything a service needs to run correctly:

```yaml
# services/inference/pacto/pacto.yaml

pactoVersion: "1.0"

service:
  name: inference
  version: 1.0.0

interfaces:
  - name: http
    type: http
    port: 8082
    contract: interfaces/openapi.json       # auto-generated

configuration:
  schema: config.schema.json                 # auto-generated

dependencies:
  - ref: oci://ghcr.io/trianalab/pacto-demo/runtime
    required: true
    compatibility: "^1.0.0"

runtime:
  workload: service
  state:
    type: stateless
  health:
    interface: http
    path: /health
```

From this, Pacto extracts:

- **Interfaces** — protocols, ports, visibility, OpenAPI endpoints
- **Configuration** — every property, type, default, and whether it's required
- **Dependencies** — OCI references with semver compatibility constraints
- **Runtime** — workload type, state model, persistence, data criticality
- **Scaling** — replica bounds and upgrade strategy

Full contract reference: [trianalab.github.io/pacto/contract-reference](https://trianalab.github.io/pacto/contract-reference/)

---

## Overrides: Runtime Simulation Without File Edits

Overrides let you test scenarios against contracts without modifying any files. Think of it as "what-if" analysis for your service topology.

```bash
# Validate with production config values
pacto validate services/runtime/pacto --values overrides/production.yaml

# Override a single field inline
pacto validate services/runtime/pacto --set service.version=2.0.0

# Diff with overrides on the new contract
pacto diff oci://ghcr.io/trianalab/pacto-demo/runtime services/runtime/pacto \
    --new-set 'interfaces[0].port=9090'
```

Use cases:
- Simulate a version bump and check for breaking changes before committing
- Validate a contract against production-specific configuration
- Test dependency compatibility with a hypothetical upgrade

Overrides are available on all commands: `validate`, `explain`, `diff`, `doc`, and `pack`.

---

## Dependency Graph

Pacto resolves OCI-based dependency trees. Each service declares its dependencies as OCI references with semver constraints, and Pacto pulls and resolves them:

```
$ pacto graph services/api/pacto

api@1.0.0
└─ inference@1.0.0
   └─ runtime@1.0.0
```

---

## Contract Bundle

Each service produces a self-contained bundle that gets packed and pushed to an OCI registry:

```
services/runtime/pacto/
├── pacto.yaml                   # contract
├── config.yaml                  # sample config (input for schema-infer)
├── config.schema.json           # JSON Schema (generated by pacto-plugin-schema-infer)
└── interfaces/openapi.json      # OpenAPI spec (generated by pacto-plugin-openapi-infer)
```

```
$ pacto pack services/runtime/pacto -o dist/runtime.tar.gz
Packed runtime@1.0.0 -> dist/runtime.tar.gz
```

Bundles are pushed to GHCR as OCI artifacts, just like container images. Other services reference them by OCI URI (`oci://ghcr.io/trianalab/pacto-demo/runtime`), and Pacto pulls and resolves them automatically.

---

## CI Integration

This repository runs the full Pacto pipeline using [pacto-actions](https://github.com/trianalab/pacto-actions). Every workflow runs automatically on push to `main` — click any workflow to see the output in the job summary.

```
generate schemas (schema-infer + openapi-infer)
        │
        ▼
validate contracts
        │
        ▼
diff against published version (detect breaking changes)
        │
        ▼
generate documentation
        │
        ▼
publish contract to OCI registry
```

| Capability | Workflow | What it demonstrates |
|------------|----------|----------------------|
| Validation | [Validate & Explain](../../actions/workflows/demo-validate.yml) | `pacto validate` + `pacto explain` on every contract |
| Dependency graph | [Dependency Graph](../../actions/workflows/demo-graph.yml) | `pacto graph` resolves the full service tree |
| Breaking changes | [Breaking Change Detection](../../actions/workflows/demo-breaking-change.yml) | Uses `--new-set` overrides to simulate breaking changes and diff against published OCI artifact |
| Documentation | [Contract Documentation](../../actions/workflows/demo-docs.yml) | `pacto doc` generates Markdown with diagrams and tables |
| Packaging | [Pack Contract Bundles](../../actions/workflows/demo-pack.yml) | `pacto pack` creates OCI-ready bundles |
| Full CI | [Pacto CI](../../actions/workflows/ci-pacto.yml) | Validate, diff, document, and push to GHCR |

---

## Architecture

```
api@1.0.0             public gateway     :8080
└─ inference@1.0.0    ML inference       :8082
   └─ runtime@1.0.0   model execution    :8081
```

Three Go services built with [Huma](https://huma.rocks). Official Pacto plugins ([pacto-plugin-schema-infer](https://github.com/TrianaLab/pacto-plugins/tree/main/plugins/pacto-plugin-schema-infer) and [pacto-plugin-openapi-infer](https://github.com/TrianaLab/pacto-plugins/tree/main/plugins/pacto-plugin-openapi-infer)) generate the JSON Schema from sample config files and extract OpenAPI specs from source code via static analysis. Both artifacts are referenced by the Pacto contract and included in the OCI bundle.

---

## Quickstart

Prerequisites: [Go 1.23+](https://go.dev/dl/), [Pacto CLI](https://github.com/trianalab/pacto)

```bash
# validate all contracts
pacto validate services/runtime/pacto
pacto validate services/inference/pacto
pacto validate services/api/pacto

# resolve the full dependency graph
pacto graph services/api/pacto

# inspect a contract
pacto explain services/inference/pacto

# generate documentation
pacto doc services/inference/pacto

# package a contract as an OCI bundle
pacto pack services/runtime/pacto -o dist/runtime.tar.gz

# detect breaking changes using overrides (no file edits needed)
pacto diff oci://ghcr.io/trianalab/pacto-demo/runtime services/runtime/pacto \
    --new-set service.version=2.0.0 \
    --new-set 'interfaces[0].port=9090'
```

---

## License

MIT
