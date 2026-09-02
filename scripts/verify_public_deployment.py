#!/usr/bin/env python3
"""Verify the deployed GoreeCloud Mesh website against reviewed repository source."""

from __future__ import annotations

import argparse
from hashlib import sha256
from pathlib import Path
import ssl
import time
from urllib.error import HTTPError, URLError
from urllib.parse import urljoin, urlparse
from urllib.request import HTTPRedirectHandler, HTTPSHandler, Request, build_opener

ROOT = Path(__file__).resolve().parents[1]
SOURCE = ROOT / "website"
PRODUCTION_URL = "https://mesh.goreecloud.com/"
EXPECTED_HOST = "mesh.goreecloud.com"
MAX_BODY_BYTES = 2_000_000
REQUEST_TIMEOUT_SECONDS = 12
RETRY_ATTEMPTS = 12
RETRY_DELAY_SECONDS = 10

REQUIRED_HEADER_MARKERS = {
    "content-security-policy": ("default-src 'self'", "frame-ancestors 'none'"),
    "strict-transport-security": ("max-age=31536000",),
    "x-content-type-options": ("nosniff",),
    "x-frame-options": ("DENY",),
    "referrer-policy": ("no-referrer",),
}


class SameHostRedirectHandler(HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        parsed = urlparse(newurl)
        if parsed.scheme != "https" or parsed.hostname != EXPECTED_HOST:
            raise URLError(f"redirect outside canonical Mesh host rejected: {newurl}")
        return super().redirect_request(req, fp, code, msg, headers, newurl)


def validate_configuration() -> None:
    parsed = urlparse(PRODUCTION_URL)
    if parsed.scheme != "https" or parsed.hostname != EXPECTED_HOST:
        raise SystemExit("production verifier must remain fixed to the canonical HTTPS Mesh host")
    if parsed.username or parsed.password or parsed.query or parsed.fragment:
        raise SystemExit("production verifier target must not contain credentials, query, or fragment")
    for relative in ("index.html", "robots.txt", "404.html", "assets/goreecloud-mesh-mark.svg"):
        path = SOURCE / relative
        if not path.is_file() or path.is_symlink():
            raise SystemExit(f"missing or unsafe reviewed website source: {relative}")


def fetch(path: str) -> tuple[int, str, dict[str, str], bytes]:
    if not path.startswith("/"):
        raise ValueError("remote path must be origin-rooted")
    url = urljoin(PRODUCTION_URL, path.lstrip("/"))
    parsed = urlparse(url)
    if parsed.scheme != "https" or parsed.hostname != EXPECTED_HOST:
        raise ValueError(f"remote URL escaped canonical Mesh host: {url}")

    request = Request(
        url,
        headers={
            "User-Agent": "GoreeCloud-Mesh-Deployment-Verifier/1.0",
            "Accept": "*/*",
            "Accept-Encoding": "identity",
            "Cache-Control": "no-cache",
        },
        method="GET",
    )
    opener = build_opener(
        SameHostRedirectHandler(),
        HTTPSHandler(context=ssl.create_default_context()),
    )
    try:
        with opener.open(request, timeout=REQUEST_TIMEOUT_SECONDS) as response:
            body = response.read(MAX_BODY_BYTES + 1)
            if len(body) > MAX_BODY_BYTES:
                raise RuntimeError(f"response too large: {path}")
            return (
                response.status,
                response.geturl(),
                {k.lower(): v for k, v in response.headers.items()},
                body,
            )
    except HTTPError as error:
        body = error.read(MAX_BODY_BYTES + 1)
        return (
            error.code,
            error.geturl(),
            {k.lower(): v for k, v in error.headers.items()},
            body,
        )


def expected_bytes(path: str) -> bytes:
    relative = "index.html" if path == "/" else path.lstrip("/")
    return (SOURCE / relative).read_bytes()


def verify_once() -> list[str]:
    errors: list[str] = []

    for path in ("/", "/robots.txt", "/assets/goreecloud-mesh-mark.svg"):
        try:
            status, final_url, headers, body = fetch(path)
        except (URLError, RuntimeError, ValueError) as error:
            errors.append(f"{path}: network verification failed: {error}")
            continue

        if status != 200:
            errors.append(f"{path}: expected HTTP 200, got {status}")
            continue
        if urlparse(final_url).hostname != EXPECTED_HOST:
            errors.append(f"{path}: final host drifted to {final_url}")
        expected = expected_bytes(path)
        if body != expected:
            errors.append(
                f"{path}: deployed bytes differ from reviewed source; "
                f"expected sha256={sha256(expected).hexdigest()} "
                f"deployed sha256={sha256(body).hexdigest()}"
            )

        if path == "/":
            text = body.decode("utf-8", errors="replace")
            for marker in (
                "Interlace · coordination fabric",
                'alt="GoreeCloud Mesh Interlace mark"',
                '<link rel="canonical" href="https://mesh.goreecloud.com/">',
            ):
                if marker not in text:
                    errors.append(f"/: missing expected Interlace deployment marker: {marker}")
            if "noindex" in headers.get("x-robots-tag", "").lower():
                errors.append("/: production site unexpectedly sends X-Robots-Tag: noindex")
            for header, markers in REQUIRED_HEADER_MARKERS.items():
                value = headers.get(header, "")
                if not value:
                    errors.append(f"/: missing required response header: {header}")
                    continue
                for marker in markers:
                    if marker.lower() not in value.lower():
                        errors.append(f"/: {header} missing required value: {marker}")

    try:
        status, _, _, body = fetch("/__goreecloud_mesh_missing_verification_path__")
        if status != 404:
            errors.append(f"404 behavior: expected HTTP 404, got {status}")
        expected_404 = expected_bytes("/404.html")
        if body != expected_404:
            errors.append(
                "404 behavior: deployed 404 body differs from reviewed 404.html; "
                f"expected sha256={sha256(expected_404).hexdigest()} "
                f"deployed sha256={sha256(body).hexdigest()}"
            )
    except (URLError, RuntimeError, ValueError) as error:
        errors.append(f"404 behavior: network verification failed: {error}")

    return errors


def verify_with_retry() -> None:
    last_errors: list[str] = []
    for attempt in range(1, RETRY_ATTEMPTS + 1):
        last_errors = verify_once()
        if not last_errors:
            print(
                "Mesh production verification passed: canonical custom domain is reachable over HTTPS, "
                "reviewed public bytes match, Interlace artwork is current, security headers are present, "
                "and 404 behavior is correct."
            )
            return
        if attempt < RETRY_ATTEMPTS:
            print(f"Verification attempt {attempt} did not yet pass; retrying.")
            time.sleep(RETRY_DELAY_SECONDS)

    print("Mesh production verification failed:")
    for error in last_errors:
        print(f"- {error}")
    raise SystemExit(1)


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--check-config", action="store_true")
    args = parser.parse_args()
    validate_configuration()
    if args.check_config:
        print("Mesh production verifier configuration is valid.")
        return
    verify_with_retry()


if __name__ == "__main__":
    main()
