# GoreeCloud Evidence Envelope

The GoreeCloud Evidence Envelope is the bounded transport contract used by GoreeCloud Mesh to carry evidence from authoritative GoreeCloud systems without becoming authoritative for the underlying domain.

The v1 contract is defined by `contracts/mesh.evidence-envelope.schema.json` and implemented in `internal/contracts/evidence_envelope.go`.

## Purpose

GoreeCloud systems increasingly produce evidence that another system must consume:

- Wardveil Security produces security, protection, trust, scan, quarantine, incident, and runtime-acceptance evidence.
- Privacy Shield produces privacy decisions, consent/lifecycle evidence, minimized receipts, and privacy-safe AI/RAG evidence.
- Everkeep produces protection, integrity, recovery, preservation, continuity, and restore-verification evidence.
- Glaze UI may produce presentation/design-conformance evidence but never security, privacy, or recovery truth.
- GoreeCloud Mesh produces coordination and governance evidence and transports the other systems' evidence without upgrading it.

The envelope standardizes provenance, freshness, subject identity, minimization, and transport metadata. It deliberately does not standardize or reinterpret each producer's domain-specific outcome vocabulary.

## Authority rule

An envelope never transfers authority.

A `wardveil-security` producer may assert only the `security` authority domain. A `privacy-shield` producer may assert only `privacy`. Everkeep may assert only its resilience/recovery/preservation/continuity domains. Glaze UI may assert only presentation or design-conformance. Mesh may assert only coordination or governance.

Mesh must reject an envelope whose declared authority domain does not belong to its producer.

## Provenance rule

Every accepted envelope must bind to:

1. a recognized first-party GoreeCloud producer;
2. that producer's canonical repository;
3. an exact 40-character Git revision;
4. the producer contract governing the assertion;
5. an opaque producer-controlled evidence source.

The envelope therefore makes source identity inspectable without treating source provenance alone as proof that the producer's assertion is correct.

## Freshness rule

Every envelope requires both `observed_at` and producer-declared `valid_until` timestamps.

Mesh does not invent a universal freshness duration. Evidence that is future-dated, expired, or whose validity window does not follow its observation time fails closed.

## Data-minimization rule

The envelope is metadata, not a telemetry lake or evidence payload container.

It must not contain raw user content, credentials, keys, tokens, secret material, message bodies, file bodies, browsing history, DNS history, AI prompt content, location history, or equivalent private payloads. The `contains_user_content` and `contains_secret_material` fields are required and must be `false`.

`data_class` is limited to:

- `public` — non-sensitive public metadata;
- `operational` — bounded service/contract metadata;
- `derived` — minimized derived state that is safe to transport under the producer contract.

A short summary is optional and capped at 512 characters. A `sha256:<digest>` may bind the envelope to an external producer-controlled payload without placing that payload in Mesh.

## Producer-defined outcomes

`assertion` and `outcome` are producer-defined values. Mesh preserves them as opaque domain semantics.

For example, Everkeep may use a restore-verification assertion whose valid outcomes are defined by an Everkeep contract. Mesh may validate the envelope and make it discoverable, but it must not turn an Everkeep-specific outcome into a Wardveil security result, a Privacy Shield decision, or a Glaze UI semantic state.

## Consumer rule

Consumers must evaluate both layers:

1. the Mesh envelope is structurally valid, current, minimized, and correctly attributed; and
2. the producer-specific contract says the assertion/outcome is acceptable for the requested purpose.

Passing layer 1 never substitutes for layer 2.

## Glaze UI presentation

Glaze UI may render evidence freshness, producer identity, source visibility, unavailable state, and domain-specific status supplied by the producer. It must not infer protection, privacy compliance, recoverability, or conformance merely because an envelope exists.

## Current implementation boundary

The v1 implementation is a source-level contract and validator. It does not yet provide a durable envelope registry, HTTP API, signature verification, distributed revocation, cross-node replication, or production acceptance. Those remain separate milestones.
