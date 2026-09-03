# Mesh Platform Registry

The Mesh Platform Registry is the normalized read model for GoreeCloud platform declarations and computed conformance evidence. It is an aggregation layer, not a replacement authority.

## Source contract

Each component owns `goreecloud.platform.yaml` in its canonical repository. Repository CI validates that declaration against a pinned revision of the canonical contract in `GoreeCloud/GoreeCloud` and computes a conformance result.

Mesh records a minimized normalized representation using `contracts/mesh.platform-record.v1.schema.json`. Every record carries:

- the producer repository;
- an immutable producer revision;
- the platform-contract schema version;
- `authority_transfer: false`;
- component identity/version/lifecycle/platforms;
- declared capabilities and dependencies;
- explicit relationships;
- all seven Platform System evaluation states;
- runtime/health/readiness state where known;
- separate backup and verified-restore state;
- portability state;
- computed conformance and missing-evidence identifiers; and
- bounded evidence references rather than evidence payload duplication.

## Authority boundary

Mesh never upgrades producer truth. A valid transport or registry record does not itself prove that a component is secure, privacy-compliant, recoverable, accessible, healthy, Glaze-conformant, Identity-integrated, or Stable-eligible. Those assertions remain owned by their authoritative producers and acceptance evidence.

The registry rejects records that request authority transfer, whose source repository does not exactly match the component repository, whose source revision is not immutable-looking, or that label a non-conformant record Stable-eligible.

Authenticated writes add a transport identity binding without transferring authority. The verified GoreeCloud Identity service principal must have `mesh.platform-registry.write`, and its `service_id` must exactly match the record's `component.id`. This prevents one scoped component from overwriting another component's platform declaration. Authentication proves who delivered the record; it does not independently prove the producer-domain assertions inside the record.

## Durable state

`internal/platformregistry.NewPersistent` stores the normalized registry in `goreecloud.mesh.platform-registry-state.v1` JSON. The runtime default is `./mesh-platform-registry.json`; an empty `--platform-registry` path keeps the registry in memory.

Durable state is fail-closed:

- the state file must be a regular non-symlink file;
- existing state must not be accessible to group or other users;
- files are atomically replaced with mode `0600`;
- persisted records are revalidated on load;
- duplicate component IDs and unsupported state schemas are rejected; and
- state size is bounded.

Only normalized coordination metadata is stored. Credentials, bearer tokens, private keys, application payloads, and other producer-private data do not belong in this state file.

## Authenticated HTTP API

The private-first `/v1/` API exposes:

- `GET /v1/platform-registry` — list normalized records; requires `mesh.platform-registry.read`;
- `POST /v1/platform-registry` — validate and persist a producer-bound record; requires `mesh.platform-registry.write` and exact service/component identity binding;
- `GET /v1/platform-registry/{id}` — read one record; requires `mesh.platform-registry.read`; and
- `GET /v1/platform-registry/{id}/dependents` — compute direct and transitive dependents from declared dependencies/required relationships; requires `mesh.platform-registry.read`.

When no GoreeCloud Identity verifier is configured, all of these endpoints fail closed with authentication unavailable. The Mesh process remains loopback-only by default; adding these source-level endpoints does not authorize public or production exposure.

## Recovery semantics

`backup_status` and `restore_status` are separate. Setting a backup to `verified` never modifies restore state. A `verified` restore requires a concrete `last_verified_restore` timestamp, and the timestamp is forbidden when restore state is not verified.

This makes the later Manager Continuity Health question — “If this server disappeared today, could GoreeCloud be restored?” — answerable from restoration evidence rather than backup-job optimism.

## Dependency graph

The registry computes dependents only from explicitly declared dependencies and required relationships. It does not infer relationships from naming, network adjacency, UI placement, or shared infrastructure.

## Current implementation boundary

The stacked platform-registry API candidate provides validated aggregation, owner-only atomic JSON persistence, authenticated/scoped ingestion, authenticated reads, and dependency-impact queries. It does **not** establish deployed GoreeCloud Identity acceptance, Wardveil runtime acceptance, Privacy Shield runtime acceptance, Everkeep backup/restore acceptance, production publication, Manager consumption, release acceptance, or Stable qualification. Those remain separate evidence-backed gates.
