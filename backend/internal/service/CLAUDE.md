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
