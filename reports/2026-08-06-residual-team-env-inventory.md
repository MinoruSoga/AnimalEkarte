# Residual team env inventory (STATUS §1)

| Field | Value |
|-------|--------|
| Date | 2026-08-06 |
| Agent | Agent-Env-Inventory (AnimalEkarte residual team) |
| Repo | `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte` |
| Branch / HEAD | `main` @ `2515079df` (`docs: point implement/task-create skills at STATUS.md SoT`) |
| Scope | STATUS.md §1 residual USER/ops only — read-only evidence |
| Forbidden ops not run | `make migrate` / `make seed` / `make reset` / volume rm / force-push / claim delete |

---

## 0. Raw environment measurements

### 0.1 Docker Compose

Command: `docker compose --env-file .env.local ps -a`

| Name | Service | Status | Ports |
|------|---------|--------|-------|
| `animalekarte-db-1` | db | **Up 29h (healthy)** | `0.0.0.0:5434->5432/tcp` |
| `animalekarte-backend-1` | backend | **Exited (1)** ~29h ago | (none) |
| `animalekarte-frontend-1` | frontend | **Created** (never started) | (none) |
| `animalekarte-backend-run-9bb08aa014fd` | backend (one-off) | Up 27h (healthy) | `8080/tcp` **not published to host**; `Cmd=["printenv"]` |

Compose services (config): `db`, `backend`, `frontend`, `codegen`.

Host listeners:

- **5434**: listening (OrbStack → db)
- **8080**: not listening on host
- **3003**: not listening on host

`.env.local`: **present** (779 bytes). Keys present (names only): `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSL_MODE`, `APP_ENV`, `PORT`, `GIN_MODE`, `LOG_LEVEL`, `JWT_SECRET`, `CORS_ALLOWED_ORIGIN`, `COOKIE_DOMAIN`, `COOKIE_CROSS_DOMAIN`, `DB_RESET`, `DEV_ADMIN_EMAIL`, `DEV_ADMIN_PASSWORD`, `LIFF_MOCK`.  
No `E2E_LOGIN_*` keys in `.env.local`.  
`.env`: missing. Makefile uses `DC = docker compose --env-file .env.local`.

**Start command (USER, not run):**

```bash
make up
# equivalent: docker compose --env-file .env.local up -d --wait --wait-timeout 1200 db backend frontend
```

**Backend exit root cause (from logs, 2026-08-05):** seed bundle `003_demo` failed:

```text
failed to load hospitalizations from .../hospitalizations.csv:
ERROR: null value in column "owner_request" of relation "hospitalizations"
violates not-null constraint (SQLSTATE 23502)
```

Static CSV probe: `hospitalizations.csv` has **2 rows**, both with **empty `owner_request`**.

### 0.2 DB connectivity (read-only)

Via `docker compose --env-file .env.local exec -T db psql` (user/db from `.env.local`; secrets not printed).

| Check | Result |
|-------|--------|
| `SELECT 1` | **OK** (1 row) |
| `schema_migrations` count | **4** |
| `schema_migrations` columns | `filename`, `checksum`, `executed_at` (PK on `filename`) |
| Recorded keys | `001_init.sql`, `seeds/002_master`, `seeds/003_demo`, `seeds/004_staging` |
| `exam_reference_ranges` table | **exists** (`to_regclass` → `exam_reference_ranges`) |
| `exam_reference_ranges` row count | **0** |
| `lab_import_jobs` | exists, count **0** |
| `lab_import_events` | exists, count **0** |
| `checkup_package_import_receipts` | exists, count **0** |
| `checkup_packages` table name | **absent** (expected: package import uses receipts + related checkup_* tables) |
| `hospitalizations` count | **2** |

Checksum reconciliation (file/bundle vs recorded):

| Key | Disk SHA-256 vs DB | Match? |
|-----|--------------------|--------|
| `001_init.sql` | `92bf199d…2c4afd9b` | **YES** |
| `seeds/002_master` | computed vs recorded | **YES** |
| `seeds/003_demo` | disk `4b2fe7f9…` vs DB `dbf14592…` | **NO — mismatch** |
| `seeds/004_staging` | computed vs recorded | **YES** |

Implication: any fresh `make migrate` / backend entrypoint migrate that re-validates `seeds/003_demo` will **fail closed on checksum mismatch**. Local recovery path is USER `make reset` per `docs/ops/deploy/LOCAL_DB_RESET.md` (not run here). After reset, seed may still fail until `hospitalizations.owner_request` empties are fixed or schema/seed aligned.

### 0.3 Seed static

| Item | Result |
|------|--------|
| `backend/migrations/seeds/003_demo/exam_reference_ranges.csv` | **present** (21 lines = header + 20 data rows) |
| Manifest entry | `003_demo/manifest.json` includes `table: exam_reference_ranges` / `csvFile: exam_reference_ranges.csv` |
| `python3 scripts/verify_seed.py` | **OK** (exit 0) — static graph checks; does **not** prove DB apply |

### 0.4 E2E_LOGIN_* (SET/UNSET only)

| Location | `E2E_LOGIN_EMAIL` | `E2E_LOGIN_PASSWORD` |
|----------|-------------------|----------------------|
| Host shell | **UNSET** | **UNSET** |
| `.env.local` | **UNSET_OR_EMPTY** | **UNSET_OR_EMPTY** |

### 0.5 Claim branches

```text
git branch -a | rg -i claim  → (no matches)
git ls-remote --heads origin 'claim/*' → empty
```

Local branches only: `main` (tracking origin), `production`, `chore/bugmd-loop-driver`.  
**No leftover `claim/*` branches** on local or origin at inventory time. Historical STATUS text still mentions retained claims in old BUG notes; live git does not.

### 0.6 Open GitHub issues

```bash
gh issue list --state open --json number --jq length
# → 18
```

Open numbers: `#89 #97 #98 #99 #201 #211 #249 #250 #252 #253 #254 #255 #256 #257 #258 #259 #261 #284`.

### 0.7 Migrations related to TASK-032 / TASK-374

On-disk DDL is consolidated: only `backend/migrations/001_init.sql` (+ `seeds/`).

| Residual ID | Schema location in `001_init.sql` | Local DB state |
|-------------|-----------------------------------|----------------|
| TASK-032 (lab import) | ~L3204+ former `005_add_lab_import_tables.sql`; `lab_import_jobs` / `lab_import_events`; later lab_import retraction/receipt tables present in `\dt` | Tables **present** (empty). DDL checksum **matches** recorded `001_init.sql` |
| TASK-374 (checkup package import) | ~L5039+ `006_checkup_package_import.sql` / `TASK-374 / #211 / DEC-59`; creates `checkup_package_import_receipts` | Table **present** (empty) |

There are **no separate top-level** `005_*.sql` / `006_*.sql` migration files anymore — apply means “this env has current `001_init.sql` + migrate runner coverage,” not an independent file to run.

---

## 1. Per-residual inventory (STATUS §1)

Legend: **agent can do now?** = non-destructive product work allowed under current gates (no migrate/seed/reset/claim-delete).

### TASK-004 — land 時 screens-drift 隔離

| Field | Content |
|-------|---------|
| Current evidence | Procedure doc: `reports/2026-07-31-task-004-005-land-proc.md`. Working tree is clean main (no open land set measured this session). Ops discipline residual each land. |
| Agent can do now? | **no** (USER land staging / commit boundaries) |
| USER next action | On next land: path-scope intentional vs foreign paths per `reports/2026-07-31-task-004-005-land-proc.md`, then commit only intentional set. |

### TASK-005 — land 前 closed-pack 回帰

| Field | Content |
|-------|---------|
| Current evidence | Same land-proc report; historical gate examples: `check-docs-symbol-drift.sh`, scoped hosp tests. Not re-run as a land gate this session (no land set). |
| Agent can do now? | **no** (USER land gate; agent may assist scoped checks only when a land set exists) |
| USER next action | Before land: re-run closed-pack checks from land-proc (symbol-drift + scoped BE/FE) and only then commit. |

### TASK-009 — 003_demo seed の DB 適用

| Field | Content |
|-------|---------|
| Current evidence | Static: CSV + manifest + `verify_seed.py` **GREEN**. DB: `exam_reference_ranges` **0 rows** while CSV has **20**. `seeds/003_demo` **checksum mismatch** (disk ≠ `schema_migrations`). Backend last exit: hospitalizations `owner_request` NOT NULL on empty CSV cells. |
| Agent can do now? | **no** (seed apply / reset forbidden; data fix may be separate product unit if USER opens it) |
| USER next action | After confirming local-only risk: fix `003_demo/hospitalizations.csv` empty `owner_request` (or approve seed fix), then `make reset` (or equivalent LOCAL_DB_RESET) so `exam_reference_ranges` and full 003_demo land; confirm `SELECT COUNT(*) FROM exam_reference_ranges` > 0. |

### TASK-010 — scenarios 要実測の残

| Field | Content |
|-------|---------|
| Current evidence | `reports/BROWSER_VERIFICATION_BACKLOG.md` — campaign batch rows still **UNREPORTED**; STATUS points to batch5 + backlog. Frontend/backend not serving localhost:3003/8080 now. |
| Agent can do now? | **no** (USER browser; env down) |
| USER next action | Bring stack up (`make up` after seed healthy), then execute backlog scenarios and fill result column. |

### TASK-020 — Playwright runtime（要 E2E_LOGIN_*）

| Field | Content |
|-------|---------|
| Current evidence | Host + `.env.local`: `E2E_LOGIN_EMAIL`/`PASSWORD` **UNSET**. Forwarding code exists per `reports/2026-07-31-task-020-env-forward.md`. App ports not published. |
| Agent can do now? | **no** (blocked on host credentials + running stack) |
| USER next action | Export `E2E_LOGIN_EMAIL`/`E2E_LOGIN_PASSWORD` on host (do not commit), start stack, re-run Playwright via `frontend/scripts/run-e2e.sh`. |

### TASK-021 — exclusion 破壊削除

| Field | Content |
|-------|---------|
| Current evidence | Prep reports exist (`reports/2026-07-31-task-021-*.md`). STATUS: wait for PO destructive approval + external use confirmation. |
| Agent can do now? | **no** (gate: PO approval) |
| USER next action | Obtain PO written approval for destructive exclusion delete, then open a single implementation unit. |

### TASK-022 — #239 S13 手動 correction / RLS 証跡

| Field | Content |
|-------|---------|
| Current evidence | Human residual only (STATUS). No agent-owned open code task. Runtime stack currently not serving UI. |
| Agent can do now? | **no** |
| USER next action | With auth env up, run S13 manual correction + capture signer/RLS runtime proof outside repo secrets. |

### TASK-023 — #254 5 フロー UAT

| Field | Content |
|-------|---------|
| Current evidence | Open issue **#254** still open (part of 18). Needs `E2E_LOGIN_*` + browser UAT. Stack down. |
| Agent can do now? | **no** |
| USER next action | Set `E2E_LOGIN_*`, start stack, complete five business-flow UAT and record non-secret results on #254. |

### TASK-024 — #256 screenshot / FAQ sign-off

| Field | Content |
|-------|---------|
| Current evidence | Open issue **#256**. Human visual/FAQ sign-off residual. |
| Agent can do now? | **no** |
| USER next action | Capture screenshots / FAQ visual sign-off per #256 and DEC-61 (no-rewrite default); store roster/receipts outside repo. |

### TASK-032-apply — lab import migration 適用 + claim 解放

| Field | Content |
|-------|---------|
| Current evidence | Lab import DDL folded into `001_init.sql`; local tables `lab_import_*` **exist**; `001` checksum **matches**. **No `claim/*` branches** remain. Product code already on main (STATUS). |
| Agent can do now? | **no** (apply/claim release are USER; local DDL already applied) |
| USER next action | On any env still pre-consolidation: `make migrate` or approved reset; if any leftover claim refs outside this clone, delete locally after integrate — this clone has none. |

### TASK-033 — #201 救急投薬 cutover

| Field | Content |
|-------|---------|
| Current evidence | STATUS **BLOCKED** until clinical input + decision SoT + DB review. Open issue **#201**. |
| Agent can do now? | **no** |
| USER next action | Clinical owner fills canonical #201 bundle (cap/warning + emergency record policy with sources); only then allow agent cutover unit. |

### TASK-374-apply — checkup package import migration 適用

| Field | Content |
|-------|---------|
| Current evidence | `006_checkup_package_import` section in `001_init.sql` (~L5039); local `checkup_package_import_receipts` **exists** (0 rows). Open issue **#211** still USER/clinical. |
| Agent can do now? | **no** |
| USER next action | Confirm migrate on each target env; clinic import/rollback evidence + clinical row approval per #211 (manifests stay outside repo). |

### TASK-378-reset — 001 統合後の環境 DB_RESET

| Field | Content |
|-------|---------|
| Current evidence | Local `001_init.sql` checksum **matches**, but **`seeds/003_demo` mismatches** → reset still required for seed currency. Runbook: `docs/ops/deploy/LOCAL_DB_RESET.md` / `make reset`. Agent must not run. |
| Agent can do now? | **no** |
| USER next action | Local only: `make reset` after accepting data loss + backup contract; fix hospitalizations seed first if postflight still fails on `owner_request`. |

### POST-PULL — 各環境 `make migrate`

| Field | Content |
|-------|---------|
| Current evidence | Local DDL already recorded and matching. Seed mismatch means migrate alone is insufficient for 003_demo refresh (checksum fail-closed). Other envs not probed. |
| Agent can do now? | **no** |
| USER next action | On each env after pull: `make migrate` if keys missing; if checksum mismatch on 001 or seeds, follow env-specific reset (local: `make reset`). |

### SCEN-OPS-CLAIM — claim ブランチ解放

| Field | Content |
|-------|---------|
| Current evidence | `git branch -a` / `git ls-remote origin 'claim/*'`: **zero claim branches**. Residual may already be cleared for this repo remote. |
| Agent can do now? | **no** (claim delete is USER-only even if none found) |
| USER next action | Optional: re-scan other machines/worktrees for stale `claim/*` and delete after merge; nothing to delete on this origin snapshot. |

### LINE-R05 — production rollout + column DROP（HOLD）

| Field | Content |
|-------|---------|
| Current evidence | `reports/2026-07-31-r05-single-sot-phase-a.md` / `phase-b.md`: CODE GREEN paths with **DROP HOLD** / rollout HOLD; STATUS restates HOLD. |
| Agent can do now? | **no** |
| USER next action | Production rollout inventory (legacy column empty) then separate packet for column DROP after PO/ops gate. |

### R6/R7 — worktree 隔離 / empty-diff COMPLETE 禁止

| Field | Content |
|-------|---------|
| Current evidence | Continuous ops discipline (STATUS). Single worktree: main checkout only (`git worktree list` → one entry). Report `reports/2026-07-31-line-r06-r07-nav-honesty.md` historical. |
| Agent can do now? | **yes** (obey discipline on future units: isolated worktree, no empty-diff COMPLETE) |
| USER next action | Enforce in reviews: require real diff + worktree isolation evidence before COMPLETE claims. |

---

## 2. Recommended USER order (revalidated)

1. **Seed data unblock** — fix `003_demo/hospitalizations.csv` empty `owner_request` (else reset/migrate seed will fail again).  
2. **TASK-378-reset / TASK-009** — local `make reset` to clear `seeds/003_demo` checksum mismatch and load `exam_reference_ranges` (expect 20 rows).  
3. **`make up`** — backend + frontend currently not serving; need healthy stack for UAT/E2E.  
4. **POST-PULL / TASK-032-apply / TASK-374-apply** — other envs: migrate or approved rebuild; local DDL already present.  
5. **E2E_LOGIN_*** → TASK-020 / TASK-023 / #254.  
6. **TASK-010** + browser backlog fill-in.  
7. **TASK-022 / TASK-024** human evidence.  
8. **TASK-033 / #201** clinical SoT then agent.  
9. **TASK-021** only after PO destructive approval.  
10. **LINE-R05** remains HOLD.  
11. **TASK-004/005 / R6/R7** every land/session.

---

## 3. Snapshot matrix

| ID | Blocker class | Stack/DB evidence | Agent now? |
|----|---------------|-------------------|------------|
| TASK-004 | USER land ops | land-proc doc only | no |
| TASK-005 | USER land ops | land-proc doc only | no |
| TASK-009 | USER seed/reset | static OK; DB 0 ranges; 003 checksum mismatch; hosp NOT NULL | no |
| TASK-010 | USER browser | backlog UNREPORTED; UI down | no |
| TASK-020 | USER secrets + stack | E2E_LOGIN UNSET; ports down | no |
| TASK-021 | PO gate | prep reports only | no |
| TASK-022 | USER human | stack down | no |
| TASK-023 | USER UAT | #254 open; login UNSET | no |
| TASK-024 | USER visual | #256 open | no |
| TASK-032-apply | USER env apply | local lab_import tables present; no claims | no |
| TASK-033 | clinical gate | #201 open | no |
| TASK-374-apply | USER env + #211 | local receipt table present | no |
| TASK-378-reset | USER local reset | 003_demo checksum mismatch | no |
| POST-PULL | USER each env | local 001 match; seed mismatch | no |
| SCEN-OPS-CLAIM | USER (maybe done) | 0 claim branches | no |
| LINE-R05 | HOLD | phase A/B HOLD | no |
| R6/R7 | ongoing discipline | 1 worktree | yes (discipline only) |

---

## 4. Commands replayed (non-destructive)

```bash
docker compose --env-file .env.local ps -a
docker compose --env-file .env.local exec -T db psql -U "$DB_USER" -d "$DB_NAME" -c 'SELECT 1'
# + COUNT/schema_migrations/to_regclass/COUNT exam_reference_ranges (read-only)
python3 scripts/verify_seed.py
git branch -a | rg -i claim
git ls-remote --heads origin 'claim/*'
gh issue list --state open --json number --jq length
# SHA-256 of 001_init.sql and seed bundleChecksum-equivalent for 002/003/004
```

---

## 5. Honesty limits

- Did **not** start containers, migrate, seed, or reset.  
- Did **not** print secret values (`DB_PASSWORD`, JWT, DEV_ADMIN_*, etc.).  
- STG/PROD connectivity not measured (local compose only).  
- One-off `backend-run` container is healthy but not a substitute for the compose `backend` service (no host port publish; `printenv` command).  
- Claim history inside STATUS BUG footnotes may lag live git; **git is evidence of record** for SCEN-OPS-CLAIM here.
