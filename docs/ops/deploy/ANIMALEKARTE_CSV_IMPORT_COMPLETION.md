# AnimalEkarte CSV seed overlay completion report（旧7表経路）

> これは2026-07-15の disposable seed-export 用7表adapterの履歴記録であり、医院カットオーバーF6の実装・運用証跡ではない。正式な21表 consumer は [CLINIC_CSV_IMPORT.md](./CLINIC_CSV_IMPORT.md)、現在のseed生成境界は [SEED_MIGRATION_OPERATIONS.md](./SEED_MIGRATION_OPERATIONS.md) を参照する。

## Completion Report

- Run status: INCOMPLETE (local generation and verification complete; commit/reset/STG intentionally deferred)

### Checklist Results

| Checklist item | Expected behavior | Actual behavior | Status | Verification method | Evidence |
|---|---|---|---|---|---|
| Source contract and schema diff | Seven source schemas, row counts, and target differences documented | Documented below; source PASS manifest read | PASS | Source header/count probe; `001_init.sql` | 7-table diff below |
| Placeholder resolution | No unresolved target placeholders | `{{CLINIC_ID}}`, `{{FALLBACK_*}}` absent from generated seeds | PASS | `rg -n '\{\{CLINIC_ID\}\}|\{\{FALLBACK_' backend/migrations/seeds` | 0 matches |
| Seven-table replacement | Source graph replaces old 003 demo owner/clinical/accounting rows | Generated 003_demo contains source counts; dependent old demo rows are empty where no source equivalent exists | PASS | Generated CSV row counts and source manifest comparison | Counts below |
| DB-to-seed generation | No CSV hand merge/copy | Read-only source mount → disposable DB → `seed-export` COPY dump | PASS | Docker run log | `✓ seed-export completed` |
| Seed verification | `verify_seed.py` green | Green | PASS | `python3 scripts/verify_seed.py` | `OK` |
| FK integrity | No required orphan references | All requested checks returned 0 | PASS | Deterministic CSV FK check | 0 orphans |
| PHI/size safety | No silent public PHI commit | 529MB local generated data remains uncommitted; USER decision required | BLOCKED | `git status`, `du -sh` | Commit deferred |
| Reset boundary | Agent must not run reset | No `make reset`, `db_reset`, volume deletion, push, or STG apply | PASS | Session command audit | None |
| Workflow orchestration | Fan-out/workflow mode | No Task/Agent/Workflow primitive was available; parallel read-only probes and phased plan used as capability fallback | BLOCKED | Session capability record | Missing native subagent primitive |

### Schema diff and resolution

| Table | Source columns/rows | Target additions or differences | Resolution |
|---|---:|---|---|
| owners | 19 / 10,370 | `name_kana`, LINE/L-step, transfer, timestamps, soft-delete fields | Source overlap copied; target defaults/NULLs retained; clinic resolved to target clinic |
| pets | 16 / 15,654 | `name_kana`, acquisition/danger/environment/insurance, timestamps, soft-delete fields | Source overlap copied; target defaults/NULLs retained; species placeholder/fallback resolved |
| medical_records | 8 / 425,544 | doctor/appointment/version/entered_by/recommendation/timestamps/soft-delete | Source clinical fields copied; omitted nullable/default columns retained |
| exams | 7 / 14,533 | doctor/status/machine/timestamps/soft-delete | Target exam type resolved to target `検査` fallback; omitted fields use defaults |
| exam_results | 6 / 1,322,503 | exam type field/result/unit/reference/abnormal/status/timestamps | Source six fields copied; target defaults/NULLs retained |
| billings | 8 / 392,105 | hospitalization/subtotal/tax/insurance/memo/timestamps/soft-delete | Source total/status/date/links copied; amount/default fields retained |
| billing_items | 9 / 1,542,422 | tax_rate/source/merchandise/treatment/appointment/trimming/discount/timestamps | Source fields copied; target nullable/default fields retained |

Target references resolved from the disposable target DB: clinic `八王子病院`; first target animal species; clinic-local exam type `検査`. No target IDs were guessed from old_db.

### Run Summary

- Changed files: `backend/internal/csvimport/import.go`, `backend/internal/csvimport/import_test.go`, `backend/cmd/seed-export/main.go`, `scripts/verify_seed.py`, `backend/migrations/seeds/003_demo/` generated CSVs, `backend/migrations/seeds/004_staging/appointment_trimming_details.csv`, `todo.md`, this report.
- Failure Signature log: delete order FK failures (estimates, medical_records, lstep owner references) and source empty/placeholder type conversion; each fixed with a distinct root-cause patch and rerun. Final export succeeded.
- Risk Tier: Local write | Safety boundary events: stopped before commit, reset, STG, push.
- Orchestration mode: phased plan + parallel read-only probes; native subagent/Workflow capability was unavailable in this session.
- Import counts (in → out): owners 10,370 → 10,370; pets 15,654 → 15,654; medical_records 425,544 → 425,544; exams 14,533 → 14,533; exam_results 1,322,503 → 1,322,503; billings 392,105 → 392,105; billing_items 1,542,422 → 1,542,422.
- Bundle targets: 002_master unchanged; 003_demo owns all seven imported tables and emptied incompatible old demo dependents; 004_staging dependent appointment trimming rows empty.
- PHI/size decision (2026-07-15): full dump stays **local-only** at `old_db/sensitive-local/animalekarte-003-demo-full/` (~529MB, may contain real PHI). Git keeps the small `003_demo` demo; adapter/tooling may be committed. Do not push files over GitHub's 100MB limit (`billing_items.csv`, `exam_results.csv`).

### Harness and execution selections

- Harness: TDD selected. `migration-seed-safety` and `tdd-workflow` were read; adapter boundary tests pass in Docker.
- Loop: sequential overlay with verification gate; repair rounds were bounded and used new hypotheses.
- De-sloppify: no source CSV copied into the repository; source was mounted read-only; no raw PHI was written to logs/reports.
- Saved Prompt Validation Gate: `prompt validation not run: chat-only output`.

### Verification outputs

`python3 scripts/verify_seed.py`:

```text
OK
verified: 7 masters, imported clinical graph (legacy treatment fixtures skipped), CHECK equivalent, procedure presence, FK, cross-tenant, appointment time window, daily distribution, audit log actor tenant, vaccination vaccine category, exam_type_field category
```

Requested orphan checks: `pets.owner_id`, `medical_records.owner_id/pet_id`, `exams.medical_record_id/pet_id`, `exam_results.exam_id`, `billings.owner_id/pet_id/medical_record_id`, `billing_items.billing_id` — all `0 orphans`.

### Remaining risks

- **RETIRED**: sensitive-localのfull dumpを `backend/migrations/seeds/003_demo/` へ直接復元する旧手順は使用しない。21表handoffから実行可能seedへ変換する専用経路は現行コードに未実装であり、実装・検証までは隔離保管に留める。
- STG/PROD application, push of PHI-scale seeds, and external issue operations remain out of scope.
- After local overlay, use `git update-index --skip-worktree` on overlaid CSVs to avoid accidental commits.
