#!/usr/bin/env python3
"""
swag が生成する swagger.yaml / swagger.json の定義名から
'model.' および 'handler.' プレフィックスを除去する。

使用方法:
  python3 scripts/clean-swagger.py docs/swagger.yaml docs/swagger.json
"""

import re
import sys


def strip_prefix(text: str) -> str:
    """model. / handler. プレフィックスを定義名と $ref から除去する。"""
    # definitions キーの置換: "model.Foo:" -> "Foo:"
    text = re.sub(r'\bmodel\.([A-Za-z][A-Za-z0-9_]*)\b', r'\1', text)
    text = re.sub(r'\bhandler\.([A-Za-z][A-Za-z0-9_]*)\b', r'\1', text)
    return text


def process_file(path: str) -> None:
    with open(path, 'r', encoding='utf-8') as f:
        content = f.read()

    cleaned = strip_prefix(content)

    if cleaned == content:
        print(f'  no change: {path}')
        return

    with open(path, 'w', encoding='utf-8') as f:
        f.write(cleaned)
    print(f'  cleaned:   {path}')


if __name__ == '__main__':
    if len(sys.argv) < 2:
        print('Usage: clean-swagger.py <file1> [file2 ...]')
        sys.exit(1)

    for path in sys.argv[1:]:
        process_file(path)
