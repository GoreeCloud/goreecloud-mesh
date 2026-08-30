#!/usr/bin/env python3
from hashlib import sha1
from pathlib import Path
import re
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
SITE = ROOT / "website"
DIST = ROOT / "dist"
REQUIRED = (
    "index.html", "style.css", "glaze-ui-2.0.0.css", "app.js", "_headers",
    "robots.txt", "sitemap.xml", "assets/goreecloud-mesh-mark.svg",
)

for relative in REQUIRED:
    path = SITE / relative
    if not path.is_file() or path.is_symlink():
        raise SystemExit(f"missing or unsafe Mesh Center source: {relative}")

html = (SITE / "index.html").read_text(encoding="utf-8")
css = (SITE / "style.css").read_text(encoding="utf-8")
headers = (SITE / "_headers").read_text(encoding="utf-8")
mark = (SITE / "assets/goreecloud-mesh-mark.svg").read_bytes()

def git_blob_sha(data: bytes) -> str:
    return sha1(f"blob {len(data)}\0".encode() + data).hexdigest()

for marker in (
    "Mesh Center — GoreeCloud", "GoreeCloud Mesh", "Mesh Center · Weave",
    "Coordinate the platform.", "Registry", "Graph", "Policy", "Events",
    "Evidence plane", "Authority boundaries", "GoreeCloud Identity",
    "Wardveil Security", "Privacy Shield", "Everkeep", "Glaze UI",
    "Production acceptance is separate", "Not yet qualified",
    'name="goreecloud-glaze-ui" content="2.0.0"', 'data-glaze-ui="2.0.0"',
):
    if marker not in html:
        raise SystemExit(f"Mesh Center marker missing: {marker}")

for forbidden in (
    "production-ready", "Production Ready", "Stable platform", "fully deployed",
    "authority transfer", "data:image", "raw.githubusercontent.com",
):
    if forbidden in html:
        raise SystemExit(f"Mesh Center publishes forbidden or misleading marker: {forbidden}")

if git_blob_sha(mark) != "0b2c6881668ce319081390b217f6d59b4298dd4d":
    raise SystemExit("Mesh Center Weave derivative does not match the approved canonical Git blob")

for src in re.findall(r'(?:src|href)=["\']([^"\']+)', html):
    if src.startswith(("http://", "https://")) and "github.com/GoreeCloud/" not in src and "goreecloud.com/" not in src:
        raise SystemExit(f"unauthorized external link/resource: {src}")

for directive in (
    "Content-Security-Policy:", "default-src 'self'", "connect-src 'none'",
    "frame-ancestors 'none'", "X-Content-Type-Options: nosniff",
):
    if directive not in headers:
        raise SystemExit(f"Mesh Center security header marker missing: {directive}")

for marker in ("prefers-reduced-motion", "prefers-reduced-transparency", "forced-colors", "--g-touch:48px"):
    if marker not in css:
        raise SystemExit(f"Mesh Center accessibility/responsiveness marker missing: {marker}")

subprocess.run([sys.executable, str(ROOT / "scripts" / "build_public_site.py")], check=True)
for relative in REQUIRED:
    if (DIST / relative).read_bytes() != (SITE / relative).read_bytes():
        raise SystemExit(f"isolated Mesh Center artifact drifted from reviewed source: {relative}")
print("Mesh Center public website validation passed")
