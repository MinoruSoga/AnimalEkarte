# CODE-QUALITY-228: vaccine_service Update メソッドの input nil チェック欠落

## 概要

`backend/internal/service/vaccine_service.go` の `Update` メソッドに
`input == nil` チェックが欠落している。
同ファイルの medicine_service / procedure_service と比べて不統一。

## 該当箇所

**ファイル:** `backend/internal/service/vaccine_service.go:145-160`

```go
func (s *vaccineService) Update(ctx context.Context, clinicID, id uint64, input *UpdateVaccineInput) (*model.Vaccine, error) {
    // ❌ input == nil チェックがない
    if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
        return nil, apperrors.Wrap(err, "failed to get vaccine")
    }
    if err := validateOptionalName(input.Name); err != nil {  // input が nil なら panic
        return nil, err
    }
    // ...
}
```

## medicine_service（正しい実装）との比較

```go
// medicine_service.go:222-226
func (s *medicineService) Update(ctx context.Context, clinicID, id uint64, input *UpdateMedicineInput) (*model.Medicine, error) {
    if input == nil {
        return nil, apperrors.WrapInvalidInput(ErrMsgInputNotNil)  // ✅ チェックあり
    }
    if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
        // ...
    }
}
```

## 修正方法

```go
func (s *vaccineService) Update(ctx context.Context, clinicID, id uint64, input *UpdateVaccineInput) (*model.Vaccine, error) {
    if input == nil {
        return nil, apperrors.WrapInvalidInput(ErrMsgInputNotNil)
    }
    if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
        // ...
    }
}
```

## 優先度

MEDIUM — handler の ShouldBindJSON が成功していれば nil にならないため実害リスクは低いが、
medicine/procedure との一貫性を保つために修正すべき。
