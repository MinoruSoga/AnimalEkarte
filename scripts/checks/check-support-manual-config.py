#!/usr/bin/env python3
"""Check initial support rollout defaults, not deployment or provider readiness."""

import argparse
import json
import pathlib
from typing import NoReturn


ROOT = pathlib.Path(__file__).resolve().parents[2]
CONFIG_PATHS = ('backend/wrangler.jsonc', 'backend/wrangler.production.jsonc')
DISABLED_FLAGS = ('SUPPORT_HYBRID_SEARCH_ENABLED', 'SUPPORT_EXTERNAL_DATA_APPROVED')


def unique_object(pairs: list[tuple[str, object]]) -> dict[str, object]:
    result: dict[str, object] = {}
    for key, value in pairs:
        if key in result:
            raise ValueError('Duplicate JSON key')
        result[key] = value
    return result


def reject_constant(_value: str) -> NoReturn:
    raise ValueError('Non-JSON numeric constant')


def decode_jsonc(text: str) -> object:
    output: list[str] = []
    state = 'normal'
    escaped = False
    index = 0
    text = text.removeprefix('\ufeff')
    while index < len(text):
        char = text[index]
        following = text[index + 1:index + 2]
        if state == 'string':
            output.append(char)
            if escaped:
                escaped = False
            elif char == '\\':
                escaped = True
            elif char == '"':
                state = 'normal'
        elif state == 'line':
            output.append(char if char in '\r\n' else ' ')
            if char in '\r\n':
                state = 'normal'
        elif state == 'block':
            if char == '*' and following == '/':
                output.append(' ')
                state = 'normal'
                index += 1
            else:
                output.append(char if char in '\r\n' else ' ')
        elif char == '/' and following in ('/', '*'):
            state = 'line' if following == '/' else 'block'
            output.append(' ')
            index += 1
        else:
            if char == '"':
                state = 'string'
                escaped = False
            elif char in ']}':
                # Only structural closers can consume a trailing comma.
                previous = len(output) - 1
                while previous >= 0 and output[previous].isspace():
                    previous -= 1
                if previous >= 0 and output[previous] == ',':
                    value_end = previous - 1
                    while value_end >= 0 and output[value_end].isspace():
                        value_end -= 1
                    if value_end < 0 or output[value_end] in '{[,:':
                        raise ValueError('Invalid trailing comma')
                    output[previous] = ' '
            output.append(char)
        index += 1
    if state in ('string', 'block'):
        raise ValueError('Unterminated JSONC token')
    return json.loads(
        ''.join(output),
        object_pairs_hook=unique_object,
        parse_constant=reject_constant,
    )


def validate_vars(config: object) -> None:
    if not isinstance(config, dict) or not isinstance(config.get('vars'), dict):
        raise ValueError('Explicit support vars are required')
    values = config['vars']
    for key in DISABLED_FLAGS:
        if values.get(key) != 'false':
            raise ValueError('Initial support rollout flags must be string false')
    model = values.get('SUPPORT_LLM_MODEL')
    if not isinstance(model, str) or not model.startswith('glm-') or len(model) <= 4:
        raise ValueError('GLM must remain the initial support model')
    if any(key.startswith('SUPPORT_') and key.endswith('_API_KEY') for key in values):
        raise ValueError('Support credentials must not be inline vars')


def validate_config(config: object) -> None:
    if not isinstance(config, dict):
        raise ValueError('Config must be an object')
    validate_vars(config)
    environments = config.get('env', {})
    if not isinstance(environments, dict):
        raise ValueError('Invalid named environment configuration')
    # Wrangler vars are non-inheritable: each named environment must be explicit.
    for environment in environments.values():
        validate_vars(environment)


def main() -> int:
    argparse.ArgumentParser(description=__doc__).parse_args()
    for relative in CONFIG_PATHS:
        try:
            path = ROOT
            for part in pathlib.PurePosixPath(relative).parts:
                path /= part
                if path.is_symlink():
                    raise ValueError('Config inputs must not use symlinks')
            validate_config(decode_jsonc(path.read_text(encoding='utf-8')))
        except (OSError, ValueError):
            print(json.dumps({
                'status': 'FAIL',
                'summary': 'Invalid initial support rollout configuration',
                'artifacts': [relative],
                'next_actions': ['Restore disabled/GLM defaults; do not auto-enable providers'],
            }))
            return 1
    print(json.dumps({
        'status': 'PASS',
        'summary': 'Checked-in support defaults remain disabled with a GLM model',
        'artifacts': list(CONFIG_PATHS),
        'next_actions': [],
    }))
    return 0


if __name__ == '__main__':
    raise SystemExit(main())
