# GoreeCloud Mesh API

The initial Mesh API is a private-first JSON HTTP contract. It is intentionally small and does not imply public exposure or production authorization.

## Common behavior

- Request and response content uses JSON unless an endpoint has no body.
- Responses include `Cache-Control: no-store` and `X-Content-Type-Options: nosniff`.
- JSON request bodies are bounded to 1 MiB and reject unknown fields.
- Identifiers are caller-supplied stable strings in the first milestone.
- Mutating and policy-evaluation routes use the GoreeCloud Identity verifier boundary and dedicated least-privilege scopes. When a verifier is unavailable, protected routes fail closed with `401`; a verified principal without the required scope receives `403`.
- Read-only inspection routes remain private-first. Production authorization and endpoint exposure remain separate acceptance work.

## Health

### `GET /healthz`

Returns process availability.

```json
{"status":"ok","service":"goreecloud-mesh"}
```

This is not a complete production-readiness or dependency-health signal.

## State

### `GET /v1/state`

Returns the current Registry and relationship state.

This endpoint is intended for administrative inspection and testing. Future production authorization may restrict or replace broad state reads with narrower views.

## Services

### `GET /v1/services`

Lists registered services in deterministic ID order.

### `POST /v1/services`

Creates or replaces a service registration. Requires `mesh.services.write`.

Required fields:

- `id`
- `name`
- `kind`

Supported health states are `unknown`, `healthy`, `degraded`, and `unavailable`.

### `GET /v1/services/{id}`

Returns one registered service or `404` when it is unknown.

## Capability discovery

### `GET /v1/capabilities/{capability}`

Returns registered services that advertise the capability and are not explicitly unavailable.

Discovery is not authorization. A discovered target may still be denied by Mesh Policy, Identity, Gateway, Network, application permissions, or Wardveil controls.

## Relationships

### `POST /v1/relationships`

Creates or replaces an explicit relationship between two registered services. Requires `mesh.relationships.write`.

Self-relationships and references to unknown services are rejected.

## Dependency impact

### `GET /v1/graph/impact?id=<service-id>`

Returns the directly and transitively affected registered services whose declared dependency or enabled required relationship points to the selected service.

Impact is advisory coordination data. Specialized Monitoring, Manager, Gateway, Network, and application evidence remain authoritative for their own runtime state.

## Policy evaluation

### `POST /v1/evaluate`

Evaluates whether a registered source has an enabled relationship that authorizes use of a capability advertised by the registered target. Requires `mesh.policy.evaluate`.

The evaluator fails closed when the source or target is unknown, the target is unavailable, the capability is missing, the target does not advertise the capability, or no enabled relationship authorizes it.

Mesh Policy is one authorization input, not a substitute for application authorization, GoreeCloud Identity, Gateway controls, Network controls, or Wardveil Security.

## Integral platform catalog

### `GET /v1/platforms`

Returns the canonical Mesh-side catalog of mandatory integral platform systems. Each entry records the system identifier, display name, repository, authority boundary, expected contract source, integration-manifest path, and whether the system is required.

The catalog identifies Glaze UI as the design/interaction/accessibility authority, Wardveil Security as the security/protection authority, Privacy Shield as the privacy/data-minimization authority, and Everkeep as the resilience/recovery/preservation authority. Mesh does not acquire those authorities merely because it coordinates their evidence.

### `GET /v1/platforms/status`

Returns one read-only integration-status record for each mandatory platform system by joining the canonical authority catalog with the latest bounded runtime contract evidence recorded in Mesh.

The response also includes `source_attestations_validated`, which reports whether all four mandatory platform manifests have separately validated source attestations. Source provenance cannot satisfy runtime evidence requirements.

Missing runtime evidence is explicitly pending and fail-closed. This endpoint does not authorize release, deployment, production acceptance, or Stable promotion.

## Platform source attestations

### `GET /v1/platforms/source-attestations`

Returns the current source-provenance attestations and `all_validated`, a fail-closed completeness result for the four mandatory integral platform systems.

### `POST /v1/platforms/source-attestations`

Records or replaces the latest source attestation for one mandatory platform. Requires `mesh.attestations.write`. Requests fail closed on catalog mismatches, malformed provenance, invalid validation state, or implied runtime/Stable acceptance.

Source attestations contain source provenance only. They do not imply runtime acceptance, production acceptance, deployment authorization, release authorization, or Stable qualification.

## Runtime platform contracts

### `GET /v1/contracts`

Returns the four mandatory integral platform systems and the currently recorded runtime evidence entries.

### `POST /v1/contracts`

Records or replaces one bounded runtime evidence entry. Requires `mesh.contracts.write`. The contract and producer repository must match the canonical platform catalog. Validated evidence requires identifiable source provenance, an exact lowercase 40-character Git revision, and producer-defined validity.

Supported states are `pending`, `validated`, and `blocked`. Unknown platform systems, non-canonical contracts, invalid states, stale evidence, and untraceable validated evidence fail closed.

### `GET /v1/contracts/stable-eligibility`

Returns the current fail-closed source-level Stable eligibility calculation together with the mandatory-system list and recorded evidence. This calculation is not production or Stable authorization.

## Producer evidence envelopes

The GoreeCloud Evidence Envelope v1 is a transport/provenance contract. It keeps domain truth owned by the producer while allowing Mesh to persist, query, and transport bounded evidence metadata. See [`evidence-envelopes.md`](evidence-envelopes.md).

### `GET /v1/evidence/envelopes`

Lists immutable evidence envelopes in newest-observation-first order. Each returned item includes a derived `fresh` boolean evaluated against the producer-declared `valid_until` boundary.

Optional exact-match query filters:

- `current=true|false` — include only currently fresh or stale/expired records.
- `producer=<producer-id>` — for example `wardveil-security`, `privacy-shield`, `everkeep`, `glaze-ui`, or `goreecloud-mesh`.
- `authority_domain=<domain>` — for example `security`, `privacy`, `recovery`, `presentation`, or `governance`.
- `subject_kind=<kind>`
- `subject_id=<id>`
- `assertion=<assertion>`

The response reports counts for the filtered result. `fresh: true` means only that the envelope is inside the producer-declared evidence validity window. It is not a positive domain verdict.

### `POST /v1/evidence/envelopes`

Records one producer-authored evidence envelope. Requires `mesh.evidence.write`.

The request must satisfy `goreecloud.evidence-envelope.v1`, including:

- a canonical producer system and repository;
- an exact lowercase 40-character source revision;
- a contract belonging to that producer;
- an authority domain the producer actually owns;
- a scoped subject;
- assertion, outcome, source, observation time, and producer-declared validity window;
- an allowed minimized data class (`public`, `operational`, or `derived`);
- explicit `contains_user_content: false` and `contains_secret_material: false`;
- an optional `sha256:<64 lowercase hex>` payload digest.

Evidence IDs are immutable. Replaying the exact same envelope is idempotent. Reusing an existing ID with different content fails closed. Fresh evidence is required at write time; expired evidence cannot be inserted as new current evidence.

### `GET /v1/evidence/envelopes/{id}`

Returns one immutable envelope and its current derived `fresh` state, or `404` when unknown.

Expired evidence remains readable after its validity window for audit/provenance purposes. Expiration does not delete history and does not prevent Mesh from restarting.

### `GET /v1/evidence/status`

Returns current/stale counts overall and by producer.

This endpoint is evidence-transport and freshness status only. It must not be interpreted as a security, privacy, recovery, continuity, or design-conformance verdict.

## Everkeep recovery evidence

### `GET /v1/everkeep/recovery-evidence`

Returns the currently recorded Mesh-owned Everkeep recovery evidence, the canonical required dimensions, and the current fail-closed recovery-readiness result.

Required dimensions are:

- `backup_coverage`
- `restore_capability`
- `portability`
- `documentation`
- `provenance`

Every usable evidence item must carry the canonical dimension, a bounded state, an authoritative source, an exact lowercase 40-character source revision, an observation time, and a producer-defined `valid_until` boundary. Expired, future-dated, malformed, missing, degraded, or unknown evidence cannot satisfy recovery readiness.

### `POST /v1/everkeep/recovery-evidence`

Records or replaces evidence for one canonical recovery dimension. Requires `mesh.everkeep.recovery.write`.

The endpoint accepts metadata and evidence state only. Passwords, tokens, recovery codes, private keys, secret values, backup contents, and application payload content must not be recorded.

### `GET /v1/everkeep/recovery-readiness`

Returns the fail-closed readiness evaluation across all five canonical Everkeep application dimensions.

A `ready: true` response means only that the recorded source-level evidence set is complete, structurally valid, currently within producer-declared validity, and marked validated. It does not prove that a production backup exists, that a target-environment restore succeeded, that Everkeep runtime acceptance is complete, or that GoreeCloud Mesh qualifies for Stable release.
