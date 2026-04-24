# BUG-399: merchandise_item の Update リクエストで UnitPrice・TaxRate の範囲バリデーション欠落

## 概要
`merchandise_item_request.go` の `createMerchandiseItemRequest` は `UnitPrice` に `min=0`、`TaxRate` に `min=0,max=1` の binding タグを持つが、`updateMerchandiseItemRequest` は同フィールドに範囲バリデーションがない。PATCH での更新時に負の単価や 100% 超の税率が通過する。

## 再現手順
1. `PATCH /masters/merchandise-items/:id` に `{"unit_price": -500}` を送信
2. **結果**: 200 OK — バリデーションなしで -500 が保存される
3. `PATCH /masters/merchandise-items/:id` に `{"tax_rate": 2.5}` を送信
4. **結果**: 200 OK — 250% という不正な税率が保存される

## 現状コード

### `backend/internal/handler/merchandise_item_request.go`
```go
// Create リクエスト（正しい）
type createMerchandiseItemRequest struct {
    Name      string   `json:"name"        binding:"required"`
    Category  string   `json:"category"    binding:"required,oneof=food goods other"`
    UnitPrice int64    `json:"unit_price"  binding:"min=0"`       // ← バリデーションあり
    TaxType   string   `json:"tax_type"    binding:"required,oneof=included excluded exempt"`
    TaxRate   float64  `json:"tax_rate"    binding:"min=0,max=1"` // ← バリデーションあり
    IsActive  *bool    `json:"is_active"`
    SortOrder *int     `json:"sort_order"`
}

// Update リクエスト（問題あり）
type updateMerchandiseItemRequest struct {
    Name      *string  `json:"name"`
    Category  *string  `json:"category"   binding:"omitempty,oneof=food goods other"`
    UnitPrice *int64   `json:"unit_price"`                        // ← min=0 が欠落
    TaxType   *string  `json:"tax_type"   binding:"omitempty,oneof=included excluded exempt"`
    TaxRate   *float64 `json:"tax_rate"`                          // ← min=0,max=1 が欠落
    IsActive  *bool    `json:"is_active"`
    SortOrder *int     `json:"sort_order"`
}
```

## 修正方針

### `merchandise_item_request.go` — Update リクエストに範囲バリデーション追加
```go
type updateMerchandiseItemRequest struct {
    Name      *string  `json:"name"`
    Category  *string  `json:"category"   binding:"omitempty,oneof=food goods other"`
    UnitPrice *int64   `json:"unit_price"  binding:"omitempty,min=0"`       // ← 追加
    TaxType   *string  `json:"tax_type"   binding:"omitempty,oneof=included excluded exempt"`
    TaxRate   *float64 `json:"tax_rate"   binding:"omitempty,min=0,max=1"` // ← 追加
    IsActive  *bool    `json:"is_active"`
    SortOrder *int     `json:"sort_order"`
}
```

**注意**: pointer 型なので `omitempty` を前置することで「nil の場合はスキップ、値がある場合は範囲チェック」という正しい動作になる。

## 優先度
**Low** — サービス層の `merchandise_item_service.go` で `UnitPrice < 0` チェック（BUG-380 の対象）があるため即座の問題にはならないが、ハンドラ層のバリデーションも統一すべき。二重防衛。

## 関連チケット
- **BUG-380**: サービス層の価格バリデーション不統一（UnitPrice < 0 チェックの統一）

## 関連ファイル
- `backend/internal/handler/merchandise_item_request.go` — 修正対象
