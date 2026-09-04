# GoreeCloud Mesh Event Contract v1

## Purpose

GoreeCloud Mesh Events provides a bounded coordination signal for Registry and relationship lifecycle changes. The current implementation is deliberately **in-process only**. It decouples local producers from local consumers without turning Mesh into a store or transport for application payloads, credentials, user activity, or producer-private data.

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

A valid Mesh event states that Mesh observed a local Registry or relationship lifecycle action. It does not independently prove authentication, authorization, security, privacy, recovery, health, conformance, or Stable eligibility owned by another GoreeCloud authority.

## Delivery semantics

The current event bus is bounded and best-effort:

- subscribers are local in-process channels;
- each subscriber chooses a bounded buffer (with a minimum size of one);
- publication does not block a Registry or relationship mutation when a subscriber buffer is full;
- an event may therefore be dropped for a slow local subscriber;
- event IDs are process-local and are not durable replay offsets;
- events are not persisted to the Mesh state store or Platform Registry;
- restart does not replay historical events; and
- no external delivery, federation, webhook, queue, cross-host ordering, or notification guarantee is claimed.

These semantics are intentional for the current foundation and must not be described as durable event delivery.

## Future durable/external boundary

Durable journals, replay, subscriber checkpoints, retry/backoff, retention policies, external subscriptions, cross-host delivery, and delivery guarantees remain a separate Mesh milestone. That future work must define and validate, before activation:

- GoreeCloud Identity-authenticated publisher and consumer identities;
- distinct least-privilege publish/read/subscribe scopes rather than ambient Mesh access;
- producer and subject binding that prevents cross-service impersonation;
- Wardveil trust and abuse-control requirements;
- Privacy Shield minimization, purpose, retention, disclosure, and deletion requirements;
- bounded retry and dead-letter behavior that cannot amplify sensitive data exposure;
- ordering and idempotency semantics;
- Everkeep backup/restore requirements for any durable event state; and
- failure isolation so an event subsystem outage does not make application-owned data irrecoverable.

No future transport should infer authority from successful delivery. Transport acceptance and producer-domain truth remain separate.
