---
paths:
  - "backend/**/*.go"
  - "frontend/src/**/*.{ts,tsx}"
---

# TDD Enforcement (Mandatory Process)

Test-first is a process gate, not optional guidance.

## Rules

1. **RED first** — For new behavior or bug fixes, write or extend a failing test before production code.
2. **GREEN minimal** — Implement only enough code to make that test pass.
3. **IMPROVE** — Refactor with tests green; do not expand scope mid-cycle.
4. **No silent skips** — Do not merge or claim done with failing, skipped-as-green, or TODO-only stub tests for the changed path.
5. **Colocate** — Place tests beside the unit under test (`*_test.go`, `*.test.ts(x)`).

## Verification

- Backend (scoped): `docker compose exec backend go test ./internal/<pkg>/...`
- Frontend (scoped): `docker compose exec frontend npx vitest run <path>`
  - Do **not** use `pnpm test:run -- <path>` (path is ignored; runs the full suite).
- Prefer coverage ratchet / package-scoped runs over full-repo suites in agent sessions.

## Exceptions

- Docs-only, pure config comment, or generated-file changes may skip RED/GREEN when no runtime behavior changes.
- Document the exception in the PR/commit body when skipping.
