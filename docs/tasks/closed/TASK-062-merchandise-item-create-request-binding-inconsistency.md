# TASK-062: merchandise_item Create request — UnitPrice/TaxType バリデーション不整合

## 優先度

LOW

---

## 概要

`createMerchandiseItemRequest` の数値・ENUM フィールドが他マスタの Create request と一貫していない。

1. `UnitPrice int64`（非ポインタ・binding なし）→ 未指定時に 0 が登録される（intent が不明）
2. `TaxType string`（`binding:"omitempty,oneof=..."` ）→ 未指定でも登録可能。他のマスタ（procedure / cage / vaccine）は `binding:"required,oneof=..."` で必須
3. `TaxRate float64`（非ポインタ・binding なし）→ 未指定時に 0.0 が登録される

---

## 問題箇所

### backend/internal/handler/merchandise_item_request.go

```go
// ❌ 現状
type createMerchandiseItemRequest struct {
    Name      string  `json:"name"      binding:"required"`
    Category  string  `json:"category"  binding:"required,oneof=food goods other"`
    UnitPrice int64   `json:"unit_price"`                                   // ← binding なし・非ポインタ
    TaxType   string  `json:"tax_type"  binding:"omitempty,oneof=included excluded exempt"` // ← omitempty で任意
    TaxRate   float64 `json:"tax_rate"`                                     // ← binding なし・非ポインタ
    IsActive  bool    `json:"is_active"`
    SortOrder int     `json:"sort_order"`
}
```

### 他マスタとの比較

| フィールド | merchandise_item | procedure / cage / vaccine |
|-----------|----------------|---------------------------|
| 単価フィールド | `UnitPrice int64`（binding なし） | `Price *int64`（ポインタ・オプション） |
| TaxType | `omitempty` | `binding:"required,oneof=..."` |

---

## 確認事項

修正前に以下の仕様を確認すること:

1. **`UnitPrice = 0` は有効か?** — 無料の商品（price=0）が存在するなら非ポインタ int64 のままで OK。ただし `binding:"min=0"` 程度は付けるべき。
2. **`TaxType` は必須か?** — 必須であれば `binding:"required,oneof=included excluded exempt"` に変更する。
3. **`TaxRate` は必須か?** — `TaxType=exempt` のとき `TaxRate=0` が自然なのでポインタ不要かもしれない。

---

## 修正方針（仕様確認後）

```go
// ✅ TaxType が必須の場合の修正例
type createMerchandiseItemRequest struct {
    Name      string  `json:"name"      binding:"required"`
    Category  string  `json:"category"  binding:"required,oneof=food goods other"`
    UnitPrice int64   `json:"unit_price" binding:"min=0"`
    TaxType   string  `json:"tax_type"  binding:"required,oneof=included excluded exempt"`
    TaxRate   float64 `json:"tax_rate"  binding:"min=0,max=1"`
    IsActive  bool    `json:"is_active"`
    SortOrder int     `json:"sort_order"`
}
```

---

## 備考

- `IsActive` と `SortOrder` が非ポインタで binding なしは他のマスタと共通のパターンなので問題なし
- TASK-058（ENUM binding oneof タグ）とは別問題。TASK-058 は ENUM フィールドに `oneof` が全くない問題、本タスクは `TaxType` の `omitempty` と `required` の選択の問題
