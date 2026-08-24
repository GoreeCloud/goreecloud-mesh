# Runtime Platform Contracts

GoreeCloud Mesh coordinates applications and services while preserving the authority of the four mandatory integral platform systems. Runtime conformance is evidence-backed and fail-closed.

## Mandatory contracts

- **Glaze UI** — presentation and interaction surfaces must identify the accepted Glaze UI contract used by a Mesh-facing console or client.
- **Wardveil Security** — security state and protection integrations must identify the Wardveil contract and evidence source used for the runtime.
- **Privacy Shield** — coordination metadata, retention, minimization, and application-adapter behavior must identify the accepted Privacy Shield contract.
- **Everkeep** — export, backup, restore, recovery, preservation, and portability behavior must identify the accepted Everkeep contract.

## Evidence model

Runtime evidence records only coordination metadata: platform, contract identifier, state, evidence source, revision, observation time, and a bounded detail field. Reusable credentials, private keys, tokens, recovery secrets, and private application content do not belong in contract evidence.

States are `pending`, `validated`, or `blocked`. Unknown systems and unknown states are rejected.

## Stable gate

Mesh is not Stable-eligible unless validated runtime evidence exists for all four mandatory systems. Missing evidence, pending evidence, or blocked evidence fails closed. Source-level declarations in `platform-conformance.json` describe requirements; they do not substitute for runtime evidence or production acceptance.

## Authority boundary

The registry records evidence; it does not replace the specialized platform systems. Glaze UI remains the design authority, Wardveil Security remains the security/protection authority, Privacy Shield remains the privacy-control authority, and Everkeep remains the resilience/preservation authority.
