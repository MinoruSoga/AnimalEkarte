# SLACK-HAC-IMPORT — Hachioji send-report versus formal-bundle gap

Campaign `remaining-ops-20260920` revision 1. Unit `SLACK-HAC-IMPORT`. Attempt `att-slack-hac-import-20260920-001`. Claim `claim/SLACK-HAC-IMPORT`. Prompt SHA `d8f01b0ab40018702f19cad3b29c5a0079a05a05deaef7bea6b50f8bc2a494b4`. Acceptance-checklist SHA `acb2c08201c840f94fa9823557fd0724486c3d0a22a771c1fbae74298ccbaf7b`.

Worktree `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte-rem-slack-hac-import` on `feat/rem-slack-hac-import-20260920` at HEAD `873685b0bea3692c2f8100b19ded660c8357f2b0`. Sheet date: 2026-09-20. Local files only.

This unit does **not** import CSV, apply DB, run `make csv-import*`, run `make stg-uat-handoff`, inspect live `_old_db_handoff/`, or treat a Slack send report as a verified bundle. **Receipt remains UNKNOWN.**

Binding sources (read-only this session):

- [todo-issue.md](../../../todo-issue.md) **SLACK-HAC-IMPORT** (L276–280)
- [todo-operations.md](../../../todo-operations.md) **H0-2 / HAC-CSV-1** (L22, L57, L113–117) and **AE-STG-UAT-LANE3-HAC** (L129–131)
- [docs/ops/deploy/CLINIC_CSV_IMPORT.md](../../ops/deploy/CLINIC_CSV_IMPORT.md)
- [docs/ops/deploy/OLD_DB_HANDOFF_LOCAL.md](../../ops/deploy/OLD_DB_HANDOFF_LOCAL.md) (八王子 wrapper 対象外)
- [LINMIG-234.md](../linmig-campaign-20260919/LINMIG-234.md) (P3 formal-bundle contract analog; **not** an HAC receipt)

## Send report is not a formal bundle

[todo-issue.md](../../../todo-issue.md) L278 records a **9月11日 BAK→CSV 送付報告** (Slack 出典 77–100, 127–159, 295–423). The same paragraph states that **現在の完全性・受領・投入結果は UNKNOWN**.

[SLACK-INTAKE.md](../todo-campaign-20260919-ready17/SLACK-INTAKE.md) L49 classifies HAC-IMPORT as existing-ID ops connection with the operator rule **送付報告≠投入完了**.

| Artifact | What it can show | What it cannot show |
|----------|------------------|---------------------|
| Slack / ops **send report** (9/11 BAK extract) | That someone reported sending extracted CSV | Completeness, producer verification, separate-channel manifest SHA, `TRUSTED_CANDIDATE`/`PASS`, positive payment/split counts, consumer receipt, apply/verify |
| Formal **COMPLETE producer bundle** | Hospital/run-fixed `manifest.json` + 21 AnimalEkarte-shaped CSVs meeting [CLINIC_CSV_IMPORT.md](../../ops/deploy/CLINIC_CSV_IMPORT.md) L81–L88 | Nothing in this session: no current bundle was received or inspected |
| Consumer **receipt** | Manifest SHA from `clinic-migration-run-report.json` on a **separate channel** (CLINIC_CSV_IMPORT.md L81), producer revision/run, eligibility, verify aggregates | Not substitutable by chat text, Linear history, or a send report |

Do **not** treat the send report as a verified bundle. Do **not** treat older “完全 KNJO 未受領” wording as a current fact ([todo-operations.md](../../../todo-operations.md) L5, L115). Current receipt is still **UNKNOWN** because this session did not collate a live producer result.

## Current HAC receipt (this session)

**UNKNOWN.** No current formal COMPLETE producer bundle, `clinic-migration-run-report.json`, live `manifest.json`, or `_old_db_handoff/hachioji/` tree was received or inspected.

| Item | This session |
|------|----------------|
| Slack 9/11 BAK→CSV send report | Cited in todo-issue.md L278 as a **report**, not as receipt |
| Formal COMPLETE producer bundle receipt | **UNKNOWN** / CLINIC_CSV_IMPORT.md L189 **未記入**（受領記録なし） |
| Manifest SHA-256 (separate-channel) | **UNKNOWN** — do not invent; do not take a self-declared value from a source directory (CLINIC_CSV_IMPORT.md L81) |
| Producer revision / run ID / `handoffEligibility` | **UNKNOWN** (H0-2 L57: producer 追加状況未照合) |
| 21-table row counts / per-CSV SHA-256 / 欠損表 | **UNKNOWN** — not inspected |
| `payments.csv` / `payment_splits.csv` positive counts | Documented KNJO is header-only (CLINIC_CSV_IMPORT.md L7, L190). Live current files not inspected → **UNKNOWN**; formal apply remains **BLOCKED** until a verified KNJO formal bundle has positive counts |
| Formal vs rehearsal purpose | **UNKNOWN** — send report does not pin eligibility |
| Import / STG load / 投入結果 | **UNKNOWN**. Not run in this unit |
| Linear BRT-42 live body | Not re-read. 9月15日読取は Needs Human; 入力そのものは未確認 (H0-2 L115). Unconnected Linear text is **not** receipt |

## H0-2 / HAC-CSV-1 gap (prep; no import)

[todo-operations.md](../../../todo-operations.md) L22: local deliverable is 完全KNJO/clean BAK の受領参照・provenance・producer結果の**不足票**. External start needs **現在の正式入力受領と検証**. 未接続 Linear の昔の本文で受領済みにしない.

H0-2 current row (L57): **UNKNOWN（producer の追加状況未照合）**. Next step is confirm current receipt/producer result; if still unresolved, wait for complete input. Previous HAC-INPUT-2 完全 KNJO 未受領 is history, not this session's receipt.

H0-2 procedure (L113–117), recorded as gaps — **not executed**:

1. Collate BRT-42 current body/receipt with the producer's **current** receive record. Do not decide “still unreceived” from the old phrase alone.
2. If unreceived: prepare a shortage list for the data owner — CHECKDB-clean new BAK **or** full 32-column KNJO, acquisition datetime, source provenance. Send/extract/restore stays in the owner's approval scope.
3. After input exists: producer runs the canonical pipeline and hands over a bundle with completeness, 21 tables, and eligibility. Deliverable: 受領参照, producer revision/run, manifest digest, 検証結果. **既知破損 BAK の再復元や PARTIAL の昇格で通さない.**

This sheet is that shortage/gap record for the send-report vs formal-bundle distinction. It is **not** a filled receipt.

## Formal-bundle necessities (Hachioji; no import)

Sources: CLINIC_CSV_IMPORT.md L5–L7, L20–L46, L53–L62, L81–L88, L183–L193; OLD_DB_HANDOFF_LOCAL.md L36–L40, L110; todo-operations.md L129–131.

A formal F6 cutover candidate is hospital/run-fixed `manifest.json` + 21 AnimalEkarte-shaped CSVs. The consumer does not connect to `old_db` and does not write `backend/migrations/seeds/`.

Accept for **formal** `make csv-import` preflight only when **all** of the following hold (unknown fields rejected). This sheet records the contract; it does not assert a live match.

1. Manifest SHA-256 from `clinic-migration-run-report.json` via a **separate channel** (CLINIC_CSV_IMPORT.md L81).
2. Manifest `PASS`, stage `animalekarte_stage`, clinic code/ordinal, run ID, 10M ID band, **21-table fixed order**, per-CSV SHA-256 (L82).
3. Manifest schema version; approved stage mapping bundle (`020_canonical.sql` + `030_stage.sql`) / CSV contract SHA-256 (L82).
4. `handoffEligibility=TRUSTED_CANDIDATE` and completeness `PASS` (L82).
5. Verified source identity; explicit empty incomplete-table list; stage build UUID; DB1/DB2/DB3 summary digests + generation order; base-load/KNJO evidence digests (L82).
6. **Positive row counts** for `payments` and `payment_splits` (L7, L82, L190). Header-only KNJO fails preflight.
7. Source directory `0700`; manifest/CSV `0600`; **no symlinks** (L83).
8. Six target seed IDs explicit (active clinic, species fallback, 検査, trimming, cash, credit_card) — no implicit name/first-row resolution (L85).
9. Target ID band empty on all 21 tables before apply; no delete/replace of existing rows (L86).
10. PHI may be in CSV/manifest: no cell values in git, logs, reports, Issues, or this sheet (L88; todo-operations.md L13).

`REHEARSAL_ONLY` / `PARTIAL` may sit under `_old_db_handoff/<clinic>/` as storage only; formal `make csv-import` preflight **rejects** them (CLINIC_CSV_IMPORT.md L53–L56). Presence of files is **not** formal eligibility (OLD_DB_HANDOFF_LOCAL.md L110).

Importer clinic codes: hachioji=1, jouto=2, shikishima=3, hakobuneco=4 (CLINIC_CSV_IMPORT.md L61–L62).

### 八王子 lane (do not use the 3-clinic wrapper)

`make stg-uat-handoff` is 城東・敷島・箱 only. **八王子は対象外** (CLINIC_CSV_IMPORT.md L58–L59; OLD_DB_HANDOFF_LOCAL.md L38; todo-operations.md L130). Do not reuse that wrapper for HAC.

AE-STG-UAT-LANE3-HAC (todo-operations.md L129–131): after H0-2 / H0-3b, collate current 八王子 data, clinic, band, and seed binding, then pick a **formal or STG UAT individual** path from CLINIC_CSV_IMPORT.md / Makefile. This unit does **not** choose or run that path. Stop if STG load would be required.

## What LINMIG-234 does not prove

[LINMIG-234.md](../linmig-campaign-20260919/LINMIG-234.md) is the P3 / #250 formal-bundle, 突合, rehearsal, restore-prep sheet. Its formal-receipt row is also **UNKNOWN**. It does **not** fill H0-2, does **not** verify the 9/11 send report, and does **not** authorize HAC import.

## Execution ticket (local prep fields; values UNKNOWN)

Shared ops fields (todo-operations.md L13). Filled only with non-secret placeholders this session:

| Field | This session |
|-------|----------------|
| ID | SLACK-HAC-IMPORT / H0-2 / HAC-CSV-1 / BRT-42 |
| 対象環境・医院 | 八王子 (`hachioji` / ordinal 1 per importer contract). Live occupancy **UNKNOWN** |
| revision・manifest 識別 | **UNKNOWN** (send report is not a manifest id) |
| 読取・変更する範囲 | This unit: this markdown only |
| 前段証拠 | Send report cited; formal bundle **UNKNOWN**; producer collation **UNKNOWN** |
| 操作者・承認者 | **UNKNOWN** (業務上の個人責任者名は USER ops; CLINIC_CSV_IMPORT.md L13) |
| 実行枠 | **UNKNOWN**. This unit does not start import |
| 中止条件 | Missing TRUSTED_CANDIDATE/PASS; header-only payments; treating send report as receipt; PARTIAL promotion; known-corrupt BAK restore; 八王子 via `make stg-uat-handoff`; STG load without eligibility |
| 復旧 | Not applicable to this docs sheet. Later import: pre-validated full backup; no importer delete |
| 証拠保存先 | Non-secret refs only; PHI/roster/secrets stay repo-external |

## What this sheet is not

| Candidate | Why it is not done |
|-----------|--------------------|
| CSV import / `make csv-import*` / `make stg-uat-handoff` | Out of scope. Local files only. 八王子 is wrapper-out. |
| Treating the 9/11 send report as a verified bundle | todo-issue.md L278–279; SLACK-INTAKE L49 送付報告≠投入完了 |
| Inventing a current bundle SHA / run ID / eligibility / counts | CLINIC_CSV_IMPORT.md L81, L189 未記入 |
| Fabricating a COMPLETE bundle to unblock apply | CLINIC_CSV_IMPORT.md L185 |
| Declaring current unreceived solely from old HAC-INPUT-2 text | todo-operations.md L5, L115 |
| Campaign ledger mutation | Controller-owned |

## Out of scope (this unit)

- CSV import, DB apply, `make migrate`, Docker app tests, STG/PROD load
- Filling live receipt fields, inspecting PHI CSV cells, Slack Vault paste
- Campaign ledger edits; Linear/GitHub live mutation
- Inventing named owners, clinic actuals, or bundle digests

Follow-up (not this unit): USER/producer collates **current** formal input (CHECKDB-clean BAK or full 32-column KNJO) against CLINIC_CSV_IMPORT.md, produces a verified 21-table bundle with separate-channel SHA, `TRUSTED_CANDIDATE`/`PASS`, and positive payment/split counts, then a **separate** approved HAC path (not the 3-clinic wrapper). Until then H0-2 stays UNKNOWN and this sheet's receipt stays UNKNOWN.
