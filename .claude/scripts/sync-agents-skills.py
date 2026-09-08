#!/usr/bin/env python3
"""Mirror canonical skill/rule bytes and source-command wrappers safely."""
import re
import subprocess
import sys
from pathlib import Path

from mirror_safety import sync_outputs

OUTPUT = re.compile(r"(?:skills/[a-zA-Z0-9_.-]+/(?:[^/]+/)*[^/]+|rules/[^/]+\.md)")


def source_files(directory: Path):
    if directory.is_symlink() or not directory.is_dir():
        raise ValueError(f"source directory missing or symlinked: {directory}")
    for path in sorted(directory.rglob("*")):
        if path.is_symlink():
            raise ValueError(f"symlinked source refused: {path}")
        if path.is_file():
            yield path
        elif not path.is_dir():
            raise ValueError(f"non-file source refused: {path}")


def command_wrapper(root: Path, source: Path) -> bytes:
    helper = root / "scripts/yaml-frontmatter-description.mjs"
    if helper.is_symlink() or not helper.is_file():
        raise ValueError("missing or symlinked frontmatter helper")
    result = subprocess.run(["node", str(helper), str(source)], capture_output=True, text=True)
    if result.returncode != 0:
        # No source contents in errors: command frontmatter may contain private text.
        raise ValueError(f"invalid command description or unavailable Node helper: {source.name}")
    name = source.stem
    if not re.fullmatch(r"[a-z][a-z0-9-]*", name):
        raise ValueError(f"invalid command name: {name}")
    # Match the previous awk exactly: remove the first two exact '---' lines,
    # retaining following separators and emitting a final newline for each line.
    body, markers = [], 0
    text = source.read_bytes().decode("utf-8")
    lines = text.split("\n")
    if text.endswith("\n"):
        lines.pop()
    for line in lines:
        if line == "---" and markers < 2:
            markers += 1
            continue
        if markers >= 2:
            body.append(line + "\n")
    if markers != 2:
        raise ValueError(f"command needs LF-delimited frontmatter: {source.name}")
    return (f'---\nname: "source-command-{name}"\ndescription: {result.stdout}\n---\n\n'
            f'# source-command-{name}\n\n'
            f'Use this skill when the user asks to run the migrated source command `{name}`.\n\n'
            '## Command Template\n' + ''.join(body)).encode("utf-8")


def regenerate(root: Path) -> None:
    desired = {}
    skills = root / ".claude/skills"
    commands = root / ".claude/commands"
    rules = root / ".claude/rules"
    for folder in (skills, commands, rules):
        if folder.is_symlink() or not folder.is_dir():
            raise ValueError(f"source directory missing or symlinked: {folder}")
    skill_count = 0
    for folder in sorted(skills.iterdir()):
        if folder.is_symlink():
            raise ValueError(f"symlinked source refused: {folder}")
        if not folder.is_dir():
            continue
        skill_count += 1
        for source in source_files(folder):
            desired[f"skills/{source.relative_to(skills).as_posix()}"] = source.read_bytes()
    command_count = 0
    for source in sorted(commands.glob("*.md")):
        if source.is_symlink() or not source.is_file():
            raise ValueError(f"invalid command source: {source}")
        key = f"skills/source-command-{source.stem}/SKILL.md"
        if any(path.startswith(f"skills/source-command-{source.stem}/") for path in desired):
            raise ValueError(f"canonical skill collides with command wrapper: {source.stem}")
        desired[key] = command_wrapper(root, source)
        command_count += 1
    rule_count = 0
    for source in sorted(rules.glob("*.md")):
        if source.is_symlink() or not source.is_file():
            raise ValueError(f"invalid rule source: {source}")
        desired[f"rules/{source.name}"] = source.read_bytes()
        rule_count += 1
    # Legacy skills/rules/wrappers were byte copies of these exact inputs.
    sync_outputs(root / ".agents", desired, desired, OUTPUT)
    print(f"regenerated: {root / '.agents/skills'}")
    print(f"  skills:   {skill_count}")
    print(f"  commands: {command_count}")
    print(f"  rules:    {rule_count} -> {root / '.agents/rules'}")


def main() -> int:
    if len(sys.argv) != 2:
        print("usage: sync-agents-skills.py <repo-root>", file=sys.stderr)
        return 2
    try:
        regenerate(Path(sys.argv[1]))
    except (OSError, ValueError, TypeError, KeyError) as error:
        print(f"ERROR: {error}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
