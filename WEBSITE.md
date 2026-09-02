# Mesh Center website

GoreeCloud Mesh owns the standalone **Mesh Center** public website source in this repository.

## Target address

- Intended public hostname: `https://mesh.goreecloud.com/`
- Cloudflare Pages project name: `goreecloud-mesh`
- Production branch: `main`
- Framework preset: `None`
- Build command: `python scripts/build_public_site.py`
- Build output directory: `dist`
- Root directory: blank

The Cloudflare project, DNS/custom-domain binding, HTTPS verification, and production acceptance remain deployment operations separate from source implementation. A successful source build or preview does not establish public production availability.

## Glaze UI baseline

The website is a ground-up consumer of **Glaze UI 2.2.0 Stable**. `website/glaze.lock.json` pins the Stable tag, promotion commit, and each browser CSS Git blob. CI checks out that exact Glaze revision and the build verifies every copied blob before creating `dist/`. Browsers consume only local built assets.

## Source layout

- `website/` — reviewed public source
- `website/assets/goreecloud-mesh-mark.svg` — byte-identical consumer derivative of the approved **Interlace** asset (`5362a52bd9fb38379f083a4d894934ed1acf9b67`)
- `website/glaze.lock.json` — Glaze UI 2.2.0 Stable consumer lock
- `scripts/build_public_site.py` — produces the isolated `dist/` artifact
- `scripts/validate_public_site.py` — validates branding, Glaze provenance, accessibility hooks, security boundaries, public truth, and built-asset integrity
- `.github/workflows/validate-website.yml` — exact-revision website acceptance gate

## Public truth boundary

Mesh Center may describe implemented Mesh source capabilities and accepted source/CI evidence. It must not convert source presence, a successful build, a preview deployment, or another system's evidence into production acceptance.

GoreeCloud Mesh remains coordination-only. Identity, Wardveil Security, Privacy Shield, Everkeep, Glaze UI, Gateway, Network, Monitoring, and application-specific authorities retain their documented domains. `authority_transfer=false` remains a substantive platform invariant, not a presentation label.
