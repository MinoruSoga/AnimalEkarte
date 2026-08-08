# Residual Closeout Ledger

> **Role:** Live residual-closeout run state only (not a second product backlog).  
> **SoT for product residual:** root `STATUS.md` + `PO-todo.md`.  
> **Policy:** `docs/work/decisions/fable-po-recommendation.md`.

## Frozen unit order (do not re-order without USER)

| Unit | Scope | External write? | Status |
|------|--------|-----------------|--------|
| **U0** | Tree hygiene + claim release report | No (local write) | **COMPLETE** (2026-08-08) |
| **U1** | PO-01 · GitHub #98 close path | Yes | pending — next |
| **U2** | PO-02 · GitHub #99 close path | Yes | pending |
| **U3** | PO-03 · #252 ↔ #257 go-live gate Yes/No | Yes (or STATUS one-liner) | pending |
| **U4** | PO-06 · TASK-023 / #254 five-flow UAT | Yes (evidence) | pending |
| **U5** | TASK-024 / #256 screenshot-FAQ sign-off | Yes (evidence) | pending |
| **U6** | TASK-022 / #239 S13 / RLS evidence | Yes (evidence) | pending |

**HOLD** (separate decision required): TASK-021 B/C/D, LINE-R05 DROP, TASK-033/#201 clinical, #249 clinical ranges.

---

## U0 — Tree hygiene + claim release report

| Field | Value |
|-------|--------|
| Date | 2026-08-08 |
| Branch / HEAD | `main` @ `264777c882c86d6ad5f9d7c3d22f047588fc23fc` |
| Run status | **COMPLETE** |
| Next unit | **U1 = PO-01 · #98** (do not auto-chain) |

### Working-tree decisions

| Path | Classification | Action |
|------|----------------|--------|
| `.gitignore` (+ `.code-review-graph/`) | Keep — local tooling ignore; not on `origin/main` | local commit (no push) |
| `backend/internal/billing/accounting_complete_appointments_test.go` | Noise — import order only (`testdb`/`reservation` swap) | path-scoped `git restore` |
| `backend/internal/billing/accounting_repository_tx_atomicity_test.go` | Noise — same import order only | path-scoped `git restore` |
| `reports/` (incl. `reports/bug-md-2agent-loop/`) | Local loop artifacts | leave **untracked**; do not stage; not gitignored (optional future ignore is PO-side memo only) |

### Claim ancestry (re-measured)

| Claim | Tip | `merge-base --is-ancestor … main` | Release |
|-------|-----|-------------------------------------|---------|
| `claim/ARCH-A4` | `26380f0b2` | exit **0** (ancestor) | USER: `git branch -D claim/ARCH-A4` |
| `claim/ARCH-A7` | `4b349d796` | exit **0** (ancestor) | USER: `git branch -D claim/ARCH-A7` |

Agent **did not** delete claim branches.

### Gate commands (verbatim evidence in Completion Report)

1. `git status -sb` / `git diff --stat` — inventory
2. Path-scoped restore of two billing tests → empty diff for those paths
3. Claim ancestry exit 0 for A4/A7
4. Post-commit: `git status -sb` shows only `?? reports/` (and no claim deletes)

### Notes for PO

- `reports/` remains untracked by design for this unit; adding a permanent `.gitignore` rule is optional and not required for U0.
- Also present local branch (not in U0 release set): `chore/bugmd-loop-driver`.
- U0 does **not** authorize starting U1 in the same session.
