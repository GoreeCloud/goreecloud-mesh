# GoreeCloud Mesh — Durable Replay Ownership Boundary

## Status

**Development source contract.** This document describes repository-local replay ownership for the durable subscriber checkpoint foundation. It does not establish deployed delivery, distributed consensus, remote acknowledgement, exactly-once processing, or production acceptance.

## Ownership model

`ReplaySubscriber` already serializes replay passes that share one `DurableSubscriberCheckpoints` instance through an in-process ownership gate.

On supported Unix platforms, replay ownership now also uses a kernel-held advisory `flock` on a persistent private rendezvous file derived from the checkpoint-state path:

`<subscriber-checkpoint-path>.replay.lock`

The lock file is a regular mode-0600 file. Its **presence does not mean the replay is owned**. Ownership exists only while a process holds the kernel lock on the open file descriptor. This avoids treating a stale filesystem marker as live authority.

A process exit or crash closes its file descriptors and releases the kernel lock. The rendezvous file intentionally remains in place for future ownership attempts.

## Fail-closed file boundary

The lock path is opened without following a final symlink. Existing lock files must be regular files with mode 0600. Symlink-backed, non-regular, or non-private lock paths fail closed.

The containing directory is created with a restrictive mode when needed. The replay ownership file carries no event payload, subscriber payload, credential, or acknowledgement content.

## Context and contention

Process-lock acquisition is non-blocking at the kernel boundary and retried with a short bounded polling interval so the caller's context remains authoritative. A cancelled or expired context stops waiting and returns the context error.

The existing in-process gate is acquired before the process lock. If process-lock acquisition fails, the in-process gate is released so the current runtime cannot deadlock itself.

Unsupported platforms fail closed instead of silently reverting to in-process-only ownership for durable replay.

## Delivery semantics boundary

Cross-process replay ownership prevents two cooperative processes using the same checkpoint-state path from intentionally running a replay pass concurrently. It does **not** establish exactly-once processing.

In particular, a handler may complete its side effect and then fail before the durable subscriber checkpoint advances, or checkpoint persistence may fail after handler success. A later replay may therefore present the same event again. Handlers that require stronger semantics must be idempotent or use a separately designed transaction/deduplication contract.

This ownership primitive also does not prove remote receipt, external side-effect persistence, acknowledgement transport, retry/backoff, dead-letter handling, distributed leases across hosts without a shared lock-capable filesystem, federation, or production availability.

## Production boundary

Production acceptance still requires the selected storage/filesystem semantics to be verified in the target environment, multi-process failure injection, crash/restart tests, operational observability, subscriber-specific delivery policy, and independently accepted authentication/authorization. Kernel replay ownership is one durability primitive, not a complete message-delivery system.
