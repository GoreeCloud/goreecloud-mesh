# GoreeCloud Mesh API

The initial Mesh API is a private-first JSON HTTP contract. It does not imply public exposure or production authorization.

## Common behavior

- Request and response content uses JSON unless an endpoint explicitly defines another media type, such as the bounded event stream.
- Responses include `Cache-Control: no-store` and `X-Content-Type-Options: nosniff`.
- JSON request bodies are bounded to 1 MiB and reject unknown fields.
- Identifiers are caller-supplied stable strings in the first milestone.
- Protected routes use the GoreeCloud Identity verifier boundary and dedicated least-privilege scopes. When a verifier is unavailable, protected routes fail closed with `401`; a verified principal without the required scope receives `403`.
- Event consumption requires `mesh.events.read`; evidence read routes require `mesh.evidence.read`; evidence delivery requires `mesh.evidence.write` plus producer-service identity binding.
- Other read-only inspection routes remain private-first until their production authorization policies are separately accepted.

## Health

### `GET /healthz`

Returns process availability.

```json
{"status":"ok","service":"goreecloud-mesh"}
```

This is not a complete production-readiness or dependency-health signal.

## State

### `GET /v1/state`

Returns the current Registry and relationship state. This endpoint is intended for administrative inspection and testing.

## Services

### `GET /v1/services`

Lists registered services in deterministic ID order.

### `POST /v1/services`

Creates or replaces a service registration. Requires `mesh.services.write`.

Required fields: `id`, `name`, and `kind`. Supported health states are `unknown`, `healthy`, `degraded`, and `unavailable`.

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

## Live lifecycle events

### `GET /v1/events/stream`

Requires the dedicated read-only `mesh.events.read` scope and the existing GoreeCloud Identity verifier boundary.

The response media type is `text/event-stream`. This endpoint is a bounded live adapter over the process-local best-effort Mesh event bus; it is not a journal, queue, or replay API.

Required query input:

- one or more `type=<registered-event-type>` parameters. At least one explicit type is required so external consumers do not silently receive every lifecycle event.

Optional bounded transport controls:

- `buffer=<1..64>` — defaults to `8`;
- `window_seconds=<1..10>` — defaults to `5`.

Only `type`, `buffer`, and `window_seconds` are accepted. Unknown parameters fail closed. `Last-Event-ID`, cursor-style requests, and replay semantics are rejected. The stream intentionally emits no SSE `id:` field. The JSON event envelope still contains its process-local `evt-<sequence>` identifier, but that identifier is not a durable offset.

Each stream closes after its bounded window. Reconnecting creates a fresh live subscription; events may be missed between windows or when a subscriber buffer fills. No acknowledgement, retry, ordering, durability, cross-host continuity, or delivery guarantee is implied. See [`events.md`](events.md).

Authentication establishes the consumer identity and event-read authority only. It does not upgrade event payloads into security, privacy, recovery, health, conformance, or producer-domain truth.

## Integral platform catalog

### `GET /v1/platforms`

Returns the canonical Mesh-side catalog of mandatory integral platform systems and their authority boundaries.

### `GET /v1/platforms/status`

Returns a read-only joined source/runtime integration-status view. Missing runtime evidence is pending and fail-closed. This endpoint does not authorize release, deployment, production acceptance, or Stable promotion.

## Platform source attestations

### `GET /v1/platforms/source-attestations`

Returns source-provenance attestations and the current fail-closed completeness result for mandatory integral platform systems.

### `POST /v1/platforms/source-attestations`

Records or replaces one source attestation. Requires `mesh.attestations.write`.

Source attestations prove reviewed source provenance only. They do not imply runtime acceptance, production acceptance, deployment authorization, release authorization, or Stable qualification.

## Runtime platform contracts

### `GET /v1/contracts`

Returns mandatory integral platform systems and currently recorded runtime evidence entries.

### `POST /v1/contracts`

Records or replaces one bounded runtime evidence entry. Requires `mesh.contracts.write`.

### `GET /v1/contracts/stable-eligibility`

Returns the current fail-closed source-level Stable eligibility calculation. It is not production or Stable authorization.

## Producer evidence envelopes

The GoreeCloud Evidence Envelope v1 keeps domain truth owned by the producer while allowing Mesh to validate, persist, query, and transport bounded evidence metadata. See [`evidence-envelope.md`](evidence-envelope.md) and [`evidence-delivery.md`](evidence-delivery.md).

### `GET /v1/evidence/envelopes`

Requires `mesh.evidence.read`.

Lists immutable evidence envelopes in newest-observation-first order. Each returned item includes a derived `fresh` boolean evaluated against the producer-declared `valid_until` boundary.

Optional exact-match query filters:

- `current=true|false`
- `producer=<producer-id>`
- `authority_domain=<domain>`
- `subject_kind=<kind>`
- `subject_id=<id>`
- `assertion=<assertion>`

`fresh: true` is timing metadata only and is never a positive domain verdict.

### `POST /v1/evidence/envelopes`

Requires `mesh.evidence.write` and a verified producer service identity.

The authenticated `service_id` must exactly match `producer.system`. Scope possession alone cannot be used by one GoreeCloud service to submit another producer's evidence.

The envelope must satisfy `goreecloud.evidence-envelope.v1`, including canonical repository/contract ownership, exact Git revision, owned authority domain, scoped subject, producer-declared validity, minimized data class, and explicit false user-content/secret-material flags.

Evidence IDs are immutable:

- first acceptance returns `201` and `replayed: false`;
- an exact idempotent replay returns `200` and `replayed: true`;
- the same ID with different content fails closed.

The delivery receipt includes the accepted envelope, acceptance time, replay state, and authenticated producer service ID. It never includes the bearer credential.

### `GET /v1/evidence/envelopes/{id}`

Requires `mesh.evidence.read`.

Returns one immutable envelope and its current derived `fresh` state, or `404` when unknown. Expired evidence remains readable for audit/provenance history.

### `GET /v1/evidence/status`

Requires `mesh.evidence.read`.

Returns current/stale counts overall and by producer. This is transport/freshness status only, not security, privacy, recovery, continuity, or design-conformance truth.

### `GET /v1/evidence/subjects/{kind}/{id}`

Requires `mesh.evidence.read`.

Returns the consumer-oriented evidence view for one subject. An optional `scope=<scope>` query parameter narrows the view.

The response contains:

- `subject`
- `transport.state` (`available` for a successful Mesh read)
- current/stale envelope counts
- independent producer/authority groups
- assertion groups containing `latest`, optional `latest_current`, and `history_count`

Mesh preserves producer outcomes verbatim and deliberately does not return an overall security/privacy/recovery/conformance verdict. This is the read model consumed by the Glaze UI 1.6 Candidate Mesh evidence consumer.

## Everkeep recovery evidence

### `GET /v1/everkeep/recovery-evidence`

Returns the currently recorded Mesh-owned Everkeep recovery evidence and fail-closed recovery-readiness result.

Required dimensions are `backup_coverage`, `restore_capability`, `portability`, `documentation`, and `provenance`.

### `POST /v1/everkeep/recovery-evidence`

Requires `mesh.everkeep.recovery.write`.

The endpoint accepts metadata and evidence state only. Passwords, tokens, recovery codes, private keys, secret values, backup contents, and application payload content must not be recorded.

### `GET /v1/everkeep/recovery-readiness`

Returns fail-closed readiness across the canonical Everkeep application dimensions.

A `ready: true` response means only that the recorded source-level evidence set is complete, structurally valid, currently within producer-declared validity, and marked validated. It does not prove production restore success or Stable qualification.
