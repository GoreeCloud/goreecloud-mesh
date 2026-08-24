# GoreeCloud Mesh Architecture

## Role

GoreeCloud Mesh occupies the Shared Platform and Integration layer. It coordinates application- and service-level relationships without taking ownership away from specialized systems or application data authorities.

Mesh answers platform questions such as:

- What applications and services exist?
- Which capabilities do they advertise?
- What explicit relationships connect them?
- Which dependencies are required?
- What integration contract governs a relationship?
- What is the current operational state of a dependency?
- What could be affected if a dependency becomes unavailable?
- Is a requested interaction explicitly permitted by the registered relationship model?

## Authority boundaries

Mesh does not become a universal administrator.

- **GoreeCloud Manager** remains the central administration and operational-management application.
- **GoreeCloud Monitoring** remains authoritative for monitoring collection and health evidence.
- **GoreeCloud Network** remains authoritative for network connectivity and network policy implementation.
- **GoreeCloud Gateway** remains authoritative for ingress, proxying, traffic routing, and service-access enforcement at its boundary.
- **GoreeCloud Identity** remains authoritative for identities, authentication, SSO, and identity-provider behavior.
- **Wardveil Security** defines platform security and evidence contracts.
- **Privacy Shield** defines privacy-control, data-minimization, and privacy-evidence contracts.
- **Everkeep** defines resilience, recovery, preservation, portability, continuity, succession, and digital-legacy contracts.
- **Glaze UI** defines the design and interaction contract for Mesh Console and other user-facing Mesh surfaces.

Mesh coordinates these systems through documented contracts and adapters; it does not duplicate their implementation.

## Core components

### Mesh Registry

Stores service identity, kind, version, endpoint metadata, capabilities, declared dependencies, health state, labels, and platform-conformance state. Registration metadata is coordination data only and should not contain application payloads, message content, user files, credentials, tokens, or other unnecessary private activity.

### Mesh Graph

Builds explicit relationship and dependency edges from Registry records. The first implementation supports reverse dependency-impact traversal so administrators and consuming systems can determine which registered components may be affected by an unavailable dependency.

### Mesh Policy

Evaluates requested service-to-service capability use against explicit enabled relationships. The initial policy model fails closed when a source or target is unknown, a target is unavailable, a capability is not advertised, or no relationship authorizes the capability.

Mesh Policy is not a replacement for Identity authorization, Gateway access controls, Network policy, application permissions, or Wardveil Security enforcement. It is the service-relationship policy layer and should compose with those authorities.

### Mesh Events

Publishes bounded lifecycle events when services and relationships change. The initial implementation is an in-process event bus. External durable delivery, replay, ordering guarantees, federation, and notification adapters remain future milestones.

### Mesh Nodes

A node is a registered application, service, client-facing backend, platform component, or other approved Mesh participant represented by a Registry service record. Future node records may include cryptographic service identity and attested runtime metadata only after Identity and Wardveil contracts are defined.

### Mesh Connections

Connections are explicit relationship records between registered nodes. A relationship contains source, target, type, optional capability, optional contract identifier, required/optional dependency semantics, enabled state, and update time.

### Mesh Console

Mesh Console will provide a Glaze UI administrative experience for the registry, graph, relationships, policy decisions, dependency impact, integration status, platform conformance, and operational state. Console completion is not claimed by this milestone.

## Initial data model

The first milestone deliberately keeps the model small and portable. Durable state is written atomically as JSON with restrictive file permissions. This avoids introducing a database dependency before the multi-node, concurrency, scale, recovery, and migration requirements justify one.

State contains:

- `services`: registered service records keyed by service ID.
- `relationships`: explicit relationship records keyed by relationship ID.

The storage implementation can be replaced later without changing the public Mesh model or API contract.

## Discovery

Capability discovery returns registered services that advertise the requested capability and are not explicitly `unavailable`. Discovery does not itself grant permission to use the service. A caller must still satisfy the relevant relationship, Identity, Gateway, Network, application, and Wardveil controls.

## Failure isolation

Mesh is designed to help the platform degrade safely rather than create more coupling. Applications remain independently deployable and recoverable. A Mesh outage must not automatically destroy application-owned data or make independently operable applications irrecoverable.

The dependency graph is descriptive and coordinating; it should not become a hidden runtime dependency for every request unless a future contract explicitly requires that behavior and provides a failure-safe design.

## Privacy and security posture

The initial API listens on loopback by default. No public exposure is authorized by this repository foundation. Authentication and authorization adapters are intentionally not faked; production use requires a real GoreeCloud Identity and Wardveil-backed trust model.

Mesh metadata must remain data-minimized. It must not become a centralized store of private application content merely because it connects applications.

## Resilience

Atomic JSON persistence provides crash-safe single-file replacement for the first milestone. Everkeep integration will define backup classification, restore evidence, migration/export requirements, corruption handling, and recovery acceptance before Stable qualification.

## Planned evolution

1. Identity-backed service identities and authenticated registration.
2. Wardveil relationship-policy and evidence integration.
3. Privacy Shield metadata classification and retention controls.
4. Everkeep backup/export/restore evidence.
5. Monitoring health adapters and health-evidence provenance.
6. Gateway and Network adapters for connection enforcement state.
7. Durable event journal and external event subscribers.
8. Versioned contract catalog and compatibility evaluation.
9. Mesh Console using the current Stable Glaze UI contract.
10. Multi-node/federated coordination only after consistency, recovery, and failure semantics are explicitly approved.
