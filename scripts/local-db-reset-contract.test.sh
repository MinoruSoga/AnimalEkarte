#!/usr/bin/env bash
# scripts/local-db-reset-contract.test.sh
#
# local-db-reset-contract.sh の回帰テスト（fixture + mock docker）。
# 誤 project / 誤 volume / snapshot・hash 失敗 / cache 削除禁止 /
# migration・seed・health 不足を非 0 にし、正常 fixture だけ 0 にする。
# 実 Docker の make reset は実行しない（USER-only）。
#
# Usage: bash scripts/local-db-reset-contract.test.sh
# Exit codes:
#   0  全ケース PASS
#   1  いずれかのケースが期待と異なる
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONTRACT="$SCRIPT_DIR/local-db-reset-contract.sh"

if [[ ! -f "$CONTRACT" ]]; then
  echo "FAIL  contract script not found at $CONTRACT"
  exit 1
fi

TMP_ROOT="$(mktemp -d)"
trap 'rm -rf "$TMP_ROOT"' EXIT

failures=0

GOOD_COMPOSE='services:
  db:
    image: postgres:18-alpine
volumes:
  postgres_data:
    name: ekarte-postgres-data
  frontend_node_modules:
    name: ekarte-frontend-node-modules
  go_mod_cache:
    name: ekarte-go-mod-cache
  go_build_cache:
    name: ekarte-go-build-cache
'

GOOD_MAKEFILE='reset:
	@bash scripts/local-db-reset-contract.sh

migrate:
	@echo migrate
'

# Build a fixture tree under $1 and write docker mock + optional curl mock.
# Environment knobs are written into $casedir/env.sh for the runner.
build_fixture() {
  local casedir="$1"
  mkdir -p "$casedir/scripts" \
    "$casedir/backend/migrations" \
    "$casedir/bin" \
    "$casedir/state" \
    "$casedir/backups"

  cp "$CONTRACT" "$casedir/scripts/local-db-reset-contract.sh"
  printf '%s\n' "$GOOD_COMPOSE" >"$casedir/docker-compose.yml"
  printf '%s\n' "$GOOD_MAKEFILE" >"$casedir/Makefile"
  printf '%s\n' 'APP_ENV=development' >"$casedir/.env.local"
  # minimal root DDL set
  printf '%s\n' '-- ddl' >"$casedir/backend/migrations/001_init.sql"
  printf '%s\n' '-- ddl' >"$casedir/backend/migrations/002_extra.sql"

  # default mock state
  printf '%s\n' "volume_present=1" >"$casedir/state/flags"
  printf '%s\n' "dump_mode=ok" >>"$casedir/state/flags"
  printf '%s\n' "postflight_mode=ok" >>"$casedir/state/flags"
  printf '%s\n' "health_code=200" >>"$casedir/state/flags"
  : >"$casedir/state/docker.log"
}

write_mock_docker() {
  local casedir="$1"
  cat >"$casedir/bin/docker" <<'MOCK'
#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
LOG="$ROOT/state/docker.log"
FLAGS="$ROOT/state/flags"
# shellcheck disable=SC1090
source "$FLAGS"

# Record every invocation for assertions (single line).
printf 'DOCKER:' >>"$LOG"
printf ' %q' "$@" >>"$LOG"
printf '\n' >>"$LOG"

# Refuse cache volume deletion if ever requested.
for arg in "$@"; do
  case "$arg" in
    ekarte-frontend-node-modules|ekarte-go-mod-cache|ekarte-go-build-cache)
      # volume rm <cache> is the only forbidden path we hard-fail on.
      if printf '%s' "$*" | grep -E 'volume[[:space:]]+rm' >/dev/null 2>&1; then
        echo "mock-docker: refusing cache volume delete: $arg" >&2
        exit 99
      fi
      ;;
  esac
done

# volume inspect
if [[ "${1:-}" == "volume" && "${2:-}" == "inspect" ]]; then
  vol="${3:-}"
  if [[ "$vol" == "ekarte-postgres-data" ]]; then
    if [[ "${volume_present:-1}" == "1" ]]; then
      echo '{"Name":"ekarte-postgres-data"}'
      exit 0
    fi
    exit 1
  fi
  # caches always "exist" for keep assertions
  echo "{\"Name\":\"$vol\"}"
  exit 0
fi

# volume rm
if [[ "${1:-}" == "volume" && "${2:-}" == "rm" ]]; then
  vol="${3:-}"
  if [[ "$vol" == "ekarte-postgres-data" ]]; then
    # record deletion
    echo "volume_present=0" >"$FLAGS.tmp"
    grep -v '^volume_present=' "$FLAGS" >>"$FLAGS.tmp" || true
    mv "$FLAGS.tmp" "$FLAGS"
    echo "removed $vol"
    exit 0
  fi
  echo "mock-docker: unexpected volume rm: $vol" >&2
  exit 98
fi

# compose subcommands — find the verb after compose flags
# argv shape: compose -p PROJECT -f FILE [--env-file FILE] VERB ...
if [[ "${1:-}" == "compose" ]]; then
  shift
  # skip flags until verb
  while [[ $# -gt 0 ]]; do
    case "$1" in
      -p|--project-name) shift 2 ;;
      -f|--file) shift 2 ;;
      --env-file) shift 2 ;;
      --profile) shift 2 ;;
      -*) shift ;;
      *) break ;;
    esac
  done
  verb="${1:-}"
  shift || true

  case "$verb" in
    ls)
      echo '[]'
      exit 0
      ;;
    config)
      # not used for name extraction; static parse is source of truth
      exit 0
      ;;
    up)
      exit 0
      ;;
    down)
      # Must not receive -v
      for a in "$@"; do
        if [[ "$a" == "-v" || "$a" == "--volumes" ]]; then
          echo "mock-docker: compose down received volume-removal flag: $a" >&2
          exit 97
        fi
      done
      exit 0
      ;;
    logs)
      if [[ "${postflight_mode:-ok}" == "ok" ]]; then
        echo "Migration completed file=001_init.sql"
        echo "Migration completed file=002_extra.sql"
        echo "Seed bundle loaded bundle=002_master"
        echo "Migration key coverage missing=0 extra=0 expected=3 recorded=3"
      elif [[ "${postflight_mode:-}" == "missing_coverage" ]]; then
        echo "Migration key coverage missing=2 extra=0 expected=5 recorded=3"
      else
        echo "no coverage line"
      fi
      exit 0
      ;;
    ps)
      if [[ "${postflight_mode:-ok}" == "unhealthy" ]]; then
        echo "NAME  STATUS"
        echo "backend  Exit 1"
      else
        echo "NAME  STATUS"
        echo "backend  Up (healthy)"
      fi
      exit 0
      ;;
    exec)
      # compose exec -T db sh -c '...'
      # Find the -c payload
      payload=""
      prev=""
      for a in "$@"; do
        if [[ "$prev" == "-c" ]]; then
          payload="$a"
        fi
        prev="$a"
      done

      if [[ "$payload" == *pg_isready* ]]; then
        exit 0
      fi

      if [[ "$payload" == *pg_dumpall* ]]; then
        case "${dump_mode:-ok}" in
          ok)
            # Non-empty SQL-ish payload for gzip
            printf '%s\n' '-- mock pg_dumpall'
            printf '%s\n' 'CREATE DATABASE animalekarte;'
            printf '%s\n' 'SELECT 1;'
            exit 0
            ;;
          empty)
            exit 0
            ;;
          fail)
            echo "pg_dumpall error" >&2
            exit 1
            ;;
        esac
      fi

      if [[ "$payload" == *schema_migrations* ]]; then
        if [[ "${postflight_mode:-ok}" == "missing_seed" ]]; then
          # DDL present, seeds absent
          if [[ "$payload" == *seeds/* ]]; then
            exit 0
          fi
          echo "1"
          exit 0
        fi
        if [[ "${postflight_mode:-ok}" == "missing_ddl" ]]; then
          if [[ "$payload" == *001_init* || "$payload" == *002_extra* ]]; then
            exit 0
          fi
          if [[ "$payload" == *COUNT* ]]; then
            echo "0"
            exit 0
          fi
          echo "1"
          exit 0
        fi
        if [[ "${postflight_mode:-ok}" == "ok" || "${postflight_mode:-ok}" == "bad_health" ]]; then
          if [[ "$payload" == *COUNT* ]]; then
            echo "5"
            exit 0
          fi
          echo "1"
          exit 0
        fi
        # default: empty answers → postflight fail
        if [[ "$payload" == *COUNT* ]]; then
          echo "0"
        fi
        exit 0
      fi

      exit 0
      ;;
    *)
      echo "mock-docker: unhandled compose verb: $verb" >&2
      exit 96
      ;;
  esac
fi

echo "mock-docker: unhandled: $*" >&2
exit 95
MOCK
  chmod +x "$casedir/bin/docker"
}

write_mock_curl() {
  local casedir="$1"
  cat >"$casedir/bin/curl" <<'CURL'
#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
# shellcheck disable=SC1090
source "$ROOT/state/flags"
# Support: curl -s -o /dev/null -w '%{http_code}' URL
echo -n "${health_code:-200}"
CURL
  chmod +x "$casedir/bin/curl"
}

set_flag() {
  local casedir="$1" key="$2" val="$3"
  local flags="$casedir/state/flags"
  local tmp="$flags.tmp"
  grep -v "^${key}=" "$flags" >"$tmp" || true
  echo "${key}=${val}" >>"$tmp"
  mv "$tmp" "$flags"
}

# Run contract under fixture. Args after expected_exit are extra env assignments
# in KEY=VAL form, then optional script args.
run_case() {
  local name="$1"
  local expected_exit="$2"
  shift 2
  local casedir="$TMP_ROOT/$name"
  build_fixture "$casedir"
  write_mock_docker "$casedir"
  write_mock_curl "$casedir"

  # Apply case-specific mutations from remaining KEY=VAL and -- script args
  local -a script_args=()
  local saw_sep=0
  local kv
  for kv in "$@"; do
    if [[ "$kv" == "--" ]]; then
      saw_sep=1
      continue
    fi
    if [[ "$saw_sep" -eq 1 ]]; then
      script_args+=("$kv")
      continue
    fi
    case "$kv" in
      COMPOSE=*)
        printf '%s\n' "${kv#COMPOSE=}" >"$casedir/docker-compose.yml"
        ;;
      MAKEFILE=*)
        printf '%s\n' "${kv#MAKEFILE=}" >"$casedir/Makefile"
        ;;
      ENVFILE=*)
        printf '%s\n' "${kv#ENVFILE=}" >"$casedir/.env.local"
        ;;
      FLAG:*)
        # FLAG:key=val
        local body="${kv#FLAG:}"
        set_flag "$casedir" "${body%%=*}" "${body#*=}"
        ;;
      PROJECT=*)
        export LOCAL_DB_RESET_EXPECTED_PROJECT="${kv#PROJECT=}"
        ;;
      COMPOSE_PROJECT_NAME=*)
        export COMPOSE_PROJECT_NAME="${kv#COMPOSE_PROJECT_NAME=}"
        ;;
      UNSET_COMPOSE_PROJECT_NAME)
        unset COMPOSE_PROJECT_NAME || true
        ;;
      *)
        echo "FAIL  [$name] unknown setup token: $kv"
        failures=$((failures + 1))
        return 0
        ;;
    esac
  done

  local out actual_exit
  set +e
  out="$(
    cd "$casedir"
    unset COMPOSE_PROJECT_NAME 2>/dev/null || true
    # re-export if set via token above — handled after unset by re-reading? 
    # We exported in the parent shell; child inherits.
    env \
      LOCAL_DB_RESET_ROOT="$casedir" \
      LOCAL_DB_RESET_DOCKER="$casedir/bin/docker" \
      LOCAL_DB_RESET_CURL="$casedir/bin/curl" \
      LOCAL_DB_RESET_BACKUP_ROOT="$casedir/backups" \
      LOCAL_DB_RESET_COMPOSE_FILE="$casedir/docker-compose.yml" \
      LOCAL_DB_RESET_ENV_FILE="$casedir/.env.local" \
      LOCAL_DB_RESET_MAKEFILE="$casedir/Makefile" \
      LOCAL_DB_RESET_MIGRATIONS_DIR="$casedir/backend/migrations" \
      LOCAL_DB_RESET_HEALTH_URL="http://127.0.0.1:8080/health" \
      LOCAL_DB_RESET_EXPECTED_PROJECT="${LOCAL_DB_RESET_EXPECTED_PROJECT:-animalekarte}" \
      bash "$casedir/scripts/local-db-reset-contract.sh" ${script_args[@]+"${script_args[@]}"} 2>&1
  )"
  actual_exit=$?
  set -e

  # clean per-case project overrides so they do not leak
  unset LOCAL_DB_RESET_EXPECTED_PROJECT || true
  unset COMPOSE_PROJECT_NAME || true

  if [[ "$actual_exit" -eq "$expected_exit" ]]; then
    echo "PASS  [$name] exit=$actual_exit (expected $expected_exit)"
  else
    echo "FAIL  [$name] exit=$actual_exit (expected $expected_exit)"
    echo "----- output -----"
    printf '%s\n' "$out"
    echo "----- docker.log -----"
    cat "$casedir/state/docker.log" 2>/dev/null || true
    echo "------------------"
    failures=$((failures + 1))
    return 0
  fi

  # Extra assertions stored as side files by callers via post_assert_* helpers
  if [[ -f "$casedir/state/assert.sh" ]]; then
    # shellcheck disable=SC1090
    if ! bash "$casedir/state/assert.sh" "$casedir" "$out"; then
      echo "FAIL  [$name] post-assert failed"
      echo "----- output -----"
      printf '%s\n' "$out"
      echo "----- docker.log -----"
      cat "$casedir/state/docker.log" 2>/dev/null || true
      failures=$((failures + 1))
    fi
  fi
}

write_assert() {
  local casedir="$1"
  cat >"$casedir/state/assert.sh"
  chmod +x "$casedir/state/assert.sh"
}

# ═══════════════════════════════════════════════════════════════════════════
# Cases
# ═══════════════════════════════════════════════════════════════════════════

# 1. Good full execute with mocks → 0
run_case "good-execute" 0

# 2. Contract-only on good fixture → 0
run_case "good-contract-only" 0 -- --contract-only

# 3. Wrong volume name in compose → non-zero (contract-only)
run_case "bad-volume" 1 \
  "COMPOSE=services:
  db:
    image: postgres:18-alpine
volumes:
  postgres_data:
    name: wrong-postgres-data
  frontend_node_modules:
    name: ekarte-frontend-node-modules
  go_mod_cache:
    name: ekarte-go-mod-cache
  go_build_cache:
    name: ekarte-go-build-cache
" \
  -- --contract-only

# 4. Missing cache volume in compose → non-zero
run_case "bad-missing-cache-volume" 1 \
  "COMPOSE=services:
  db:
    image: postgres:18-alpine
volumes:
  postgres_data:
    name: ekarte-postgres-data
  go_mod_cache:
    name: ekarte-go-mod-cache
  go_build_cache:
    name: ekarte-go-build-cache
" \
  -- --contract-only

# 5. Wrong expected project (staging-like) → non-zero on execute
run_case "bad-project-staging" 1 PROJECT=animalekarte-staging

# 6. COMPOSE_PROJECT_NAME mismatch → non-zero
# Need to set COMPOSE_PROJECT_NAME in the child. Extend run_case via env file trick:
build_fixture "$TMP_ROOT/bad-compose-project-name"
write_mock_docker "$TMP_ROOT/bad-compose-project-name"
write_mock_curl "$TMP_ROOT/bad-compose-project-name"
{
  set +e
  out="$(
    cd "$TMP_ROOT/bad-compose-project-name"
    env \
      COMPOSE_PROJECT_NAME=some-other-project \
      LOCAL_DB_RESET_ROOT="$TMP_ROOT/bad-compose-project-name" \
      LOCAL_DB_RESET_DOCKER="$TMP_ROOT/bad-compose-project-name/bin/docker" \
      LOCAL_DB_RESET_CURL="$TMP_ROOT/bad-compose-project-name/bin/curl" \
      LOCAL_DB_RESET_BACKUP_ROOT="$TMP_ROOT/bad-compose-project-name/backups" \
      LOCAL_DB_RESET_COMPOSE_FILE="$TMP_ROOT/bad-compose-project-name/docker-compose.yml" \
      LOCAL_DB_RESET_ENV_FILE="$TMP_ROOT/bad-compose-project-name/.env.local" \
      LOCAL_DB_RESET_MAKEFILE="$TMP_ROOT/bad-compose-project-name/Makefile" \
      LOCAL_DB_RESET_MIGRATIONS_DIR="$TMP_ROOT/bad-compose-project-name/backend/migrations" \
      bash "$TMP_ROOT/bad-compose-project-name/scripts/local-db-reset-contract.sh" 2>&1
  )"
  actual_exit=$?
  set -e
  if [[ "$actual_exit" -eq 1 ]]; then
    echo "PASS  [bad-compose-project-name] exit=$actual_exit (expected 1)"
  else
    echo "FAIL  [bad-compose-project-name] exit=$actual_exit (expected 1)"
    printf '%s\n' "$out"
    failures=$((failures + 1))
  fi
}

# 7. APP_ENV=production → non-zero
run_case "bad-app-env-production" 1 "ENVFILE=APP_ENV=production"

# 8. Snapshot dump empty → non-zero AND no volume rm
build_fixture "$TMP_ROOT/bad-empty-dump"
write_mock_docker "$TMP_ROOT/bad-empty-dump"
write_mock_curl "$TMP_ROOT/bad-empty-dump"
set_flag "$TMP_ROOT/bad-empty-dump" dump_mode empty
{
  set +e
  out="$(
    cd "$TMP_ROOT/bad-empty-dump"
    env \
      LOCAL_DB_RESET_ROOT="$TMP_ROOT/bad-empty-dump" \
      LOCAL_DB_RESET_DOCKER="$TMP_ROOT/bad-empty-dump/bin/docker" \
      LOCAL_DB_RESET_CURL="$TMP_ROOT/bad-empty-dump/bin/curl" \
      LOCAL_DB_RESET_BACKUP_ROOT="$TMP_ROOT/bad-empty-dump/backups" \
      LOCAL_DB_RESET_COMPOSE_FILE="$TMP_ROOT/bad-empty-dump/docker-compose.yml" \
      LOCAL_DB_RESET_ENV_FILE="$TMP_ROOT/bad-empty-dump/.env.local" \
      LOCAL_DB_RESET_MAKEFILE="$TMP_ROOT/bad-empty-dump/Makefile" \
      LOCAL_DB_RESET_MIGRATIONS_DIR="$TMP_ROOT/bad-empty-dump/backend/migrations" \
      bash "$TMP_ROOT/bad-empty-dump/scripts/local-db-reset-contract.sh" 2>&1
  )"
  actual_exit=$?
  set -e
  if [[ "$actual_exit" -ne 0 ]] && ! grep -E 'volume([[:space:]]|.*)+rm([[:space:]]|.*)+ekarte-postgres-data' "$TMP_ROOT/bad-empty-dump/state/docker.log" >/dev/null 2>&1; then
    echo "PASS  [bad-empty-dump] exit=$actual_exit and no volume rm"
  else
    echo "FAIL  [bad-empty-dump] exit=$actual_exit (want non-zero, no volume rm)"
    printf '%s\n' "$out"
    cat "$TMP_ROOT/bad-empty-dump/state/docker.log"
    failures=$((failures + 1))
  fi
}

# 9. Snapshot dump fail → non-zero AND no volume rm
build_fixture "$TMP_ROOT/bad-dump-fail"
write_mock_docker "$TMP_ROOT/bad-dump-fail"
write_mock_curl "$TMP_ROOT/bad-dump-fail"
set_flag "$TMP_ROOT/bad-dump-fail" dump_mode fail
{
  set +e
  out="$(
    cd "$TMP_ROOT/bad-dump-fail"
    env \
      LOCAL_DB_RESET_ROOT="$TMP_ROOT/bad-dump-fail" \
      LOCAL_DB_RESET_DOCKER="$TMP_ROOT/bad-dump-fail/bin/docker" \
      LOCAL_DB_RESET_CURL="$TMP_ROOT/bad-dump-fail/bin/curl" \
      LOCAL_DB_RESET_BACKUP_ROOT="$TMP_ROOT/bad-dump-fail/backups" \
      LOCAL_DB_RESET_COMPOSE_FILE="$TMP_ROOT/bad-dump-fail/docker-compose.yml" \
      LOCAL_DB_RESET_ENV_FILE="$TMP_ROOT/bad-dump-fail/.env.local" \
      LOCAL_DB_RESET_MAKEFILE="$TMP_ROOT/bad-dump-fail/Makefile" \
      LOCAL_DB_RESET_MIGRATIONS_DIR="$TMP_ROOT/bad-dump-fail/backend/migrations" \
      bash "$TMP_ROOT/bad-dump-fail/scripts/local-db-reset-contract.sh" 2>&1
  )"
  actual_exit=$?
  set -e
  if [[ "$actual_exit" -ne 0 ]] && ! grep -E 'volume rm' "$TMP_ROOT/bad-dump-fail/state/docker.log" >/dev/null 2>&1; then
    echo "PASS  [bad-dump-fail] exit=$actual_exit and no volume rm"
  else
    echo "FAIL  [bad-dump-fail] exit=$actual_exit (want non-zero, no volume rm)"
    printf '%s\n' "$out"
    cat "$TMP_ROOT/bad-dump-fail/state/docker.log"
    failures=$((failures + 1))
  fi
}

# 10. Makefile still uses volume wipe in reset → contract-only fails
run_case "bad-makefile-volume-wipe" 1 \
  "MAKEFILE=reset:
	docker compose down -v
	docker compose up -d --wait db backend frontend
" \
  -- --contract-only

# 11. Postflight missing migration coverage → non-zero (after delete; still fail-closed)
run_case "bad-postflight-coverage" 1 "FLAG:postflight_mode=missing_coverage"

# 12. Postflight missing seed keys → non-zero
run_case "bad-postflight-seed" 1 "FLAG:postflight_mode=missing_seed"

# 13. Postflight missing DDL keys → non-zero
run_case "bad-postflight-ddl" 1 "FLAG:postflight_mode=missing_ddl"

# 14. /health not 200 → non-zero
run_case "bad-health" 1 "FLAG:health_code=503" "FLAG:postflight_mode=bad_health"

# 15. backend unhealthy → non-zero
run_case "bad-backend-unhealthy" 1 "FLAG:postflight_mode=unhealthy"

# 16. Good execute asserts: volume rm only for postgres; snapshot files exist; caches never rm'd
build_fixture "$TMP_ROOT/good-execute-assert"
write_mock_docker "$TMP_ROOT/good-execute-assert"
write_mock_curl "$TMP_ROOT/good-execute-assert"
{
  set +e
  out="$(
    cd "$TMP_ROOT/good-execute-assert"
    env \
      LOCAL_DB_RESET_ROOT="$TMP_ROOT/good-execute-assert" \
      LOCAL_DB_RESET_DOCKER="$TMP_ROOT/good-execute-assert/bin/docker" \
      LOCAL_DB_RESET_CURL="$TMP_ROOT/good-execute-assert/bin/curl" \
      LOCAL_DB_RESET_BACKUP_ROOT="$TMP_ROOT/good-execute-assert/backups" \
      LOCAL_DB_RESET_COMPOSE_FILE="$TMP_ROOT/good-execute-assert/docker-compose.yml" \
      LOCAL_DB_RESET_ENV_FILE="$TMP_ROOT/good-execute-assert/.env.local" \
      LOCAL_DB_RESET_MAKEFILE="$TMP_ROOT/good-execute-assert/Makefile" \
      LOCAL_DB_RESET_MIGRATIONS_DIR="$TMP_ROOT/good-execute-assert/backend/migrations" \
      bash "$TMP_ROOT/good-execute-assert/scripts/local-db-reset-contract.sh" 2>&1
  )"
  actual_exit=$?
  set -e

  log="$TMP_ROOT/good-execute-assert/state/docker.log"
  ok=1
  if [[ "$actual_exit" -ne 0 ]]; then
    ok=0
  fi
  if ! grep -E 'volume rm ekarte-postgres-data|volume rm '\''ekarte-postgres-data'\''' "$log" >/dev/null 2>&1 \
     && ! grep -F 'volume rm ekarte-postgres-data' "$log" >/dev/null 2>&1 \
     && ! grep 'volume' "$log" | grep 'rm' | grep 'ekarte-postgres-data' >/dev/null 2>&1; then
    # docker.log uses %q so args are separate words: volume rm ekarte-postgres-data
    if ! grep -E 'DOCKER:.*volume.*rm.*ekarte-postgres-data' "$log" >/dev/null 2>&1; then
      ok=0
      echo "missing volume rm for db"
    fi
  fi
  if grep -E 'DOCKER:.*volume.*rm.*(ekarte-frontend-node-modules|ekarte-go-mod-cache|ekarte-go-build-cache)' "$log" >/dev/null 2>&1; then
    ok=0
    echo "cache volume was deleted"
  fi
  # snapshot present with gzip + sha256 + manifest
  snap_count="$(find "$TMP_ROOT/good-execute-assert/backups" -name 'pg_dumpall.sql.gz' 2>/dev/null | wc -l | tr -d ' ')"
  sha_count="$(find "$TMP_ROOT/good-execute-assert/backups" -name 'pg_dumpall.sql.gz.sha256' 2>/dev/null | wc -l | tr -d ' ')"
  man_count="$(find "$TMP_ROOT/good-execute-assert/backups" -name 'manifest.txt' 2>/dev/null | wc -l | tr -d ' ')"
  if [[ "$snap_count" -lt 1 || "$sha_count" -lt 1 || "$man_count" -lt 1 ]]; then
    ok=0
    echo "snapshot artifacts missing snap=$snap_count sha=$sha_count man=$man_count"
  fi
  # manifest mentions target volume and seed keys
  if ! grep -R -q 'target_volume=ekarte-postgres-data' "$TMP_ROOT/good-execute-assert/backups" 2>/dev/null; then
    ok=0
    echo "manifest missing target_volume"
  fi
  if ! grep -R -q '002_master' "$TMP_ROOT/good-execute-assert/backups" 2>/dev/null; then
    ok=0
    echo "manifest missing seed keys"
  fi
  # umask 077 → dir mode 700
  snap_dir="$(find "$TMP_ROOT/good-execute-assert/backups" -mindepth 1 -maxdepth 1 -type d | head -1)"
  if [[ -n "$snap_dir" ]]; then
    mode="$(stat -f '%Lp' "$snap_dir" 2>/dev/null || stat -c '%a' "$snap_dir" 2>/dev/null || echo '')"
    if [[ "$mode" != "700" ]]; then
      ok=0
      echo "backup dir mode=$mode want 700"
    fi
  fi

  if [[ "$ok" -eq 1 ]]; then
    echo "PASS  [good-execute-assert] exit=0 snapshot+db-only-rm+caches-kept"
  else
    echo "FAIL  [good-execute-assert]"
    printf '%s\n' "$out"
    cat "$log"
    failures=$((failures + 1))
  fi
}

# 17. Real repo contract-only should pass (static check against worktree)
{
  set +e
  out="$(bash "$CONTRACT" --contract-only 2>&1)"
  actual_exit=$?
  set -e
  # May fail until Makefile is updated — record as soft until then.
  if [[ "$actual_exit" -eq 0 ]]; then
    echo "PASS  [repo-contract-only] exit=0"
  else
    echo "FAIL  [repo-contract-only] exit=$actual_exit (Makefile/compose must satisfy contract)"
    printf '%s\n' "$out"
    failures=$((failures + 1))
  fi
}

echo "----"
if [[ "$failures" -gt 0 ]]; then
  echo "RESULT  $failures case(s) failed"
  exit 1
fi
echo "RESULT  all cases passed"
exit 0
