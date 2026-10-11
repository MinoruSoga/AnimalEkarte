#!/usr/bin/env python3
"""Opt-in backend launch contract; not managed policy or a runtime sandbox."""
from pathlib import Path
import json
import math
import os
import sys

DENIED_OPTIONS = {
    "--settings",
    "--setting-sources",
    "--bare",
    "--safe-mode",
    "--restricted",
    "--dangerously-skip-permissions",
    "--allow-dangerously-skip-permissions",
}
HOOK_REQUIRED_FIELDS = {
    "command": "command",
    "prompt": "prompt",
    "agent": "prompt",
    "http": "url",
}
ENV_TOGGLES = ("CLAUDE_CODE_SIMPLE", "CLAUDE_CODE_SAFE_MODE")

root = Path(__file__).resolve().parents[2]
shared = root / ".claude/settings.json"
backend = root / "backend/.claude/settings.local.json"
if not shared.is_file() or not backend.is_file():
    sys.exit("Required repository settings are missing; Claude was not launched.")


def reject_constant(value: str) -> None:
    raise ValueError(f"non-finite JSON constant: {value}")


def load_settings(path: Path) -> dict:
    try:
        settings = json.loads(path.read_text(), parse_constant=reject_constant)
    except (OSError, UnicodeDecodeError):
        sys.exit("Required repository settings are unreadable; Claude was not launched.")
    except (ValueError, RecursionError):
        sys.exit("Required repository settings are invalid; Claude was not launched.")
    if not isinstance(settings, dict):
        sys.exit("Required repository settings are invalid; Claude was not launched.")
    return settings


def valid_hook_entry(entry: object) -> bool:
    if not isinstance(entry, dict):
        return False
    hook_type = entry.get("type")
    if not isinstance(hook_type, str):
        return False
    required = HOOK_REQUIRED_FIELDS.get(hook_type)
    if required is None:
        return False
    field = entry.get(required)
    if not isinstance(field, str) or not field.strip():
        return False
    if "timeout" in entry:
        timeout = entry["timeout"]
        if isinstance(timeout, bool) or not isinstance(timeout, (int, float)):
            return False
        if isinstance(timeout, int):
            if timeout <= 0:
                return False
        elif not math.isfinite(timeout) or timeout <= 0:
            return False
    return True


def valid_matcher_entry(entry: object) -> bool:
    if not isinstance(entry, dict):
        return False
    if "matcher" in entry and not isinstance(entry["matcher"], str):
        return False
    hooks = entry.get("hooks")
    if not isinstance(hooks, list):
        return False
    return all(valid_hook_entry(hook) for hook in hooks)


def valid_shape(settings: dict) -> bool:
    if "permissions" in settings:
        permissions = settings["permissions"]
        if not isinstance(permissions, dict):
            return False
        for key in ("allow", "deny", "ask"):
            if key in permissions:
                rules = permissions[key]
                if not isinstance(rules, list) or any(
                    not isinstance(rule, str) for rule in rules
                ):
                    return False
    if "hooks" in settings:
        hooks = settings["hooks"]
        if not isinstance(hooks, dict) or any(
            not isinstance(events, list) for events in hooks.values()
        ):
            return False
        if any(
            not valid_matcher_entry(entry)
            for events in hooks.values()
            for entry in events
        ):
            return False
    if "disabledMcpjsonServers" in settings:
        disabled = settings["disabledMcpjsonServers"]
        if not isinstance(disabled, list) or any(
            not isinstance(name, str) for name in disabled
        ):
            return False
    return True


for path in (shared, backend):
    if not valid_shape(load_settings(path)):
        sys.exit("Required repository settings are invalid; Claude was not launched.")

for toggle in ENV_TOGGLES:
    if os.environ.get(toggle, "").strip().lower() in ("1", "true"):
        sys.exit("Launch safeguards are disabled by environment; Claude was not launched.")

args = sys.argv[1:]
for index, arg in enumerate(args):
    if arg == "--":
        break
    if not arg.startswith("--"):
        continue
    name, separator, value = arg.partition("=")
    if name in DENIED_OPTIONS:
        sys.exit("Launch arguments override required safeguards; Claude was not launched.")
    if name == "--permission-mode":
        mode = value if separator else (args[index + 1] if index + 1 < len(args) else "")
        if mode == "bypassPermissions":
            sys.exit("Launch arguments override required safeguards; Claude was not launched.")

# Shared settings auto-load from primary cwd. Explicit backend overlay retains
# its disabled MCP list without duplicating root permissions/hooks.
os.chdir(root)
os.execvp("claude", ["claude", "--settings", str(backend), *args])
