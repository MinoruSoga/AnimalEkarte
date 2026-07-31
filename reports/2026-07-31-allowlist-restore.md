# Allowlist restore — `TODO-MD-ALLOWLIST-RESTORE-20260731`

**Date:** 2026-07-31  
**HEAD:** `b5df79bc88d491f70e29308b56dfda2ac842d0c7` (unchanged; no commit)

## Action

Path-scoped restore of overspill thrash from orchestrated unit `TODO-MD-ALL-OPEN-ORCHESTRATED-20260731` (OLDDB/csv-import narrative outside packet allowlists).

```bash
git restore --source=HEAD --staged --worktree -- \
  backend/migrations/CLAUDE.md \
  docs/ops/deploy/ANIMALEKARTE_CSV_IMPORT_COMPLETION.md \
  docs/ops/deploy/CLINIC_CSV_IMPORT.md \
  docs/ops/deploy/README.md \
  docs/ops/deploy/SEED_MIGRATION_OPERATIONS.md
```

Note: plain `git restore -- <path>` was insufficient when index was staged; used `--source=HEAD --staged --worktree`.

## Result

- Five thrash paths: empty `git diff` / empty `git diff --cached` / absent from porcelain
- Intentional HOSP/docs/todo/reports packets preserved
- Foreign `backend/migrations/seeds/003_demo/line_reservation_settings.csv` still modified
- Claims not deleted; no migrate/push
