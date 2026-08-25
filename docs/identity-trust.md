# Identity and Trust Boundary

GoreeCloud Mesh requires authenticated service identities before privileged coordination operations are accepted.

## Authority

GoreeCloud Identity remains the service-identity and authentication authority. Wardveil Security remains the trust-boundary, protection-state, and security-evidence authority. Mesh consumes verified identity and security evidence; it does not mint production identities or replace either system.

## Principal model

The trust contract defines a verified principal with:

- stable GoreeCloud service identity;
- explicit authorization scopes;
- identity issuer;
- optional issuer subject identifier.

Scopes are normalized and evaluated explicitly. Unknown or absent scopes fail closed.

## Verifier boundary

`trust.Verifier` is the adapter contract for GoreeCloud Identity integration. A verifier accepts an HTTP request and returns a verified principal or an error. The interface intentionally does not prescribe bearer tokens, cookies, mutual TLS, signed requests, or another credential format before the Identity and Wardveil contracts select and validate the production mechanism.

Mesh now enforces this verifier boundary on mutating API routes. If no verifier is installed, privileged mutations fail closed with an authentication error rather than accepting anonymous writes. A verifier failure or invalid principal is rejected as unauthenticated; a verified principal without the required scope is rejected as forbidden.

## Least-privilege scopes

The current privileged API scopes are:

- `mesh.services.write` for service registration;
- `mesh.relationships.write` for relationship mutation;
- `mesh.policy.evaluate` for policy evaluation requests;
- `mesh.attestations.write` for source-attestation writes; and
- `mesh.contracts.write` for runtime contract-evidence writes.

Read-only inspection endpoints remain separate from these mutation scopes in this source increment. Their eventual production exposure remains governed by the private-first API policy and applicable Identity, Wardveil Security, Gateway, and Network controls.

## Runtime boundary

The current Mesh process does not install a production GoreeCloud Identity verifier. Consequently, privileged mutations through the default process fail closed until an authoritative verifier adapter is supplied. This source enforcement must not be represented as production Identity integration, runtime authentication acceptance, or permission to invent a temporary credential format.

Development and tests may inject bounded verifier implementations directly through the authorized server constructor. Caller-supplied identity headers or arbitrary tokens must not be treated as verified production identity.

## Sensitive information

Mesh trust state must not persist reusable credentials, raw tokens, private keys, recovery secrets, or unnecessary identity claims. Logs should record only bounded operational identity metadata needed for auditing and diagnostics.

## Production gate

Production authentication remains incomplete until an authoritative GoreeCloud Identity verifier implementation and the applicable Wardveil Security trust/protection acceptance are implemented, tested, and accepted. Source-level authorization enforcement and unit tests do not constitute runtime security acceptance, production authorization, deployment approval, release authorization, or Stable qualification.
