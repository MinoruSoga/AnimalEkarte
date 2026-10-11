#!/usr/bin/env python3
"""Keep standalone instruction files bound to one canonical safety policy."""
import argparse
import hashlib
import os
import pathlib
import re
import subprocess

SOURCE = '.claude/rules/agent-trust-boundaries.md'
START = '<!-- AUTO:agent-trust-boundaries sha256='
END = '<!-- AUTO:END:agent-trust-boundaries -->'
BLOCK = re.compile(r'<!-- AUTO:agent-trust-boundaries sha256=([a-f0-9]{64}) -->\n(.*?)' + re.escape(END), re.S)


def targets(root):
    result = subprocess.run(['git', 'ls-files', '-z'], cwd=root, capture_output=True, check=True)
    paths = result.stdout.decode().split('\0')
    return sorted(path for path in paths if path and path != SOURCE and (
        pathlib.PurePosixPath(path).name in ('CLAUDE.md', 'AGENTS.md')
        or (path.startswith(('.claude/agents/', '.claude/rules/')) and path.endswith('.md'))))


def render(content, body):
    """Preserve foreign text and refuse edited/incomplete managed blocks."""
    matches = list(BLOCK.finditer(content))
    if content.count(START) != len(matches) or content.count(END) != len(matches) or len(matches) > 1:
        raise ValueError('malformed or duplicate safety block')
    if matches:
        match = matches[0]
        if hashlib.sha256(match[2].encode()).hexdigest() != match[1]:
            raise ValueError('edited safety block; preserve and review owner changes')
        replacement = START + hashlib.sha256(body.encode()).hexdigest() + ' -->\n' + body + END
        return content[:match.start()] + replacement + content[match.end():]
    block = START + hashlib.sha256(body.encode()).hexdigest() + ' -->\n' + body + END + '\n'
    return content + ('\n' if content.endswith('\n') else '\n\n') + block


def sync(root, paths, check=False):
    root = pathlib.Path(root).resolve()
    source = root / SOURCE
    if source.is_symlink() or not source.resolve().is_relative_to(root):
        raise ValueError('symlinked canonical source refused')
    canonical = source.read_text()
    pending = []
    for path in paths:
        target = root / path
        if not target.resolve().is_relative_to(root) or target.is_symlink():
            raise ValueError('outside-root or symlinked target refused')
        link = os.path.relpath(source, target.parent).replace(os.sep, '/')
        body = f'\nSource: [safety policy]({link}). Regenerate after editing the source.\n\n' + canonical
        current = target.read_bytes().decode()
        updated = render(current, body)
        if current != updated:
            pending.append((target, updated))
    # Preflight every target before any write; read-only --check is used in CI.
    if check:
        return len(pending)
    for target, updated in pending:
        target.write_bytes(updated.encode())
    return 0


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--check', action='store_true')
    args = parser.parse_args()
    root = pathlib.Path(__file__).resolve().parents[2]
    try:
        count = sync(root, targets(root), args.check)
    except (ValueError, OSError, subprocess.SubprocessError):
        print('BLOCKED: safety block generation refused; review canonical source and target ownership')
        return 2
    if count:
        print(f'FAIL: {count} instruction files need safety block regeneration')
        return 1
    print('PASS: instruction safety blocks match the canonical source')
    return 0


if __name__ == '__main__':
    raise SystemExit(main())
