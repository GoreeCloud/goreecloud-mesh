# GoreeCloud Mesh — Durable Event Journal Foundation

**Roadmap:** FR-008  
**Lifecycle:** Development source candidate

This source slice introduces bounded single-runtime durable event journal and subscriber-checkpoint primitives as the first persistence layer for future Mesh replay and checkpoint delivery. It does not change the existing live SSE endpoint into a durable stream and does not claim cross-host delivery.

The journal assigns a monotonic durable offset separate from the existing process-local `evt-<sequence>` event ID. Existing event envelopes remain governed by `goreecloud.mesh.event.v1`, remain closed/privacy-minimized, and keep `authority_transfer: false`.

The journal persists its state atomically to a private file, reloads and validates retained entries after restart, preserves monotonic offsets, bounds retained history, and supports replay after an explicit durable checkpoint. A checkpoint older than retained history fails with an explicit checkpoint-too-old state instead of silently skipping missing events. A checkpoint ahead of the journal also fails closed. Corrupt state, schema drift, invalid retained events, or implicit retention-policy changes are rejected at open time.

## Durable subscriber checkpoint state

`DurableSubscriberCheckpoints` adds a separate private, atomic checkpoint store for future durable consumers. A subscriber may acknowledge only a checkpoint that is not ahead of the journal and is not older than the retained replay boundary. Checkpoints are monotonic per canonical subscriber identifier: repeats are idempotent and regressions fail closed.

Checkpoint state survives restart, is strictly decoded, rejects symlink-backed or non-private state, is bounded by subscriber count and file size, and does not mutate the event journal itself.

This is a **state primitive only**. Calling `Acknowledge` records a replay position supplied by a future authorized delivery path; it does not prove that a subscriber actually received, processed, persisted, or acted on an event. No acknowledgement endpoint or remote caller authority is established by this source slice.

## Remaining acceptance gates

This foundation intentionally does **not** establish an external replay endpoint, external subscriber acknowledgement transport, retry/backoff, dead-letter handling, exactly-once or at-least-once delivery, external publisher authority, multi-process locking, shared-filesystem correctness, cross-host ordering, federation, production retention approval, production deployment, or Stable qualification. Those remain separate FR-008 acceptance gates.
