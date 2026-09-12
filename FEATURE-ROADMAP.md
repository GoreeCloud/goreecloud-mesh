# GoreeCloud Mesh — Feature Roadmap

**Status:** Active roadmap control  
**As of:** 2026-09-11  
**Authoritative project record:** Project Specification — Mesh  
**Canonical repository:** GoreeCloud/goreecloud-mesh  
**Drive control:** `GoreeCloud/Feature Roadmap/GoreeCloud Mesh/FEATURE-ROADMAP.docx`

## Purpose

This file is the repository-side feature roadmap control for GoreeCloud Mesh. It records current planned and recommended feature work without replacing the authoritative project record, implementation evidence, release gates, or GoreeCloud Tasks Management.

## Roadmap

| ID | Feature / obligation | Priority | Current state |
| --- | --- | --- | --- |
| FR-001 | Reconcile and maintain every current planned or recommended GoreeCloud Mesh feature from the authoritative project record and verified repository evidence in this roadmap. | High | Ongoing control |
| FR-002 | Move actionable feature obligations into GoreeCloud Tasks Management when required, preserving priority, dependency, and lifecycle disposition. | High | Ongoing control |
| FR-003 | Do not mark features implemented, complete, cancelled, or superseded without authoritative evidence and synchronized repository/Drive roadmap updates. | High | Ongoing control |
| FR-004 | Complete production-grade GoreeCloud Identity and Wardveil trust for Mesh service access: validated Identity verifier, accepted least-privilege scopes, approved signed/verified service registration where applicable, Wardveil security-evidence integration, authenticated API enforcement, and protected administration. | P0 | In Development. Source-level Principal/Verifier contracts are integrated. Draft PR #41 adds a bounded `mesh.events.read` Identity-authenticated external lifecycle-event consumer and passed exact-head Mesh CI `34440782436` plus Platform Contract `34440782819` at `bc83dcb832be77f59c79a819d7705c8a4ce3636e`. Live Identity issuance/JWKS trust, production authorization, Wardveil runtime acceptance, protected administration, and production acceptance remain pending. |
| FR-005 | Complete Monitoring, Gateway, and Network production adapters with accepted evidence transport, enforcement-state visibility, mismatch detection, runtime provenance validation, freshness handling, and controlled coordination without granting Mesh universal administrator authority. | P1 | Partially Source Implemented. Authority-specific source contracts and fail-closed evidence freshness semantics are integrated; production adapters and target-runtime acceptance remain pending. |
| FR-006 | Complete Privacy Shield and Everkeep runtime integration for privacy-safe metadata, retention enforcement, telemetry/logging boundaries, backup/export/restore, corruption handling, portability, recovery evidence, and continuity acceptance. | P1 | Partially Source Implemented. Metadata classification, payload exclusion, retention classes, recovery-evidence structures, and fail-closed source-level recovery-readiness contracts are integrated. Runtime retention/privacy controls, production backup/restore, portability validation, and accepted evidence remain pending. |
| FR-007 | Implement Mesh Center using the current Stable Glaze UI with Registry, Graph, Policy, Events, Nodes, Connections, dependency impact, compatibility and conformance visibility, health context, accessibility, responsive behavior, and explicit degraded/offline/error states. | P1 | Planned. The public Mesh website has separate production-verified Glaze evidence, but that does not implement or accept the Mesh Center product/admin experience. |
| FR-008 | Implement durable Mesh Events and external delivery: durable event journal, replay/checkpoints, bounded retention, delivery guarantees, ordering/idempotency semantics, external subscribers, cross-process/cross-host transport, retry/backoff, dead-letter handling where approved, and explicit Privacy Shield/Wardveil/Identity/Everkeep acceptance. | P1 | In Development foundation only. The merged/local event foundation and Draft PR #41 provide bounded in-process/live SSE behavior with explicit no-replay/no-durable-offset semantics. Durable storage, replay, delivery guarantees, cross-host transport, retained subscriber state, and production acceptance remain pending. |
| FR-009 | Implement the versioned contract catalog, compatibility analysis, controlled breaking-change transitions, lifecycle coordination, and dependency compatibility visibility required by Milestone 6. | P1 | Planned. `/v1/` remains the current API namespace; no complete compatibility-analysis/catalog subsystem is established as accepted. |
| FR-010 | Complete the mandatory GoreeCloud Manager and GoreeCloud Mesh operational relationship with bounded documented APIs, authority separation, authentication, privacy/security/recovery contracts, runtime evidence, and production acceptance. | P1 | Planned/incomplete. The authoritative Mesh specification requires the Manager relationship for Stable qualification; current source-level relationships do not constitute production acceptance. |
| FR-011 | Define and, only when justified by validated requirements, implement the next persistence architecture for scale, consistency, concurrency, recovery, migration, backup, and multi-node operation while preserving replaceability and fail-closed integrity. | P2 | Planned/conditional. The current atomic JSON persistence remains the Development foundation. A database or distributed-state implementation must not be selected before the explicit requirements named in the authoritative specification are established. |
| FR-012 | Complete production-readiness acceptance: target-environment runtime validation, authenticated service identity, current seven-system platform acceptance, privacy-conscious observability, recovery and rollback proof, dependency/degraded-mode testing, failure isolation, performance evaluation, deployment documentation, explicit production approval, and exact accepted-revision evidence. | P0 | Production/Stable gate — Pending. GoreeCloud Mesh remains Development; no release, production deployment, or Stable qualification is implied by source validation or website publication. |

## Maintenance and synchronization

This roadmap and the corresponding Drive `FEATURE-ROADMAP.docx` must remain materially synchronized with one another and with the authoritative project or service record. Update both copies whenever feature scope, priority, dependency, implementation status, cancellation, supersession, recommendation, or verification state materially changes.

No feature may be represented as complete or Stable solely because it appears in this roadmap. Completion and lifecycle claims require the applicable authoritative implementation, validation, review, release, and production evidence.

## Reconciliation rule

At each material feature change, reconcile this roadmap against the current authoritative project record, repository implementation state, applicable platform-system requirements, and GoreeCloud Tasks Management. Missing obligations, stale status, duplicated work, roadmap drift, or undocumented disposition changes are defects to correct.
