# TASK-030: trimming 死亡ペット guard（finalPetID 常時検証）

- Date: 2026-08-01
- Claim: `claim/TASK-030`（release は USER 専権）
- HEAD at start: `ba920ae07`
- Risk tier: Local write
- Runtime / LINE / LIFF / migration: **未実行**（unit + source 実測のみ）

## Root cause

`createDetailForExistingAppointment` と `Update` で owner/pet link は `finalPetID` を検証していた一方、死亡検証だけが `if input.PetID != nil` に囲まれ `input.PetID` を渡していた。

`input.PetID` を省略した通常経路では予約上の `locked.PetID`（= `finalPetID`）が死亡していても business write と audit が実行された。

`ValidateReservationPetNotDeceased` は `petID == nil` なら no-op、非 nil なら `DeceasedAt != nil` で InvalidInput（`backend/internal/reservation/reservation_service.go:186-191` / `sharedkernel/pet_not_deceased.go`）。

## Minimal patch

両関数で死亡検証の `if input.PetID != nil` を外し、引数を `finalPetID` に変更。

### Before（欠陥）

`createDetailForExistingAppointment`（旧 ~497-501）:

```go
if input.PetID != nil {
    if err := reservation.ValidateReservationPetNotDeceased(txCtx, s.reservation, clinicID, input.PetID); err != nil {
        return err
    }
}
```

`Update`（旧 ~653-657）: 同型。

### After

`backend/internal/trimming/trimming_service.go:497-500`（createDetail）:

```go
// Always validate the effective pet (appointment pet when request omits pet_id).
if err := reservation.ValidateReservationPetNotDeceased(txCtx, s.reservation, clinicID, finalPetID); err != nil {
    return err
}
```

`backend/internal/trimming/trimming_service.go:652-655`（Update）: 同型。

`Create`（:279）は着手時点で無条件検証済みのため未変更。

## Added tests

| # | Function | Path under test |
|---|----------|-----------------|
| ① | `TestTrimmingService_CreateExistingDetail_RejectsDeceasedPetWhenPetIDOmitted` | detail 作成・`PetID` 省略・予約ペット死亡 |
| ② | `TestTrimmingService_Update_RejectsDeceasedPetWhenPetIDOmitted` | Update・`PetID` 省略・予約ペット死亡 |
| ③ | `TestTrimmingService_Update_RejectsDeceasedPetReplacement` | Update・差し替え先死亡 |
| ④ | `TestTrimmingService_Create_RejectsDeceasedPet` | 通常 Create・ペット死亡 |

File: `backend/internal/trimming/trimming_deceased_pet_test.go`

Write/audit zero assert examples:

- create detail: lines 93-95 (`detailCreateCalls`, `appointmentUpdateCalls`, `audit.entries`)
- update omit: lines 145-147
- replacement: lines 199-201
- create: lines 244-246

## RED (before fix)

Command:

```bash
docker compose exec -T backend go test -p 1 ./internal/trimming -run 'TestTrimmingService_.*Deceased' -count=1
```

Verbatim:

```
--- FAIL: TestTrimmingService_CreateExistingDetail_RejectsDeceasedPetWhenPetIDOmitted (0.00s)
    trimming_deceased_pet_test.go:69: appointment must not be updated for a deceased appointment pet when pet_id is omitted
--- FAIL: TestTrimmingService_Update_RejectsDeceasedPetWhenPetIDOmitted (0.00s)
    trimming_deceased_pet_test.go:120: appointment must not be updated for a deceased appointment pet when pet_id is omitted
2026/08/01 15:59:07 ERROR failed to update trimming appointment error="死亡したペットは予約できません: invalid input"
2026/08/01 15:59:07 ERROR failed to create trimming appointment error="死亡したペットは予約できません: invalid input"
FAIL
FAIL	github.com/animal-ekarte/backend/internal/trimming	0.003s
FAIL
```

Nil pet_id の 2 経路が FAIL。Create と explicit replacement は修正前から PASS（既にガードあり）。

## GREEN (after fix)

Same command. Verbatim:

```
=== RUN   TestTrimmingService_CreateExistingDetail_RejectsDeceasedPetWhenPetIDOmitted
2026/08/01 15:59:30 ERROR failed to create trimming detail for existing appointment error="死亡したペットは予約できません: invalid input" appointment_id=77 clinic_id=1
--- PASS: TestTrimmingService_CreateExistingDetail_RejectsDeceasedPetWhenPetIDOmitted (0.00s)
=== RUN   TestTrimmingService_Update_RejectsDeceasedPetWhenPetIDOmitted
2026/08/01 15:59:30 ERROR failed to update trimming appointment error="死亡したペットは予約できません: invalid input"
--- PASS: TestTrimmingService_Update_RejectsDeceasedPetWhenPetIDOmitted (0.00s)
=== RUN   TestTrimmingService_Update_RejectsDeceasedPetReplacement
2026/08/01 15:59:30 ERROR failed to update trimming appointment error="死亡したペットは予約できません: invalid input"
--- PASS: TestTrimmingService_Update_RejectsDeceasedPetReplacement (0.00s)
=== RUN   TestTrimmingService_Create_RejectsDeceasedPet
2026/08/01 15:59:30 ERROR failed to create trimming appointment error="死亡したペットは予約できません: invalid input"
--- PASS: TestTrimmingService_Create_RejectsDeceasedPet (0.00s)
PASS
ok  	github.com/animal-ekarte/backend/internal/trimming	0.003s
```

## Regression (baseline vs after)

Command:

```bash
docker compose exec -T backend go test -p 1 ./internal/trimming ./internal/sharedkernel ./internal/reservation -count=1
```

| Package | Baseline | After |
|---------|----------|-------|
| trimming | ok 3.935s | ok 3.843s |
| sharedkernel | ok 0.002s | ok 0.003s |
| reservation | ok 12.998s | ok 11.992s |

FAIL sets: baseline empty, after empty → **new failures: 0**.

## phase2.html sync

「死亡ペット直 API の一括監査」を 2026-08-01 source 再実測に同期:

| Path | Measured |
|------|----------|
| trimming Create | `ValidateReservationPetNotDeceased` unconditional on `input.PetID` |
| trimming createDetail / Update | was gated; TASK-030 → always `finalPetID` |
| admin Create | `ValidateReservationPetNotDeceased` at `:140` (not missing) |
| LIFF resolveReservationPetID | no sharedkernel call; `livingPets` in-memory only |

Runtime not claimed green.

## Out of scope (confirmed)

- No edits to `appointment_admin_service.go` / `liff_service_reservations.go`
- No migration / seed / frontend
- No `todo.md` / `bug.md` / foreign WIP

## Independent review

code-reviewer subagent: **APPROVE**, no CRITICAL/HIGH/MEDIUM.
