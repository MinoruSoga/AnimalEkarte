# [実装] BUG-002 — deceased pet FE hard block on medical-records/new

## Scope
- Only BUG-002: `/medical-records/new?petId=deceased` must not show full editable form (UAT showed form with 【死亡】 banner).
- BE Create reject ALREADY on main + unit PASS. Strengthen FE only unless BE gap.
- Out of scope: other BUGs, merge, push, migrate.

## SoT
- MedicalRecordForm + pet selection / new route.
- bug.md BUG-002.

## DoD
1. Deceased petId on new: hard stop (redirect or non-editable block + clear message); no create mutation.
2. Living pet OK.
3. Vitest coverage for deceased new route.
4. Docker FE tests with --entrypoint '' / vitest as per CLAUDE.md.
5. Handoff + review-required. No merge/push.

## Docker
docker compose run --rm --no-deps --entrypoint '' -v <WT>/frontend/src:/app/src frontend npx vitest run <paths>
Never omit --entrypoint '' on backend go test.
