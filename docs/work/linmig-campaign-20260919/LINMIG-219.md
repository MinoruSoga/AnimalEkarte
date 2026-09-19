# LINMIG-219 — legacy data count specification (no live values)

Campaign `linmig-ops-prep-20260919` revision 1. Unit `LINMIG-219`. Claim id `LINMIG-219`. Linear issue `LINMIG-219` was not found at sheet authoring; current Linear state is **UNKNOWN**.

This file is a **count specification only**. It names aggregation dimensions, join rules, record fields, and the PASS / FAIL / UNKNOWN result-branch procedure for P3 / #250 day-of 突合. It does **not** query STG or PROD, run SQL, export PHI, invent counts, or treat rehearsal PASS as production done.

| Field | Value at authoring |
| --- | --- |
| STG / PROD read scope | **UNKNOWN** |
| Operator / 読取担当 | **UNKNOWN** |
| Target clinic code / ordinal / `clinic_id` | **UNKNOWN** — importer accepts only hachioji=1, jouto=2, shikishima=3, hakobuneco=4 ([CLINIC_CSV_IMPORT.md](../../ops/deploy/CLINIC_CSV_IMPORT.md) L61–L62) |
| Manifest SHA-256 / run ID / revision | **UNKNOWN** |
| Formal COMPLETE bundle receipt | **UNKNOWN** (same document L189: 未記入) |
| Environment (rehearsal / STG / PROD) | **UNKNOWN** |
| Actual table counts / clinic_id counts / amount totals | **未実行** |
| Sample visual check | **未実行** |

Callers: campaign controller and operators reading `docs/work/linmig-campaign-20260919/`. No application import. Net-new file.

## Binding sources

| Source | Role for this sheet |
| --- | --- |
| [todo-operations.md](../../../todo-operations.md) P3 / PROD-DATA-MIGRATION (L171–L175) | Same manifest/revision 件数・clinic・参照・金額 突合; rehearsal PASS is not production; unknown commit or 突合 FAIL blocks P4/P8 |
| [CLINIC_CSV_IMPORT.md](../../ops/deploy/CLINIC_CSV_IMPORT.md) | 21-table mapping, `clinic_id` isolation, manifest `RowCount`, aggregate count on reports, payment graph, `CUTOVER_REF_ROW_COUNT` / `CUTOVER_REF_CLINIC_ISOLATION` |
| [GOLIVE_RUNBOOK.md](../../delivery/GOLIVE_RUNBOOK.md) day-of T+1:15 (L60) and restore rehearsal (L82) | テーブル別件数・`clinic_id` 別件数・金額合計・サンプル目視; restore rehearsal records the same three aggregates |
| [UAT-Q2-VACCINE-SPECIES.md](../todo-campaign-20260918/UAT-Q2-VACCINE-SPECIES.md) | Clinic-scoped 1-row-1-count join: `vaccinations.clinic_id = pets.clinic_id = vaccines.clinic_id`; 参照異常 separated from normal counts |
| [UAT-Q4-UNPAID-TRIAGE.md](../todo-campaign-20260918/UAT-Q4-UNPAID-TRIAGE.md) | Billing 1:1 unit; do not inflate counts/amounts with payment JOIN rows; `billings.total_amount` vs payment `total_amount` |
| `backend/internal/csvimport` (`CutoverRefRowCount`, copy/verify mismatch, payment graph) | Executable contract for row-count equality and billing/payment `total_amount` identity |

## Scope and exclusions

| In | Out |
| --- | --- |
| Count dimensions, join rules, record fields, result branch | Live STG/PROD reads, SQL execution, invented numbers |
| Same-manifest source vs target 突合 procedure | CSV apply, DB write, `make migrate`, production cutover |
| UNKNOWN / 未実行 cells for unobserved facts | PHI (names, cell values, owner/pet identifiers) |
| Fail-closed stop when operator/scope/commit outcome missing | Treating KNJO header-only payments as 正件数 PASS |

## Count dimensions

Day-of gate ([GOLIVE_RUNBOOK.md](../../delivery/GOLIVE_RUNBOOK.md) L60) and P3 artifact ([todo-operations.md](../../../todo-operations.md) L175) require four independent dimensions. Each dimension is recorded separately; a PASS on one does not fill another.

| ID | Dimension | Unit of count | Source of expected | Target of actual | Notes |
| --- | --- | --- | --- | --- | --- |
| D1 | テーブル別件数 | One row in one of the 21 formal cutover tables | Manifest `Tables[].RowCount` for that table | Target band row count after apply/verify | Mismatch → `CUTOVER_REF_ROW_COUNT` ([CLINIC_CSV_IMPORT.md](../../ops/deploy/CLINIC_CSV_IMPORT.md) L51; `cutover_import_copy.go` committed-row check) |
| D2 | `clinic_id` 別件数 | Rows whose clinic isolation matches the run's target clinic | Manifest hospital code/ordinal + expected clinic assignment | Target rows grouped by `clinic_id` (or parent-FK clinic for tables without the column) | Cross-clinic join is forbidden. Isolation failure → `CUTOVER_REF_CLINIC_ISOLATION` |
| D3 | 金額合計 | Sum of amount fields on the billing/payment graph, not JOIN-inflated row counts | Manifest / source CSV aggregates for `billings.total_amount` and payment `total_amount` | Target sums on the same keys | Payment graph requires billing/payment `total_amount` identity ([CLINIC_CSV_IMPORT.md](../../ops/deploy/CLINIC_CSV_IMPORT.md) L126, L146, L156) |
| D4 | 参照整合 | Rows that survive clinic-scoped FK checks | Source graph after parent-FK isolation | Target validated FKs | Orphans fail-closed; do not move them into D1/D3 |

D4 is the P3 「参照」 component. UAT-Q2/Q4 designs are **investigation templates** for clinic-scoped joins; they are not additional production cutover tables.

### D1 — 21-table row-count set

Fixed order from [CLINIC_CSV_IMPORT.md](../../ops/deploy/CLINIC_CSV_IMPORT.md) L20–L46 (`csvimport.CutoverTableSpecs()` is the executable list). Actual cells stay 未実行.

| # | Target table | Isolation | Manifest row count | Target row count | Result |
| ---: | --- | --- | --- | --- | --- |
| 1 | staffs | `clinic_id` column | 未実行 | 未実行 | UNKNOWN |
| 2 | procedures | `clinic_id` column | 未実行 | 未実行 | UNKNOWN |
| 3 | merchandise_items | `clinic_id` column | 未実行 | 未実行 | UNKNOWN |
| 4 | owners | `clinic_id` column | 未実行 | 未実行 | UNKNOWN |
| 5 | pets | `clinic_id` column | 未実行 | 未実行 | UNKNOWN |
| 6 | medical_records | `clinic_id` column | 未実行 | 未実行 | UNKNOWN |
| 7 | inquiries | id band + parent FK | 未実行 | 未実行 | UNKNOWN |
| 8 | clinical_plans | id band + parent FK | 未実行 | 未実行 | UNKNOWN |
| 9 | vital_records | `clinic_id` column | 未実行 | 未実行 | UNKNOWN |
| 10 | appointments | `clinic_id` column | 未実行 | 未実行 | UNKNOWN |
| 11 | appointment_trimming_details | `clinic_id` column | 未実行 | 未実行 | UNKNOWN |
| 12 | billings | `clinic_id` column | 未実行 | 未実行 | UNKNOWN |
| 13 | billing_items | id band + parent FK | 未実行 | 未実行 | UNKNOWN |
| 14 | payments | `clinic_id` column | 未実行 | 未実行 | UNKNOWN |
| 15 | payment_splits | `clinic_id` column | 未実行 | 未実行 | UNKNOWN |
| 16 | estimates | `clinic_id` column | 未実行 | 未実行 | UNKNOWN |
| 17 | estimate_items | id band + parent FK | 未実行 | 未実行 | UNKNOWN |
| 18 | exams | `clinic_id` column | 未実行 | 未実行 | UNKNOWN |
| 19 | exam_results | id band + parent FK | 未実行 | 未実行 | UNKNOWN |
| 20 | vaccines | `clinic_id` column | 未実行 | 未実行 | UNKNOWN |
| 21 | vaccinations | `clinic_id` column | 未実行 | 未実行 | UNKNOWN |
| — | aggregate count (report field) | run-level | 未実行 | 未実行 | UNKNOWN |

Empty incomplete tables must be the explicit empty list in the manifest ([CLINIC_CSV_IMPORT.md](../../ops/deploy/CLINIC_CSV_IMPORT.md) L82). Header-only `payments.csv` / `payment_splits.csv` is **not** 正件数; preflight refuses a formal manifest until both payment tables have 正件数 (L7, L190). Do not write 0 as a live observed count here.

### D2 — clinic_id 別件数

Record one row per approved clinic in the run. Do not invent clinic names or IDs.

| Clinic label (anonymous) | code / ordinal | `clinic_id` | D1 sum for this clinic | Cross-clinic remainder | Result |
| --- | --- | --- | --- | --- | --- |
| UNKNOWN until operator names the run | UNKNOWN | UNKNOWN | 未実行 | 未実行 | UNKNOWN |

Cross-clinic remainder must be 0 for the imported band. Same numeric ID in another clinic is **not** a join key (UAT-Q2 L14).

### D3 — 金額合計

Amount totals are independent of D1. Payment JOIN row counts must not be used as billing counts (UAT-Q4 L14, L32).

| Metric | Source expected | Target actual | Result |
| --- | --- | --- | --- |
| `billings.total_amount` sum (clinic-scoped, non-deleted population used by the authorized run) | 未実行 | 未実行 | UNKNOWN |
| `payments.total_amount` sum (active, clinic-scoped; 1 payment per billing on the cutover graph) | 未実行 | 未実行 | UNKNOWN |
| Billing vs payment `total_amount` identity on completed graph | 未実行 | 未実行 | UNKNOWN |
| Split arithmetic (`payment_splits` sum vs parent payment; received/change) | 未実行 | 未実行 | UNKNOWN |
| Unpaid residual (UAT-Q4 formula; **not** a day-of gate substitute) | 未実行 | 未実行 | UNKNOWN |

Cutover verify requires billing/payment `total_amount` identity and split totals ([CLINIC_CSV_IMPORT.md](../../ops/deploy/CLINIC_CSV_IMPORT.md) L156). UAT-Q4 unpaid bands (`0` / `1–9,999` / `10,000+`) are a later investigation axis; they are not required to close D3 on go-live day, and they stay 未実行.

## Join rules

Apply before any numeric comparison. If a join rule cannot be confirmed, the dimension is UNKNOWN, not PASS.

1. **Tenant first.** Restrict every table to the approved clinic. Tables with `clinic_id` filter on that column. Tables without `clinic_id` (inquiries, clinical_plans, billing_items, estimate_items, exam_results) isolate by id band **and** parent FK to a clinic-owned parent ([CLINIC_CSV_IMPORT.md](../../ops/deploy/CLINIC_CSV_IMPORT.md) L26–L46). Do not walk a broken FK into another clinic.
2. **Do not join on ID alone across clinics.** UAT-Q2: `vaccinations.clinic_id = pets.clinic_id = vaccines.clinic_id = 対象医院`. Missing, deleted, or other-clinic references leave the normal count and enter 参照異常 (UAT-Q2 L14–L16).
3. **One business row, one count.** UAT-Q4: billing 1 row = 1 count. Count and sum after uniqueness on the billing key. Duplicate JOIN rows from payments are not added to N or T (UAT-Q4 L14, L36). Cutover graph: `payments.billing_id` UNIQUE; splits are 1–2 per payment, parented by `billing_id`.
4. **Stop buckets stay out of normal totals.** Multiple active payments, missing payment parent, other-clinic payment, or orphan payment: count separately; do not compute unpaid into D3 (UAT-Q4 L14–L15, L37). Orphan payments are counted outside the billing population and are not mixed into amount totals.
5. **Deleted / out-of-window rows.** UAT-Q2 uses 未削除 vaccinations; UAT-Q4 uses 非削除 billing/payment. Day-of cutover counts the imported band, not application soft-delete UI. Date column, window, and timezone for any investigation slice are UNKNOWN until the operator fixes them (UAT-Q2 L14; UAT-Q4 L13).
6. **Source vs target.** Compare the same manifest/revision only ([todo-operations.md](../../../todo-operations.md) L175). After preflight, CSV SHA-256 must still match; a changed source is rollback, not a new expected count ([CLINIC_CSV_IMPORT.md](../../ops/deploy/CLINIC_CSV_IMPORT.md) L87).
7. **No PHI in joins or outputs.** Reports may carry status/timestamp, manifest digest, clinic/run/target metadata, ID band, aggregate count, six seed IDs, failure stage — not CSV cell values ([CLINIC_CSV_IMPORT.md](../../ops/deploy/CLINIC_CSV_IMPORT.md) L17, L88, L144).

This sheet does not include executable SQL. Authorized operators use `make csv-import-preflight` / `make csv-import-verify` (read-only) and the importer's in-transaction checks. Ad-hoc SQL against STG/PROD is out of scope.

## Record fields

Every authorized 突合 run records the following. Values in this sheet remain UNKNOWN or 未実行.

### Run identity (fill before any read)

| Field | This sheet | Required to leave UNKNOWN |
| --- | --- | --- |
| Environment (disposable rehearsal / STG / PROD) | UNKNOWN | Named environment + approval that this run may read it |
| Operator | UNKNOWN | Named 読取担当 |
| Read scope | UNKNOWN | Clinic + band + tables + whether STG or PROD is in scope |
| Manifest SHA-256 / run ID / producer revision | UNKNOWN | Same values on the run sheet and the report (not self-declared from the source directory) |
| Clinic code / ordinal / target `clinic_id` | UNKNOWN | One of the four importer codes; seed IDs confirmed |
| Window / timezone (if slicing UAT-Q2/Q4) | UNKNOWN | Operator-fixed column and bounds |
| Backup / rollback receipt (PROD) | UNKNOWN | [GOLIVE_RUNBOOK.md](../../delivery/GOLIVE_RUNBOOK.md) §3.1 |

### Per-table / per-clinic observation (fill only after authorized read)

| Field | This sheet |
| --- | --- |
| Manifest `RowCount` | 未実行 |
| Target band count | 未実行 |
| `clinic_id` histogram (or parent-FK clinic) | 未実行 |
| `CUTOVER_REF_ROW_COUNT` / `CUTOVER_REF_CLINIC_ISOLATION` | 未実行 (absent until verify runs) |
| Report aggregate count | 未実行 |
| `billings.total_amount` sum | 未実行 |
| `payments.total_amount` sum | 未実行 |
| Split / received / change check | 未実行 |
| 参照異常 count (not mixed into D1/D3) | 未実行 |
| Sample visual check (non-PHI identifiers only) | 未実行 |
| Commit outcome (`PASS` / `FAILED_*` / `COMMIT_OUTCOME_UNKNOWN`) | UNKNOWN |

### Sample visual check (D1–D3 companion, not a count)

[GOLIVE_RUNBOOK.md](../../delivery/GOLIVE_RUNBOOK.md) L60 requires サンプル目視. Procedure only: operator picks a pre-approved non-PHI seed set (the six seed IDs already allowed on reports). Confirm clinic, type, and amount fields match the source without copying cell PHI into git. Result cell: **未実行**.

## Result-branch procedure

Reconcile after the record fields for a **single** manifest/revision are complete. Do not mix rehearsal numbers into the production row.

```
inputs: run identity + D1 + D2 + D3 + D4 + commit outcome
if operator OR read scope OR environment OR manifest identity is missing
    → UNKNOWN (do not read STG/PROD to fill this sheet)
if commit outcome is COMMIT_OUTCOME_UNKNOWN or STARTED / missing report
    → UNKNOWN; stop re-apply, restore, and go-live ([CLINIC_CSV_IMPORT.md] L162–L163)
if payments/payment_splits are not 正件数 on the formal manifest
    → FAIL-closed for apply; D3 cannot PASS ([CLINIC_CSV_IMPORT.md] L7, L190)
if any D1 table: target count ≠ manifest RowCount
    → FAIL (CUTOVER_REF_ROW_COUNT)
if any imported row fails clinic isolation / parent-FK clinic
    → FAIL (CUTOVER_REF_CLINIC_ISOLATION)
if D3 billing/payment total_amount identity or split arithmetic fails
    → FAIL
if D4 orphan / cross-clinic FK remains
    → FAIL
if classified buckets do not re-sum to pre-classification N and T
    → FAIL (UAT-Q4 L36–L37; do not explain or patch from this FAIL)
if D1–D4 match AND verify report PASS AND sample visual PASS
    AND environment is the intended one
    → PASS for that environment only
rehearsal PASS
    → not production PASS ([todo-operations.md] L175)
FAIL or UNKNOWN on production day-of
    → T+3:00 No-Go; P4/P8 must not start; rollback criterion 5 ([GOLIVE_RUNBOOK.md] L60, L129)
```

| Status | Meaning on this sheet | Allowed next step |
| --- | --- | --- |
| **PASS** | Authorized verify showed D1–D4 match for the named environment and manifest | That environment only. Never copy to another env |
| **FAIL** | Count, clinic, amount, or 参照 mismatch after an authorized read | Stop. Restore from verified backup if already applied. Do not P4/P8 |
| **UNKNOWN** | Missing operator, read scope, bundle, or commit outcome; or this sheet's 未実行 cells | Do not query live DBs from this unit. Wait for USER-authorized ops |
| **未実行** | Actual-value cell not filled | Not evidence of zero rows |

Restore rehearsal ([GOLIVE_RUNBOOK.md](../../delivery/GOLIVE_RUNBOOK.md) L82) must record 件数・`clinic_id` 別件数・金額合計 after restore. Those cells are also **未実行** here.

## Stop conditions (do not upgrade this sheet)

- STG/PROD access, SQL, or PHI export would be required to fill 未実行 cells → **stop**; leave UNKNOWN.
- Formal COMPLETE bundle not received → apply remains BLOCKED; do not invent 正件数.
- `make stg-uat-handoff` skip/PASS of a previous hospital is not current integrity ([CLINIC_CSV_IMPORT.md](../../ops/deploy/CLINIC_CSV_IMPORT.md) L63–L68).
- Wrapper exit 0 is not all-hospital completion (same, L68).
- Linear `LINMIG-219` is not a live issue at authoring; do not close an external ticket from this file.

## Verification of this document (docs-only)

This unit verifies the sheet against cited docs, not against databases. Commands belong in the Completion Report, not as live counts.
