# GoreeCloud Mesh Evidence Envelopes

GoreeCloud Mesh Evidence Envelope v1 is the shared transport and provenance boundary for minimized evidence emitted by GoreeCloud platform systems.

Its purpose is not to centralize platform authority. It allows Mesh to verify who produced an assertion, which source revision and producer contract it came from, what scope it represents, and whether the producer-declared evidence window is still current while leaving the represented domain truth with the producer.

## Authority model

The currently recognized producer authority domains are:

| Producer | Allowed authority domains |
| --- | --- |
| GoreeCloud Mesh | coordination, governance |
| Glaze UI | presentation, design-conformance |
| Wardveil Security | security |
| Privacy Shield | privacy |
| Everkeep | resilience, recovery, preservation, continuity |

A producer cannot submit an envelope using another producer's authority domain or contract namespace. Mesh validates that boundary and rejects authority escalation.

A structurally accepted envelope does not mean the assertion is positive. Examples of valid current envelopes include a Privacy Shield `DENY`, an Everkeep failed restore verification, a Wardveil attention state, or a Mesh transport validation result.

## Envelope invariants

Every envelope must include:

- version `goreecloud.evidence-envelope.v1`;
- immutable evidence ID;
- canonical producer system and repository;
- exact lowercase 40-character producer Git revision;
- a producer-owned contract;
- producer-owned authority domain;
- scoped subject kind and ID;
- assertion and outcome;
- opaque/bounded source reference;
- evidence observation time;
- producer-declared `valid_until` time;
- minimized data class: `public`, `operational`, or `derived`;
- explicit declarations that raw user content and secret material are absent.

An optional `payload_digest` can bind the minimized envelope to producer-held evidence without placing the underlying payload in Mesh. The only accepted format is `sha256:<64 lowercase hex characters>`.

## Data minimization

Mesh evidence envelopes are not a telemetry lake or private-content replication mechanism.

They must not contain raw user content, credentials, tokens, private keys, passwords, protected file/message bodies, prompts, model context, retrieved private documents, backup/archive payloads, or similar private material. Producer adapters should prefer derived state, reason codes, scope identifiers, opaque evidence references, validity windows, and digests.

A digest can prove correspondence to producer-held material but does not authorize disclosure of that material.

## Immutability and replay

Evidence IDs are immutable.

- First insertion of a valid fresh envelope creates the durable record.
- Exact replay of the same envelope is idempotent.
- Reuse of an existing ID with different content is rejected.

This prevents silent mutation of historical evidence while allowing safe retry behavior from producers.

## Freshness and durability

Freshness is evaluated from the producer-declared `valid_until` boundary.

- New writes must be fresh.
- Current queries exclude expired records when `current=true` is requested.
- Expired evidence remains durably stored and queryable for audit/provenance.
- Normal expiration never turns historical evidence into invalid storage and must not block Mesh startup.
- Future-dated observations, malformed provenance, invalid authority, duplicate persisted IDs, or noncanonical producer metadata fail closed during load.

Freshness is not domain success. A current envelope can represent `DENY`, `fail`, `blocked`, `attention`, `degraded`, or another negative producer outcome.

## Persistence

The default store is `./mesh-evidence-envelopes.json` and can be changed with:

```text
-evidence-envelopes <path>
```

An empty path disables persistence for tests or ephemeral operation.

The store uses a versioned JSON envelope, atomic temporary-file replacement, and restrictive file permissions. It is deliberately separate from service registry state, source attestations, mandatory runtime contract evidence, and Everkeep recovery-evidence state so those evidence classes retain independent lifecycle and acceptance semantics.

## API

Mesh exposes:

- `GET /v1/evidence/envelopes`
- `POST /v1/evidence/envelopes`
- `GET /v1/evidence/envelopes/{id}`
- `GET /v1/evidence/status`

Writes require the least-privilege Identity scope `mesh.evidence.write`.

Read filters include producer, authority domain, subject kind, subject ID, assertion, and current/stale state. These remain private-first administrative interfaces in the current milestone.

See `docs/api.md` for request semantics.

## Producer adapters

The envelope contract is language-neutral. Each producer remains responsible for converting its native evidence into a minimized envelope and for deciding which native assertions are eligible for Mesh transport.

Initial producer adapters are being maintained in their canonical repositories:

- Wardveil Security — security evidence adapter.
- Privacy Shield — privacy evidence adapter, including AI/RAG minimization boundaries.
- Everkeep — restore-verification evidence adapter.
- Glaze UI — presentation/design-conformance evidence profile and Candidate evidence presentation surfaces.

Producer adapter source does not itself establish runtime delivery or production acceptance. A deployed producer, authenticated delivery path, Mesh ingestion, durable observation, and product-specific acceptance remain separate gates.

## Presentation boundary

Glaze UI 1.6 Candidate defines how a consumer can display producer state, authority identity, evidence freshness, and Mesh transport state without merging them into a generic safety status.

Mesh should expose the evidence necessary for that presentation, but it must not transform:

- transport available into secure;
- current evidence into compliant;
- backup existence into recoverable;
- an accepted envelope into producer acceptance;
- multiple authority domains into a single averaged verdict.

## Current acceptance boundary

The registry and API are source-level foundations. They do not independently prove that producer adapters are deployed, that authenticated runtime delivery is active, that evidence is arriving from production, or that any GoreeCloud application is production-ready or Stable-qualified.
