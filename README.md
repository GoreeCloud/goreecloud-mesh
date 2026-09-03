# GoreeCloud Mesh

GoreeCloud Mesh is the application and service coordination fabric of the GoreeCloud ecosystem. It provides a native GoreeCloud-owned operational layer for service and capability discovery, relationship and dependency modeling, policy-aware coordination, lifecycle state, events, health-aware impact analysis, evidence transport, and application interoperability.

Mesh does not replace GoreeCloud Manager, Monitor, Network, Gateway, GoreeCloud Identity, Glaze UI, Wardveil Security, Privacy Shield, or Everkeep. It coordinates application- and service-level relationships across those specialized systems while preserving their authority boundaries.

## Foundation scope

The native foundation provides:

- **Mesh Registry** — durable registry of applications, services, capabilities, dependencies, endpoints, health, and platform-conformance state.
- **Mesh Graph** — explicit service relationships and transitive dependency-impact analysis.
- **Mesh Policy** — fail-closed evaluation of registered relationships and requested capabilities.
- **Mesh Events** — bounded in-process event publication for registry and relationship lifecycle changes.
- **Mesh Nodes and Connections** — service/node identity, operational state, and explicit versionable relationships.
- **Mesh Platform Catalog and Status** — authority/contract metadata and joined integration status for GoreeCloud Identity, Glaze UI, Wardveil Security, Privacy Shield, and Everkeep without transferring their authority into Mesh.
- **Mesh Source Attestations** — durable exact-source provenance kept independent from runtime and Stable acceptance.
- **Mesh Runtime Evidence** — bounded runtime contract evidence bound to canonical repositories, contracts, exact revisions, and observation time.
- **Mesh Evidence Envelopes** — durable immutable transport records for minimized producer-authoritative evidence with canonical producer/contract/authority validation, freshness, scoped subjects, minimization flags, and optional digest binding.
- **Authenticated Evidence Delivery** — `mesh.evidence.write` plus exact producer-service identity binding for ingestion, and `mesh.evidence.read` for inspection/consumer views.
- **Evidence Subject Views** — producer/authority/assertion-separated consumer models that expose latest/current evidence without manufacturing an overall security, privacy, recovery, continuity, identity, authorization, or conformance verdict.
- **Mesh API** — private-first HTTP interface for discovery, registration, graph inspection, platform evidence, producer evidence, and policy evaluation.
- **Mesh Center** — planned Glaze UI administrative experience; not yet represented as complete.

## Architecture principles

- Native GoreeCloud implementation from the ground up.
- Stable, documented APIs instead of direct database or filesystem coupling.
- Each participating application remains authoritative for its own data and business rules.
- Least privilege and explicit relationships rather than ambient trust.
- Privacy-conscious metadata: Mesh records only coordination and minimized evidence information necessary to operate relationships and evidence transport.
- Durable, atomic local state with no external database dependency in the first milestone.
- Source provenance remains separate from runtime evidence and cannot satisfy Stable gates by itself.
- Evidence transport validity never creates or upgrades security, privacy, recovery, continuity, identity, authentication, authorization, or design-conformance truth.
- A write scope is insufficient to impersonate another producer: the verified service identity must match the envelope producer.
- Authentication proves the bound producer identity and granted Mesh scope; it does not prove the producer-domain assertion carried by an evidence envelope.
- Expired evidence remains auditable but cannot satisfy current-evidence queries; normal expiry must not prevent restart.
- Required GoreeCloud Manager, GoreeCloud Identity, Glaze UI, Wardveil Security, Privacy Shield, and Everkeep platform relationships or contracts remain release gates where applicable; this Development foundation does **not** claim Stable conformance.

## Integral platform authority model

Mesh is the coordination authority within the seven-system GoreeCloud Integral Platform System model. GoreeCloud Manager remains the central administration and operational-management authority; GoreeCloud Identity remains the identity/authentication/authorization authority; Privacy Shield remains the privacy and data-use authority; Wardveil Security remains the security evaluation and protection authority; Everkeep remains the continuity/recovery/preservation authority; and Glaze UI remains the presentation and design-conformance authority. Mesh may coordinate relationships and validate or transport bounded evidence from applicable systems, but it cannot manufacture, transfer, or upgrade their domain truth.

The platform evidence-plane contract is `contracts/mesh.platform-evidence-plane.v1.json`. It binds each evidence producer to its canonical repository and producer-owned Mesh evidence profile while keeping `authority_transfer` false. GoreeCloud Manager is an administrative consumer/authority relationship rather than an evidence producer in that five-producer evidence-plane contract.

The repository-root `goreecloud.platform.yaml` is the current platform-wide machine-readable declaration for Mesh under GoreeCloud Platform Contract v0.1. `docs/platform-conformance.json` remains supplemental Mesh-specific conformance evidence and acceptance detail; it must not compete with or silently weaken the root Platform Contract declaration.

## Run

```bash
go run ./cmd/mesh \
  -listen 127.0.0.1:8787 \
  -state ./mesh-state.json \
  -source-attestations ./mesh-source-attestations.json \
  -runtime-evidence ./mesh-runtime-evidence.json \
  -evidence-envelopes ./mesh-evidence-envelopes.json \
  -everkeep-recovery-evidence ./mesh-everkeep-recovery-evidence.json
```

The default listen address is loopback-only. Persistent stores are written atomically to separate restrictive-permission JSON files. Passing an empty persistence path disables that store for test or ephemeral operation.

## Evidence API

Authenticated evidence routes include:

- `GET /v1/evidence/envelopes` — requires `mesh.evidence.read`.
- `POST /v1/evidence/envelopes` — requires `mesh.evidence.write` and a producer-matching verified service identity.
- `GET /v1/evidence/envelopes/{id}` — requires `mesh.evidence.read`.
- `GET /v1/evidence/status` — requires `mesh.evidence.read`.
- `GET /v1/evidence/subjects/{kind}/{id}` — requires `mesh.evidence.read` and returns the authority-separated consumer view.

A first accepted envelope returns `201`; an exact immutable replay returns `200` with `replayed: true`.

See [`docs/api.md`](docs/api.md), [`docs/evidence-envelope.md`](docs/evidence-envelope.md), [`docs/evidence-delivery.md`](docs/evidence-delivery.md), and [`docs/architecture.md`](docs/architecture.md).

## Validation

```bash
gofmt -w .
go vet ./...
go test ./...
```

CI runs formatting validation, vetting, tests, build validation, and platform contract checks.

## Release boundary

This repository is in **Development**. The source now contains the durable evidence registry, authenticated read/write enforcement, producer-bound delivery receipts, consumer subject views, and reference integration contracts. It does not independently establish a deployed GoreeCloud Identity verifier, production Gateway/Network/TLS routing, DNS publication, multi-node replication, target-environment producer delivery, Mesh Center completion, or production/Stable acceptance of the integral platform systems.

## License

AGPL-3.0-only. See [`LICENSE`](LICENSE).
