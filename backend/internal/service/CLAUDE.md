# Service Layer — P1 / P8 / P10 / P11 / P13 / P17

## P1: FindByID before Delete/Update (MANDATORY)

```go
// ✅
func (s *vaccineService) Update(ctx context.Context, clinicID, id uint64, input UpdateVaccineInput) (*model.Vaccine, error) {
    if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
        return nil, apperrors.Wrap(err, "failed to find vaccine")
    }
    // ...
}

// ❌ Update/Delete without FindByID first
```

## P8: apperrors.Wrap on all error returns (MANDATORY)

```go
// ✅
if err != nil {
    return nil, apperrors.Wrap(err, "failed to find vaccine")
}

// ❌
if err != nil {
    return nil, err  // 未ラップ
}
```

## P10: FK dependency check before Delete (MANDATORY)

```go
// ✅ — check references, return 409 if in use
count, err := s.repo.CountUsageByVaccineID(ctx, id)
if err != nil {
    return apperrors.Wrap(err, "failed to count usage")
}
if count > 0 {
    return apperrors.WrapConflict(fmt.Errorf("vaccine %d is in use", id))
}
```

**注意**: 末端エンティティ（他から FK 参照されない）は依存チェック不要。

## P11: slog.ErrorContext before repo error return (MANDATORY)

```go
// ✅
vaccines, err := s.repo.FindAll(ctx, clinicID)
if err != nil {
    slog.ErrorContext(ctx, "failed to find vaccines", "error", err)
    return nil, apperrors.Wrap(err, "failed to find vaccines")
}

// ❌ ログなし
if err != nil {
    return nil, apperrors.Wrap(err, "failed to find vaccines")
}
```

**除外（ログ不要）**: `WrapInvalidInput` / NotFound 存在確認 / `WrapConflict`

## P13: Definition order in service file (MANDATORY)

```
1. const
2. buildFunc (buildXxxUpdateFields etc.)
3. interface
4. struct
5. constructor (NewXxxService)
6. methods
```

## P17: Input struct naming (MANDATORY)

```go
// ✅
type CreateVaccineInput struct { ... }
type UpdateVaccineInput struct { ... }

// ❌
type VaccineCreateRequest struct { ... }  // 順序逆
type CreateVaccineParams struct { ... }   // Params は違反
```

## カルテ子エンティティ書込（MANDATORY）

カルテ（`medical_record`）の子エンティティ（treatment / examination / vital / prescription /
checkup_field_result）への書込は、確定(finalize)と子エンティティ書込が競合するレースを防ぐため、
以下の不変条件を必ず守る（BE-refactor.md X-11 由来）:

1. tx 内で `medicalRecordRepo.LockByIDForUpdate(ctx, clinicID, medicalRecordID)`（BE-refactor.md
   R31 で `LockDraftByID` からリネーム）を呼び、返却された `record.Status` を確認してから
   finalized チェックを行う。**名前に反して status 不問で行ロックする** — finalized 判定は
   呼び出し側（service）の責務であり、ロック取得自体は draft 限定ではない。
2. 子 repo（treatment/examination/vital/prescription/checkup_field_result）の Create/Update は
   `dbOrTx(ctx, r.db)` で ambient tx に参加させる。参加させないと、`LockByIDForUpdate` の
   `FOR UPDATE` 行ロックと子テーブルの `medical_record_id` FK チェックがデッドロックする。

既存 5 サービス（`treatment_service.go` / `examination_service.go` / `vital_service.go` /
`prescription_service.go` / `checkup_field_result_service.go`）が先例。検証は
`medical_record_finalize_lock_concurrency_test.go`（`LockByIDForUpdate` の行ロック自体の並行性）
と、個別 repo の tx atomicity test（`examination_repository_tx_atomicity_test.go` /
`checkup_field_result_tx_atomicity_test.go` 等）が担う。
