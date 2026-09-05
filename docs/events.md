# GoreeCloud Mesh Event Contract v1

## Purpose

GoreeCloud Mesh Events provides a bounded coordination signal for Registry and relationship lifecycle changes. The underlying bus remains deliberately **in-process, best-effort, non-durable, and non-replayable**. The source implementation now also exposes a narrowly scoped GoreeCloud Identity-authenticated HTTP streaming adapter for external consumers; that adapter transports only the same closed lifecycle metadata and does not turn Mesh into a durable event broker or a store for application payloads, credentials, user activity, or producer-private data.

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

## Envelope rules

Every current event contains:

- `schema` — exactly `goreecloud.mesh.event.v1`;
- `id` — a process-local monotonic `evt-<sequence>` identifier;
- `type` — one of the registered v1 event names;
- `source` — the coordination producer/source identifier;
- `subject` — the service or relationship identifier affected by the event;
- `data` — a type-specific closed payload;
- `created_at` — the UTC event time; and
- `authority_transfer` — always `false`.

The Go validator rejects unsupported event types, unexpected payload fields, malformed field types, control characters, oversized identity/value fields, and any attempt to set `authority_transfer` to true.

## Privacy and security boundary

Mesh lifecycle events are metadata, not application messages. The v1 payloads are intentionally closed so a caller cannot opportunistically add bearer tokens, authorization headers, private keys, browsing or DNS history, user content, recovery secrets, raw security findings, or other producer-private material.

Local consumers may use `SubscribeTypes` to request only the registered lifecycle event types they actually need. External consumers must authenticate through the existing GoreeCloud Identity verifier boundary, possess the dedicated read-only `mesh.events.read` scope, and explicitly name at least one registered event type. Possession of unrelated Mesh read or write scopes does not authorize event consumption.

The external adapter rejects unknown query parameters, unsupported event types, ambiguous repeated bounds, replay requests, and `Last-Event-ID`. Its query surface contains only event-type filters and bounded transport controls; credentials remain in the normal Identity authorization boundary rather than event payloads or query parameters.

A valid Mesh event states that Mesh observed a local Registry or relationship lifecycle action. Authentication proves the consumer identity and read scope only. Neither authentication nor successful delivery independently proves security, privacy, recovery, health, conformance, Stable eligibility, or producer-domain truth owned by another GoreeCloud authority.

## Delivery semantics

The current event path is bounded and best-effort:

- the authoritative event bus remains process-local;
- `Subscribe` receives all currently registered lifecycle event types for backward-compatible local behavior;
- `SubscribeTypes` permits explicit event-type minimization and rejects unsupported event types;
- local subscriber buffers have a minimum size of one and a hard maximum of 64 events;
- publication does not block a Registry or relationship mutation when a subscriber buffer is full;
- an event may therefore be dropped for a slow subscriber;
- `GET /v1/events/stream` exposes a source-level external Server-Sent Events adapter protected by `mesh.events.read`;
- the external consumer must supply one or more repeated `type=<registered-event-type>` filters;
- external requested buffers are bounded to 1–64 events;
- each HTTP stream window is bounded to 1–10 seconds and closes deliberately, preserving the existing server write-timeout boundary;
- the adapter emits no SSE `id:` field and rejects `Last-Event-ID`, cursor, or other replay-style parameters;
- reconnecting creates a new live subscription and can miss events between stream windows;
- event IDs inside the JSON envelope are process-local identifiers, not durable replay offsets;
- events are not persisted to the Mesh state store or Platform Registry;
- restart does not replay historical events; and
- no durable external delivery, federation, webhook, queue, cross-host ordering, acknowledgement, retry, or notification guarantee is claimed.

These semantics must not be described as durable event delivery. The external adapter establishes a source-level authenticated read boundary for live lifecycle metadata only.

## Remaining durable/external boundary

Durable journals, replay, subscriber checkpoints, retry/backoff, retention policies, continuous delivery guarantees, cross-host/federated transport, and delivery acceptance remain a separate Mesh milestone. That future work must define and validate, before activation:

- GoreeCloud Identity-authenticated publisher identities in addition to the current consumer read boundary;
- distinct least-privilege publish/read scopes and producer/subject binding that prevents cross-service impersonation;
- Wardveil trust, event-integrity, connection-security, and abuse-control requirements;
- Privacy Shield minimization, purpose, retention, disclosure, and deletion requirements;
- bounded retry and dead-letter behavior that cannot amplify sensitive data exposure;
- ordering and idempotency semantics;
- Everkeep backup/restore requirements for any durable event state; and
- failure isolation so an event subsystem outage does not make application-owned data irrecoverable.

No future transport should infer authority from successful delivery. Transport acceptance and producer-domain truth remain separate.
