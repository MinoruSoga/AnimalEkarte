#!/usr/bin/env python3
"""Copy staff-attach shared secret into frontend/.env.local as VITE_DEMO_LOGIN_PASSWORD.

Local DEV maintainer helper only. Never prints the secret value.
Does not commit anything. Expects gitignored paths:

  sensitive-local/stg-uat-staff-secrets.json
  frontend/.env.local

Usage (repo root):
  python3 scripts/sync-vite-demo-login-password.py
"""
from __future__ import annotations

import json
import os
import re
import stat
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
SECRETS_PATH = ROOT / "sensitive-local" / "stg-uat-staff-secrets.json"
ENV_PATH = ROOT / "frontend" / ".env.local"
KEY = "VITE_DEMO_LOGIN_PASSWORD"
# Shared staff-attach secret length observed on local UAT files (do not hardcode value).
MIN_LEN = 16
MAX_LEN = 128


def _load_shared_password(path: Path) -> str:
    if not path.is_file():
        raise SystemExit(f"FAIL  secrets file missing: {path}")
    mode = path.stat().st_mode & 0o777
    if mode & 0o077:
        raise SystemExit(f"FAIL  secrets file mode too open: {oct(mode)} (want 0600)")
    data = json.loads(path.read_text(encoding="utf-8"))
    entries = data.get("secrets")
    passwords: list[str] = []
    if isinstance(entries, list):
        for row in entries:
            if not isinstance(row, dict):
                continue
            pw = row.get("password")
            if isinstance(pw, str) and pw != "":
                passwords.append(pw)
    elif isinstance(entries, dict):
        for pw in entries.values():
            if isinstance(pw, str) and pw != "":
                passwords.append(pw)
    else:
        raise SystemExit("FAIL  secrets JSON shape unsupported")
    if not passwords:
        raise SystemExit("FAIL  no passwords found in secrets file")
    unique = set(passwords)
    if len(unique) != 1:
        raise SystemExit(
            f"FAIL  expected a single shared password, found {len(unique)} distinct values"
        )
    password = next(iter(unique))
    if not (MIN_LEN <= len(password) <= MAX_LEN):
        raise SystemExit(
            f"FAIL  shared password length {len(password)} outside {MIN_LEN}..{MAX_LEN}"
        )
    return password


def _upsert_env(path: Path, key: str, value: str) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    existing = path.read_text(encoding="utf-8") if path.is_file() else ""
    line = f"{key}={value}"
    pattern = re.compile(rf"^{re.escape(key)}=.*$", re.MULTILINE)
    if pattern.search(existing):
        updated = pattern.sub(line, existing, count=1)
    else:
        if existing and not existing.endswith("\n"):
            existing += "\n"
        updated = existing + line + "\n"
    path.write_text(updated, encoding="utf-8")
    os.chmod(path, stat.S_IRUSR | stat.S_IWUSR)


def main() -> int:
    password = _load_shared_password(SECRETS_PATH)
    _upsert_env(ENV_PATH, KEY, password)
    # Intentionally report length only — never the value.
    print(
        f"OK  wrote {KEY} to {ENV_PATH.relative_to(ROOT)} "
        f"(length={len(password)}, source=staff-attach secrets)"
    )
    print("NOTE restart/rebuild frontend so Vite picks up env changes")
    return 0


if __name__ == "__main__":
    sys.exit(main())
