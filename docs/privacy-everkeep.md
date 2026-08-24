# Privacy Shield and Everkeep Governance

GoreeCloud Mesh coordinates application and service relationships while minimizing retained data and preserving recoverability of Mesh-owned state. Privacy Shield and Everkeep remain separate platform authorities; this source increment establishes the contracts Mesh must satisfy without claiming production acceptance.

## Privacy Shield boundary

Mesh governance records are limited to coordination metadata, operational evidence, and narrowly justified sensitive metadata. Application payload content does not belong in Mesh governance records. Message bodies, user files, notes, browsing content, reusable credentials, authentication tokens, private keys, recovery secrets, and equivalent application-owned content must remain outside this model.

Each governed record identifies its purpose, data class, retention class, whether it is exportable, and—when retention is bounded—a positive maximum age. Retention states are `ephemeral`, `bounded`, and `preserved`. A bounded record without an explicit maximum age fails validation.

This model is a source-level privacy contract. Actual retention enforcement, erasure workflows, user-facing privacy controls, telemetry acceptance, and Privacy Shield conformance evidence remain separate implementation and acceptance work.

## Everkeep boundary

Everkeep governs resilience, backup, restore, portability, preservation, and recovery evidence. Mesh defines four source-level recovery capabilities that must be evidenced before recovery readiness can be considered satisfied: export, backup, restore, and verification.

Recovery evidence records capability, state, source, optional revision, and observation time. Recovery readiness fails closed unless validated evidence exists for all four required capabilities. A true source-level result does not prove that a production backup exists or that a target-environment restore has succeeded.

## Authority and lifecycle

Privacy Shield remains authoritative for privacy-control contracts and data minimization. Everkeep remains authoritative for resilience and preservation. Mesh consumes these contracts and records bounded evidence; it does not supersede either system.

This increment does not create production retention jobs, backup schedules, backup repositories, encryption keys, recovery credentials, export destinations, restore jobs, or production acceptance. GoreeCloud Mesh remains in Development until the applicable runtime integrations and acceptance gates are completed.
