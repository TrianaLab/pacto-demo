# Pacto Demo

Production failures rarely happen because the code is wrong. They happen because a port changed, a required config key was missing, or a dependency upgraded and broke an assumption no one documented.

**Pacto** is a runtime contract system. It captures what a service needs to run correctly — interfaces, configuration, dependencies, and runtime assumptions — in a single, machine-readable contract. That contract is versioned, published to an OCI registry, validated in CI, and diffed on every pull request.

> If it passes Pacto, it won't surprise you in production.

> **[github.com/trianalab/pacto](https://github.com/trianalab/pacto)** — Install the CLI, read the docs, and learn more.

---

## Mental Model

Think of Pacto as **OpenAPI, but for the entire service** — not just the HTTP interface.

- API contracts (ports, protocols, endpoints)
- Configuration validation (schema, required values, defaults)
- Dependency resolution (services, versions, compatibility)
- Runtime guarantees (health, state, scaling, persistence)

One contract. One validation step. One place to check before deploying.

---

## What Pacto Replaces

Today, these checks are scattered across your stack:

- **CI scripts** that validate env vars or config files
- **Helm values validation** buried in templates
- **Runbooks and tribal knowledge** about which services depend on what
- **Integration tests** that catch port or protocol mismatches too late

Pacto consolidates all of this into a single contract per service, validated automatically on every change.

---

## Breaking Change Detection

A port changes from `8081` to `9090`. In most systems, you find out when the deployment fails. With Pacto, you find out before code is merged.

Simulate contract changes inline — no files to copy or edit. Use `--new-set` on `pacto diff` to apply hypothetical changes and see how they would be classified:

```
$ pacto diff oci://ghcr.io/trianalab/pacto-demo/runtime oci://ghcr.io/trianalab/pacto-demo/runtime \
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
$ pacto diff oci://ghcr.io/trianalab/pacto-demo/runtime oci://ghcr.io/trianalab/pacto-demo/runtime \
    --new-values overrides/breaking-changes.yaml
```

Both approaches produce the same result — choose whichever fits your workflow.

---

## What Gets Validated

A single `pacto.yaml` declares the full operational surface of a service:

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

## Scenario Testing with Overrides

Overrides let you run "what-if" simulations against contracts without touching any files. Test how a change would behave before you commit it.

```bash
# Validate with production config values
pacto validate oci://ghcr.io/trianalab/pacto-demo/runtime --values overrides/production.yaml

# Override a single field inline
pacto validate oci://ghcr.io/trianalab/pacto-demo/runtime --set service.version=2.0.0

# Diff with overrides on the new contract
pacto diff oci://ghcr.io/trianalab/pacto-demo/runtime oci://ghcr.io/trianalab/pacto-demo/runtime \
    --new-set 'interfaces[0].port=9090'
```

Use cases:
- Simulate a version bump and check for breaking changes before committing
- Validate a contract against production-specific configuration
- Test dependency compatibility with a hypothetical upgrade

Overrides are available on all commands: `validate`, `explain`, `diff`, `doc`, and `pack`.

---

## Dependency Graph

Services don't run in isolation. A missing or incompatible dependency is a production incident waiting to happen.

Pacto resolves OCI-based dependency trees before deployment. Each service declares its dependencies as OCI references with semver constraints, and Pacto pulls, resolves, and validates compatibility across the chain:

```
$ pacto graph oci://ghcr.io/trianalab/pacto-demo/api

api@1.0.0
└─ inference@1.0.0
   └─ runtime@1.0.0
```

If a dependency introduces a breaking change, `pacto diff` catches it. If a version constraint can't be satisfied, `pacto graph` fails. Both happen in CI, not in production.

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

Every change is validated before it reaches production. This repository runs the full Pacto pipeline using [pacto-actions](https://github.com/trianalab/pacto-actions) — every workflow runs automatically on push to `main`.

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

Prerequisites: [Pacto CLI](https://github.com/trianalab/pacto)

```bash
# validate all contracts
pacto validate oci://ghcr.io/trianalab/pacto-demo/runtime
pacto validate oci://ghcr.io/trianalab/pacto-demo/inference
pacto validate oci://ghcr.io/trianalab/pacto-demo/api

# resolve the full dependency graph
pacto graph oci://ghcr.io/trianalab/pacto-demo/api

# inspect a contract
pacto explain oci://ghcr.io/trianalab/pacto-demo/inference

# generate documentation
pacto doc oci://ghcr.io/trianalab/pacto-demo/inference

# package a contract as an OCI bundle (requires local clone)
pacto pack services/runtime/pacto -o dist/runtime.tar.gz

# detect breaking changes using overrides (no file edits needed)
pacto diff oci://ghcr.io/trianalab/pacto-demo/runtime oci://ghcr.io/trianalab/pacto-demo/runtime \
    --new-set service.version=2.0.0 \
    --new-set 'interfaces[0].port=9090'
```

---

## License

MIT
