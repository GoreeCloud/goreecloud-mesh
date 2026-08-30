# Mesh Center website

GoreeCloud Mesh owns the standalone **Mesh Center** public website source in this repository.

## Target address

- Public hostname: `https://mesh.goreecloud.com/`
- Cloudflare Pages project name: `goreecloud-mesh`
- Production branch: `main`
- Framework preset: `None`
- Build command: `python scripts/build_public_site.py`
- Build output directory: `dist`
- Root directory: blank

The Cloudflare project and DNS/custom-domain binding are deployment operations separate from source implementation. The website must not be described as publicly deployed until those operations and production verification are complete.

## Source layout

- `website/` — reviewed public source
- `website/assets/goreecloud-mesh-mark.svg` — byte-identical consumer derivative of the approved Weave asset from `GoreeCloud/goreecloud-branding-assets`
- `scripts/build_public_site.py` — creates the isolated `dist/` artifact
- `scripts/validate_public_site.py` — validates branding provenance, Glaze UI markers, security headers, truth boundaries, and artifact identity
- `.github/workflows/validate-website.yml` — exact-revision CI gate

## Public truth boundary

Mesh Center may describe implemented GoreeCloud Mesh source capabilities and accepted source/CI evidence. It must not convert source presence, a successful build, a preview deployment, or another system's evidence into production acceptance.

GoreeCloud Mesh remains coordination-only. GoreeCloud Identity, Wardveil Security, Privacy Shield, Everkeep, Glaze UI, Gateway, Network, Monitoring, and application-specific authorities retain their documented domains. `authority_transfer=false` remains a substantive platform rule, not a presentation label.
