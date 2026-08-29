# Authenticated GoreeCloud Mesh Evidence Delivery

This contract defines the runtime delivery boundary for `goreecloud.evidence-envelope.v1` after an authoritative producer has already created a minimized envelope.

It does not move identity, authentication, authorization, account, device, credential, session, security, privacy, recovery, continuity, presentation, or design-conformance authority into GoreeCloud Mesh.

## Identity boundary

Evidence API access uses the existing GoreeCloud Identity verifier boundary in Mesh.

A protected request must resolve to a valid service principal with:

- a non-empty service identity;
- a trusted issuer accepted by the configured verifier;
- a non-empty subject for producer delivery;
- the exact least-privilege scope required by the route.

Mesh does not mint producer credentials.

Authentication proves the service principal and its granted Mesh scope. It does not prove the producer-domain assertion carried by an envelope, including when the producer is GoreeCloud Identity itself.

## Producer write scope

`POST /v1/evidence/envelopes` requires `mesh.evidence.write`.

Scope possession alone is not sufficient. The authenticated `service_id` must exactly match the envelope producer:

| Envelope producer | Required authenticated service ID |
| --- | --- |
| `goreecloud-identity` | `goreecloud-identity` |
| `wardveil-security` | `wardveil-security` |
| `privacy-shield` | `privacy-shield` |
| `everkeep` | `everkeep` |
| `glaze-ui` | `glaze-ui` |
| `goreecloud-mesh` | `goreecloud-mesh` |

A scoped Privacy Shield principal therefore cannot submit a Wardveil envelope, a scoped Wardveil principal cannot submit Everkeep recovery evidence, and a scoped GoreeCloud Identity principal cannot use Identity authentication to submit evidence under another producer's authority.

This identity binding is separate from envelope validation. Both must pass.

## Evidence read scope

The evidence inspection and consumer-view routes require `mesh.evidence.read`:

- `GET /v1/evidence/envelopes`
- `GET /v1/evidence/envelopes/{id}`
- `GET /v1/evidence/status`
- `GET /v1/evidence/subjects/{kind}/{id}`

A read credential does not confer evidence-write authority.

## Delivery result and replay behavior

A new accepted envelope returns HTTP `201`.

An exact retry of the same immutable envelope ID and identical content returns HTTP `200` with `replayed: true`. This is transport idempotency, not a new producer observation.

Reusing an existing evidence ID with different content fails closed.

The delivery receipt binds the accepted envelope to the authenticated producer service ID. It does not contain or return the bearer credential.

## Freshness

New evidence must be current at ingestion time according to its producer-declared `valid_until` boundary.

Expired evidence remains durable and inspectable after it was legitimately accepted. Historical expiration does not prevent Mesh restart.

Freshness remains timing metadata only. `fresh: true` does not mean authenticated, authorized, protected, private, recoverable, compliant, healthy, or conformant.

## Consumer subject view

`GET /v1/evidence/subjects/{kind}/{id}` provides the consumer-oriented read model used by Glaze UI and product surfaces.

The view:

- reports Mesh transport availability and current/stale counts;
- groups evidence by producer and authority domain;
- groups each producer's evidence by assertion;
- exposes the latest observation and latest currently valid observation separately;
- preserves the producer's outcome vocabulary verbatim;
- never calculates an overall identity, authentication, authorization, security, privacy, recovery, continuity, or conformance verdict.

A successful Everkeep restore verification may therefore appear alongside a Wardveil security finding and a GoreeCloud Identity authentication result without one cancelling or upgrading the others.

## Credential handling in producer clients

Reference delivery clients in Wardveil Security, Privacy Shield, and Everkeep accept a short-lived GoreeCloud Identity bearer credential supplied by the calling runtime.

Those clients:

- place the credential only in the HTTP `Authorization` header;
- do not copy it into the evidence envelope;
- do not return it in the delivery receipt;
- do not persist it;
- reject non-loopback plaintext HTTP destinations;
- reject redirects while carrying the credential.

GoreeCloud Identity is now a recognized bounded evidence producer in the Mesh schema and validator, but an Identity-owned runtime delivery client is not yet implemented. The producer profile therefore continues to record Identity delivery implementation and production acceptance as false.

Credential issuance, renewal, revocation, audience policy, and production token verification remain GoreeCloud Identity responsibilities.

## Transport security boundary

Reference producer and Glaze consumer clients require HTTPS for non-loopback destinations. Loopback HTTP is allowed only for local development and tests.

This source contract does not claim that a public or private production Mesh endpoint, TLS route, Gateway policy, DNS record, or Cloudflare deployment is already active.

## Production acceptance boundary

This milestone establishes source-level authenticated delivery enforcement, producer identity binding, bounded Identity producer acceptance in the envelope validator, and tested producer/consumer clients where implemented. Production acceptance still requires the deployed Mesh runtime to be configured with a real GoreeCloud Identity verifier and the intended Gateway/Network/TLS controls, followed by runtime evidence that producer identities, scopes, replay behavior, and consumer reads work in the target environment. Identity evidence delivery additionally requires an Identity-owned delivery implementation and its own target-environment acceptance.
