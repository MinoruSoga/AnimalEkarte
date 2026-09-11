#!/usr/bin/env python3
"""Generate static Codex role/command artifacts without overwriting foreign WIP.

Runtime role discovery is verified separately in the active Codex session.
All destinations are checked before writes; only hash-owned or exact legacy
canonical outputs may be replaced/removed. Unknown unrelated files survive.
"""
import json
import re
import sys
import tomllib
from pathlib import Path

from mirror_safety import sync_outputs

RESERVED = {"architect", "debugger", "formatter", "reviewer", "planner", "implementer", "test-strategist"}
NAME = re.compile(r"[a-z][a-z0-9-]*")
OUTPUT = re.compile(r"(?:agents/[a-z][a-z0-9-]*\.toml|commands/[a-zA-Z0-9_.-]+\.md)")


def read_frontmatter_and_body(path: Path) -> tuple[str, str]:
    text = path.read_text(encoding="utf-8")
    parts = text.split("---\n")
    if not text.startswith("---\n") or len(parts) < 3:
        return "", text
    frontmatter = parts[1]
    body = "---\n".join(parts[2:]).lstrip("\n")
    match = re.search(r'^description:\s*(.*)$', frontmatter, re.MULTILINE)
    desc = match.group(1).strip() if match else ""
    if desc.startswith('"') and desc.endswith('"'):
        desc = desc[1:-1]
    return desc, body


def toml_quote(text: str) -> str:
    return text.replace('\\', '\\\\').replace('"', '\\"')


def legacy_agent(name: str, desc: str, body: str) -> str:
    """Exact former generator bytes, solely for one-time ownership recognition."""
    return (f'name = "{toml_quote(name)}"\n'
            f'description = "{toml_quote(desc)}"\n'
            f'developer_instructions = """\n{toml_quote(body)}\n"""\n')


def load_roles(root: Path) -> list[dict]:
    manifest = json.loads((root / ".claude/codex-agent-manifest.json").read_text())
    if not isinstance(manifest, dict) or set(manifest) != {"version", "roles"} or manifest["version"] != 1:
        raise ValueError("invalid agent manifest version/fields")
    if not isinstance(manifest["roles"], list):
        raise ValueError("manifest roles must be a list")
    names, sources = set(), set()
    for role in manifest["roles"]:
        if not isinstance(role, dict) or not {"name", "source"} <= set(role) <= {"name", "source", "sandbox_mode"}:
            raise ValueError("invalid role fields")
        name, source = role["name"], role["source"]
        if not isinstance(name, str) or not NAME.fullmatch(name) or name in RESERVED or name in names:
            raise ValueError(f"invalid, reserved or duplicate role name: {name}")
        if not isinstance(source, str) or not re.fullmatch(r"[a-z][a-z0-9-]*\.md", source) or source in sources:
            raise ValueError(f"invalid or duplicate role source: {source}")
        if "sandbox_mode" in role and role["sandbox_mode"] != "read-only":
            raise ValueError("roles may only restrict to read-only or inherit user sandbox")
        path = root / ".claude/agents" / source
        if path.is_symlink() or not path.is_file():
            raise ValueError(f"missing or symlinked role source: {source}")
        names.add(name)
        sources.add(source)
    return manifest["roles"]


def build_outputs(root: Path, roles: list[dict]) -> tuple[dict[str, bytes], dict[str, bytes]]:
    desired, legacy = {}, {}
    for source in sorted((root / ".claude/agents").glob("*.md")):
        if source.is_symlink() or not source.is_file():
            raise ValueError(f"invalid source: {source}")
        desc, body = read_frontmatter_and_body(source)
        legacy[f"agents/{source.stem}.toml"] = legacy_agent(source.stem, desc, body).encode()
    for role in roles:
        desc, body = read_frontmatter_and_body(root / ".claude/agents" / role["source"])
        fields = {"name": role["name"], "description": desc, "developer_instructions": body}
        if "sandbox_mode" in role:
            fields["sandbox_mode"] = role["sandbox_mode"]
        text = "".join(f"{key} = {json.dumps(value, ensure_ascii=False)}\n" for key, value in fields.items())
        tomllib.loads(text)
        desired[f"agents/{role['name']}.toml"] = text.encode()
    for source in sorted((root / ".claude/commands").glob("*.md")):
        if source.is_symlink() or not source.is_file():
            raise ValueError(f"invalid source: {source}")
        key = f"commands/{source.name}"
        desired[key] = legacy[key] = source.read_bytes()
    return desired, legacy


def regenerate(root: Path) -> None:
    for folder in (root / ".claude/agents", root / ".claude/commands"):
        if folder.is_symlink() or not folder.is_dir():
            raise ValueError(f"source directory missing or symlinked: {folder}")
    roles = load_roles(root)
    desired, legacy = build_outputs(root, roles)
    base = root / ".codex"
    sync_outputs(base, desired, legacy, OUTPUT)
    print(f"regenerated: {base / 'agents'} ({len(roles)} agents)")
    print(f"regenerated: {base / 'commands'} ({sum(key.startswith('commands/') for key in desired)} commands)")


def main() -> int:
    if len(sys.argv) != 2:
        print("usage: sync-codex-mirror.py <repo-root>", file=sys.stderr)
        return 2
    try:
        regenerate(Path(sys.argv[1]))
    except (OSError, ValueError, TypeError, KeyError, tomllib.TOMLDecodeError) as error:
        print(f"ERROR: {error}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
