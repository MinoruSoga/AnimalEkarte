#!/usr/bin/env python3
"""
P11 compliance fixer: insert slog.ErrorContext before every apperrors.Wrap in service files.

Rules:
- Insert ONLY when the line before `return ..., apperrors.Wrap(err, "msg")` lacks slog.ErrorContext
- Skip lines inside WrapInvalidInput / WrapConflict / WrapNotFound blocks
- Extract message from apperrors.Wrap and use it for slog.ErrorContext
- Keep existing indentation
"""

import re
import os
import sys

SERVICE_DIR = os.path.join(os.path.dirname(__file__), "../internal/service")

# Patterns to skip (no slog needed)
SKIP_WRAP_PATTERNS = [
    r'apperrors\.WrapInvalidInput',
    r'apperrors\.WrapConflict',
    r'apperrors\.WrapNotFound',
]

WRAP_PATTERN = re.compile(
    r'^(\s*)return\s+(.+?),\s*apperrors\.Wrap\(err,\s*"([^"]+)"\)'
)

def already_has_slog(lines, idx):
    """Check if the preceding non-empty line already contains slog.ErrorContext."""
    i = idx - 1
    while i >= 0:
        line = lines[i].rstrip()
        if line == '':
            i -= 1
            continue
        return 'slog.ErrorContext' in line
    return False

def should_skip(line):
    for pat in SKIP_WRAP_PATTERNS:
        if re.search(pat, line):
            return True
    return False

def fix_file(path):
    with open(path, 'r') as f:
        lines = f.readlines()

    new_lines = []
    changed = 0

    i = 0
    while i < len(lines):
        line = lines[i]
        stripped = line.rstrip('\n')

        if should_skip(stripped):
            new_lines.append(line)
            i += 1
            continue

        m = WRAP_PATTERN.match(stripped)
        if m:
            indent = m.group(1)
            msg = m.group(3)

            if not already_has_slog(new_lines, len(new_lines)):
                slog_line = f'{indent}slog.ErrorContext(ctx, "{msg}", "error", err)\n'
                new_lines.append(slog_line)
                changed += 1

        new_lines.append(line)
        i += 1

    if changed > 0:
        with open(path, 'w') as f:
            f.writelines(new_lines)
        print(f"  {os.path.basename(path)}: +{changed} slog.ErrorContext lines")

    return changed

def main():
    total = 0
    files = sorted(f for f in os.listdir(SERVICE_DIR)
                   if f.endswith('.go') and not f.endswith('_test.go'))

    print(f"Scanning {len(files)} service files for P11 violations...")
    for fname in files:
        path = os.path.join(SERVICE_DIR, fname)
        n = fix_file(path)
        total += n

    print(f"\nDone. Total insertions: {total}")

if __name__ == '__main__':
    main()
