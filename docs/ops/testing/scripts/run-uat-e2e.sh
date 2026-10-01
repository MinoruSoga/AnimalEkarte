#!/usr/bin/env bash
# UAT E2E runner — loads E2E_LOGIN_* from .env.local without echoing values,
# then delegates to frontend/scripts/run-e2e.sh.
#
# Extraction mirrors check-uat-env.sh: read only the two wanted keys via python,
# never source the dotenv (it is not guaranteed to be shell-safe), never place
# values in argv or logs. Usage:
#
#   docs/ops/testing/scripts/run-uat-e2e.sh e2e/owners-search.spec.ts
#   docs/ops/testing/scripts/run-uat-e2e.sh --auth-smoke
set -euo pipefail

readonly EXIT_NOT_READY=1
ROOT="$(cd "$(dirname "$0")/../../../.." && pwd)"
cd "$ROOT"

if [ ! -f .env.local ]; then
  echo "run-uat-e2e.sh: .env-local file missing" >&2
  exit "$EXIT_NOT_READY"
fi

extract_login_env() {
  python3 -c '
import sys

wanted = {"E2E_LOGIN_EMAIL", "E2E_LOGIN_PASSWORD"}
for raw_line in open(sys.argv[1], encoding="utf-8"):
    line = raw_line.strip()
    if not line or line.startswith("#") or "=" not in line:
        continue
    name, raw_value = line.split("=", 1)
    name = name.strip()
    if name.startswith("export "):
        name = name[7:].strip()
    if name not in wanted:
        continue
    value = raw_value.strip()
    q1, q2 = chr(39), chr(34)
    if len(value) >= 2 and value[0] == value[-1] and value[0] in (q1, q2):
        value = value[1:-1]
    sys.stdout.buffer.write(name.encode("utf-8") + b"\0")
    sys.stdout.buffer.write(value.encode("utf-8") + b"\0")
' ".env.local"
}

loaded_email=""
loaded_password=""
while IFS= read -r -d '' name && IFS= read -r -d '' value; do
  case "$name" in
    E2E_LOGIN_EMAIL) loaded_email="$value" ;;
    E2E_LOGIN_PASSWORD) loaded_password="$value" ;;
  esac
done < <(extract_login_env)

if [ -z "$loaded_email" ] || [ -z "$loaded_password" ]; then
  echo "run-uat-e2e.sh: E2E_LOGIN_EMAIL / E2E_LOGIN_PASSWORD not found in dotenv" >&2
  exit "$EXIT_NOT_READY"
fi

export E2E_LOGIN_EMAIL="$loaded_email"
export E2E_LOGIN_PASSWORD="$loaded_password"
exec bash frontend/scripts/run-e2e.sh "$@"
