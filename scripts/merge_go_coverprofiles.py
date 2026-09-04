#!/usr/bin/env python3
"""Merge Go coverprofiles from isolated CI shards without double-counting blocks."""

from __future__ import annotations

import argparse
from dataclasses import dataclass
from pathlib import Path


@dataclass
class CoverageBlock:
    statements: int
    count: int


def read_profile(path: Path) -> tuple[str, dict[str, CoverageBlock]]:
    lines = path.read_text(encoding="utf-8").splitlines()
    if not lines or not lines[0].startswith("mode: "):
        raise ValueError(f"{path}: missing coverage mode header")

    mode = lines[0].removeprefix("mode: ").strip()
    if mode not in {"set", "count", "atomic"}:
        raise ValueError(f"{path}: unsupported coverage mode {mode!r}")

    blocks: dict[str, CoverageBlock] = {}
    for line_number, line in enumerate(lines[1:], start=2):
        if not line.strip():
            continue
        try:
            location, statements_text, count_text = line.rsplit(maxsplit=2)
            statements = int(statements_text)
            count = int(count_text)
        except (ValueError, TypeError) as exc:
            raise ValueError(f"{path}:{line_number}: invalid coverage block") from exc
        if statements < 0 or count < 0:
            raise ValueError(f"{path}:{line_number}: negative coverage value")
        existing = blocks.get(location)
        if existing is not None and existing.statements != statements:
            raise ValueError(
                f"{path}:{line_number}: statement count mismatch for {location}"
            )
        if existing is None:
            blocks[location] = CoverageBlock(statements=statements, count=count)
        elif mode == "set":
            existing.count = max(existing.count, count)
        else:
            existing.count += count
    return mode, blocks


def merge_profiles(paths: list[Path]) -> tuple[str, dict[str, CoverageBlock]]:
    if not paths:
        raise ValueError("at least one coverage profile is required")

    merged_mode: str | None = None
    merged: dict[str, CoverageBlock] = {}
    for path in paths:
        mode, blocks = read_profile(path)
        if merged_mode is None:
            merged_mode = mode
        elif mode != merged_mode:
            raise ValueError(
                f"coverage mode mismatch: expected {merged_mode!r}, got {mode!r} in {path}"
            )
        for location, block in blocks.items():
            existing = merged.get(location)
            if existing is not None and existing.statements != block.statements:
                raise ValueError(f"statement count mismatch for {location}")
            if existing is None:
                merged[location] = CoverageBlock(block.statements, block.count)
            elif mode == "set":
                existing.count = max(existing.count, block.count)
            else:
                existing.count += block.count

    assert merged_mode is not None
    return merged_mode, merged


def write_profile(path: Path, mode: str, blocks: dict[str, CoverageBlock]) -> None:
    output = [f"mode: {mode}"]
    output.extend(
        f"{location} {block.statements} {block.count}"
        for location, block in sorted(blocks.items())
    )
    path.write_text("\n".join(output) + "\n", encoding="utf-8")


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--output", required=True, type=Path)
    parser.add_argument("profiles", nargs="+", type=Path)
    args = parser.parse_args()

    mode, blocks = merge_profiles(args.profiles)
    write_profile(args.output, mode, blocks)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
