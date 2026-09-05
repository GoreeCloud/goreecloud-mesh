# GoreeCloud Mesh Event Contract v1

## Purpose

GoreeCloud Mesh Events provides a bounded coordination signal for Registry and relationship lifecycle changes. The underlying bus remains deliberately **in-process, best-effort, non-durable, and non-replayable**. The source implementation also exposes a narrowly scoped GoreeCloud Identity-authenticated HTTP streaming adapter for external consumers; that adapter transports only the same closed lifecycle metadata and does not turn Mesh into a durable event broker or a store for application payloads, credentials, user activity, or producer-private data.

The machine-readable contract is `contracts/mesh.event.v1.schema.json` and the Go contract identifier is `goreecloud.mesh.event.v1`.

## Current event types

### `mesh.service.upserted.v1`

Emitted after a service registration or update is durably accepted by the Mesh Registry.

The event source and subject are the registered service ID. The payload is closed to:

- `health` — the Registry health value (`unknown`, `healthy`, `degraded`, or `unavailable`).

This health value is coordination context from the registered service record. It is not a substitute for GoreeCloud Monitoring evidence and does not transfer Monitoring authority into Mesh.

### `mesh.relationship.upserted.v1`

Emitted after a relationship registration or update is durably accepted.

The source is the relationship's source service, the subject is the relationship ID, and the payload is closed to:

- `target` — the target service ID; and
- `type` — the registered relationship type.

Capability, contract contents, credentials, policy decisions, request payloads, and application data are intentionally not copied into this lifecycle event.

## Event naming convention

Mesh-owned event type names use a lowercase dot-separated authority namespace and an explicit major version:

`mesh.<subject>[.<qualifier>...].<action>.v<major>`

Requirements:

- the first segment is `mesh` only for events whose coordination fact is produced by GoreeCloud Mesh;
- at least a subject and action segment appear between `mesh` and the version;
- non-version segments use lowercase ASCII letters, digits after the first character, and hyphens only;
- the version segment is `v` followed by a positive integer;
- a producer-specific domain event must not be renamed into the `mesh` namespace merely because Mesh transports it; producer authority remains producer-owned; and
- every active type must be explicitly registered in code, the JSON Schema enum, tests, and this contract documentation before use.

The Go validator enforces the Mesh-owned naming shape before type-specific validation. Unknown but well-shaped names still fail closed because naming conformity does not register an event type.

## Versioning rules

The envelope version and event-type version are separate compatibility dimensions:

- `schema: goreecloud.mesh.event.v1` versions the common envelope fields and their semantics;
- the `.v1` suffix on each current event type versions that type's subject/source meaning and closed `data` payload;
- breaking envelope changes require a new envelope schema major version;
- removing or renaming a payload field, changing a field's meaning/type, changing source/subject binding, or changing the authority represented by an existing event requires a new event-type major version;
- because payloads are closed, adding a field to an existing event type is treated as a compatibility change and must not be silently introduced under the same type identity;
- a new independent event type begins at `.v1` and may share the current envelope only when the envelope itself is unchanged; and
- old and new event-type versions may coexist only when their schemas and consumer semantics remain explicitly distinguishable.

A Git commit, release label, transport change, or implementation refactor does not itself change an event contract version.

## Schema registration and change control

`contracts/mesh.event.v1.schema.json` is the machine-readable source contract for the current envelope and registered event types. It is intentionally closed:

- top-level additional properties are rejected;
- `type` is an explicit enum;
- each current type has a closed type-specific `data` object;
- required envelope fields are explicit;
- `authority_transfer` is fixed to `false`; and
- field sizes and control-character restrictions are enforced again by the Go validator where applicable.

Regression tests verify that the JSON Schema identifier, closed-envelope setting, registered type enum, Go event constants, and `authority_transfer:false` rule remain aligned. A new type or version is incomplete until code, schema, tests, and documentation all agree on the same exact identity and payload.

## Envelope rules

Every current event contains:

- `schema` — exactly `goreecloud.mesh.event.v1`;
- `id` — a process-local monotonic `evt-<sequence>` identifier;
- `type` — one of the registered versioned event names;
- `source` — the coordination producer/source identifier;
- `subject` — the service or relationship identifier affected by the event;
- `data` — a type-specific closed payload;
- `created_at` — the UTC event time; and
- `authority_transfer` — always `false`.

The Go validator rejects unsupported schemas and event types, nonconforming Mesh-owned type names, unexpected payload fields, malformed field types, control characters, oversized identity/value fields, and any attempt to set `authority_transfer` to true.

## Publisher and consumer authorization requirements

### Current publishers

Current lifecycle events are emitted only by GoreeCloud Mesh after the corresponding Registry or relationship mutation has been accepted. There is no external arbitrary-event publication endpoint in this contract. An authenticated caller's ability to mutate another Mesh resource does not grant a separate right to publish fabricated lifecycle events.

External/domain event publication remains prohibited until a separately governed producer contract exists. Before activation, that future contract must require at minimum:

- a GoreeCloud Identity-authenticated workload principal;
- a dedicated write permission distinct from `mesh.events.read` and from unrelated Mesh mutation scopes;
- exact producer/source binding so one service cannot impersonate another producer;
- event-type ownership/allowlisting so a producer cannot publish another authority's types;
- subject-binding rules appropriate to the event type;
- `authority_transfer:false` and closed schema validation before dispatch or persistence; and
- Wardveil and Privacy Shield acceptance for the intended transport and data class.

No write scope is activated merely by documenting these future requirements.

### Current consumers

External live consumers must:

- authenticate through the existing GoreeCloud Identity verifier boundary;
- carry the dedicated read-only `mesh.events.read` scope;
- explicitly request one or more registered event types; and
- remain subject to the endpoint's bounded stream window and buffer limits.

Possession of evidence, Platform Registry, policy-evaluation, or mutation scopes does not imply event-read permission. Authentication establishes consumer identity and authorization only; it does not upgrade event payloads into producer-domain truth.

## Privacy and data-minimization requirements

Mesh lifecycle events are metadata, not application messages. The v1 payloads are intentionally closed so a caller cannot opportunistically add bearer tokens, authorization headers, private keys, browsing or DNS history, user content, recovery secrets, raw security findings, prompts, arbitrary request bodies, or other producer-private material.

Current requirements are:

- every event type declares the minimum fields needed for its coordination purpose;
- arbitrary additional payload fields fail closed;
- credentials and secret material are never event data;
- producer-domain detail stays with the authoritative producer instead of being copied into Mesh by convenience;
- `authority_transfer` is always false;
- local consumers should use `SubscribeTypes` to minimize received types;
- external consumers must explicitly filter event types rather than defaulting to all events; and
- adding a new data class or durable event state requires Privacy Shield review of purpose, minimization, disclosure, retention, and deletion before activation.

A valid Mesh event states only that Mesh observed the governed coordination action represented by that event type. It does not independently prove security, privacy, recovery, health, conformance, Stable eligibility, or another producer's domain verdict.

## Retry and delivery requirements

The current event path is explicitly best-effort and does not retry delivery:

- the authoritative event bus remains process-local;
- `Subscribe` receives all currently registered lifecycle event types for backward-compatible local behavior;
- `SubscribeTypes` permits explicit event-type minimization and rejects unsupported event types;
- local subscriber buffers have a minimum size of one and a hard maximum of 64 events;
- publication does not block or roll back a successful Registry/relationship mutation when a subscriber buffer is full;
- a slow subscriber may therefore miss an event;
- `GET /v1/events/stream` exposes a source-level external SSE adapter protected by `mesh.events.read`;
- external requested buffers are bounded to 1–64 events;
- each HTTP stream window is bounded to 1–10 seconds and closes deliberately, preserving the existing server write-timeout boundary;
- reconnecting creates a new live subscription and can miss events between windows;
- the adapter emits no SSE `id:` field and rejects `Last-Event-ID`, cursor, or other replay-style parameters;
- no acknowledgement, retry/backoff, dead-letter, exactly-once, at-least-once, cross-host ordering, federation, or notification guarantee exists.

These are requirements, not temporary hidden behavior: callers must not rely on delivery for correctness, authorization, recovery, or irreversible domain actions. Any future retry/delivery guarantee requires a separately versioned transport contract with bounded retry/backoff, idempotency, ordering, dead-letter handling, privacy controls, and failure isolation.

## Retention requirements

Current lifecycle events have **no authorized durable retention**:

- events are not written to the Mesh state store or Platform Registry;
- event IDs are process-local identifiers, not durable offsets;
- process restart loses event history;
- the bounded external stream does not create a retained subscriber session; and
- Mesh does not currently retain an event journal for replay, audit, analytics, or notification history.

Durable retention must remain disabled until a separate accepted design defines the exact retained data class and purpose, maximum retention/deletion behavior under Privacy Shield, access controls and evidence integrity under Identity/Wardveil, backup/restore and recovery requirements under Everkeep where applicable, and the relationship between retained transport evidence and producer-domain records. Absence of an accepted retention policy is a fail-closed reason not to persist event history.

## Remaining durable/external boundary

Durable journals, replay, subscriber checkpoints, retry/backoff, dead-letter handling, retention implementation, continuous delivery guarantees, cross-host/federated transport, and delivery acceptance remain separate Mesh milestones. That future work must define and validate, before activation:

- GoreeCloud Identity-authenticated publisher identities in addition to the current consumer read boundary;
- distinct least-privilege publish/read scopes and producer/subject binding that prevents cross-service impersonation;
- Wardveil trust, event-integrity, connection-security, and abuse-control requirements;
- Privacy Shield minimization, purpose, retention, disclosure, and deletion requirements;
- bounded retry and dead-letter behavior that cannot amplify sensitive data exposure;
- ordering and idempotency semantics;
- Everkeep backup/restore requirements for any durable event state; and
- failure isolation so an event subsystem outage does not make application-owned data irrecoverable.

No future transport should infer authority from successful delivery. Transport acceptance and producer-domain truth remain separate.
