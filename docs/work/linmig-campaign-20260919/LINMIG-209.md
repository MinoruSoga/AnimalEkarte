# LINMIG-209 — Per-environment migration apply-or-not, schema-diff, backup/restore prep

Campaign `linmig-ops-prep-20260919` revision 1. Claim ID `LINMIG-209` (Linear issue **not found**; keep as claim only). Maps to [production setup](../../ops/infra/production/setup.md) migrate order, [production runbook](../../ops/infra/production/runbook.md) backup/restore, [STG PlanetScale seed runbook](../../ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md) stop gates, and AUTH-V-D1 in [todo-verification.md](../../../todo-verification.md).

Worktree: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte-linmig-209` on `feat/linmig-209-ops-prep` at HEAD `aac697645df92fd24611c7c13bf0f7dda12a6e08`. Sheet date: 2026-09-19. Repo evidence only.

**This unit does not** run `make migrate`, apply SQL, query shared databases, dispatch deploy, `wrangler secret put`, `pg_dump`/`pg_restore`, or change STG/PROD. Agents must not auto-apply migrations ([.claude/CLAUDE.md](../../../.claude/CLAUDE.md) Agent Migration Authority). STG runbook L4 additionally forbids using that runbook as a basis for agent DB query/write/reset or deploy.

**Target environment: UNKNOWN. Live revision / `headSha`: UNKNOWN.** Local checkout HEAD is not a production or STG Worker/DB revision.

Truth source order used: `backend/cmd/migrate` + `backend/migrations/*.sql` + `seeds/002_master/manifest.json` over current runbooks over this sheet. Live schema, backup existence, and restore success are external facts and stay **UNKNOWN**.

## 1. Why this sheet exists

Operators need a single local checklist that answers, per environment:

1. Apply migrate, skip, or stop?
2. What repo artifacts to diff against `schema_migrations`?
3. What backup acquisition fields must exist before any write?
4. What restore-prep identity assertions must pass on a disposable target?

This sheet is **prep**, not apply authorization. Unchecked or UNKNOWN items are No-Go / stop.

Callers: campaign controller and operators reading `docs/work/linmig-campaign-20260919/` (same pattern as `docs/work/todo-campaign-20260918/*.md`). No application import. Net-new file; no prior `LINMIG-209.md`.

## 2. Apply-or-not (do not apply from this session)

Decision for this session: **do not apply**. No target env, live revision, backup receipt, or operator approval was observed.

When a future human-approved run is considered, use the matrix. Empty cells stay UNKNOWN until filled from a protected operator record — do not invent host/db names as live facts.

### 2.1 Shared stop rules (all environments)

Stop (do **not** apply) if any of:

| ID | Stop if | Source |
|---|---|---|
| S-AGENT | The actor is an agent session, or the command is unattended `make migrate` / SQL against a shared DB | project CLAUDE.md Agent Migration Authority; STG runbook L4 (no agent query/write/reset/deploy from that runbook) |
| S-ENV | Target database/branch, data owner, operator, and window are not fixed on the run sheet | STG runbook L21; setup.md L8–L19 |
| S-REV | Reviewed git ref / intended `headSha` is not recorded (setup.md L72; runbook.md L75). This worktree HEAD is not a live Worker/DB revision (setup.md L19: do not infer runtime from the draft) | setup.md L19, L72; runbook.md L3, L75 |
| S-BACKUP | Backup acquisition record is incomplete (any field missing = No-Go) | runbook.md L44–L46; STG runbook L22 |
| S-PUBLIC | `public` schema presence is unknown or missing (prerequisite failure, not a seed failure) | STG runbook L28–L32 |
| S-EMPTY | Empty `schema_migrations` with existing `clinics` (no baseline checksum write) | `backend/migrations/CLAUDE.md` L64; `cmd/migrate` empty-history guard |
| S-CHECKSUM | DDL/CSV checksum mismatch vs recorded history | STG runbook L67; migrations CLAUDE.md L75 |
| S-MISSING | Coverage `missing>0` on last completed run, or last outcome partial/unknown | STG runbook L68, L70 |
| S-HEALTH | `/health` 200 is the only evidence (process health, not DB) | setup.md L73; runbook.md L13; STG runbook L60 |
| S-RESET | Proposed path is `DB_RESET=true`, `DROP SCHEMA`, or workflow-as-reset | STG runbook L80–L84; Makefile migrate is not reset |
| S-FALLBACK | Proposed recovery is `\copy` / partial INSERT / writing checksums into `schema_migrations` | STG runbook L72–L76 |
| S-AUTH | AUTH-V-D1 preflight still BLOCKED and the run would be used to bootstrap first admin without schema/owner confirmation | todo-verification.md AUTH-V-D1-PREFLIGHT |

### 2.2 Local disposable Compose

| Apply? | When | Path | Must not |
|---|---|---|---|
| **USER may apply** after pull if migrations/seeds changed | Developer-owned disposable DB; not shared STG/PROD | `make migrate` → `$(DC) run --rm --entrypoint go backend run ./cmd/migrate` where `DC = docker compose --env-file .env.local` (Makefile L7, L114–L115). Agents surface the command; they do not run it | `make db` / host `go run` / `docker compose exec db psql`; treating local success as STG/PROD |
| **Stop / rebuild instead** | Pre-integration `001_init.sql` checksum on an existing volume | USER-only `LOCAL_DB_RESET.md`. No no-reset upgrade path (migrations CLAUDE.md L58) | Inventing a baseline row in `schema_migrations` |

Post-pull rule: after pulling a commit that adds or changes migrations, the **developer** runs `make migrate` before using the updated app. This session did not pull a new migration and did not run it.

### 2.3 Shared STG (PlanetScale)

Documented intended identity in the rebuild section is PlanetScale Postgres `animalekarte-stg` / branch `main` (org `noah-animalekarte`) — **live confirmation UNKNOWN** (STG runbook L88, L105).

| Apply? | When | Path | Must not |
|---|---|---|---|
| **Normal apply: operator only**, after §2 stop gates | Reviewed `main -> staging` or backend workflow dispatch (STG runbook L38); always repository `cmd/migrate` | Cloudflare workflow order: deploy → `POST /_internal/migrate` → post-migrate health (STG runbook L13). Path filter auto-starts that workflow only for backend-path changes; it is not a migrate-skip decision | Agent query/write; Hyperdrive as manual migrate path (advisory lock); `pscale role reset-default` on app `postgres` password (L15, L34) |
| **Skip migrate** | No backend schema/seed change in the reviewed artifact; last coverage `missing=0` already matches current plan | Record skip reason + artifact SHA. Non-backend path changes may not auto-start the workflow (L13 path filter) | Treating skip or “workflow did not start” as “schema verified” without a dated coverage line |
| **Rebuild / `DB_RESET`: separate approval** | Target, data loss, backup/restore, downtime, operator, approval all present | STG runbook §6. `backend-deploy.yml` has **no** `db_reset` input; Worker migrate does **not** pass `DB_RESET` — workflow is **not** a reset path (L84) | Partial table repair, restoring retired `003_demo`/`004_staging`, stacking reruns on unknown outcome (L70, L82) |

Pre-deploy stop gates (any unmet → stop deploy/rebuild) — STG runbook L19–L26:

1. Target database/branch, data owner, operator, maintenance window fixed on the run sheet.
2. What to keep, verified backup/restore, rollback decision confirmed.
3. Target Wrangler `secrets.required` names and vars confirmed **without displaying values**.
4. Artifact contains top-level DDL, `002_master/manifest.json`, and every listed CSV.
5. If legacy keys exist: current master-only translation vs target master completeness, plus reviewed recovery plan.
6. Relevant backend STG/production gate and billing are green.

### 2.4 Production

| Apply? | When | Path | Must not |
|---|---|---|---|
| **Do not apply now** | setup.md L3, L19, L68: do not execute until approved go-live date and Linear delivery item are confirmed; stop if any fail-closed item is unknown; at source review `7c6592f9f`, `backend-deploy.yml` lacked production trigger/Environment gate | Future activated workflow only, after setup.md §§1–6 are **implemented and verified**: **deploy → migrate → `/health` → optional smoke** (setup.md L66; runbook.md L11) | `make migrate` against production; inferring live resources from drafts (setup.md L5); staging secret reuse unless explicitly approved (setup.md L13); `APP_ENV=development\|test\|staging` / login/demo seed on production (setup.md L50); `DB_RESET` via workflow (STG runbook L84 — workflow is not a reset path) |

Fail-closed preconditions (setup.md L8–L19) — all need dated human evidence before **any** production deploy/migrate:

- target account/zone/project and billing approved
- production database, role, backup/restore owner, R2, DNS, certificate verified
- GitHub Environment with required reviewers; exact case-sensitive name matches the workflow
- production secrets environment-scoped; staging values not reused unless explicitly approved
- frontend production settings verified (`Production` Environment; `VERCEL_ENV=production` selects API)
- workflow/config tests and `actionlint` green on the reviewed change
- current Linear delivery state and go-live date confirmed externally

Production seed contract (setup.md L47–L50; runbook.md L13, L65): CSV is **master-only** (`002_master`) in every environment. Production draft omits `APP_ENV`, so login seeding and the shared demo-password shortcut stay disabled by the empty-value gate. Production must not set a development/test/staging `APP_ENV`. Synthetic users need separate approved provisioning (owner, expiry, cleanup) — not migrate demo seed. Clinical 21-table CSV is not `cmd/migrate`.

Unchecked go-live boxes mean stop (runbook.md L69–L80). This document does not assert current external state.

## 3. Schema-diff inputs (repo plan vs live UNKNOWN)

Do not query live `schema_migrations` from an agent session. Operators may run **approved read-only** comparison after target confirmation. Checksums and secrets are not copied onto this sheet (STG runbook L59).

### 3.1 Inventory derivation (repository)

1. Enumerate top-level `backend/migrations/*.sql` filenames (STG runbook L56; migrations CLAUDE.md L50–L54).
2. Read `backend/migrations/seeds/002_master/manifest.json` table/file order; every listed file exists (L57).
3. After an approved migrate, require log coverage `missing=0` for current DDL + `seeds/002_master` (L12, L58). Login key `seeds/003_login` only if login seed actually applied (L12).
4. Compare live `schema_migrations` keys to that plan; then required master rows per manifest table (L59).
5. `/health` 200 is liveness only (L60).

Current top-level DDL in this checkout (`ls backend/migrations/*.sql`):

| Order | Expected `schema_migrations` key |
|---|---|
| 1 | `001_init.sql` |
| 2 | `002_medical_records_entered_by_staff_fk.sql` |
| 3 | `003_appointments_created_by_staff_fk.sql` |

`seeds/` directories and `live_insert_lab_device_clinic2.sql` are **not** top-level DDL. Count is not frozen; re-run `ls backend/migrations/*.sql` at execution time (migrations CLAUDE.md L50–L54).

`002_master` manifest (12 tables; SSOT; do not freeze checksums/row counts):

| # | table | csvFile in manifest |
|---|---|---|
| 1 | `companies` | `companies.csv` |
| 2 | `animal_species` | `animal_species.csv` |
| 3 | `lstep_auto_managed_prefixes` | `lstep_auto_managed_prefixes.csv` |
| 4 | `lstep_condition_tag_mappings` | `lstep_condition_tag_mappings.csv` |
| 5 | `lstep_send_purpose_tag_prefixes` | `lstep_send_purpose_tag_prefixes.csv` |
| 6 | `clinics` | `clinics.csv` |
| 7 | `clinic_settings` | `clinic_settings.csv` |
| 8 | `exam_types` | `exam_types.csv` |
| 9 | `reservation_types` | `reservation_types.csv` |
| 10 | `payment_methods` | `payment_methods.csv` |
| 11 | `permission_groups` | `permission_groups.csv` (file lives under `accounts/`) |
| 12 | `permission_group_rules` | `permission_group_rules.csv` (file lives under `accounts/`) |

CSV bundle order is `002_master` only in every environment. `003_demo` / `004_staging` are deleted and must not be restored (setup.md L47–L48; STG runbook L9).

Fresh expected history keys (STG runbook L12; migrations CLAUDE.md L62):

```text
001_init.sql
002_medical_records_entered_by_staff_fk.sql
003_appointments_created_by_staff_fk.sql
seeds/002_master
seeds/003_login   # only when login seed actually applied
```

Expected log shape (STG runbook L40–L48). `applied`/`skipped` may vary on already-applied DBs; **do not freeze those numbers**. Judge on current plan, `missing=0`, no checksum mismatch, and workflow exit. `extra` keys are informational (integrated/deleted history), not fail by themselves.

### 3.2 Live diff result

| Field | This session |
|---|---|
| Target env | **UNKNOWN** |
| Live `schema_migrations` keys | **UNKNOWN** (not queried) |
| Live checksums vs files | **UNKNOWN** |
| Coverage `missing` / `extra` / `expected` / `recorded` | **UNKNOWN** |
| `public` schema present | **UNKNOWN** |
| Legacy stub keys (`002_seed_master.sql`, retired demo bundles) | **UNKNOWN** |
| Whether login seed is in history | **UNKNOWN** (must not be on production) |

Do not treat the repo inventory above as “already applied on STG/PROD.”

## 4. Backup confirmation (HOLD until method recorded)

Backup existence and restore success are **external facts**. Acquisition is **HOLD** until a human-approved method is recorded. Missing any field is **No-Go**. Never invent commands, destinations, or credentials from this document (runbook.md L44–L46).

Required acquisition-record fields (runbook.md L46):

| # | Field | Live value this session |
|---|---|---|
| 1 | owner | **UNKNOWN** |
| 2 | method (managed backup **or** reviewed `pg_dump` procedure) | **UNKNOWN** |
| 3 | exact target identity | **UNKNOWN** |
| 4 | protected repo-external destination | **UNKNOWN** (none named in the runbook) |
| 5 | acquisition timestamp | **UNKNOWN** |
| 6 | size | **UNKNOWN** |
| 7 | SHA-256 checksum | **UNKNOWN** |
| 8 | retention/expiry | **UNKNOWN** |
| 9 | receipt ID | **UNKNOWN** |
| 10 | restore-rehearsal link | **UNKNOWN** |

STG pre-deploy also requires verified backup/restore and a rollback decision (STG runbook L22). Production go-live checkbox: “backup and isolated restore rehearsal evidence reviewed” (runbook.md L74). Notification/backup **schedules** are required contracts, not achieved facts (runbook.md L66).

## 5. Restore-prep (disposable target only)

Schema rollback is **not** automatic. If a `GOOD_SHA` app rollback is incompatible with applied migrations, use a forward-compatible fix or a **separately approved restore plan** (runbook.md L42). AWS/RDS is not a rollback target (runbook.md L22; STG runbook L80).

Restore only to a **disposable isolated** target. Before any destructive restore option, enforce identity assertion (runbook.md L48–L59):

```bash
: "${ISOLATED_DB_HOST:?}"
: "${ISOLATED_DB_NAME:?}"
test "$ISOLATED_DB_NAME" = "$EXPECTED_DISPOSABLE_DB_NAME" || exit 1
test "$ISOLATED_DB_HOST" = "$EXPECTED_DISPOSABLE_DB_HOST" || exit 1
# Operator also verifies provider project/id is the approved disposable target.
pg_restore --clean --if-exists -h "$ISOLATED_DB_HOST" -U "$ISOLATED_DB_USER" -d "$ISOLATED_DB_NAME" '<snapshot>'
```

This example is **not** a production restore command. Allowlisted host/database and disposable provider ID must come from the approved rehearsal record, not visual similarity. Stop on mismatch or nonzero exit.

Restore-prep confirmation (all UNKNOWN this session):

| Assertion | Status |
|---|---|
| Approved rehearsal record exists | **UNKNOWN** |
| `EXPECTED_DISPOSABLE_DB_HOST` / `EXPECTED_DISPOSABLE_DB_NAME` | **UNKNOWN** (not in repo) |
| `ISOLATED_DB_HOST` / `ISOLATED_DB_NAME` set and equal to expected | **UNKNOWN** |
| Disposable provider project/id verified | **UNKNOWN** |
| `ISOLATED_DB_USER` approved value | **UNKNOWN** (appears in example only; not in `:?` tests) |
| Snapshot path | **UNKNOWN** — `'<snapshot>'` is a placeholder |
| Post-restore: schema, representative **non-sensitive** data, app compatibility, timing, cleanup | **UNKNOWN** |
| Production restore authorized from this example | **No** (explicit prohibition) |

After restore rehearsal: do not leave the disposable target as a confused alias of STG/PROD. Cleanup is part of the rehearsal (runbook.md L59).

## 6. AUTH-V-D1 coupling (schema is an input, not a migrate apply)

AUTH-V-D1 is first-admin grant/mail on a **named environment**. It is BLOCKED until environment, operator, approval, existing staff/home clinic, impact, and rollback are fixed ([todo-verification.md](../../../todo-verification.md) L42–L44, L238). Local implementation or SQL fixtures are not target-env success.

Relevance to this sheet:

- PREFLIGHT requires **applied schema** on the chosen path, safe connection identity, maintenance window, audit, and comms-loss recovery as a gap table (L246).
- Existing system_admin → do **not** re-run bootstrap; use restore/normal add paths (L246).
- Shared `ekarte_db` / `old-db-postgres` / STG / PROD are not test DBs; **do not auto-run production grant, mail, or migration** (L242).
- APPLY after PREFLIGHT + explicit approval: one transaction, COMMIT, audit receipt, then normal login — not this unit (L248–L250).
- Communication loss or unknown COMMIT: do not re-issue; read-only receipt. Do not DELETE account/staff/audit to undo success (L250).

Do not treat “migrations look current in git” as AUTH-V-D1 schema confirmation. Target env and applied schema remain **UNKNOWN**.

## 7. Operator fill-in (I-MIGRATE gaps)

Values, hosts, credentials, PHI, and invented receipts are not recorded. Unsupplied stays **UNKNOWN**.

| ID | Required input | Why | This session |
|---|---|---|---|
| I-ENV | Named target: local disposable **or** STG **or** PROD, plus human approval | S-ENV | **UNKNOWN** |
| I-REV | Reviewed git SHA / workflow `headSha` for that env | S-REV | **UNKNOWN** (worktree HEAD `aac697645` is not live) |
| I-OWNER | Data owner + operator + window | STG L21; AUTH-V-D1 | **UNKNOWN** |
| I-BACKUP | Acquisition fields 1–10 complete | runbook.md L46 | **UNKNOWN** / HOLD |
| I-RESTORE | Disposable identity assertion vars + rehearsal link | runbook.md L48–L59 | **UNKNOWN** |
| I-DIFF | Read-only `schema_migrations` vs §3 plan | STG L59 | **UNKNOWN** |
| I-PUBLIC | `public` present on target | STG L28–L32 | **UNKNOWN** |
| I-LEGACY | Legacy key translation vs master completeness | STG L25, L69 | **UNKNOWN** |
| I-AUTH | AUTH-V-D1 env/admin/staff/home-clinic | todo-verification.md | **UNKNOWN** / BLOCKED |
| I-LINEAR | Live Linear LINMIG-209 / go-live item | setup.md L16; this campaign | **not found** / **UNKNOWN** |

## 8. What this session did not do

- Did not run `make migrate`, `cmd/migrate`, SQL, `psql`, `pscale`, `pg_dump`, `pg_restore`, Wrangler, or workflow dispatch.
- Did not apply, skip, or rebuild any environment.
- Did not mutate the campaign ledger.
- Did not post to Linear/GitHub.

When I-ENV through I-BACKUP are filled by a human, re-evaluate §2. Until then: **do not apply.**
