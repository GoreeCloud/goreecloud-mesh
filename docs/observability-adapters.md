# Monitoring, Gateway, and Network Adapter Boundaries

GoreeCloud Mesh consumes evidence from specialized GoreeCloud authorities without replacing them or silently acquiring their administrative privileges.

## GoreeCloud Monitoring

Monitoring remains authoritative for collected health and observability evidence. Mesh may consume a validated `HealthObservation` containing a service identifier, approved health state, source authority, observation time, optional revision, and bounded detail. Mesh must preserve evidence provenance and must reject health observations attributed to another authority.

Mesh may use accepted health state for dependency impact, capability discovery, policy context, and administrative presentation. It must not represent a caller-supplied value as observed Monitoring evidence.

## GoreeCloud Gateway

Gateway remains authoritative for ingress, reverse proxying, service access, traffic routing, and gateway enforcement. Mesh may consume an `EnforcementObservation` describing whether a registered relationship is represented in Gateway configuration or observed enforcement state.

A Gateway observation does not give Mesh authority to mutate Gateway configuration. Future write-capable coordination requires a separately approved least-privilege contract.

## GoreeCloud Network

Network remains authoritative for connectivity and network-policy implementation. Mesh may consume Network enforcement evidence associated with a registered relationship. This supports comparison between intended Mesh relationships and observed/configured network state without granting Mesh unrestricted network administration.

## Evidence freshness

Observations carry an explicit observation time. Callers choose an approved maximum age for the use case. Missing timestamps, non-positive freshness windows, stale observations, and observations dated in the future fail closed.

## Mismatch detection

A later milestone may combine Mesh Graph relationships with Gateway and Network observations to identify conditions such as:

- an enabled required Mesh relationship with no corresponding enforcement path;
- an enforcement path for a relationship that Mesh considers disabled or unknown;
- stale evidence that cannot support a current operational conclusion; or
- a dependency whose Monitoring health evidence indicates degradation or unavailability.

Mismatch detection remains advisory until the applicable specialized authority confirms enforcement state. Mesh must not silently reconfigure another system in response to a mismatch.

## Sensitive-information boundary

Adapter evidence is coordination metadata. It must not include credentials, API keys, access tokens, private keys, request payloads, user content, private application records, recovery secrets, or unnecessary traffic contents. Evidence detail fields must remain bounded and suitable for privacy-conscious administrative diagnostics.

## Current implementation state

This source increment defines adapter interfaces, evidence structures, validation, and freshness semantics only. No production Monitoring, Gateway, or Network adapter is enabled, and no external system configuration is changed by these contracts.
