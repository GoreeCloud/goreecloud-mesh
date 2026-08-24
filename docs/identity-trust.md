# Identity and Trust Boundary

GoreeCloud Mesh requires authenticated service identities before production registration, relationship mutation, policy evaluation, or other privileged coordination operations are accepted.

## Authority

GoreeCloud Identity remains the service-identity and authentication authority. Wardveil Security remains the trust-boundary, protection-state, and security-evidence authority. Mesh consumes verified identity and security evidence; it does not mint production identities or replace either system.

## Principal model

The initial trust contract defines a verified principal with:

- stable GoreeCloud service identity;
- explicit authorization scopes;
- identity issuer;
- optional issuer subject identifier.

Scopes are normalized and evaluated explicitly. Unknown or absent scopes fail closed.

## Verifier boundary

`trust.Verifier` is an adapter contract for a future GoreeCloud Identity integration. A verifier accepts an HTTP request and returns a verified principal or an error. The interface intentionally does not prescribe bearer tokens, cookies, mutual TLS, signed requests, or another credential format before the Identity and Wardveil contracts select and validate the production mechanism.

The current Mesh process does not install a production verifier. Its listener therefore remains loopback/private-first. Development code must not treat caller-supplied identity headers or arbitrary tokens as verified production identity.

## Sensitive information

Mesh trust state must not persist reusable credentials, raw tokens, private keys, recovery secrets, or unnecessary identity claims. Logs should record only bounded operational identity metadata needed for auditing and diagnostics.

## Production gate

Production authentication remains incomplete until a GoreeCloud Identity adapter and Wardveil Security trust contract are implemented, tested, and accepted. Source-level trust interfaces and unit tests do not constitute runtime security acceptance, production authorization, or Stable qualification.
