# GoreeCloud Mesh — Durable Event Journal Foundation

**Roadmap:** FR-008  
**Lifecycle:** Development source candidate

This source slice introduces bounded single-runtime durable event journal, subscriber-checkpoint, and replay-runner primitives as the first persistence layer for future Mesh replay and checkpoint delivery. It does not change the existing live SSE endpoint into a durable stream and does not claim cross-host delivery.

The journal assigns a monotonic durable offset separate from the existing process-local `evt-<sequence>` event ID. Existing event envelopes remain governed by `goreecloud.mesh.event.v1`, remain closed/privacy-minimized, and keep `authority_transfer: false`.

The journal persists its state atomically to a private file, reloads and validates retained entries after restart, preserves monotonic offsets, bounds retained history, and supports replay after an explicit durable checkpoint. A checkpoint older than retained history fails with an explicit checkpoint-too-old state instead of silently skipping missing events. A checkpoint ahead of the journal also fails closed. Corrupt state, schema drift, invalid retained events, or implicit retention-policy changes are rejected at open time.

## Durable subscriber checkpoint state

`DurableSubscriberCheckpoints` adds a separate private, atomic checkpoint store for future durable consumers. A subscriber may acknowledge only a checkpoint that is not ahead of the journal and is not older than the retained replay boundary. Checkpoints are monotonic per canonical subscriber identifier: repeats are idempotent and regressions fail closed.

Checkpoint state survives restart, is strictly decoded, rejects symlink-backed or non-private state, is bounded by subscriber count and file size, and does not mutate the event journal itself.

## Bounded subscriber replay runner

`ReplaySubscriber` composes the journal and checkpoint store for one bounded in-process consumer pass. It loads the subscriber's durable checkpoint, replays a bounded batch, invokes the supplied handler in offset order, and advances the durable checkpoint only after that handler returns successfully.

If a handler fails, that event is **not acknowledged**. A later call may therefore present the same event again before later offsets. If checkpoint persistence fails, the runner stops immediately rather than reporting progress that was not durably recorded. If the subscriber checkpoint has fallen behind retained history, replay fails closed before the handler runs.

This runner establishes neither exactly-once nor at-least-once delivery. Successful handler return is only the local callback boundary supplied to this source primitive; it does not prove external persistence, side-effect completion, downstream acknowledgement, or cross-process delivery.

## Acceptance boundary

These are **single-runtime state and replay primitives only**. They do not establish an external replay endpoint, remote acknowledgement authority, authenticated subscriber transport, retry/backoff scheduling, dead-letter handling, exactly-once or at-least-once delivery, multi-process locking, shared-filesystem correctness, cross-host ordering, federation, production retention approval, production deployment, production acceptance, or Stable qualification. Those remain separate FR-008 acceptance gates.
