# GoreeCloud Mesh API

The initial Mesh API is a private-first JSON HTTP contract. It is intentionally small and does not imply public exposure or production authorization.

## Common behavior

- Request and response content uses JSON unless an endpoint has no body.
- Responses include `Cache-Control: no-store` and `X-Content-Type-Options: nosniff`.
- JSON request bodies are bounded to 1 MiB and reject unknown fields.
- Identifiers are caller-supplied stable strings in the first milestone.
- Authentication is not yet implemented. Bind to loopback only until the GoreeCloud Identity and Wardveil trust contract is implemented and validated.

## Health

### `GET /healthz`

Returns process availability.

```json
{"status":"ok","service":"goreecloud-mesh"}
```

This is not a complete production-readiness or dependency-health signal.

## State

### `GET /v1/state`

Returns the current Registry and relationship state.

This endpoint is intended for administrative inspection and testing. Future production authorization may restrict or replace broad state reads with narrower views.

## Services

### `GET /v1/services`

Lists registered services in deterministic ID order.

### `POST /v1/services`

Creates or replaces a service registration.

Required fields:

- `id`
- `name`
- `kind`

Example:

```json
{
  "id": "goreecloud-identity",
  "name": "GoreeCloud Identity",
  "kind": "platform",
  "version": "development",
  "endpoint": "https://identity.goreecloud.com",
  "capabilities": ["identity.authenticate", "identity.resolve"],
  "dependencies": [],
  "health": "healthy",
  "labels": {"layer": "identity-security"},
  "conformance": {
    "glaze_ui": "pending",
    "wardveil": "pending",
    "privacy_shield": "pending",
    "everkeep": "pending"
  }
}
```

Supported health states are `unknown`, `healthy`, `degraded`, and `unavailable`.

### `GET /v1/services/{id}`

Returns one registered service or `404` when it is unknown.

## Capability discovery

### `GET /v1/capabilities/{capability}`

Returns registered services that advertise the capability and are not explicitly unavailable.

Discovery is not authorization. A discovered target may still be denied by Mesh Policy, Identity, Gateway, Network, application permissions, or Wardveil controls.

## Relationships

### `POST /v1/relationships`

Creates or replaces an explicit relationship between two registered services.

Required fields:

- `id`
- `from`
- `to`
- `type`

Example:

```json
{
  "id": "manager-identity-auth",
  "from": "goreecloud-manager",
  "to": "goreecloud-identity",
  "type": "consumes",
  "capability": "identity.authenticate",
  "contract": "identity-auth-v1",
  "required": true,
  "enabled": true
}
```

Self-relationships and references to unknown services are rejected.

## Dependency impact

### `GET /v1/graph/impact?id=<service-id>`

Returns the directly and transitively affected registered services whose declared dependency or enabled required relationship points to the selected service.

Example response:

```json
{
  "service": "goreecloud-identity",
  "affected": ["goreecloud-manager", "goreecloud-notes"]
}
```

Impact is advisory coordination data. Specialized Monitoring, Manager, Gateway, Network, and application evidence remain authoritative for their own runtime state.

## Policy evaluation

### `POST /v1/evaluate`

Evaluates whether a registered source has an enabled relationship that authorizes use of a capability advertised by the registered target.

Example request:

```json
{
  "source": "goreecloud-manager",
  "target": "goreecloud-identity",
  "capability": "identity.authenticate"
}
```

Example response:

```json
{
  "allowed": true,
  "reason": "enabled relationship authorizes capability"
}
```

The initial evaluator fails closed when:

- the source is unknown;
- the target is unknown;
- the target is unavailable;
- the capability is missing;
- the target does not advertise the capability; or
- no enabled relationship authorizes it.

Mesh Policy is one authorization input, not a substitute for application authorization, GoreeCloud Identity, Gateway controls, Network controls, or Wardveil Security.

## Integral platform catalog

### `GET /v1/platforms`

Returns the canonical Mesh-side catalog of mandatory integral platform systems. Each entry records the system identifier, display name, repository, authority boundary, expected contract source, integration-manifest path, and whether the system is required.

The catalog currently identifies:

- Glaze UI as the authority for design, interaction, accessibility, adaptive behavior, and visual presentation;
- Wardveil Security as the authority for security/protection status and evidence semantics;
- GoreeCloud Privacy Shield as the authority for privacy controls, privacy status, data minimization, retention expectations, and privacy-capability governance; and
- Everkeep as the authority for resilience, recovery, preservation, portability, succession, and digital-legacy evidence.

The endpoint is descriptive and coordination-oriented. Mesh does not acquire the technical authority of a listed system merely because it catalogs or consumes that system's contract.

### `GET /v1/platforms/status`

Returns one read-only integration-status record for each mandatory platform system by joining the canonical authority catalog with the latest bounded runtime contract evidence recorded in Mesh.

The response also includes `source_attestations_validated`, which reports whether all four mandatory platform manifests have separately validated source attestations. That field is independent of `stable_eligible`; source provenance cannot satisfy runtime evidence requirements.

Missing runtime evidence is represented explicitly as `pending` with `stable_gate_satisfied: false`; it is never omitted or treated as healthy. This endpoint does not convert Mesh into the authority for design, security, privacy, or resilience, and it does not authorize release, deployment, production acceptance, or Stable promotion.

## Platform source attestations

### `GET /v1/platforms/source-attestations`

Returns the current source-provenance attestations and `all_validated`, a fail-closed completeness result for the four mandatory integral platform systems.

A source attestation binds a reviewed platform integration manifest to:

- its canonical platform and producer repository;
- the exact full producer Git revision;
- the canonical integration-manifest path;
- the SHA-256 digest of the exact reviewed manifest bytes;
- source-validation state;
- validation workflow and run evidence when validated; and
- the observation time.

Source attestations contain source provenance only. They do not imply runtime acceptance, production acceptance, deployment authorization, release authorization, or Stable qualification.

### `POST /v1/platforms/source-attestations`

Records or replaces the latest source attestation for one mandatory platform. The request is validated against the canonical Mesh platform catalog and fails closed on repository mismatches, manifest-path mismatches, malformed revisions or digests, invalid states, missing validated-workflow evidence, or any request that implies runtime or Stable acceptance.

Example:

```json
{
  "platform": "glaze-ui",
  "repository": "GoreeCloud/glaze-ui",
  "revision": "0123456789abcdef0123456789abcdef01234567",
  "manifest_path": "contracts/mesh.integration.json",
  "manifest_sha256": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
  "state": "validated",
  "validation_workflow": "Glaze UI CI",
  "validation_run_id": 123456789,
  "observed_at": "2026-08-24T23:45:00Z",
  "runtime_acceptance_implied": false,
  "stable_acceptance_implied": false
}
```

## Runtime platform contracts

### `GET /v1/contracts`

Returns the four mandatory integral platform systems and the currently recorded runtime evidence entries.

The mandatory systems are Glaze UI, Wardveil Security, Privacy Shield, and Everkeep.

### `POST /v1/contracts`

Records or replaces one bounded runtime evidence entry. The `contract` value must exactly match that platform's `contract_source` from `GET /v1/platforms`; arbitrary aliases or synthetic contract identifiers are rejected. A `validated` record additionally requires non-empty `source` and `revision` fields so a validated state is tied to identifiable runtime evidence rather than an untraceable assertion.

Example:

```json
{
  "platform": "wardveil-security",
  "contract": "contracts/wardveil.status.schema.json",
  "state": "validated",
  "source": "wardveil-adapter",
  "revision": "example-revision",
  "detail": "runtime security contract accepted"
}
```

Supported states are `pending`, `validated`, and `blocked`. Unknown platform systems, non-canonical contract identifiers, invalid states, and validated records without source/revision evidence are rejected. Evidence must not contain reusable credentials, private keys, tokens, recovery secrets, or private application content.

### `GET /v1/contracts/stable-eligibility`

Returns the current fail-closed source-level Stable eligibility calculation together with the mandatory-system list and recorded evidence.

`stable_eligible` is `true` only when validated runtime evidence exists for all four mandatory systems. This endpoint does not itself authorize a release, production deployment, production acceptance, or Stable promotion. Those remain separate governed transitions requiring their applicable evidence and approvals.

## Compatibility

The `/v1/` namespace is the first public contract boundary inside the repository. Breaking changes must move through a controlled version transition instead of silently changing consumer behavior.
