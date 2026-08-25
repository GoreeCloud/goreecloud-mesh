# Privacy Shield and Everkeep Governance

GoreeCloud Mesh coordinates application and service relationships while minimizing retained data and preserving recoverability of Mesh-owned state. Privacy Shield and Everkeep remain separate platform authorities; this source increment establishes the contracts Mesh must satisfy without claiming production acceptance.

## Privacy Shield boundary

Mesh governance records are limited to coordination metadata, operational evidence, and narrowly justified sensitive metadata. Application payload content does not belong in Mesh governance records. Message bodies, user files, notes, browsing content, reusable credentials, authentication tokens, private keys, recovery secrets, and equivalent application-owned content must remain outside this model.

Each governed record identifies its purpose, data class, retention class, whether it is exportable, and—when retention is bounded—a positive maximum age. Retention states are `ephemeral`, `bounded`, and `preserved`. A bounded record without an explicit maximum age fails validation.

This model is a source-level privacy contract. Actual retention enforcement, erasure workflows, user-facing privacy controls, telemetry acceptance, and Privacy Shield conformance evidence remain separate implementation and acceptance work.

## Everkeep boundary

Everkeep governs resilience, backup, restore, portability, preservation, and recovery evidence. Mesh recovery evidence now uses the same five required dimensions declared by `docs/everkeep.acceptance.json`: `backup_coverage`, `restore_capability`, `portability`, `documentation`, and `provenance`.

Each recovery-evidence record identifies its canonical dimension, state, authoritative source, exact lowercase 40-character source revision, observation time, and producer-defined validity boundary. Evidence with an unknown dimension or state, missing source, non-exact revision, future observation time, missing validity, non-forward validity, or expired validity fails closed for readiness.

Recovery readiness requires current validated evidence for all five dimensions. Missing, degraded, unknown, malformed, or expired evidence cannot satisfy the source-level gate. Mesh does not define a universal freshness duration; `valid_until` remains supplied by the authoritative producer or applicable policy.

A true source-level readiness result still does not prove that a production backup exists, that a target-environment restore has succeeded, or that Everkeep runtime acceptance has been granted. `docs/everkeep.acceptance.json` remains explicitly non-ready until target-runtime and exact-revision acceptance are completed.

## Authority and lifecycle

Privacy Shield remains authoritative for privacy-control contracts and data minimization. Everkeep remains authoritative for resilience and preservation. Mesh consumes these contracts and records bounded evidence; it does not supersede either system.

This increment does not create production retention jobs, backup schedules, backup repositories, encryption keys, recovery credentials, export destinations, restore jobs, or production acceptance. GoreeCloud Mesh remains in Development until the applicable runtime integrations and acceptance gates are completed.
