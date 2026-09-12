# GoreeCloud Mesh Architecture

## Governing role

GoreeCloud Mesh is the GoreeCloud Integral Platform System for **private networking, connectivity, reachability, service discovery, and service communication**.

Mesh answers questions such as:

- How can an authorized GoreeCloud device, service, user, site, or infrastructure component securely reach another?
- Which service endpoints and reachability capabilities are currently advertised?
- Which approved transport or network path can carry a request or bounded evidence record?
- What network/reachability context is available to an authorized consumer?
- Which service-communication dependencies may be affected when a path or endpoint becomes unavailable?

Mesh transports information. It does not become authoritative for the meaning or synchronization state of the information it carries.

## GoreeCloud Sync boundary

GoreeCloud Sync is a separate Integral Platform System. Sync owns the platform-wide synchronization and state-coordination responsibilities that were historically mixed into parts of the Mesh design.

Sync, not Mesh, owns authority for:

- cross-device and application-to-application synchronization;
- offline-first state replication;
- change tracking and delta synchronization;
- synchronization queues, retry/recovery behavior, and synchronization health;
- synchronization policy and selective/permission-aware replication;
- shared synchronization event contracts and event propagation;
- conflict detection and resolution;
- authoritative-version and reconciliation decisions;
- cross-device continuity and synchronization state restoration.

A Mesh path may carry Sync traffic or Sync evidence. That transport does not transfer Sync authority into Mesh.

## Other authority boundaries

Mesh does not become a universal administrator or a substitute for another system.

- **GoreeCloud Manager** remains the central administration and operational-management authority.
- **GoreeCloud Monitoring** remains authoritative for monitoring collection and health evidence.
- **GoreeCloud Network** remains authoritative for network connectivity implementation and network policy where that product boundary applies.
- **GoreeCloud Gateway** remains authoritative for ingress, proxying, traffic routing, and service-access enforcement at its boundary.
- **GoreeCloud Identity** remains authoritative for identities, authentication, authorization, SSO, accounts, devices, sessions, credentials, and delegated authority.
- **Wardveil Security** defines platform security, trust, protection, and security-evidence contracts.
- **Privacy Shield** defines privacy-control, data-minimization, purpose/retention, and privacy-evidence contracts.
- **Everkeep** defines resilience, recovery, preservation, portability, continuity, succession, and recovery-evidence contracts.
- **Glaze UI** defines the design and interaction contract for Mesh Center and other user-facing Mesh surfaces.
- **GoreeCloud Sync** defines synchronization, state coordination, reconciliation, and cross-device continuity contracts.

Mesh may expose or transport bounded information from those systems through documented contracts. It must not manufacture, merge, strengthen, or silently reinterpret their authority.

## Current Development source transition

The repository was originally built around a broader coordination-fabric model. Current source therefore contains Registry, Graph, Policy, and Events functions whose historical names and behavior extend beyond the adopted Mesh connectivity/reachability target.

Those functions remain real Development-stage source and must not be described as already removed. They are subject to controlled decomposition or reclassification:

- functionality necessary for service discovery, reachability, endpoint metadata, transport, network-path context, or bounded evidence delivery may remain in Mesh;
- synchronization state, replication, reconciliation, cross-device continuity, shared synchronization event contracts, and conflict resolution belong to GoreeCloud Sync;
- administration and broad operational control belong to GoreeCloud Manager;
- producer-domain security, privacy, identity, recovery, and presentation truth remains with the applicable producer authority.

No compatibility-preserving migration may use the existence of historical Mesh code as justification to keep obsolete authority boundaries indefinitely.

## Current source components

### Mesh Registry

The current Registry stores service identity, kind, version, endpoint metadata, capabilities, declared dependencies, health context, labels, and platform-conformance state.

For the adopted architecture, the Registry is primarily a **service discovery and reachability registry**. Registration metadata must remain minimized and must not contain application payloads, message content, user files, credentials, tokens, synchronization state, or unnecessary private activity.

### Mesh Graph

The existing Graph builds explicit relationship and dependency edges and supports reverse dependency-impact traversal. This remains a Development source capability.

Graph information may continue where it describes reachability, endpoint/service dependencies, or transport impact. Application-state dependency orchestration, synchronization topology, and reconciliation semantics must migrate to the appropriate authority rather than expanding Mesh.

### Mesh Policy

The existing policy evaluator checks registered service relationships and advertised capabilities and fails closed on unknown/unavailable sources or targets and missing enabled relationships.

Mesh Policy may serve bounded connectivity/service-communication admission metadata, but it is not a replacement for Identity authorization, Gateway access controls, Network policy, application permissions, Privacy Shield decisions, Wardveil Security enforcement, or GoreeCloud Sync synchronization policy.

### Mesh Events

The existing in-process event bus publishes Registry and relationship lifecycle events. Those local lifecycle signals are implementation behavior, not a grant of platform-wide state-coordination authority.

Durable synchronization event propagation, shared synchronization event contracts, replay/reconciliation behavior, and cross-application state propagation belong to GoreeCloud Sync. Mesh may later provide an approved transport path for such events without becoming their semantic authority.

### Mesh Nodes and Connections

Nodes and connections represent service participants and explicit communication relationships. Their target role is to support discovery, reachability, transport, and observable service-communication context.

They must not silently become a second synchronization graph or a general-purpose application workflow engine.

### Mesh Evidence Transport

Mesh currently contains evidence-envelope, producer binding, receipt, and consumer-view source foundations. Evidence transport may remain a Mesh service-communication capability when it is bounded, authenticated, minimized, and producer-authoritative.

Transport validity never turns Mesh into the authority for security, privacy, identity, recovery, synchronization, or Glaze UI claims.

### Mesh Center

Mesh Center is the planned Glaze UI surface for Mesh-specific connectivity, discovery, reachability, service communication, evidence transport, health/path context, and migration/conformance visibility.

GoreeCloud Manager remains the broader administration authority. Mesh Center must not become a competing universal administration plane.

## Discovery and reachability

Service discovery returns registered services/endpoints that advertise the requested capability and are not explicitly unavailable. Discovery does not grant authorization.

A caller must still satisfy all applicable Identity, Gateway, Network, application, Wardveil Security, Privacy Shield, Manager, and Sync boundaries. Reachability means that a path or endpoint is available; it does not mean the caller is authorized to use application data or synchronize state.

## Failure isolation

Mesh must improve connectivity without becoming an unnecessary universal failure domain. Applications remain independently deployable and recoverable where their contracts allow it.

A Mesh outage must not automatically destroy application-owned data, erase synchronization state owned by Sync, or make independently operable applications irrecoverable. Any hard runtime dependency on Mesh requires explicit availability, degradation, recovery, and failure-safe design.

## Privacy and security posture

The current API listens on loopback by default. No public runtime exposure is authorized by this source foundation.

Production use requires accepted GoreeCloud Identity, Wardveil Security, Privacy Shield, and applicable Gateway/Network controls. Mesh metadata must remain data-minimized and must not become a centralized store of private application content merely because it connects services.

## Resilience

Atomic JSON persistence provides crash-safe single-file replacement for current Development stores. Everkeep integration must define backup classification, restore evidence, migration/export requirements, corruption handling, and recovery acceptance before Stable qualification.

GoreeCloud Sync is not a backup authority, and Mesh is not a synchronization or backup authority.

## Platform Contract 0.3 boundary

The repository-root `goreecloud.platform.yaml` is the controlling machine-readable declaration under GoreeCloud Platform Contract 0.3.

Mesh implements the Mesh authority itself, so the Mesh self-slot is `not-applicable-justified`. GoreeCloud Sync is a distinct required platform-system dimension and remains blocked until an explicit Mesh-to-Sync reachability/transport integration is implemented, validated, and accepted.

The repository remains Development/nonconformant. Contract adoption does not establish production runtime acceptance.

## Planned migration and evolution

1. Preserve and harden private networking, reachability, discovery, service communication, and bounded evidence-transport functions that belong in Mesh.
2. Inventory Registry/Graph/Policy/Event source behavior against the adopted eight-system authority matrix.
3. Move or replace synchronization/state-coordination semantics with explicit GoreeCloud Sync contracts rather than duplicate them in Mesh.
4. Integrate Identity-backed service identities and authenticated reachability/service registration.
5. Integrate Wardveil trust and security evidence at the transport/reachability boundary.
6. Integrate Privacy Shield metadata classification, minimization, and retention controls.
7. Complete Everkeep backup/export/restore evidence for Mesh-owned state.
8. Complete Monitoring, Gateway, and Network adapters for observed connectivity/enforcement state where applicable.
9. Define a bounded Mesh-to-Sync transport/reachability contract with no authority transfer.
10. Reconcile Mesh Center to the current Stable Glaze UI contract and the Manager administration boundary.
11. Add multi-node/federated reachability only after consistency, recovery, security, and failure semantics are explicitly approved.

Every migration must preserve verified current behavior long enough for a controlled transition, but obsolete coordination authority must not survive merely for compatibility convenience.
