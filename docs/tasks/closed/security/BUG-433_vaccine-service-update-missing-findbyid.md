# BUG-433: vaccine_service の Update で FindByID 存在確認が欠落

## 概要

BUG-424/BUG-430 と同種の問題が `vaccine_service.go` の `Update` メソッドにも存在する。
`FindByID` による事前存在確認を行わずに `UpdateFields` を直接呼び出しており、
他クリニックのワクチンマスタを上書きできる可能性がある。

## 問題箇所

```go
// vaccine_service.go:90-111
func (s *vaccineService) Update(ctx context.Context, clinicID, id uint64, input *UpdateVaccineInput) (*model.Vaccine, error) {
    if err := validateOptionalName(input.Name); err != nil {
        return nil, err
    }
    if err := validateNonNegativePrice(input.Price, "金額"); err != nil {
        return nil, err
    }
    if input.Species != nil {
        if err := validateVaccineSpecies(*input.Species); err != nil {
            return nil, err
        }
    }
    fields := buildVaccineUpdateFields(input)
    if len(fields) == 0 {
        return nil, apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)
    }
    vaccine, err := s.repo.UpdateFields(ctx, clinicID, id, fields)  // ← FindByID 存在確認なし
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to update vaccine")
    }
    slog.InfoContext(ctx, "vaccine updated", slog.Uint64("clinic_id", clinicID), slog.Uint64("vaccine_id", id))
    return vaccine, nil
}
```

## 標準パターン（cage_service.go）

```go
func (s *cageService) Update(ctx context.Context, clinicID, id uint64, input *UpdateCageInput) (*model.Cage, error) {
    // ✅ Step 1: 存在確認（テナント検証含む）
    existing, err := s.repo.FindByID(ctx, clinicID, id)
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to get cage")
    }
    // Step 2: フィールド更新 ...
}
```

## 修正方針

Update メソッドの先頭に存在確認を追加する。

```go
func (s *vaccineService) Update(ctx context.Context, clinicID, id uint64, input *UpdateVaccineInput) (*model.Vaccine, error) {
    // ← 追加: 存在確認・テナント検証
    if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
        return nil, apperrors.Wrap(err, "failed to get vaccine")
    }

    if err := validateOptionalName(input.Name); err != nil {
        return nil, err
    }
    // ... 以降は変更なし
}
```

## 影響ファイル

- `backend/internal/service/vaccine_service.go` — 行 90-111（Update メソッド）

## 優先度

**High** — マルチテナント境界の保護が不完全。他クリニックのワクチンマスタを上書きできる可能性。

## 関連チケット

- BUG-424（reservation_type / trimming_master×2 / diagnosis_name の同種問題）
- BUG-430（reservation_type_group_service の同種問題）
