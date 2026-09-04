#!/usr/bin/env python3
from pathlib import Path
import hashlib
import json
import os
import shutil
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[1]
SOURCE = ROOT / "website"
DIST = ROOT / "dist"
LOCK = json.loads((SOURCE / "glaze.lock.json").read_text(encoding="utf-8"))
PUBLIC_FILES = ("index.html", "404.html", "_headers", "robots.txt")
LOCAL_ASSETS = ("mesh-theme.css", "site.css", "site.js")

if LOCK.get("version") != "1.1.0" or LOCK.get("product") != "GLAZE UI V1.1":
    raise SystemExit("Glaze consumer lock must target GLAZE UI V1.1 / 1.1.0 Stable")
if LOCK.get("release_commit") != "15cc76d2bcd4065552dc31c77145b63f34d9e7b2":
    raise SystemExit("unexpected GLAZE UI V1.1 Stable release commit")
if LOCK.get("entrypoint") != "css/glaze-v1.1.0.css":
    raise SystemExit("unexpected GLAZE UI V1.1 Stable entrypoint")

def git_blob_sha(data: bytes) -> str:
    prefix = b"blob " + str(len(data)).encode("ascii") + b"\0"
    return hashlib.sha1(prefix + data).hexdigest()

def require_file(path: Path) -> Path:
    if not path.is_file() or path.is_symlink():
        raise SystemExit(f"missing or unsafe public source: {path.relative_to(ROOT)}")
    return path

def checkout_glaze_source() -> tuple[Path, tempfile.TemporaryDirectory[str] | None]:
    source_root = os.environ.get("GLAZE_UI_SOURCE")
    if source_root:
        return Path(source_root), None

    # Cloudflare Pages and other isolated builders should perform one immutable
    # Git fetch instead of thirteen independent unauthenticated raw-file requests.
    # Every consumed file is still verified below by its canonical Git blob SHA.
    temp = tempfile.TemporaryDirectory(prefix="goreecloud-glaze-ui-")
    checkout = Path(temp.name) / "source"
    checkout.mkdir()
    repository = f"https://github.com/{LOCK['repository']}.git"
    try:
        subprocess.run(
            ["git", "-C", str(checkout), "init", "--quiet"],
            check=True,
            stdout=subprocess.DEVNULL,
            stderr=subprocess.PIPE,
            text=True,
        )
        subprocess.run(
            [
                "git",
                "-C",
                str(checkout),
                "fetch",
                "--quiet",
                "--depth=1",
                repository,
                LOCK["release_commit"],
            ],
            check=True,
            stdout=subprocess.DEVNULL,
            stderr=subprocess.PIPE,
            text=True,
        )
        subprocess.run(
            ["git", "-C", str(checkout), "checkout", "--quiet", "--detach", "FETCH_HEAD"],
            check=True,
            stdout=subprocess.DEVNULL,
            stderr=subprocess.PIPE,
            text=True,
        )
    except (OSError, subprocess.CalledProcessError) as error:
        temp.cleanup()
        detail = getattr(error, "stderr", None) or str(error)
        raise SystemExit(f"unable to fetch locked GLAZE UI source: {detail.strip()}") from None
    return checkout, temp

def read_glaze(source_root: Path, upstream_path: str, expected_sha: str) -> bytes:
    data = require_file(source_root / upstream_path).read_bytes()
    actual_sha = git_blob_sha(data)
    if actual_sha != expected_sha:
        raise SystemExit(f"Glaze integrity mismatch for {upstream_path}: {actual_sha} != {expected_sha}")
    return data

if DIST.exists():
    shutil.rmtree(DIST)
(DIST / "assets" / "glaze").mkdir(parents=True)

for name in PUBLIC_FILES:
    shutil.copy2(require_file(SOURCE / name), DIST / name)
if (SOURCE / "sitemap.xml").exists():
    shutil.copy2(require_file(SOURCE / "sitemap.xml"), DIST / "sitemap.xml")
for name in LOCAL_ASSETS:
    shutil.copy2(require_file(SOURCE / name), DIST / "assets" / name)
for asset in sorted((SOURCE / "assets").iterdir()):
    if asset.is_symlink() or not asset.is_file():
        raise SystemExit(f"unsafe public asset: {asset.relative_to(ROOT)}")
    shutil.copy2(asset, DIST / "assets" / asset.name)

glaze_source, temporary_glaze_source = checkout_glaze_source()
try:
    for upstream_path, expected_sha in LOCK["files"].items():
        name = Path(upstream_path).name
        (DIST / "assets" / "glaze" / name).write_bytes(read_glaze(glaze_source, upstream_path, expected_sha))
finally:
    if temporary_glaze_source is not None:
        temporary_glaze_source.cleanup()

print(
    f"Built public site with {LOCK['product']} / {LOCK['version']} Stable "
    f"pinned to {LOCK['release_commit']}"
)
