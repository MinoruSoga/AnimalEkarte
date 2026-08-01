# TASK-028: Closing settings standard PATCH integrity

Date: 2026-08-01  
Issue: #252 technical gap (value apply remains USER gate)  
Claim: `claim/TASK-028` (agent must not delete)

## Problem (before)

`UpdateStandard` performed unlocked read-modify-save via full-column UPSERT:

- no HH:MM / order / closed_weekdays validation
- no transaction or row lock → concurrent partial PATCH lost updates
- no authenticated actor propagation
- no application audit

## Design decisions

### Serialization: parent clinic row lock (`FOR UPDATE`)

Chosen over:

| Option | Rejected because |
|--------|------------------|
| CAS / version column | Requires migration (out of scope) |
| `clinic_settings` FOR UPDATE only | Row may be absent on first upsert; lock would not serialize |
| Advisory lock only | Works, but package already has `ClinicRepository.LockByIDForUpdate` fail-closed pattern |

**Selected:** `LockByIDForUpdate` on `clinics` **before** settings read, inside `Transactor.WithTx`.

Evidence: `closing_settings_service.go:209-215` (lock) then `:215` (read).

### Validation

`validateStandardClosingTimes` reuses `sharedkernel.ParseHHMM` (same primitive as `validateSpecialPeriodTimes`). Final merged state must have boundary strictly before weekday end and sunday end. `closed_weekdays` ∈ [0,6], unique.

### Audit

Fail-closed ambient-tx `LogEntryTx` with non-secret metadata only:

- `fields.*.present` markers + `closed_weekdays_count`
- `changed_fields` list
- **no raw clock strings**

## Files changed

| Path | Role |
|------|------|
| `backend/internal/clinic/closing_settings_service.go` | validation, tx, lock, audit, actor arg |
| `backend/internal/clinic/closing_settings_handler.go` | `ExtractStaffID` → service |
| `backend/internal/clinic/closing_settings_update_standard_integrity_test.go` | 4-family integrity tests (new) |
| `backend/internal/clinic/closing_settings_service_test.go` | constructor / UpdateStandard actor+deps |
| `backend/internal/clinic/closing_settings_handler_test.go` | actor required |
| `backend/cmd/api/composition_clinic.go` | wire deps + audit bridge |
| `backend/cmd/api/composition_runtime.go` | pass `auditKernel` into clinic composition |
| `backend/cmd/api/composition_clinic_test.go` | arity update |
| `reports/2026-08-01-task-028-closing-settings-integrity.md` | this report |

Allowlist expansion (compile/wire necessity): `composition_runtime.go`, `*_handler_test.go`, `composition_clinic_test.go`.

## Added tests (4 families)

1. `TestClosingSettingsService_UpdateStandard_Validation`
2. `TestClosingSettingsService_UpdateStandard_ConcurrentSameClinic_NoLostUpdate` (+ other-clinic)
3. `TestClosingSettingsService_UpdateStandard_Audit`
4. `TestClosingSettingsService_UpdateStandard_Rollback`

Concurrency / rollback tests use **real Postgres** via `setupClinicSettingsTestDB` + `persistence.NewTransactor` (not mock locks).

## RED evidence (naive UpdateStandard body restored temporarily)

Command:

```bash
docker compose exec -T backend go test -p 1 ./internal/clinic \
  -run 'TestClosingSettingsService_UpdateStandard.*(Validation|Concurrent|Audit|Rollback)' -count=1 -v
```

Verbatim summary from `/tmp/task028-red-targeted.txt`:

```
--- FAIL: TestClosingSettingsService_UpdateStandard_Validation (0.00s)
    --- FAIL: .../invalid_time_format_on_boundary
    --- FAIL: .../time_order_reversed:_boundary_after_weekday_end_(partial_final_state)
    --- FAIL: .../partial_update_leaves_sunday_end_before_new_boundary
    --- FAIL: .../closed_weekdays_out_of_range
    --- FAIL: .../closed_weekdays_duplicate
--- PASS: TestClosingSettingsService_UpdateStandard_ConcurrentSameClinic_NoLostUpdate (0.24s)
--- PASS: TestClosingSettingsService_UpdateStandard_ConcurrentOtherClinic_NoContention (0.06s)
--- FAIL: TestClosingSettingsService_UpdateStandard_Audit (0.00s)
--- FAIL: TestClosingSettingsService_UpdateStandard_Rollback (0.05s)
FAIL
```

Note: single-shot concurrent same-clinic race did not fail under RED (timing-dependent). Fixed implementation uses 15 contention rounds + parent-row `FOR UPDATE`; GREEN multi-round passes. Validation / audit / rollback RED are deterministic.

## GREEN evidence

Same command after restore — from `/tmp/task028-green-v.txt`:

```
--- PASS: TestUpdateClosingSettings (0.00s)
--- PASS: TestClosingSettingsService_UpdateStandard_Validation (0.00s)
--- PASS: TestClosingSettingsService_UpdateStandard_ConcurrentSameClinic_NoLostUpdate (0.25s)
--- PASS: TestClosingSettingsService_UpdateStandard_ConcurrentOtherClinic_NoContention (0.07s)
--- PASS: TestClosingSettingsService_UpdateStandard_Audit (0.00s)
--- PASS: TestClosingSettingsService_UpdateStandard_Rollback (0.06s)
ok  	github.com/animal-ekarte/backend/internal/clinic	0.387s
```

Regression:

```bash
docker compose exec -T backend go test -p 1 ./internal/clinic ./cmd/api -count=1
```

| | FAIL set |
|--|----------|
| baseline | `TestClinicHolidayRepository_FindAllByYearMonth`, `TestClinicHolidayRepository_FindByDate` |
| after | _(none)_ |

New failures vs baseline: **0**.

## Independent review

go-reviewer probe on the 4 integrity focus points: **PASS** (no CRITICAL/HIGH/MEDIUM).

## Out of scope / unexecuted

- No production closing-time value apply
- No migration / seed / runtime env apply
- No OpenAPI edit (`backend/docs/api.yaml` foreign WIP)
- No GitHub / push / PR
- No `todo.md` edit
