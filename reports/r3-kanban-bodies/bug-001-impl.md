# [実装] BUG-001 — deceased pet FE block on accounting new (direct URL)

## Scope (single BUG)
- Only BUG-001: `/accounting/new?petId=...` for deceased pets must not show a working settlement form; create/complete already rejected on BE (ALREADY on main).
- Companion BE already: assertAccountingPetNotDeceased + unit PASS. Do not re-land BE unless gap found.
- Out of scope: BUG-002/033/034/035, other accounting features, migrate, merge, push.

## SoT
- Repo worktree from this task. bug.md BUG-001 section.
- CLAUDE.md: Docker only tests. No make migrate. No PII in logs.

## DoD
1. Direct URL with deceased petId: FE blocks UI (message + no editable create path) before save.
2. Living pet path unchanged.
3. Vitest for guard; optional BE regression if touched.
4. Docker tests MUST use:
   docker compose run --rm --no-deps --entrypoint '' -v <worktree>/backend:/app -e TZ=Asia/Tokyo backend go test …
   Without --entrypoint '', air starts and hangs.
5. Handoff under CorpVault/55_Handoff with branch tip SHA.
6. End in review-required (orchestrator will complete for verify). Do not merge/push.

## Forbidden
- Other BUGs, force-push, migrate apply, staging, secrets in task output.
