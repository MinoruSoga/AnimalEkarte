#!/bin/bash
# Canonical .claude skills/rules plus source-command-* wrappers -> .agents.
# Preserve foreign files; preflight all targets before replacing hash-owned or
# exact legacy canonical outputs. No directory wipes. Paths remain .claude/.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
python3 -B "$(dirname "$0")/test_sync_agents_skills.py"
python3 -B "$(dirname "$0")/sync-agents-skills.py" "$ROOT"
