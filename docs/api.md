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

Returns the canonical Mesh-side catalog of mandatory integral platform systems. Each entry records the system identifier, display name, repository, authority boundary, expected contract source, and whether the system is required.

The catalog currently identifies:

- Glaze UI as the authority for design, interaction, accessibility, adaptive behavior, and visual presentation;
- Wardveil Security as the authority for security/protection status and evidence semantics;
- GoreeCloud Privacy Shield as the authority for privacy controls, privacy status, data minimization, retention expectations, and privacy-capability governance; and
- Everkeep as the authority for resilience, recovery, preservation, portability, succession, and digital-legacy evidence.

The endpoint is descriptive and coordination-oriented. Mesh does not acquire the technical authority of a listed system merely because it catalogs or consumes that system's contract.

## Runtime platform contracts

### `GET /v1/contracts`

Returns the four mandatory integral platform systems and the currently recorded runtime evidence entries.

The mandatory systems are Glaze UI, Wardveil Security, Privacy Shield, and Everkeep.

### `POST /v1/contracts`

Records or replaces one bounded runtime evidence entry.

Example:

```json
{
  "platform": "wardveil-security",
  "contract": "wardveil-runtime-v1",
  "state": "validated",
  "source": "wardveil-adapter",
  "revision": "example-revision",
  "detail": "runtime security contract accepted"
}
```

Supported states are `pending`, `validated`, and `blocked`. Unknown platform systems, missing contract identifiers, and invalid states are rejected. Evidence must not contain reusable credentials, private keys, tokens, recovery secrets, or private application content.

### `GET /v1/contracts/stable-eligibility`

Returns the current fail-closed source-level Stable eligibility calculation together with the mandatory-system list and recorded evidence.

`stable_eligible` is `true` only when validated evidence exists for all four mandatory systems. This endpoint does not itself authorize a release, production deployment, production acceptance, or Stable promotion. Those remain separate governed transitions requiring their applicable evidence and approvals.

## Compatibility

The `/v1/` namespace is the first public contract boundary inside the repository. Breaking changes must move through a controlled version transition instead of silently changing consumer behavior.
