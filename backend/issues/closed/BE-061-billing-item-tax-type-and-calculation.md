# BE-061: BillingItem / EstimateItem 課税区分 API + 税額計算ロジック

**Status**: Closed
**Priority**: High
**Affects**: accounting_handler, billing_service, estimate_handler, estimate_service
**Date Created**: 2026-03-25
**Related**: TASK-029, BE-058（先行必須）, FE-123

## Summary

`billing_items` と `estimate_items` に `tax_type` フィールドを追加し、
税額計算（外税/内税/非課税）をサービス層に実装する。
会計精算画面で課税区分・税率・税額を確認・編集できるようにするための API。

## 現状のコード

```go
// backend/internal/model/accounting.go:83（BillingItem）
TaxRate float64 `gorm:"type:numeric(3,2);default:0.10" json:"tax_rate"`
// TaxType フィールドなし
// 税額計算ロジックなし（フロントで計算？）

// billing_items テーブルの現状
// unit_price, quantity, tax_rate が存在
// tax_type が存在しない
// tax_amount（明細ごとの税額）が存在しない
```

## 必要な変更

### 1. 税額計算ロジックの実装（service 層）

```go
// backend/internal/service/billing_service.go（新規追加）

// CalculateTaxAmount: 課税区分に応じた税額を計算する
// 外税: 税額 = 単価 × 数量 × 税率
// 内税: 税額 = 単価 × 数量 × 税率 ÷ (1 + 税率)
// 非課税: 税額 = 0
func CalculateTaxAmount(unitPrice int64, quantity float64, taxType model.TaxType, taxRate float64) int64 {
    subtotal := float64(unitPrice) * quantity
    switch taxType {
    case model.TaxTypeExcluded:
        return int64(math.Round(subtotal * taxRate))
    case model.TaxTypeIncluded:
        return int64(math.Round(subtotal * taxRate / (1 + taxRate)))
    case model.TaxTypeExempt:
        return 0
    default:
        return 0
    }
}

// CalculateTotals: 全 billing_items から subtotal, tax_total, total_amount を再計算
func CalculateBillingTotals(items []model.BillingItem) (subtotal, taxTotal, totalAmount int64) {
    for _, item := range items {
        itemSubtotal := int64(float64(item.UnitPrice) * item.Quantity)
        taxAmount := CalculateTaxAmount(item.UnitPrice, item.Quantity, item.TaxType, item.TaxRate)
        subtotal += itemSubtotal
        taxTotal += taxAmount
    }
    totalAmount = subtotal + taxTotal
    // 内税の場合は tax が subtotal に含まれているため total_amount = subtotal
    // ※ 混在する場合は各 item で計算して集計する
    return
}
```

### 2. BillingItem の Create/Update リクエストへの tax_type 追加

```go
// backend/internal/handler/accounting_request.go

// CreateBillingItemRequest: tax_type を追加
type CreateBillingItemRequest struct {
    Category  string  `json:"category"  binding:"required"`
    Name      string  `json:"name"      binding:"required"`
    UnitPrice int64   `json:"unit_price" binding:"required,min=0"`
    Quantity  float64 `json:"quantity"  binding:"required,min=0.1"`
    TaxType   string  `json:"tax_type"  binding:"required,oneof=included excluded exempt"`
    TaxRate   float64 `json:"tax_rate"  binding:"required,oneof=0.10 0.08"`
    IsInsuranceApplicable bool `json:"is_insurance_applicable"`
    Source    string  `json:"source"    binding:"required"`
}

// UpdateBillingItemRequest: PATCH 用ポインタ型
type UpdateBillingItemRequest struct {
    UnitPrice *int64   `json:"unit_price"`
    Quantity  *float64 `json:"quantity"`
    TaxType   *string  `json:"tax_type" binding:"omitempty,oneof=included excluded exempt"`
    TaxRate   *float64 `json:"tax_rate" binding:"omitempty,oneof=0.10 0.08"`
    IsInsuranceApplicable *bool `json:"is_insurance_applicable"`
}
```

### 3. BillingItem の Response への tax_type + 税額 追加

```go
// backend/internal/handler/accounting_response.go

type BillingItemResponse struct {
    ID        string  `json:"id"`
    BillingID string  `json:"billing_id"`
    Name      string  `json:"name"`
    UnitPrice int64   `json:"unit_price"`
    Quantity  float64 `json:"quantity"`
    TaxType   string  `json:"tax_type"`   // "included" | "excluded" | "exempt"
    TaxRate   float64 `json:"tax_rate"`
    TaxAmount int64   `json:"tax_amount"` // 計算済み税額（表示用）
    Subtotal  int64   `json:"subtotal"`   // unit_price × quantity（税抜）
    // ...
}

func toBillingItemResponse(item *model.BillingItem) BillingItemResponse {
    taxAmount := service.CalculateTaxAmount(item.UnitPrice, item.Quantity, item.TaxType, item.TaxRate)
    subtotal := int64(float64(item.UnitPrice) * item.Quantity)
    return BillingItemResponse{
        // ...
        TaxType:   string(item.TaxType),
        TaxRate:   item.TaxRate,
        TaxAmount: taxAmount,
        Subtotal:  subtotal,
    }
}
```

### 4. Billing 合計の再計算トリガー

BillingItem が更新されたとき、親 Billing の `subtotal`, `tax_total`, `total_amount` を再計算して更新する。

```go
// billing_service.go に追加
func (s *BillingService) RecalculateBillingTotals(ctx context.Context, billingID uint64) error {
    items, err := s.repo.GetBillingItems(ctx, billingID)
    if err != nil {
        return fmt.Errorf("failed to get billing items: %w", err)
    }
    subtotal, taxTotal, totalAmount := CalculateBillingTotals(items)
    return s.repo.UpdateBillingTotals(ctx, billingID, subtotal, taxTotal, totalAmount)
}
```

### 5. EstimateItem も同様に対応

estimate_handler.go / estimate_service.go に同じパターンで tax_type を追加。

## API レスポンス形式

```json
// GET /v1/billings/:id（BillingItem を含む）
{
  "id": "1",
  "subtotal": 10000,
  "tax_total": 800,
  "total_amount": 10800,
  "items": [
    {
      "id": "1",
      "name": "診察料",
      "unit_price": 5000,
      "quantity": 1.0,
      "tax_type": "excluded",
      "tax_rate": 0.10,
      "tax_amount": 500,
      "subtotal": 5000
    },
    {
      "id": "2",
      "name": "薬剤費",
      "unit_price": 3000,
      "quantity": 1.0,
      "tax_type": "exempt",
      "tax_rate": 0.10,
      "tax_amount": 0,
      "subtotal": 3000
    }
  ]
}

// PATCH /v1/billing-items/:id
{
  "tax_type": "exempt",
  "tax_rate": 0.10
}
```

## フロントエンド影響

- FE-123 で billing_items の tax_type/tax_rate/tax_amount を表示・編集する UI を実装する

## 完了条件

- [ ] `CalculateTaxAmount()` が外税/内税/非課税の3ケースで正しく計算できる
- [ ] BillingItem の GET レスポンスに `tax_type`, `tax_amount`, `subtotal` が含まれる
- [ ] POST /v1/billing-items で tax_type が保存できる
- [ ] PATCH /v1/billing-items/:id で tax_type/tax_rate が更新でき、Billing の totals が再計算される
- [ ] EstimateItem も同様に対応されている
- [ ] `CalculateTaxAmount` の単体テストが存在する（外税/内税/非課税の3ケース）
- [ ] `docker compose exec backend go test ./... -v` が通る
