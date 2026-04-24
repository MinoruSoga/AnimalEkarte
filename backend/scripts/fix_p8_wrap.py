#!/usr/bin/env python3
"""
P8 compliance fixer: replace bare `return X, err` with apperrors.Wrap in service files.
Also adds slog.ErrorContext where needed (P11) for non-validation error returns.

Rules:
- Replace `return nil, err` / `return 0, err` / etc. with `return X, apperrors.Wrap(err, "msg")`
- Derive message from:
  1. The err-generating call found by scanning backward
  2. Outer function name as fallback
- Add slog.ErrorContext ONLY for non-validation calls (not for validate* helpers)
- Skip if line already has apperrors.Wrap (shouldn't happen, but guard)
- Add "log/slog" import to any file that gains slog lines
"""

import re
import os

SERVICE_DIR = os.path.join(os.path.dirname(__file__), "../internal/service")

BARE_RETURN = re.compile(r'^(\s*)return (nil|-1|0|false|""), err$')

VALIDATE_CALL = re.compile(r'\bvalidate\w*\s*\(', re.IGNORECASE)
TX_PATTERN = re.compile(r'\.(WithTx|InTx)\s*\(')
ANY_DOT_CALL = re.compile(r'\.(\w+)\s*\(')
ERR_ASSIGN = re.compile(r'\berr\b\s*(?:,\s*\w+)?\s*:?=\s*(.+)')
IF_ERR_ASSIGN = re.compile(r'if\s+\w+\s*:?=\s*(.+?)\s*;')


def camel_to_snake(name):
    s = re.sub(r'([A-Z]+)([A-Z][a-z])', r'\1 \2', name)
    s = re.sub(r'([a-z\d])([A-Z])', r'\1 \2', s)
    return s.lower()


def entity_from_path(filepath):
    name = os.path.basename(filepath).replace('_service.go', '').replace('_validators.go', '')
    return name.replace('_', ' ')


def find_outer_func(lines, idx):
    i = idx
    while i >= 0:
        line = lines[i].strip()
        m = re.match(r'func\s+\(\s*\w+\s+\*?(\w+)\s*\)\s+(\w+)\s*\(', line)
        if m:
            recv = m.group(1)  # e.g. petService
            fname = m.group(2)  # e.g. Create
            entity = re.sub(r'[Ss]ervice$|[Vv]alidators?$', '', recv)
            entity = camel_to_snake(entity).strip()
            action = camel_to_snake(fname).strip()
            return entity, action
        i -= 1
    return '', ''


def find_err_source(lines, idx):
    """
    Scan backward to find what assigned to err.
    Returns ('validate', hint) | ('tx', '') | ('method', method_name) | ('other', '')
    """
    i = idx - 1
    while i >= 0:
        raw = lines[i].rstrip()
        stripped = raw.strip()

        if not stripped:
            i -= 1
            continue

        if re.match(r'\s*func\s', raw):
            break

        # Look for err assignment on this line
        has_err = re.search(r'\berr\s*:?=', stripped) or re.search(r'if\s+(?:\w+\s*,\s*)?err\s*:?=', stripped)
        if has_err:
            rhs_m = re.search(r':?=\s*(.+)', stripped)
            rhs = rhs_m.group(1).strip() if rhs_m else stripped

            if VALIDATE_CALL.search(rhs):
                m = re.search(r'\bvalidate(\w*)\s*\(', rhs, re.IGNORECASE)
                hint = camel_to_snake(m.group(1)).strip() if m and m.group(1) else ''
                return ('validate', hint)

            if TX_PATTERN.search(rhs):
                return ('tx', '')

            m = ANY_DOT_CALL.search(rhs)
            if m:
                return ('method', m.group(1))

            return ('other', '')

        i -= 1
    return ('other', '')


def derive_message_and_slog(source_type, call_name, entity, action):
    if source_type == 'validate':
        hint = call_name.strip() if call_name else ''
        if hint:
            return f"failed to validate {hint}", False
        return "failed to validate", False

    if source_type == 'tx':
        msg = f"failed to {action}".strip()
        if entity:
            msg = f"failed to {action} {entity}".strip()
        return msg, True

    if source_type == 'method':
        method_words = camel_to_snake(call_name).strip()
        if entity:
            msg = f"failed to {method_words} {entity}"
        else:
            msg = f"failed to {method_words}"
        return msg, True

    # other
    msg = f"failed to {action} {entity}".strip() if entity else "operation failed"
    return msg, True


def already_has_slog(new_lines):
    for line in reversed(new_lines):
        stripped = line.strip()
        if not stripped:
            continue
        return 'slog.ErrorContext' in stripped or 'slog.Error' in stripped
    return False


def fix_file(path):
    with open(path, 'r') as f:
        lines = f.readlines()

    entity = entity_from_path(path)
    new_lines = []
    n_wrap = 0
    n_slog = 0

    i = 0
    while i < len(lines):
        line = lines[i]
        stripped = line.rstrip('\n')

        m = BARE_RETURN.match(stripped)
        if m:
            indent = m.group(1)
            ret_val = m.group(2)

            src_type, call_name = find_err_source(lines, i)
            ent, action = find_outer_func(lines, i)
            ent = ent if ent else entity

            msg, needs_slog = derive_message_and_slog(src_type, call_name, ent, action)

            if needs_slog and not already_has_slog(new_lines):
                new_lines.append(f'{indent}slog.ErrorContext(ctx, "{msg}", "error", err)\n')
                n_slog += 1

            new_lines.append(f'{indent}return {ret_val}, apperrors.Wrap(err, "{msg}")\n')
            n_wrap += 1
            i += 1
            continue

        new_lines.append(line)
        i += 1

    if n_wrap > 0:
        content = ''.join(new_lines)

        # Ensure log/slog imported if slog lines were added
        if n_slog > 0 and '"log/slog"' not in content:
            # Insert after first stdlib import in the import block
            # Strategy: find the first `"` inside `import (` block and insert before it
            def insert_slog_import(text):
                # Find import block
                m = re.search(r'(import\s*\(\n)', text)
                if not m:
                    return text
                pos = m.end()
                return text[:pos] + '\t"log/slog"\n' + text[pos:]

            content = insert_slog_import(content)
            new_lines = content.splitlines(True)

        with open(path, 'w') as f:
            f.writelines(new_lines)
        print(f"  {os.path.basename(path)}: +{n_wrap} Wrap, +{n_slog} slog")

    return n_wrap, n_slog


def main():
    total_w = 0
    total_s = 0
    files = sorted(f for f in os.listdir(SERVICE_DIR)
                   if f.endswith('.go') and not f.endswith('_test.go'))
    print(f"Scanning {len(files)} service files for P8 bare returns...")
    for fname in files:
        w, s = fix_file(os.path.join(SERVICE_DIR, fname))
        total_w += w
        total_s += s
    print(f"\nDone. Wrap insertions: {total_w}, slog insertions: {total_s}")


if __name__ == '__main__':
    main()
