#!/usr/bin/env bash
# scripts/check-agent-security-policy.sh
#
# Fail if Cursor agent policy files grant unrestricted execution:
#   - .cursor/permissions.json: approvalMode "unrestricted"
#   - .cursor/permissions.json: mcpAllowlist entry "*:*"
#   - .cursor/sandbox.json: networkPolicy.default "allow"
#
# Usage: bash scripts/check-agent-security-policy.sh [repo-root]
# Exit: 0 = PASS, 1 = FAIL
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="${1:-$SCRIPT_DIR/..}"
ROOT="$(cd "$ROOT" && pwd)"

if [[ ! -d "$ROOT" ]]; then
  echo "FAIL  repo root not found: $ROOT"
  exit 1
fi

PERM="$ROOT/.cursor/permissions.json"
SANDBOX="$ROOT/.cursor/sandbox.json"
errors=0

# permissions.json is required for this policy surface.
if [[ ! -f "$PERM" ]]; then
  echo "FAIL  missing $PERM"
  exit 1
fi

# Strip // line comments and /* */ blocks so JSONC is greppable.
strip_jsonc() {
  # remove // comments (not perfect for strings, sufficient for policy keys)
  sed -E 's|//.*$||' "$1" | sed -E 's|/\*.*\*/||g'
}

perm_body="$(strip_jsonc "$PERM")"

if echo "$perm_body" | grep -Eq '"approvalMode"[[:space:]]*:[[:space:]]*"unrestricted"'; then
  echo "FAIL  approvalMode is unrestricted in .cursor/permissions.json"
  errors=$((errors + 1))
fi

# mcpAllowlist must not contain the global wildcard "*:*"
if echo "$perm_body" | grep -Eq '"\*:\*"'; then
  echo "FAIL  mcpAllowlist contains \"*:*\" in .cursor/permissions.json"
  errors=$((errors + 1))
fi

if [[ ! -f "$SANDBOX" ]]; then
  echo "FAIL  missing $SANDBOX"
  exit 1
fi

sandbox_body="$(strip_jsonc "$SANDBOX")"

# networkPolicy.default must not be "allow" (deny or non-allow only).
# Match "default": "allow" inside the file; sandbox.json is small and dedicated.
if echo "$sandbox_body" | grep -Eq '"default"[[:space:]]*:[[:space:]]*"allow"'; then
  echo "FAIL  networkPolicy.default is allow in .cursor/sandbox.json"
  errors=$((errors + 1))
fi

if [[ "$errors" -gt 0 ]]; then
  echo "----"
  printf 'RESULT  %d agent security policy violation(s)\n' "$errors"
  exit 1
fi

echo "PASS  agent security policy (no unrestricted approval, no *:* MCP, network default non-allow)"
exit 0
