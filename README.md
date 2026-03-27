# Pacto Demo

Production breaks don't come from bad code. They come from a renamed env var, a removed API path, a dependency that silently became required. Assumptions no one wrote down.

**Pacto** turns those assumptions into versioned, machine-readable contracts -- published to OCI, diffed automatically, validated in CI.

This repo is a live demo. Every contract is published to `ghcr.io/trianalab/pacto-demo`. Nothing to clone.

> **[github.com/trianalab/pacto](https://github.com/trianalab/pacto)** -- Install the CLI.

---

## Try It Now

```bash
pacto dashboard --repo oci://ghcr.io/trianalab/pacto-demo/pacto-demo
```

One command. No setup. The CLI pulls the root contract from OCI, resolves the full dependency graph, and opens a local dashboard with every service, interface, dependency, and config in the platform.

---

## Demo

![Demo](assets/pacto.mp4)

---

## The Graph

`pacto-demo` is a single root contract. Its only dependency is `frontend`. Everything else is resolved recursively through the graph.

From a single contract, the entire platform unfolds:

```bash
pacto graph oci://ghcr.io/trianalab/pacto-demo/pacto-demo
```

```
pacto-demo@1.0.0
└─ frontend@1.0.0
   └─ api-gateway@1.1.0
      ├─ auth-service@1.0.0
      │  └─ redis@7.2.0
      ├─ orders-service@1.1.0
      │  ├─ postgresql@16.0.0
      │  ├─ payments-service@1.2.0 (shared)
      │  └─ notification-worker@1.0.0
      │     └─ email-provider@1.0.0
      └─ payments-service@1.2.0
         ├─ postgresql@16.0.0 (shared)
         ├─ stripe-api@2024.01.01
         └─ fraud-service@1.0.0
            └─ redis@7.2.0 (shared)
```

This is not a static diagram. Every node is a versioned contract pulled live from OCI.

---

## Breaking Change Detection

The `payments-service` evolves through 4 published versions:

| Version | What changes |
|---------|-------------|
| **v1.0.0** | Baseline. `POST /charges`, `POST /refunds`, two events. |
| **v1.1.0** | Adds webhook endpoint, `payment.failed` event, optional `WEBHOOK_SECRET`. |
| **v1.2.0** | Optional `fraud-service` dep, `risk_score` field, `FRAUD_CHECK_ENABLED`. |
| **v2.0.0** | **Breaking.** Removes `/charges`, adds `/payment-intents`, renames config, `fraud-service` required. |

Diff v1.2.0 against v2.0.0:

```bash
pacto diff \
  oci://ghcr.io/trianalab/pacto-demo/payments-service:1.2.0 \
  oci://ghcr.io/trianalab/pacto-demo/payments-service:2.0.0
```

```
Classification: BREAKING
Changes (18):
  [BREAKING] openapi.paths[/charges] (removed): API path /charges removed [- /charges]
  [BREAKING] openapi.paths[/charges/{id}] (removed): API path /charges/{id} removed [- /charges/{id}]
  [BREAKING] schema.properties[STRIPE_API_KEY] (removed): configuration property STRIPE_API_KEY removed [- STRIPE_API_KEY]
  [BREAKING] schema.properties[ENABLE_REFUNDS] (removed): configuration property ENABLE_REFUNDS removed [- ENABLE_REFUNDS]
  [POTENTIAL_BREAKING] runtime.lifecycle.upgradeStrategy (modified): runtime.lifecycle.upgradeStrategy modified [rolling -> recreate]
  [POTENTIAL_BREAKING] dependencies.required (modified): dependencies.required modified [oci://ghcr.io/trianalab/pacto-demo/fraud-service: required=false -> oci://ghcr.io/trianalab/pacto-demo/fraud-service: required=true]
  [NON_BREAKING] openapi.paths[/payment-intents] (added): API path /payment-intents added [+ /payment-intents]
  [NON_BREAKING] openapi.paths[/payment-intents/{id}] (added): API path /payment-intents/{id} added [+ /payment-intents/{id}]
  [NON_BREAKING] openapi.paths[/payment-intents/{id}/confirm] (added): API path /payment-intents/{id}/confirm added [+ /payment-intents/{id}/confirm]
  [NON_BREAKING] openapi.paths[/payment-intents/{id}/cancel] (added): API path /payment-intents/{id}/cancel added [+ /payment-intents/{id}/cancel]
  ... and 8 more non-breaking changes
```

Removed API paths, renamed config, tightened dependencies, changed lifecycle -- all detected from the contracts.

---

## What's In a Contract

A Pacto contract is not just a dependency list. It captures interfaces, config, runtime, dependencies, and metadata in one file.

```bash
pacto explain oci://ghcr.io/trianalab/pacto-demo/payments-service:2.0.0
```

```
Service: payments-service@2.0.0
Owner: team/payments
Pacto Version: 1.0

Runtime:
  Workload: service
  State: stateful
  Persistence: shared/persistent
  Data Criticality: high

Interfaces (2):
  - http (http, port 8083, internal)
  - events (event, internal)

Dependencies (3):
  - oci://ghcr.io/trianalab/pacto-demo/postgresql (^16.0.0, required)
  - oci://ghcr.io/trianalab/pacto-demo/stripe-api (^2024.01.01, required)
  - oci://ghcr.io/trianalab/pacto-demo/fraud-service (^1.0.0, required)

Scaling: 2-4
```

The demo covers HTTP, gRPC, and event interfaces. Public and internal visibility. Stateless, stateful, and hybrid services. Long-running and scheduled workloads. Shared infra. Reusable policy refs. Local config schemas with platform-level properties. Multiple versions with real diff history.

---

## More Commands

```bash
# resolve the full graph
pacto graph oci://ghcr.io/trianalab/pacto-demo/pacto-demo

# inspect any service
pacto explain oci://ghcr.io/trianalab/pacto-demo/orders-service:1.1.0

# diff any two versions
pacto diff \
  oci://ghcr.io/trianalab/pacto-demo/payments-service:1.0.0 \
  oci://ghcr.io/trianalab/pacto-demo/payments-service:1.1.0

# validate a contract
pacto validate oci://ghcr.io/trianalab/pacto-demo/auth-service
```

---

## Local Development

```bash
make validate        # validate all bundles
make breaking-change # diff payments-service v1.2.0 vs v2.0.0
make dashboard       # launch the dashboard locally
make pack            # package all bundles
```

---

## CI

Runs the full Pacto pipeline on every change via [pacto-actions](https://github.com/trianalab/pacto-actions). See [workflows](.github/workflows/).

---

If you've ever had a deploy break because of a missing env var, a silent API change, or an undocumented dependency -- this is what fixes that.

---

## License

MIT
