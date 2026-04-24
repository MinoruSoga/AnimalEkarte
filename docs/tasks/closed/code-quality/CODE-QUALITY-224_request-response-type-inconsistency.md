# CODE-QUALITY-224: Request/Response 型の不整合 2 件

## 問題1: medicine_request の DefaultQuantity が非ポインタ型

**ファイル:** `backend/internal/handler/medicine_request.go:33`

### 現状

```go
type updateMedicineRequest struct {
    Name            *string  `json:"name"`
    ParentID        *uint64  `json:"parent_id"`
    ClearParentID   bool     `json:"clear_parent_id"`
    Price           *int64   `json:"price"`
    IsActive        *bool    `json:"is_active"`
    Description     *string  `json:"description"`
    DosageForm      *string  `json:"dosage_form" binding:"omitempty,oneof=tablet liquid injection topical powder"`
    MedicineUnit    *string  `json:"medicine_unit"`
    InventoryID     *uint64  `json:"inventory_id"`
    DefaultQuantity float64  `json:"default_quantity"`  // ❌ 非ポインタ
    SortOrder       *int     `json:"sort_order"`
    TaxType         *string  `json:"tax_type" binding:"omitempty,oneof=included excluded exempt"`
    TaxRate         *float64 `json:"tax_rate" binding:"omitempty,min=0,max=1"`
}
```

### 問題

PATCH（部分更新）では全フィールドをポインタ型にしないと、ゼロ値（`0.0`）と
「フィールド未指定」の区別ができない。

- `DefaultQuantity float64` の場合、`0.0` を送っても `buildMedicineUpdateFields` で
  `if input.DefaultQuantity != nil` できないため、`0.0` への変更が正しく伝達されない
- service 側の `UpdateMedicineInput.DefaultQuantity` は `*float64` であるため、
  handler → service の型変換で `&req.DefaultQuantity` を渡すと
  「省略された（nil）」と「明示的に0.0」を区別できない

### service の期待型（medicine_service.go:44）

```go
type UpdateMedicineInput struct {
    // ...
    DefaultQuantity *float64  // ← ポインタ型
    // ...
}
```

### 修正案

```go
type updateMedicineRequest struct {
    // ...
    DefaultQuantity *float64 `json:"default_quantity"`  // ✅ ポインタに変更
    // ...
}
```

---

## 問題2: procedure_response と medicine_response の型表現不統一

### 現状

**`medicine_response.go:21-22`** — 明示的に string 変換:
```go
type medicineResponse struct {
    // ...
    TaxType string `json:"tax_type"`  // ← string で明示
    // ...
}

// 変換時
TaxType: string(m.TaxType),  // ← 明示的キャスト
```

**`procedure_response.go:17`** — model 型をそのまま使用:
```go
type procedureResponse struct {
    // ...
    Anesthesia model.AnesthesiaType `json:"anesthesia"`  // ← model 型
    TaxType    model.TaxType        `json:"tax_type"`    // ← model 型
    // ...
}
```

### 問題

- どちらも JSON 出力は同じ文字列になるが、response struct の型が混在
- response struct は外部 API 契約であり、内部型（`model.*`）への依存を持つべきでない
- 新規開発者が response struct を見たとき、どちらが正しいパターンか判断できない

### 修正方針

`medicine_response.go` の `string` 型へ統一する（model 型を除去）:

```go
// ✅ 統一後
type procedureResponse struct {
    // ...
    Anesthesia string `json:"anesthesia"`  // model 型ではなく string
    TaxType    string `json:"tax_type"`    // model 型ではなく string
    // ...
}

// 変換時
Anesthesia: string(p.Anesthesia),
TaxType:    string(p.TaxType),
```

同様のパターンを持つ全 response ファイルで `model.*` 型の直接使用を確認し、
string 変換に統一すること。

## 優先度

- 問題1: HIGH — `DefaultQuantity = 0.0` への更新が意図通り機能しない可能性
- 問題2: MEDIUM — 機能上の問題はないが、response struct の API 契約として内部型漏洩
