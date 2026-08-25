# GoreeCloud Mesh

GoreeCloud Mesh is the application and service coordination fabric of the GoreeCloud ecosystem. It provides a native GoreeCloud-owned operational layer for service and capability discovery, relationship and dependency modeling, policy-aware coordination, lifecycle state, events, health-aware impact analysis, and application interoperability.

Mesh does not replace GoreeCloud Manager, Monitor, Network, Gateway, Identity, Glaze UI, Wardveil Security, Privacy Shield, or Everkeep. It coordinates application- and service-level relationships across those specialized systems while preserving their authority boundaries.

## Foundation scope

The initial native foundation provides:

- **Mesh Registry** — durable registry of applications, services, capabilities, dependencies, endpoints, health, and platform-conformance state.
- **Mesh Graph** — explicit service relationships and transitive dependency-impact analysis.
- **Mesh Policy** — fail-closed evaluation of registered relationships and requested capabilities.
- **Mesh Events** — bounded in-process event publication for registry and relationship lifecycle changes.
- **Mesh Nodes** — service/node identity and operational-state records through the registry model.
- **Mesh Connections** — explicit, versionable relationships between registered components.
- **Mesh Platform Catalog** — machine-readable authority and contract-source metadata for Glaze UI, Wardveil Security, Privacy Shield, and Everkeep without transferring those systems' authority into Mesh.
- **Mesh Platform Status** — read-only joined authority/evidence status that makes missing and non-validated platform evidence explicit and fail-closed.
- **Mesh Source Attestations** — durable exact-source provenance for reviewed platform integration manifests, kept independent from runtime and Stable acceptance.
- **Mesh Runtime Evidence** — durable bounded runtime contract evidence for mandatory integral platform systems, kept separate from source provenance and still subject to fail-closed Stable eligibility rules.
- **Mesh API** — private-first HTTP interface for discovery, registration, graph inspection, platform catalog/status inspection, source-provenance inspection, runtime-evidence inspection, and policy evaluation.
- **Mesh Console** — planned Glaze UI administrative experience; not yet represented as complete.

## Architecture principles

- Native GoreeCloud implementation from the ground up.
- Stable, documented APIs instead of direct database or filesystem coupling.
- Each participating application remains authoritative for its own data and business rules.
- Least privilege and explicit relationships rather than ambient trust.
- Health-aware failure isolation and dependency-impact visibility.
- Privacy-conscious metadata: Mesh records only coordination information necessary to operate relationships.
- Durable, atomic local state with no external database dependency in the first milestone.
- Source provenance remains separate from runtime evidence and cannot satisfy Stable gates by itself.
- Runtime evidence persistence does not create or infer runtime acceptance; evidence states remain explicit and fail closed.
- Mandatory Glaze UI, Wardveil Security, Privacy Shield, and Everkeep contracts are tracked as release gates; this foundation does **not** claim Stable conformance yet.

## Run

```bash
go run ./cmd/mesh \
  -listen 127.0.0.1:8787 \
  -state ./mesh-state.json \
  -source-attestations ./mesh-source-attestations.json \
  -runtime-evidence ./mesh-runtime-evidence.json
```

The default listen address is loopback-only. Source attestations and runtime contract evidence are persisted atomically to separate restrictive-permission JSON state files by default. Passing an empty value for either persistence flag disables that store for test or ephemeral operation. Public exposure is not part of this milestone.

## API

Key initial endpoints:

- `GET /healthz`
- `GET /v1/state`
- `GET /v1/services`
- `POST /v1/services`
- `GET /v1/services/{id}`
- `GET /v1/capabilities/{capability}`
- `POST /v1/relationships`
- `GET /v1/graph/impact?id=<service-id>`
- `POST /v1/evaluate`
- `GET /v1/platforms`
- `GET /v1/platforms/status`
- `GET /v1/platforms/source-attestations`
- `POST /v1/platforms/source-attestations`
- `GET /v1/contracts`
- `POST /v1/contracts`
- `GET /v1/contracts/stable-eligibility`

See [`docs/api.md`](docs/api.md) and [`docs/architecture.md`](docs/architecture.md).

## Validation

```bash
gofmt -w .
go vet ./...
go test ./...
```

CI runs formatting validation, vetting, tests, and the platform-conformance contract validator.

## Release boundary

This repository is in **Development**. The native coordination core is an initial source foundation only. Production deployment, public/private DNS publication, Gateway routing, Identity-backed authentication, external event delivery, Monitoring adapters, persistent multi-node coordination, Mesh Console completion, and runtime acceptance of Glaze UI, Wardveil Security, Privacy Shield, and Everkeep remain separate acceptance gates.

## License

AGPL-3.0-only. See [`LICENSE`](LICENSE).
