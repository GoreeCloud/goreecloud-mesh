# Mesh Platform Registry

The Mesh Platform Registry is the normalized read model for GoreeCloud platform declarations and computed conformance evidence. It is an aggregation layer, not a replacement authority.

## Source contract

Each component owns `goreecloud.platform.yaml` in its canonical repository. Repository CI validates that declaration against a pinned immutable revision of the canonical Platform Contract in `GoreeCloud/GoreeCloud` and computes a revision-bound conformance result.

Mesh records a minimized normalized representation using `contracts/mesh.platform-record.v1.schema.json`. Platform Record v1 consumes the current Platform Contract v0.2 vocabulary. Every record carries:

- the producer repository and exact 40-character producer revision;
- Platform Contract schema `0.2`;
- `authority_transfer: false`;
- the v0.2 component ID, type, version, lifecycle, and supported-platform vocabulary;
- declared capabilities, dependencies, and explicit relationships;
- all seven Platform System results using the exact v0.2 machine states;
- runtime, health, and readiness state where known;
- separate backup and verified-restore state;
- portability state;
- declared and computed conformance state;
- the exact `GoreeCloud/GoreeCloud` evaluator revision and evaluation time;
- Stable eligibility, blockers, and missing-evidence identifiers; and
- bounded evidence references rather than evidence payload duplication.

## Authority boundary

Mesh never upgrades producer truth. A valid transport or registry record does not itself prove that a component is secure, privacy-compliant, recoverable, accessible, healthy, Glaze-conformant, Identity-integrated, or Stable-eligible. Those assertions remain owned by their authoritative producers and acceptance evidence.

The registry rejects records that request authority transfer, whose source repository does not exactly match the component repository, whose source revision is not an exact Git revision, whose Platform Contract vocabulary is not v0.2, whose evaluator provenance is not bound to `GoreeCloud/GoreeCloud`, or that label a nonconformant/unverified record Stable-eligible.

Mesh rejects unknown Platform System result vocabulary instead of silently accepting display labels or legacy states. This keeps coordination records machine-comparable with the canonical contract without transferring authority to Mesh.

Authenticated writes add a transport identity binding without transferring authority. The verified GoreeCloud Identity service principal must have `mesh.platform-registry.write`, and its `service_id` must exactly match the record's `component.id`. This prevents one scoped component from overwriting another component's platform declaration. Authentication proves who delivered the record; it does not independently prove the producer-domain assertions inside the record.

## Monotonic update and replay boundary

An authenticated producer may retry the exact same record idempotently, but it may not move accepted platform state backward.

Mesh keeps two independent monotonic clocks for each component:

- producer-domain facts are bound to `observed_at`; and
- canonical conformance facts are bound to `conformance.evaluated_at`.

An update is rejected when either timestamp is older than the currently stored value. Producer-owned state changes require a strictly newer `observed_at`. Canonical conformance changes require a strictly newer `evaluated_at`. This means the canonical evaluator can legitimately publish a later evaluation for the same producer observation, and a producer can publish a newer runtime/health observation without manufacturing a new canonical evaluation.

A same-time mutation that changes producer or conformance content is rejected rather than being treated as an idempotent retry. The authenticated HTTP API returns `409 Conflict` for these regressive or ambiguous updates and leaves the stored record unchanged.

This is a registry ordering rule, not a truth upgrade. Newer evidence may still be blocked, nonconformant, unhealthy, unverified, or otherwise unfavorable and must be preserved exactly as supplied by the appropriate authority.

## Durable state

`internal/platformregistry.NewPersistent` stores the normalized registry in `goreecloud.mesh.platform-registry-state.v1` JSON. The runtime default is `./mesh-platform-registry.json`; an empty `--platform-registry` path keeps the registry in memory.

Durable state is fail-closed:

- the state file must be a regular non-symlink file;
- existing state must not be accessible to group or other users;
- files are atomically replaced with mode `0600`;
- persisted records are revalidated on load;
- duplicate component IDs and unsupported state schemas are rejected; and
- state size is bounded.

Only normalized coordination metadata is stored. Credentials, bearer tokens, private keys, application payloads, raw user activity, browsing history, DNS history, and other producer-private data do not belong in this state file.

## Authenticated HTTP API

The private-first `/v1/` API exposes:

- `GET /v1/platform-registry` — list normalized records; requires `mesh.platform-registry.read`;
- `POST /v1/platform-registry` — validate and persist a producer-bound, monotonic record; requires `mesh.platform-registry.write` and exact service/component identity binding;
- `GET /v1/platform-registry/{id}` — read one record; requires `mesh.platform-registry.read`; and
- `GET /v1/platform-registry/{id}/dependents` — compute direct and transitive dependents from declared dependencies/required relationships; requires `mesh.platform-registry.read`.

When no GoreeCloud Identity verifier is configured, all of these endpoints fail closed with authentication unavailable. The Mesh process remains loopback-only by default; adding these source-level endpoints does not authorize public or production exposure.

The read scope is intentionally distinct from write scope so Manager and other authorized consumers can aggregate platform state without gaining producer-write authority.

## Recovery semantics

`backup_status` and `restore_status` are separate. Setting a backup to `verified` never modifies restore state. A `verified` restore requires a concrete `last_verified_restore` timestamp, and the timestamp is forbidden when restore state is not verified.

This makes the later Manager Continuity Health question — “If this server disappeared today, could GoreeCloud be restored?” — answerable from restoration evidence rather than backup-job optimism.

## Conformance provenance

The registry preserves both the repository's declared conformance and the canonical evaluator's computed conformance. It also carries the exact evaluator revision and evaluation time. Manager and other consumers may present these facts, but they must not reinterpret a failed, blocked, nonconformant, unverified, or stale result as a stronger state.

A component with lifecycle `stable` is rejected unless the computed result is `conformant` and `stable_eligible` is true.

## Dependency graph

The registry computes dependents only from explicitly declared dependencies and required relationships. It does not infer relationships from naming, network adjacency, UI placement, or shared infrastructure.

## Current implementation boundary

The stacked platform-registry API candidate provides validated aggregation, owner-only atomic JSON persistence, authenticated/scoped ingestion, replay-safe monotonic updates, authenticated reads, and dependency-impact queries on top of the v0.2 record contract. It does **not** establish deployed GoreeCloud Identity acceptance, Wardveil runtime acceptance, Privacy Shield runtime acceptance, Everkeep backup/restore acceptance, production publication, Manager consumption, release acceptance, or Stable qualification. Those remain separate evidence-backed gates.
