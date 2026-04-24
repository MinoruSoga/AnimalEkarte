# BE-060: 全マスタ商品 tax_type / tax_rate API 対応

**Status**: Closed
**Priority**: High
**Affects**: consultation_handler, procedure_handler, medicine_handler, merchandise_item_handler, hospitalization_plan_handler
**Date Created**: 2026-03-25
**Related**: TASK-029, BE-058（先行必須）, FE-122

## Summary

全マスタ商品（診察・処置・薬剤・入院プラン・物販）の CRUD API に `tax_type` と `tax_rate`（薬剤・診察・処置・入院プランは新規）を追加する。
リクエスト/レスポンスの型更新と service の buildUpdateFields への追加が主な作業。

## 現状のコード

```go
// 各マスタの現状（consultations を例に）
// backend/internal/model/consultation.go
// TaxType: なし
// TaxRate: なし（price カラムのみ）

// merchandise_items のみ TaxRate は既存
// backend/internal/model/merchandise_item.go:16
TaxRate float64 `gorm:"type:numeric;not null;default:0.10" json:"tax_rate"`
// TaxType: なし
```

## 必要な変更

### 対象テーブル・ファイル一覧

| マスタ | handler ファイル | 追加フィールド |
|-------|----------------|-------------|
| consultations | consultation_handler.go | tax_type, tax_rate（両方新規） |
| procedures | procedure_handler.go | tax_type, tax_rate（両方新規） |
| medicines | medicine_handler.go | tax_type, tax_rate（両方新規） |
| hospitalization_plans | hospitalization_plan_handler.go | tax_type, tax_rate（両方新規） |
| merchandise_items | merchandise_item_handler.go | tax_type のみ（tax_rate は既存） |

### 1. Response 型への追加（全マスタ共通パターン）

```go
// 例: consultation_response.go
type ConsultationResponse struct {
    ID       string  `json:"id"`
    Name     string  `json:"name"`
    Price    *int64  `json:"price"`
    TaxType  string  `json:"tax_type"` // "included" | "excluded" | "exempt"
    TaxRate  float64 `json:"tax_rate"` // 0.10 or 0.08
    IsActive bool    `json:"is_active"`
    // ... 既存フィールド
}

func toConsultationResponse(c *model.Consultation) ConsultationResponse {
    return ConsultationResponse{
        // ... 既存フィールド
        TaxType: string(c.TaxType),
        TaxRate: c.TaxRate,
    }
}
```

### 2. Create リクエストへの追加

```go
// 例: CreateConsultationRequest
type CreateConsultationRequest struct {
    Name     string   `json:"name"     binding:"required"`
    Price    *int64   `json:"price"`
    TaxType  string   `json:"tax_type" binding:"required,oneof=included excluded exempt"`
    TaxRate  float64  `json:"tax_rate" binding:"required,oneof=0.10 0.08"`
    // ... 既存フィールド
}
```

### 3. Update リクエストへの追加（PATCH: ポインタ型）

```go
// 例: UpdateConsultationRequest
type UpdateConsultationRequest struct {
    Name    *string  `json:"name"`
    Price   *int64   `json:"price"`
    TaxType *string  `json:"tax_type" binding:"omitempty,oneof=included excluded exempt"`
    TaxRate *float64 `json:"tax_rate" binding:"omitempty,oneof=0.10 0.08"`
    // ... 既存フィールド
}
```

### 4. Service Input + buildUpdateFields への追加

```go
// 例: service/consultation_service.go
type UpdateConsultationInput struct {
    // ... 既存フィールド
    TaxType *model.TaxType
    TaxRate *float64
}

func buildConsultationUpdateFields(input UpdateConsultationInput) map[string]any {
    fields := make(map[string]any)
    // ... 既存フィールド
    if input.TaxType != nil {
        fields["tax_type"] = *input.TaxType
    }
    if input.TaxRate != nil {
        fields["tax_rate"] = *input.TaxRate
    }
    return fields
}
```

### 5. MerchandiseItem の差分（tax_rate は既存のため TaxType のみ追加）

```go
// merchandise_item_response.go: TaxType を追加（TaxRate は既存）
// merchandise_item_request.go: UpdateMerchandiseItemRequest に TaxType *string を追加
// buildMerchandiseItemUpdateFields() に TaxType の分岐を追加
```

## API レスポンス形式

```json
// GET /v1/consultations/:id
{
  "id": "1",
  "name": "初診料",
  "price": 3000,
  "tax_type": "excluded",
  "tax_rate": 0.10,
  "is_active": true
}

// POST /v1/consultations
{
  "name": "初診料",
  "price": 3000,
  "tax_type": "excluded",
  "tax_rate": 0.10
}
```

## フロントエンド影響

- FE-122 で各マスタフォームに tax_type/tax_rate セレクタを追加する

## 完了条件

- [ ] consultations, procedures, medicines, hospitalization_plans, merchandise_items の GET レスポンスに tax_type/tax_rate が含まれる
- [ ] 各マスタの POST で tax_type/tax_rate が保存できる
- [ ] 各マスタの PATCH で tax_type/tax_rate が更新できる
- [ ] tax_type が "included"|"excluded"|"exempt" 以外の場合 400 が返る
- [ ] tax_rate が 0.10/0.08 以外の場合 400 が返る
- [ ] `docker compose exec backend go test ./... -v` が通る
