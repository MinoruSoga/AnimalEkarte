#!/usr/bin/env python3
"""Generate/check complete column and FK reference docs without executing SQL."""

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path

from erd_render import GENERATOR, MARKER, render_documents
from erd_schema import load_schema


def managed_markdown(directory: Path) -> set[str]:
    return {p.name for p in directory.glob("*.md") if p.read_text().startswith(MARKER)}


def verify(files: dict[str, str], output: Path) -> list[str]:
    problems = []
    for name, text in files.items():
        p = output / name
        if p.is_symlink():
            problems.append(f"symlink at generated path: {name}")
        elif not p.exists():
            problems.append(f"missing: {name}")
        elif p.read_bytes() != text.encode():
            problems.append(f"drift: {name}")
    problems.extend(f"extra generated page: {name}" for name in sorted(managed_markdown(output) - set(files)))
    return problems


def write(files: dict[str, str], output: Path) -> None:
    # Validate every destination before the first write, preserving foreign documents.
    for name in files:
        p = output / name
        if p.is_symlink():
            raise ValueError(f"refusing to overwrite symlink: {p}")
        if not p.exists():
            continue
        if name == "manifest.json":
            manifest = json.loads(p.read_text())
            if not isinstance(manifest, dict) or manifest.get("generator") != GENERATOR:
                raise ValueError(f"refusing to overwrite foreign manifest: {p}")
        elif not p.read_text().startswith(MARKER):
            raise ValueError(f"refusing to overwrite human document: {p}")
    stale = managed_markdown(output) - set(files)
    output.mkdir(parents=True, exist_ok=True)
    for name, text in files.items():
        p = output / name
        if not p.exists() or p.read_bytes() != text.encode():
            p.write_text(text)
    for name in sorted(stale):
        (output / name).unlink()


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--root", type=Path, default=Path(__file__).resolve().parents[2])
    mode = parser.add_mutually_exclusive_group(required=True)
    mode.add_argument("--check", action="store_true", help="compare only; never write")
    mode.add_argument("--write", action="store_true", help="update owned generated docs")
    args = parser.parse_args()
    try:
        root = args.root.resolve()
        schema = load_schema(root / "backend/migrations")
        files = render_documents(schema, Path(__file__).resolve().parent)
        output = root / "docs/architecture/erd-detail"
        for directory in (output, *output.parents):
            if directory == root:
                break
            if directory.is_symlink():
                raise ValueError(f"refusing symlink in output directory: {directory}")
        if args.check:
            problems = verify(files, output)
            if problems:
                print("ERD detail drift:\n  " + "\n  ".join(problems), file=sys.stderr)
                return 1
        else:
            write(files, output)
        print(f"PASS: ERD detail {'check' if args.check else 'write'}; {len(schema.tables)} tables; {len(files)} files")
        return 0
    except (ValueError, OSError, KeyError) as error:
        print(f"ERD detail BLOCKED: {error}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    sys.exit(main())
