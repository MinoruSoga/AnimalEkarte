# CODE-QUALITY-233: service Update input nil チェック欠落 (insurance / chief_complaint)

## 概要

`insurance_service.go` と `chief_complaint_service.go` の `Update` メソッドに
`input == nil` チェックが欠落している。nil が渡されると `validateOptionalName(input.Name)` で panic する。

medicine / procedure / vaccine (CODE-QUALITY-228) / hospitalization_plan (CODE-QUALITY-229) と
同じパターンの問題。

---

## 問題1（HIGH）: insurance_service — Update の input nil チェック欠落

**ファイル:** `backend/internal/service/insurance_service.go:125-145`

### 現状コード

```go
func (s *insuranceService) Update(ctx context.Context, clinicID, id uint64, input *UpdateInsuranceInput) (*model.Insurance, error) {
    // ❌ input == nil チェックなし
    if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
        return nil, apperrors.Wrap(err, "failed to get insurance")
    }
    if err := validateOptionalName(input.Name); err != nil {  // input が nil なら panic
        return nil, err
    }
    // ...
}
```

### 修正案

```go
func (s *insuranceService) Update(ctx context.Context, clinicID, id uint64, input *UpdateInsuranceInput) (*model.Insurance, error) {
    if input == nil {
        return nil, apperrors.WrapInvalidInput(ErrMsgInputNotNil)  // ← 追加（FIRST CHECK）
    }
    if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
        return nil, apperrors.Wrap(err, "failed to get insurance")
    }
    // ...
}
```

---

## 問題2（HIGH）: chief_complaint_service — Update の input nil チェック・FindByID 欠落

**ファイル:** `backend/internal/service/chief_complaint_service.go:110-128`

### 現状コード

```go
func (s *chiefComplaintTypeService) Update(ctx context.Context, clinicID, id uint64, input *UpdateChiefComplaintTypeInput) (*model.ChiefComplaintType, error) {
    // ❌ input == nil チェックなし
    // ❌ FindByID 存在確認なし（404 が返らない）
    if err := validateOptionalName(input.Name); err != nil {  // input が nil なら panic
        return nil, err
    }
    fields := buildChiefComplaintTypeUpdateFields(input)
    if len(fields) == 0 {
        return nil, apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)
    }
    result, err := s.repo.UpdateFields(ctx, clinicID, id, fields)
    // ...
}
```

### 問題点
1. `input == nil` チェックなし → panic リスク
2. `FindByID` による存在確認なし → 存在しないレコードの更新で 404 でなく 500 か空レスポンスになる

### 修正案（medicine_service, inquiry_template_service を参照）

```go
func (s *chiefComplaintTypeService) Update(ctx context.Context, clinicID, id uint64, input *UpdateChiefComplaintTypeInput) (*model.ChiefComplaintType, error) {
    if input == nil {
        return nil, apperrors.WrapInvalidInput(ErrMsgInputNotNil)
    }
    if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
        return nil, apperrors.Wrap(err, "failed to get chief complaint type")
    }
    if err := validateOptionalName(input.Name); err != nil {
        return nil, err
    }
    // ...
}
```

---

## 優先度

| 問題 | 優先度 | 理由 |
|------|--------|------|
| insurance Update nil チェック欠落 | HIGH | panic リスク |
| chief_complaint Update nil チェック・FindByID 欠落 | HIGH | panic リスク + 404 応答不正 |
