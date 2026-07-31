# W-020-ENV: E2E_LOGIN_* forward into Playwright Docker

**Task**: TODO-MD-NEXT-ORCH-WAVE-20260731 / W-020-ENV  
**Branch**: `agent/w-020-env`  
**Worktree**: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte-w020`  
**Date**: 2026-07-31  
**Claim**: `claim/W-020-ENV` (user release after merge)

## Summary

`frontend/scripts/run-e2e.sh` now forwards host auth env vars into the Playwright
Docker container **only when set**, using name-only `-e VARNAME` (no `=value` on argv).
No secrets were invented or printed. Authenticated E2E re-run remains **BLOCKED** by
missing host credentials.

## Before / after

### Before

`docker run` always injected only:

```text
-e PLAYWRIGHT_TEST_BASE_URL="${BASE_URL}"
```

`E2E_LOGIN_EMAIL`, `E2E_LOGIN_PASSWORD`, and `E2E_AUTH_STATE_PATH` set on the host
were **not** visible inside the Playwright container. Authenticated specs depending on
`helpers/auth.ts` could not see host credentials when run via `./scripts/run-e2e.sh`.

### After

Build a `DOCKER_ENV` flag list after `BASE_URL=`:

```sh
DOCKER_ENV="-e PLAYWRIGHT_TEST_BASE_URL=${BASE_URL}"
if [ -n "${E2E_LOGIN_EMAIL:-}" ]; then DOCKER_ENV="$DOCKER_ENV -e E2E_LOGIN_EMAIL"; fi
if [ -n "${E2E_LOGIN_PASSWORD:-}" ]; then DOCKER_ENV="$DOCKER_ENV -e E2E_LOGIN_PASSWORD"; fi
if [ -n "${E2E_AUTH_STATE_PATH:-}" ]; then DOCKER_ENV="$DOCKER_ENV -e E2E_AUTH_STATE_PATH"; fi
# shellcheck disable=SC2086 # intentional for DOCKER_ENV flag list only
docker run ... $DOCKER_ENV ... -- "$@"
```

- Set vars: forwarded as `-e VARNAME` (Docker copies host value; no secret on argv).
- Unset vars: not injected (no empty-string overwrite inside container).
- Test args remain injection-safe via `"$@"`.
- No hardcoded passwords.

`frontend/e2e/README.md` adds a one-line honesty note under the Docker runner section.

## rg evidence (`E2E_LOGIN` in script)

```text
$ rg -n 'E2E_LOGIN' frontend/scripts/run-e2e.sh
32:if [ -n "${E2E_LOGIN_EMAIL:-}" ]; then DOCKER_ENV="$DOCKER_ENV -e E2E_LOGIN_EMAIL"; fi
33:if [ -n "${E2E_LOGIN_PASSWORD:-}" ]; then DOCKER_ENV="$DOCKER_ENV -e E2E_LOGIN_PASSWORD"; fi
```

Related (optional path + DOCKER_ENV usage):

```text
$ rg -n 'E2E_LOGIN|E2E_AUTH|DOCKER_ENV' frontend/scripts/run-e2e.sh
31:DOCKER_ENV="-e PLAYWRIGHT_TEST_BASE_URL=${BASE_URL}"
32:if [ -n "${E2E_LOGIN_EMAIL:-}" ]; then DOCKER_ENV="$DOCKER_ENV -e E2E_LOGIN_EMAIL"; fi
33:if [ -n "${E2E_LOGIN_PASSWORD:-}" ]; then DOCKER_ENV="$DOCKER_ENV -e E2E_LOGIN_PASSWORD"; fi
34:if [ -n "${E2E_AUTH_STATE_PATH:-}" ]; then DOCKER_ENV="$DOCKER_ENV -e E2E_AUTH_STATE_PATH"; fi
38:# shellcheck disable=SC2086 # intentional for DOCKER_ENV flag list only
41:  $DOCKER_ENV \
```

## Host credentials (names only — no values)

| Variable | Host status |
|----------|-------------|
| `E2E_LOGIN_EMAIL` | **EMAIL_UNSET** |
| `E2E_LOGIN_PASSWORD` | **PASSWORD_UNSET** |
| `E2E_AUTH_STATE_PATH` | AUTH_STATE_UNSET |

## Runtime

> **BLOCKED credentials** — host has EMAIL_UNSET / PASSWORD_UNSET.  
> Do **not** invent secrets. Do **not** re-run authenticated e2e with fake credentials.  
> Env-forward patch is complete; authenticated runtime verification remains blocked until
> a human injects real local/demo credentials into the host shell/CI secrets store.

## Files changed (allowlist)

| Path | Change |
|------|--------|
| `frontend/scripts/run-e2e.sh` | Conditional name-only `-e` forward for auth env |
| `frontend/e2e/README.md` | Honesty line: Docker runner forwards when set |
| `reports/2026-07-31-task-020-env-forward.md` | This report |

## Commands used

```bash
git branch --list 'claim/W-020-ENV'   # empty → acquire
git branch claim/W-020-ENV
# edit run-e2e.sh + README + report
rg -n 'E2E_LOGIN' frontend/scripts/run-e2e.sh
git diff --name-only
# host status: EMAIL_UNSET / PASSWORD_UNSET
# (no e2e re-run; credentials BLOCKED)
git check-ignore -v -- <paths>   # expect exit 1
git add -- <paths only>
git commit -m "fix(e2e): forward E2E_LOGIN_* into Playwright docker when set"
# no push
```

## Integration note (coordinator)

- Merge branch: **`agent/w-020-env`**
- Claim to release (user only, after main integration): `git branch -D claim/W-020-ENV`
- No secrets in tree; no migration; docs + script only.
- Downstream authenticated e2e still needs host-injected `E2E_LOGIN_EMAIL` / `E2E_LOGIN_PASSWORD`.
