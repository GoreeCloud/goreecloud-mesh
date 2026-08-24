# Platform source attestations

GoreeCloud Mesh uses source attestations to bind a bilateral platform integration manifest to the exact producer revision and exact manifest bytes that were reviewed.

A source attestation is provenance evidence only. It does not prove that Glaze UI, Wardveil Security, Privacy Shield, or Everkeep is integrated at runtime, does not transfer any authority into Mesh, does not authorize deployment or release, and does not satisfy a Stable acceptance gate by itself.

## Required provenance

A validated attestation records:

- the mandatory platform identifier;
- the canonical producer repository from the Mesh platform catalog;
- a full lowercase 40-character Git revision;
- the canonical integration-manifest path from the Mesh platform catalog;
- a lowercase SHA-256 digest of the exact manifest bytes;
- the source validation state;
- the validation workflow name and positive workflow-run identifier for validated evidence; and
- the UTC observation time.

Pending or blocked records may exist without completed workflow evidence. A record may not claim validated state without a workflow name and positive workflow-run identifier.

## Fail-closed boundaries

Mesh rejects an attestation when the platform, repository, or manifest path disagrees with the canonical catalog; when revision or digest identity is malformed; when the state is unknown; when the observation time is absent; when a validated record lacks workflow evidence; or when the record implies runtime or Stable acceptance.

The attestation contract intentionally does not include credentials, tokens, private keys, recovery material, private application content, or other sensitive information.

## Digest binding

`ManifestSHA256` computes the lowercase SHA-256 digest from the exact manifest bytes. Consumers of this contract must calculate the digest from the reviewed bytes rather than from reformatted, reconstructed, or semantically equivalent JSON.

This permits Mesh to distinguish source provenance from runtime evidence while retaining an auditable link to the exact bilateral contract that was validated.
