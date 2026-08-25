# Runtime evidence freshness

GoreeCloud Mesh does not invent a platform-wide freshness duration for Glaze UI, Wardveil Security, Privacy Shield, or Everkeep evidence.

Validated runtime evidence must instead carry a producer-declared `valid_until` timestamp derived from the authoritative producer contract or applicable policy. Mesh treats validated evidence as current only while the evaluation time is not after that timestamp.

The producer remains authoritative for how the validity deadline is determined. Mesh only validates the supplied boundary, preserves it with the evidence record, and fails closed after expiration.

A validated record is rejected when `valid_until` is absent, is not later than `observed_at`, or is already expired at evaluation time. Pending and blocked records may omit `valid_until` because they cannot satisfy a Stable gate.

This mechanism does not create runtime acceptance, production acceptance, deployment authorization, release authorization, authority transfer, or Stable qualification.
