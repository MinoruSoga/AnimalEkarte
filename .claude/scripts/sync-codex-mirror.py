#!/usr/bin/env python3
"""Regenerate .codex/agents/*.toml and .codex/commands/*.md from .claude/agents and .claude/commands.

Invoked by sync-codex-mirror.sh (which is what the commit-time hook and humans call).
Exits non-zero on any failure so the caller can fail loudly instead of silently.
"""
import re
import shutil
import sys
import tempfile
import tomllib
from pathlib import Path


def read_frontmatter_and_body(path: Path) -> tuple[str, str]:
    text = path.read_text(encoding="utf-8")
    parts = text.split("---\n")
    # Expect: "" , frontmatter, body...  (leading empty string before first ---)
    if len(parts) < 3:
        return "", text
    frontmatter = parts[1]
    body = "---\n".join(parts[2:]).lstrip("\n")
    desc_match = re.search(r'^description:\s*(.*)$', frontmatter, re.MULTILINE)
    desc = desc_match.group(1).strip() if desc_match else ""
    if desc.startswith('"') and desc.endswith('"'):
        desc = desc[1:-1]
    return desc, body


def toml_quote(s: str) -> str:
    return s.replace('\\', '\\\\').replace('"', '\\"')


def main() -> int:
    if len(sys.argv) != 2:
        print("usage: sync-codex-mirror.py <repo-root>", file=sys.stderr)
        return 2
    root = Path(sys.argv[1])
    src_agents = root / ".claude" / "agents"
    src_commands = root / ".claude" / "commands"
    dest_agents = root / ".codex" / "agents"
    dest_commands = root / ".codex" / "commands"

    if not src_agents.is_dir() or not src_commands.is_dir():
        print(f"ERROR: source directories missing under {root}", file=sys.stderr)
        return 1

    with tempfile.TemporaryDirectory() as tmp:
        tmp_agents = Path(tmp) / "agents"
        tmp_commands = Path(tmp) / "commands"
        tmp_agents.mkdir()
        tmp_commands.mkdir()

        agent_count = 0
        for f in sorted(src_agents.glob("*.md")):
            name = f.stem
            desc, body = read_frontmatter_and_body(f)
            # TOML basic multi-line strings interpret backslashes and quotes.
            # Escape the complete body so generated agent roles remain valid
            # while parsing back to the original instructions.
            safe_body = toml_quote(body)
            toml = (
                f'name = "{toml_quote(name)}"\n'
                f'description = "{toml_quote(desc)}"\n'
                f'developer_instructions = """\n{safe_body}\n"""\n'
            )
            tomllib.loads(toml)
            (tmp_agents / f"{name}.toml").write_text(toml, encoding="utf-8")
            agent_count += 1

        command_count = 0
        for f in sorted(src_commands.glob("*.md")):
            shutil.copy2(f, tmp_commands / f.name)
            command_count += 1

        if dest_agents.exists():
            shutil.rmtree(dest_agents)
        if dest_commands.exists():
            shutil.rmtree(dest_commands)
        dest_agents.parent.mkdir(parents=True, exist_ok=True)
        shutil.move(str(tmp_agents), str(dest_agents))
        shutil.move(str(tmp_commands), str(dest_commands))

    print(f"regenerated: {dest_agents} ({agent_count} agents)")
    print(f"regenerated: {dest_commands} ({command_count} commands)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
