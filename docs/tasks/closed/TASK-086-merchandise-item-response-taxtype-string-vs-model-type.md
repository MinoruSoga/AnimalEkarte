# TASK-086: merchandise_item_response — TaxType が string 型（model.TaxType に統一すべき）

## 優先度

LOW

---

## 概要

`merchandise_item_response.go` の TaxType フィールドが生の `string` 型で定義されており、
参照実装（`procedure_response.go`）の `model.TaxType` と型が不統一。

JSON 出力は同一だが、Go の型システムによる不正値の防止が機能しない。

---

## 問題箇所

### merchandise_item_response.go:15

```go
// ❌ 生の string 型
type merchandiseItemResponse struct {
    ID          uint64    `json:"id"`
    ClinicID    uint64    `json:"clinic_id"`
    Name        string    `json:"name"`
    Category    string    `json:"category"`   // ← model.MerchandiseCategory ではない
    TaxType     string    `json:"tax_type"`   // ← model.TaxType ではない
    // ...
}

func toMerchandiseItemResponse(item *model.MerchandiseItem) merchandiseItemResponse {
    return merchandiseItemResponse{
        // ...
        Category: string(item.Category),   // 明示的キャスト（冗長）
        TaxType:  string(item.TaxType),    // 明示的キャスト（冗長）
    }
}
```

---

## 参照実装（procedure_response.go）

```go
// ✅ model 型を直接使用
type procedureResponse struct {
    // ...
    Anesthesia model.AnesthesiaType `json:"anesthesia"`
    TaxType    model.TaxType        `json:"tax_type"`   // model 型
    // ...
}

func toProcedureResponse(p *model.Procedure) procedureResponse {
    return procedureResponse{
        // ...
        Anesthesia: p.Anesthesia,   // キャスト不要
        TaxType:    p.TaxType,      // キャスト不要
    }
}
```

---

## 修正方針

```go
// ✅ 修正後
type merchandiseItemResponse struct {
    ID          uint64                    `json:"id"`
    ClinicID    uint64                    `json:"clinic_id"`
    Name        string                    `json:"name"`
    Category    model.MerchandiseCategory `json:"category"`
    TaxType     model.TaxType             `json:"tax_type"`
    // ...
}

func toMerchandiseItemResponse(item *model.MerchandiseItem) merchandiseItemResponse {
    return merchandiseItemResponse{
        // ...
        Category: item.Category,   // キャスト不要
        TaxType:  item.TaxType,    // キャスト不要
    }
}
```

---

## 修正ファイル

| ファイル | 修正内容 |
|---------|---------|
| `merchandise_item_response.go` | `TaxType string` → `model.TaxType`、`Category string` → `model.MerchandiseCategory` に変更。変換関数のキャストを削除 |
