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

## Recovery semantics

`backup_status` and `restore_status` are separate. Setting a backup to `verified` never modifies restore state. A `verified` restore requires a concrete `last_verified_restore` timestamp, and the timestamp is forbidden when restore state is not verified.

This makes the later Manager Continuity Health question — “If this server disappeared today, could GoreeCloud be restored?” — answerable from restoration evidence rather than backup-job optimism.

## Dependency graph

The registry computes dependents only from explicitly declared dependencies and required relationships. It does not infer relationships from naming, network adjacency, UI placement, or shared infrastructure.

## Current implementation boundary

`internal/platformregistry` provides validated in-memory aggregation and dependency-impact queries. Persistence, authenticated ingestion, API exposure, and Manager consumption remain separate integration steps and must preserve the same producer-authority boundary when implemented.
