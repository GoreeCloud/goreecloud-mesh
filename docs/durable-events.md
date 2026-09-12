# GoreeCloud Mesh — Durable Event Journal Foundation

**Roadmap:** FR-008  
**Lifecycle:** Development source candidate

This source slice introduces a bounded single-runtime durable event journal as the first persistence primitive for future Mesh replay and checkpoint delivery. It does not change the existing live SSE endpoint into a durable stream and does not claim cross-host delivery.

The journal assigns a monotonic durable offset separate from the existing process-local `evt-<sequence>` event ID. Existing event envelopes remain governed by `goreecloud.mesh.event.v1`, remain closed/privacy-minimized, and keep `authority_transfer: false`.

The journal persists its state atomically to a private file, reloads and validates retained entries after restart, preserves monotonic offsets, bounds retained history, and supports replay after an explicit durable checkpoint. A checkpoint older than retained history fails with an explicit checkpoint-too-old state instead of silently skipping missing events. A checkpoint ahead of the journal also fails closed. Corrupt state, schema drift, invalid retained events, or implicit retention-policy changes are rejected at open time.

This foundation intentionally does **not** establish an external replay endpoint, subscriber acknowledgement, retry/backoff, dead-letter handling, exactly-once or at-least-once delivery, external publisher authority, multi-process locking, shared-filesystem correctness, cross-host ordering, federation, production retention approval, production deployment, or Stable qualification. Those remain separate FR-008 acceptance gates.
