# Pacto Demo

Production failures rarely happen because the code is wrong. They happen because a port changed, a required config key was missing, or a dependency upgraded and broke an assumption no one documented.

**Pacto** is a runtime contract system. It captures what a service needs to run correctly -- interfaces, configuration, dependencies, and runtime assumptions -- in a single, machine-readable contract. That contract is versioned, published to an OCI registry, validated in CI, and diffed on every pull request.

> If it passes Pacto, it won't surprise you in production.

> **[github.com/trianalab/pacto](https://github.com/trianalab/pacto)** -- Install the CLI, read the docs, and learn more.

---

## Try It Now

Install the [Pacto CLI](https://github.com/trianalab/pacto), then run:

```bash
pacto dashboard --repo oci://ghcr.io/trianalab/pacto-demo/pacto-demo
```

This launches an interactive dashboard that resolves the full dependency graph from a single entry point and lets you explore every service, interface, dependency, and configuration in the platform.

No cloning required. Everything is pulled live from the OCI registry.

---

## What This Demo Shows

This repository models a **complete e-commerce platform** with 15 services across 5 tiers. It's designed to demonstrate every feature of the Pacto schema in a realistic context.

```
pacto-demo@1.0.0  (entry point)
└─ frontend@1.0.0  (public, HTTP)
   └─ api-gateway@1.1.0  (public, HTTP, local policy)
      ├─ auth-service@1.0.0  (internal, gRPC, hybrid state)
      │  └─ redis@7.2.0  (infra, stateful)
      ├─ orders-service@1.1.0  (internal, HTTP + events, stateful)
      │  ├─ postgresql@16.0.0  (infra, stateful)
      │  ├─ payments-service@2.0.0  (internal, HTTP + events, stateful)
      │  │  ├─ postgresql@16.0.0
      │  │  ├─ stripe-api@2024.01.01  (external)
      │  │  └─ fraud-service@1.0.0  (internal, gRPC, stateless)
      │  │     └─ redis@7.2.0
      │  └─ notification-worker@1.0.0  (scheduled worker, events)
      │     └─ email-provider@1.0.0  (external)
      └─ payments-service@2.0.0
```

One entry point. Everything else discovered recursively.

---

## 10 Things You'll See

| # | Feature | Where to look |
|---|---------|---------------|
| 1 | **A service is NOT just dependencies** -- it has interfaces, config, runtime, scaling, owner, image, metadata | Any `pacto.yaml` (e.g. `bundles/payments-service/v2.0.0/pacto.yaml`) |
| 2 | **Graph has depth AND meaning** -- recursive deps, optional vs required, infra vs external | `pacto graph oci://ghcr.io/trianalab/pacto-demo/pacto-demo` |
| 3 | **Real version history** -- 4 versions of payments-service showing evolution | `bundles/payments-service/v{1.0.0,1.1.0,1.2.0,2.0.0}/` |
| 4 | **Meaningful diff** -- API + config + deps + runtime changes | `pacto diff bundles/payments-service/v1.2.0 bundles/payments-service/v2.0.0` |
| 5 | **Interface diversity** -- HTTP, gRPC, and event-driven | auth-service (gRPC), payments-service (HTTP + events), notification-worker (events) |
| 6 | **Visibility** -- public vs internal | frontend/api-gateway (public) vs orders-service/auth-service (internal) |
| 7 | **State diversity** -- stateless, stateful, hybrid | fraud-service (stateless), postgresql (stateful), auth-service (hybrid) |
| 8 | **Shared configuration schemas** -- services share platform-level config properties | orders-service, frontend, api-gateway include platform-app-config properties in local schemas |
| 9 | **Reusable policy** via `policy.ref` | Most services ref `platform-http-policy`; api-gateway uses local `policy.schema` |
| 10 | **Workload diversity** -- service vs scheduled | notification-worker is `scheduled`, everything else is `service` |

---

## Breaking Change Detection

The **payments-service** is the hero of this demo. It evolves through 4 versions:

| Version | What happens |
|---------|-------------|
| **v1.0.0** | `POST /charges`, `POST /refunds`, events: `payment.completed`, `payment.refunded` |
| **v1.1.0** | Adds `POST /webhooks/stripe`, `payment.failed` event, optional `WEBHOOK_SECRET` |
| **v1.2.0** | Optional `fraud-service` dependency, optional `risk_score` field, `FRAUD_CHECK_ENABLED` |
| **v2.0.0** | **BREAKING**: removes `/charges`, introduces `/payment-intents`, renames `STRIPE_API_KEY` to `STRIPE_SECRET_KEY`, `WEBHOOK_SECRET` now required, `fraud-service` now required |

Diff v1.2.0 against v2.0.0 to see every breaking change detected:

```bash
pacto diff \
  oci://ghcr.io/trianalab/pacto-demo/payments-service:1.2.0 \
  oci://ghcr.io/trianalab/pacto-demo/payments-service:2.0.0
```

Changes detected:
- **API**: `/charges` removed, `/payment-intents` introduced
- **Config**: `STRIPE_API_KEY` renamed, `WEBHOOK_SECRET` becomes required, `ENABLE_REFUNDS` removed
- **Dependencies**: `fraud-service` changes from optional to required
- **Runtime**: upgrade strategy changes from `rolling` to `recreate`
- **Events**: `payment.completed`/`payment.failed` replaced by `payment.intent.*`

---

## Architecture

### Tiers

| Tier | Services |
|------|----------|
| **Edge** | `frontend` (Next.js, public), `api-gateway` (Go, public, local policy) |
| **Domain** | `auth-service` (gRPC), `orders-service` (HTTP + events), `payments-service` (HTTP + events), `fraud-service` (gRPC), `notification-worker` (scheduled, events) |
| **Infra** | `postgresql`, `redis` |
| **External** | `stripe-api`, `email-provider` |
| **Platform** | `platform-http-policy`, `platform-app-config`, `platform-worker-config` |

### Configuration Patterns

All services use local configuration schemas. Services that share platform-level config include those properties alongside service-specific ones:

**Domain-specific schema** (payments-service v2.0.0):
```yaml
configuration:
  schema: configuration/schema.json
  values:
    STRIPE_SECRET_KEY: "${STRIPE_SECRET_KEY}"
    PAYMENTS_DB_URL: postgresql://postgres:5432/payments
```

**Platform + domain schema** (orders-service):
```yaml
configuration:
  schema: configuration/schema.json  # includes platform-app-config properties
  values:
    OTEL_SERVICE_NAME: orders-service
    DATABASE_URL: postgresql://postgres:5432/orders
```

### Policy Patterns

**Ref-based** (most services):
```yaml
policy:
  ref: oci://ghcr.io/trianalab/pacto-demo/platform-http-policy
```

**Local schema** (api-gateway):
```yaml
policy:
  schema: policy/schema.json
```

---

## Bundle Structure

Each service produces a self-contained bundle:

```
bundles/payments-service/v2.0.0/
  pacto.yaml                     # contract
  interfaces/
    openapi.json                 # OpenAPI 3.1 spec
    events.json                  # AsyncAPI 2.6 spec
  configuration/
    schema.json                  # JSON Schema for config
  docs/
    overview.md                  # human-readable docs
```

---

## More CLI Examples

```bash
# resolve the full dependency graph from the root
pacto graph oci://ghcr.io/trianalab/pacto-demo/pacto-demo

# inspect any service
pacto explain oci://ghcr.io/trianalab/pacto-demo/payments-service:2.0.0

# validate a contract
pacto validate oci://ghcr.io/trianalab/pacto-demo/auth-service

# generate documentation
pacto doc oci://ghcr.io/trianalab/pacto-demo/orders-service:1.1.0
```

### Local Development

```bash
# validate all bundles
make validate

# diff payments-service breaking change
make breaking-change

# diff non-breaking evolution
make evolution

# launch the interactive dashboard
make dashboard

# package all bundles
make pack
```

---

## CI Integration

Every change is validated before it reaches production. This repository runs the full Pacto pipeline using [pacto-actions](https://github.com/trianalab/pacto-actions).

| Capability | Workflow | What it demonstrates |
|------------|----------|----------------------|
| Validation | [Validate & Explain](../../actions/workflows/demo-validate.yml) | `pacto validate` + `pacto explain` on all 20 contracts |
| Breaking changes | [Breaking Change Detection](../../actions/workflows/demo-breaking-change.yml) | payments-service v1.2.0 vs v2.0.0 breaking change diff |
| Documentation | [Contract Documentation](../../actions/workflows/demo-docs.yml) | `pacto doc` generates docs for every contract |
| Packaging | [Pack Contract Bundles](../../actions/workflows/demo-pack.yml) | `pacto pack` creates OCI-ready bundles |
| Full CI | [Pacto CI](../../actions/workflows/ci-pacto.yml) | Validate, diff, document, and push to GHCR |

---

## License

MIT
