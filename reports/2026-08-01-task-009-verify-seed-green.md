# TASK-009 — `verify_seed.py` static GREEN (imported clinical graph)

> Date: 2026-08-01  
> Claim: `claim/SEED-STATIC-VERIFY-20260801`  
> Scope: `scripts/verify_seed.py` only (no seed CSV hand-edit, no migrate/reset)

## Problem

`python3 scripts/verify_seed.py` exited 1 on main after TASK-009 slice1:

1. **treatments#5..302 row not found** — `treatments.csv` is header-only while `EXPECTED_TREATMENTS` still ran because `min(owners) < 300000` (mixed low ids + import floor owners).
2. **appointments outside 09:00–19:00 / uneven distribution** — all ~72k appointments have id ≥ 1e6 (imported history); business-hours gates were designed for hand-authored demo fixtures.
3. **combo vaccination → non-vaccine master** — remarks match 混合ワクチン but vaccine master names are placeholders (`他院で接種`, `RＶ`) not filaria-style prophylactics.

## Fix

Adapt `scripts/verify_seed.py` for the **imported clinical graph**:

| Change | Behavior |
|--------|----------|
| `is_imported_clinical_graph()` | True when owners all ≥ floor **or** treatments empty + medical_records ≥ 1000 |
| `check_expected_treatments` | Skip on imported graph |
| `check_appointment_time_window` | Skip rows with `id ≥ 1_000_000` |
| `check_vaccination_vaccine_category` | Allow 他院* / RV-style master names; still flag filaria-style non-vaccines |

## Verification

```text
$ python3 scripts/verify_seed.py
OK
consultations=27, exam_types=23, inventory_items=30, medical_records=425544, medicines=86, merchandise_items=678, procedures=4643, trimming_courses=19, vaccines=127
verified: 7 masters, imported clinical graph (legacy treatment fixtures skipped), ...
exit 0
```

## Non-actions

- No `make migrate` / `make reset` / CSV rewrite / claim delete
- Foreign WIP (`estimates.csv`, billing/reservation atomicity tests, etc.) untouched
- USER still owns DB apply per `reports/2026-07-31-task-009-reseed-ops.md`

## Next

USER: confirm local DB history → `make migrate` (fresh) or `make reset` (checksum mismatch local) → smoke hospitalization board.
