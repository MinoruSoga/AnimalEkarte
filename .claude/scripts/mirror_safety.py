"""Shared ownership-aware file mirroring; callers construct outputs in memory.

This prevents preflight conflicts from causing partial writes. Concurrent writers
must still use isolated worktrees; multi-file filesystem writes are not atomic.
"""
import hashlib
import json
import os
import re
import tempfile
from pathlib import Path, PurePosixPath


def digest(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def validate_destination(base: Path, key: str, pattern: re.Pattern) -> Path:
    parts = PurePosixPath(key).parts
    if not pattern.fullmatch(key) or not parts or any(part in {".", ".."} for part in key.split("/")):
        raise ValueError(f"invalid output path: {key}")
    target = base / key
    current = base
    for part in (None, *parts):
        if part is not None:
            current = current / part
        if current.is_symlink():
            raise ValueError(f"symlink destination refused: {current}")
        if current.exists() and current != target and not current.is_dir():
            raise ValueError(f"non-directory output parent: {current}")
    if target.exists() and not target.is_file():
        raise ValueError(f"non-file destination refused: {target}")
    return target


def plan_updates(base: Path, desired: dict[str, bytes], legacy: dict[str, bytes], pattern: re.Pattern):
    ledger = base / ".mirror-ownership.json"
    if base.is_symlink() or ledger.is_symlink():
        raise ValueError("symlinked output root or ownership ledger")
    owned = {}
    if ledger.exists():
        record = json.loads(ledger.read_text())
        if not isinstance(record, dict) or set(record) != {"version", "files"} or record["version"] != 1 or not isinstance(record["files"], dict):
            raise ValueError("invalid ownership ledger")
        owned = record["files"]
        for key, value in owned.items():
            validate_destination(base, key, pattern)
            if not isinstance(value, str) or not re.fullmatch(r"[0-9a-f]{64}", value):
                raise ValueError("invalid ownership digest")
    snapshot, removals = {}, []
    for key in sorted(set(desired) | set(legacy) | set(owned)):
        target = validate_destination(base, key, pattern)
        current = target.read_bytes() if target.exists() else None
        snapshot[key] = current
        if current is not None:
            recognized = (key in owned and digest(current) == owned[key]) or (key in legacy and current == legacy[key])
            if not recognized and current != desired.get(key):
                raise ValueError(f"foreign or modified generated file; preserve and reconcile explicitly: {target}")
            if key not in desired:
                removals.append(key)
    return ledger, snapshot, removals


def atomic_write(path: Path, data: bytes) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with tempfile.NamedTemporaryFile(dir=path.parent, delete=False) as file:
        temporary = Path(file.name)
        file.write(data)
    try:
        os.replace(temporary, path)
    finally:
        temporary.unlink(missing_ok=True)


def sync_outputs(base: Path, desired: dict[str, bytes], legacy: dict[str, bytes], pattern: re.Pattern) -> None:
    ledger, snapshot, removals = plan_updates(base, desired, legacy, pattern)
    for key, expected in snapshot.items():
        target = validate_destination(base, key, pattern)
        current = target.read_bytes() if target.exists() else None
        if current != expected:
            raise ValueError(f"output changed during preflight: {target}")
    for key, data in sorted(desired.items()):
        if snapshot.get(key) != data:
            atomic_write(base / key, data)
    for key in removals:
        (base / key).unlink()
    record = {"version": 1, "files": {key: digest(data) for key, data in sorted(desired.items())}}
    data = (json.dumps(record, indent=2) + "\n").encode()
    if not ledger.exists() or ledger.read_bytes() != data:
        atomic_write(ledger, data)
