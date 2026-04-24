# TASK-182: CreateAccounting / UpdateAccounting Handler のポインタリテラル違反と型変換ロジック

## 優先度: Medium

## 概要
`CreateAccounting` Handler でローカル変数 `createInput` を構築してからポインタ渡しをしており、
ポインタリテラルパターン違反。また `model.BillingStatus(input.Status)` / `model.BillingStatus(*input.Status)` の
型変換がHandler 内に存在し、型マッピングは Service 入力 DTO 側で吸収すべき。

## 対象ファイル
`backend/internal/handler/accounting_handler.go`

## 現状コード（CreateAccounting, L103〜126）

```go
// ❌ ローカル変数に構築してからポインタ渡し（ポインタリテラル違反）
createInput := service.CreateAccountingInput{
    ClinicID:          clinicID,
    MedicalRecordID:   input.MedicalRecordID,
    ...
}
if input.Status != "" {
    // ❌ Handler 内で model 型変換
    createInput.Status = model.BillingStatus(input.Status)
}

ctx := c.Request.Context()
created, err := h.svc.Accounting.Create(ctx, &createInput)
```

```go
// UpdateAccounting (L163〜193) も同様
updateInput := service.UpdateAccountingInput{
    ID:       id,
    ClinicID: clinicID,
    ...
}
if input.Status != nil {
    // ❌ Handler 内で model 型変換
    s := model.BillingStatus(*input.Status)
    updateInput.Status = &s
}
if input.PaymentMethod != nil {
    // ❌ Handler 内で model 型変換
    m := model.PaymentMethod(*input.PaymentMethod)
    updateInput.PaymentMethod = &m
}
updated, err := h.svc.Accounting.Update(ctx, &updateInput)
```

## 修正後コード（CreateAccounting）

```go
// ✅ ポインタリテラルで直接渡す
// ✅ 型変換ロジックを Service DTO の入力として文字列のまま渡し、Service 側で変換する

// Service DTO を string → BillingStatus 変換に対応させる場合:
// CreateAccountingInput.Status を string 型に変更し Service 内で変換するか、
// または Handler でのリテラル構築時に型変換を含めてもよいが、
// ローカル変数への分解は不要。

created, err := h.svc.Accounting.Create(ctx, &service.CreateAccountingInput{
    ClinicID:          clinicID,
    MedicalRecordID:   input.MedicalRecordID,
    HospitalizationID: input.HospitalizationID,
    OwnerID:           input.OwnerID,
    PetID:             input.PetID,
    Subtotal:          input.Subtotal,
    TaxTotal:          input.TaxTotal,
    TotalAmount:       input.TotalAmount,
    HasInsurance:      input.HasInsurance,
    Status:            model.BillingStatus(input.Status),
    ScheduledDate:     input.ScheduledDate,
    CompletedAt:       input.CompletedAt,
    Memo:              input.Memo,
})
```

```go
// ✅ UpdateAccounting: status, paymentMethod も含めてポインタリテラルで構築
// （nil チェックが必要な場合はヘルパー関数で変換する）

func billingStatusPtr(s *string) *model.BillingStatus {
    if s == nil {
        return nil
    }
    v := model.BillingStatus(*s)
    return &v
}

func paymentMethodPtr(s *string) *model.PaymentMethod {
    if s == nil {
        return nil
    }
    v := model.PaymentMethod(*s)
    return &v
}

updated, err := h.svc.Accounting.Update(ctx, &service.UpdateAccountingInput{
    ID:                id,
    ClinicID:          clinicID,
    StaffID:           &staffID,
    ...
    Status:            billingStatusPtr(input.Status),
    PaymentMethod:     paymentMethodPtr(input.PaymentMethod),
})
```

## 補足
- `model` import は Handler 層でも許容されているが、型変換ヘルパーを file-level で定義して
  構築を一箇所に集約することが重要。
- ポインタリテラルパターン（`&service.XxxInput{...}`）により中間変数を排除する。
