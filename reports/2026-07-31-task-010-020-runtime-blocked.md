# TASK-010 / TASK-020 — Runtime verification BLOCKED

> **Date**: 2026-07-31  
> **Status**: **BLOCKED** (docker app environment not verified in this agent session)  
> **Scope**: Documentation of blocker + exact USER commands. No browser/Playwright execution performed.

---

## 1. Why BLOCKED

| Check | This session |
|:---|:---|
| `docker compose ps` / healthy frontend+backend+db | **Not verified** |
| Login / seed data availability for browser scenarios | **Not verified** |
| Playwright browser binaries / baseURL | **Not verified** |
| Chrome DevTools MCP target `http://127.0.0.1:9222` | **Not used** |

Project rules forbid agents from auto-running full-stack bring-up (`docker compose up`, full e2e suites, DB reset). Without a confirmed healthy app stack, TASK-010 browser measurements and TASK-020 full Playwright runtime would produce false failures or empty evidence.

**Disposition**: both tasks remain **open**; evidence must land in a future `reports/YYYY-MM-DD-<env>.md` after USER verifies env.

---

## 2. TASK-010 — scenarios 【要実測】 browser-test

### Intent

- Residual: **SCEN-BROWSER-001**  
- ~**78** `【要実測】` marks under `docs/ops/testing/scenarios` (re-count at start; exclude historical `scenarios/reports/**` if desired)  
- Measure → elevate expected result **or** BUG **or** 要PO  
- Prefer post **TASK-009** seed for clinical search-dependent cases  

### USER commands (suggested)

```bash
# A) Confirm stack
cd /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte
docker compose ps
# Expect frontend / backend / db healthy (exact service names per compose file)

# B) If migrations/seeds pending after pull (USER only):
make migrate

# C) Open app (typical local URLs — confirm against compose/README if different)
# Frontend: http://localhost:3000  (or project-documented port)
# API:      http://localhost:8080  (or project-documented port)

# D) List open measurement marks
rg -n '【要実測】' docs/ops/testing/scenarios --glob '!**/reports/**'

# E) Execute with browser-test skill / manual scenario packs
# Record results ONLY in:
#   reports/YYYY-MM-DD-local.md
# Do not paste long run logs into scenario markdown bodies.
```

Optional Chrome DevTools path (project MCP): browser on `http://127.0.0.1:9222` with app already running.

### Exit criteria (from todo)

1. Marks = 0 **or** each remaining item is PO待ち / BUG-xxx with table in reports  
2. PASS items have marks removed + expected results written  
3. Execution report exists under `reports/`

---

## 3. TASK-020 — ui-design-compliance Playwright (84 product pages)

### Intent

- Residual: **SPEC-TOP-E2E-RUNTIME-84**  
- Static inventory already expects **84** pages (`route-inventory.test.tsx`, e2e inventory)  
- Last full runtime cited in `docs/spec/ui-design-compliance.md`: **2026-07-23** (83 pages / 92 tests) — **84-leaf runtime deferred**  
- Expectation after 84 inventory: **93** tests (doc claim; re-confirm when running)

### USER commands (suggested)

```bash
cd /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte

# 1) Stack healthy (same as TASK-010)
docker compose ps

# 2) From frontend container (preferred project pattern):
docker compose exec frontend pnpm test:e2e -- e2e/ui-design-compliance-readonly.spec.ts --workers=1

# If playwright must run on host (only if project docs allow; Docker preferred):
# cd frontend && pnpm test:e2e -- e2e/ui-design-compliance-readonly.spec.ts --workers=1
```

Also useful static gate (not a substitute for runtime):

```bash
docker compose exec frontend pnpm test:run -- src/app/routes/route-inventory.test.tsx
```

### Exit criteria (from todo)

1. Runtime PASS count or failure list recorded once (reports or ui-design-compliance runtime date)  
2. No contradiction with inventory 84  
3. Failures not silent — branch to BUG / inventory fix TASK  

---

## 4. Shared prerequisites checklist (USER)

- [ ] Docker Desktop (or engine) running  
- [ ] `docker compose up` already done by USER (agent must not auto-up)  
- [ ] `make migrate` applied if migrations/seeds changed since last pull  
- [ ] Demo login credentials available (from local seed docs — do not commit secrets)  
- [ ] Playwright deps installed **inside** the frontend image/volume as per project  
- [ ] No parallel agent wiping worktree / foreign WIP  

---

## 5. What this agent did / did not do

| Did | Did not |
|:---|:---|
| Document BLOCKED with commands | Run Playwright |
| Point to mark census / inventory | Run browser-test scenarios |
| Link seed design / PO packets | `docker compose up/down`, full `pnpm test:e2e`, DB reset |

---

## 6. Related reports

| Report | Relation |
|:---|:---|
| `reports/2026-07-31-task-009-seed-design.md` | Seed before clinical browser cases |
| `reports/2026-07-31-task-015-audit-medium.md` | MEDIUM ↔ TASK-010 mapping |
| `reports/2026-07-31-todo-po-decisions.md` | Marks that should become 要PO not “failed tests” |
