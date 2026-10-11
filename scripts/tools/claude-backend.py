#!/usr/bin/env python3
"""Opt-in backend launch contract; not managed policy or a runtime sandbox."""
from pathlib import Path
import os
import sys

root = Path(__file__).resolve().parents[2]
shared = root / ".claude/settings.json"
backend = root / "backend/.claude/settings.local.json"
if not shared.is_file() or not backend.is_file():
    sys.exit("Required repository settings are missing; Claude was not launched.")
# Shared settings auto-load from primary cwd. Explicit backend overlay retains
# its disabled MCP list without duplicating root permissions/hooks.
os.chdir(root)
os.execvp("claude", ["claude", "--settings", str(backend), *sys.argv[1:]])
