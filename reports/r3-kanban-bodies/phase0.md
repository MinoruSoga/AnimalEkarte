# Phase 0 classify (2026-08-11) — evidence

## Environment
- main tip: 8077f00f3 (matches origin/main, ff-only clean)
- Running: frontend/backend containers Created ~2026-08-09; images frontend 2026-07-12 / backend 2026-07-31
- UAT host localhost:3003 = stale images vs Aug main (ENV_STALE for fixes after image build)

## Unit on current main (Docker --entrypoint '')
PASS billing deceased create/complete package
PASS medicalrecord deceased create + examination completed seal locks
Package ok seconds-level.

## Classification

| ID | Class | Action |
|----|-------|--------|
| BUG-001 | BE ALREADY + FE STILL_OPEN + ENV_STALE | FE deceased block on /accounting/new |
| BUG-002 | BE ALREADY + FE display STILL_OPEN + ENV_STALE | FE hard block /medical-records/new deceased |
| BUG-033 | ENV_STALE (code ALREADY on main) | rebuild only; no impl |
| BUG-034 | STILL_OPEN | treatment policy reload after finalize |
| BUG-035 | STILL_OPEN | finalized MR inputs still editable (fieldset contents) |
| BUG-027 | SPEC | do not implement |

## Human ENV rebuild
cd /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte
docker compose build frontend backend
docker compose up -d frontend backend
# make migrate only if new migrations — human only

## Campaign auth
normal merge + git push origin main OK. No force-push, no auto migrate, no staging.
