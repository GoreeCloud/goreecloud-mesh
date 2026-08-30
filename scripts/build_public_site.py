#!/usr/bin/env python3
from pathlib import Path
import shutil

ROOT = Path(__file__).resolve().parents[1]
DIST = ROOT / "dist"
FILES = {
    ROOT / "website" / "index.html": DIST / "index.html",
    ROOT / "website" / "style.css": DIST / "style.css",
    ROOT / "website" / "glaze-ui-2.0.0.css": DIST / "glaze-ui-2.0.0.css",
    ROOT / "website" / "app.js": DIST / "app.js",
    ROOT / "website" / "_headers": DIST / "_headers",
    ROOT / "website" / "robots.txt": DIST / "robots.txt",
    ROOT / "website" / "sitemap.xml": DIST / "sitemap.xml",
    ROOT / "website" / "assets" / "goreecloud-mesh-mark.svg": DIST / "assets" / "goreecloud-mesh-mark.svg",
}

if DIST.exists():
    shutil.rmtree(DIST)
for source, target in FILES.items():
    if not source.is_file() or source.is_symlink():
        raise SystemExit(f"missing or unsafe public source: {source.relative_to(ROOT)}")
    target.parent.mkdir(parents=True, exist_ok=True)
    shutil.copy2(source, target)
print(f"Built Mesh Center public site: {len(FILES)} files -> dist/")
