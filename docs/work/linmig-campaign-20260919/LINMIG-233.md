# LINMIG-233 — Clinic closing-time input plan + S09 boundary-test prep (docs-only)

> This local sheet is supporting execution/evidence material, not an independent task ledger. Current clinic closing-time task: Plane `EMR-45` (BRT-43 / #252). S09 fixture verification: Plane `EMR-126` (`TODO-V-S09 / QA-UAT-S09-FIXTURE`). Use those Plane items for current status; this document preserves the cited plan and safety boundaries.

Campaign `linmig-ops-prep-20260919` revision 1. Unit `LINMIG-233`. Attempt `att-linmig-233-20260919-001`. Claim `claim/LINMIG-233`. The 2026-09-19 local inventory did not find Linear issue `LINMIG-233`; later reconciliation matched the clinic-input work to Plane `EMR-45` and the S09 fixture work to `EMR-126`. Prompt SHA `a91f9328eadd8b21ed3dc44e34a8e68b20c5b1c3f9626f6165a3eb4b08ac8e13`.

Maps to GitHub [#252](https://github.com/MinoruSoga/AnimalEkarte/issues/252) and [GOLIVE_RUNBOOK.md](../../delivery/GOLIVE_RUNBOOK.md) **§1 item 10** (全院の締め時間設定). Binding sources for this unit:

- [docs/delivery/GOLIVE_RUNBOOK.md](../../delivery/GOLIVE_RUNBOOK.md) item 10 (L43)
- [docs/spec/screens/settings/closing-time-settings.md](../../spec/screens/settings/closing-time-settings.md) 全院投入手順 #252 / BRT-43 (L60–L74)
- [docs/ops/testing/scenarios/S09-closing-time-boundaries.md](../../ops/testing/scenarios/S09-closing-time-boundaries.md)
- [docs/ops/testing/S09-FIXTURE-DESIGN.md](../../ops/testing/S09-FIXTURE-DESIGN.md)
- [todo-verification.md](../../../todo-verification.md) **TODO-V-S09 / QA-UAT-S09-FIXTURE** (L109–L125)

Worktree `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte-linmig-233` on `feat/linmig-233-ops-prep` at HEAD `aac697645df92fd24611c7c13bf0f7dda12a6e08`. Sheet date: 2026-09-20. Local files only.

This unit does **not** apply clinic settings, PATCH `/api/v1/closing-settings`, invent per-clinic actuals, run S09 against real clinics, start Docker, or edit the campaign ledger.

## Binding rules (cited)

| Rule | Source | This unit |
|------|--------|-----------|
| Item 10 complete condition: confirm AM start **09:00**; input AM/PM boundary **12:00** and weekday/Sunday end **18:30** clinic-wide; preview / EMG 日跨ぎ / 過去締め非再計算 receipt | GOLIVE_RUNBOOK L43 | Plan those ruled values. Do not apply. Receipt remains 確定待ち. |
| PO 裁定 (2026-07-15, #252): 全院を城東同値. **本セッションでは本番投入しない。** Values are Issue 本文の転記, not invented | closing-time-settings.md L62–L70 | Ruled table copied. No production apply. |
| `closing_am_start` 09:00 is **編集不可** (no PATCH field; DB default only) | closing-time-settings.md L14, L66 | Confirm only. |
| `closed_weekdays` は変更しない（#252 対象外） | closing-time-settings.md L70 | Out of this input plan. |
| #252 裁定値は historical decision input. 固定 clinic/seed の現在値を前提にしない | closing-time-settings.md L72 | Current clinic settings = **UNKNOWN**. |
| 共有 STG/本番への PATCH は named target・USER 承認・dated receipt がある場合だけ | closing-time-settings.md L72 | This session has none. No PATCH. |
| S09 は承認済み disposable clinic へ **合成設定** を投入して検証する | closing-time-settings.md L72; S09 L9–L11 | S09 uses fixture times, not clinic go-live values. |
| S09 premises: AM start 09:00, AM/PM boundary **13:30**, weekday end **19:00** | S09 L10; S09-FIXTURE-DESIGN L30; todo-verification L115 | **Fixture, not clinic values.** Do not input 13:30 / 19:00 as #252 clinic settings. |
| 実医院の締め設定には S09 の5時刻を適用しない | todo-verification L115 | Explicit. |
| Item 10 runbook 状態 is 確定待ち: 投入完了証跡なし | GOLIVE_RUNBOOK L43 | Gate remains **HOLD** until USER apply + receipt. |

## Ruled clinic input values (plan only)

Source: closing-time-settings.md L64–L70 and GOLIVE_RUNBOOK L43. These are the only clinic-input times this sheet plans.

| Field | Ruled value (#252 / item 10) | Screen / API | This session |
|-------|------------------------------|--------------|--------------|
| `closing_am_start` | **09:00** | `/settings/closing-time` — 編集不可. Confirm preview only | Not applied. Current clinic value **UNKNOWN** |
| `closing_am_pm_boundary` | **12:00** | PATCH `/api/v1/closing-settings` | Planned. Not applied |
| `closing_weekday_end` | **18:30** | same | Planned. Not applied |
| `closing_sunday_end` | **18:30** (資料 2 に曜日区別なし) | same | Planned. Not applied |
| `closed_weekdays` | 変更しない | #252 対象外 | Do not change |

Do **not** substitute S09 synthetic **13:30** (boundary) or **19:00** (weekday end) into this table.

### Derived range preview if USER later inputs the ruled values

Half-open intervals from S09 L32 (`AM=[am_start, boundary)`, `PM=[boundary, pmEnd)`, `EMG=[pmEnd, next am_start)`). Using ruled 09:00 / 12:00 / 18:30:

| Period | Ruled range (JST) |
|--------|-------------------|
| AM | `[09:00, 12:00)` |
| PM (weekday and Sunday; sunday_end = 18:30) | `[12:00, 18:30)` |
| EMG | `[18:30, next-day 09:00)` — 越日. 0:00〜09:00 is previous-day EMG (#215) |

This preview is a **plan of expected UI after USER apply**. It is not a live clinic observation.

## Current clinic settings

**UNKNOWN.** This session did not GET `/api/v1/closing-settings` for any clinic, did not read seed/runtime DB, and did not invent 八王子 / 城東 / other per-clinic actuals.

closing-time-settings.md L16 documents product **defaults** (boundary 14:00, weekday 18:30, Sunday 17:30). Defaults are not current clinic settings. Seed/default 14:00 / 17:30 must not be written as live values.

GOLIVE item 10 状態 remains `（確定待ち: 投入完了証跡なし）`. Named apply owner **UNKNOWN**.

## Input plan (USER; not this session)

Source order: closing-time-settings.md L74. Stop if any step would write shared STG/PROD without named target + USER approval + dated receipt.

1. Confirm target env (**UNKNOWN** this session). Local files only here.
2. Confirm `closing_am_start` preview shows **09:00** and the field stays 編集不可.
3. On each target clinic, input `closing_am_pm_boundary=12:00`, `closing_weekday_end=18:30`, `closing_sunday_end=18:30`. Leave `closed_weekdays` unchanged.
4. Preview AM/PM/EMG matches the ruled range table above.
5. If `closing_am_start` ≠ 09:00, stop for prior approval + DB path (L74). Do not invent a DB write here.
6. Record that changing settings does **not** recompute past closes (#215 / 仕様正本 29 §1).
7. Store a non-secret receipt (clinic ids without PII, before/after times, preview screenshot or equivalent). Until that exists, item 10 stays HOLD.

This session executes **none** of steps 2–7 against a clinic.

## Two clocks (do not mix)

| Clock | AM start | AM/PM boundary | Weekday end | Where it is used |
|-------|----------|----------------|-------------|------------------|
| **A — clinic go-live (#252 / item 10)** | 09:00 | **12:00** | **18:30** (Sunday also 18:30) | Planned clinic input. Not applied here. |
| **B — S09 synthetic fixture** | 09:00 | **13:30** | **19:00** | Disposable clinic / helper / S09 #2–#6 only. |

Clock B is **fixture, not clinic values**. Clock A is **not** the S09 scenario premise. Running S09 expected buckets (13:30 on-boundary = PM) against Clock A would be a mixed-clock error: under Clock A, 13:30 is interior PM (after 12:00), not the half-open boundary sample.

## S09-style boundary-test prep (not executed)

Sources: S09 L7–L32; S09-FIXTURE-DESIGN L1–L30, L49–L53; todo-verification L109–L125.

Environment contract (design, not a run): local disposable clinic only; STG/PROD forbidden; clinic 1/2 excluded; cash-register-close attached account; no direct DB/`completed_at` UPDATE; no system-clock change. Helper HTTP/CLI may exist; browser #2–#6 remains 未 in S09-FIXTURE-DESIGN L3/L13/L51.

### Fixture settings to use when S09 later runs

Use Clock B on the disposable clinic: AM start 09:00, boundary **13:30**, weekday end **19:00**. Do not PATCH Clock B onto a real hospital.

### Attribution expectation (TODO-V-S09 L117–L125)

Target day D, JST. Counts: AM 1 / PM 2 / EMG 2. Amounts from fixture amounts (not invented here). Actual buckets this unit = **未実行**.

| Synthetic `completed_at` | Expected bucket (Clock B) | Actual this unit |
|--------------------------|---------------------------|------------------|
| D 10:00 | D 午前 only | **未実行** |
| D 13:30:00 | D 午後; not also 午前 (half-open; boundary is PM) | **未実行** |
| D 14:00 | D 午後 | **未実行** |
| D 20:00 | D 緊急 | **未実行** |
| D+1 02:00 | D 緊急; not D+1 緊急 | **未実行** |

S09 #1 range preview must match Clock B, not Clock A. S09 #7–#10 (confirm close, post-close warning, history) stay out of this session. Helper GREEN without browser #2–#6 is not S09 PASS.

### Prep checklist (this unit documents; does not run)

- [ ] Dedicated local Docker + allowed `APP_ENV` (S09-FIXTURE-DESIGN L24–L27)
- [ ] Synthetic password supplied out of repo; not written here
- [ ] New disposable clinic (not 1/2)
- [ ] Clock B settings on that clinic only
- [ ] Five synthetic completions at the table times
- [ ] Browser #2–#6 vs expected buckets
- [ ] Cleanup token teardown; incomplete cleanup recorded
- [ ] No shared STG/PROD; no apply of Clock B to real clinics

All boxes remain unchecked here.

## What this sheet is not

| Candidate | Why it is not done |
|-----------|--------------------|
| Applying 09:00 / 12:00 / 18:30 to any clinic | Out of scope. Local files only. |
| Treating 13:30 / 19:00 as clinic go-live values | S09 fixture only (S09 L10; todo L115). |
| Current 八王子 / 城東 / seed live times | **UNKNOWN**. Do not invent. |
| Item 10 receipt / GOLIVE green | Runbook 確定待ち. |
| S09 PASS | Browser #2–#6 未. This is prep. |
| V04 §6 form persist | Form persist is not attribution (V04 L138). |
| Linear `LINMIG-233` Done | Issue not found; claim only. |
| Campaign ledger mutation | Controller-owned. |

## Out of scope (this unit)

- Clinic settings apply / PATCH / seed writes / `make migrate` / Docker app tests
- Running S09, Playwright, synthetic-closing helper against a live env
- Filling GOLIVE item 10 receipt
- Campaign ledger edits
- Linear/GitHub live mutation
- Inventing named apply owners or per-clinic actuals

Follow-up (not this unit): USER confirms env + current settings (today UNKNOWN), inputs Clock A on named clinics with approval, stores non-secret receipt; separately, run S09 on a disposable clinic with Clock B and cleanup. Until then item 10 stays HOLD and S09 stays unexecuted.
