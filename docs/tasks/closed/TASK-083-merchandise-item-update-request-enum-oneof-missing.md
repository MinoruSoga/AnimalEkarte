# TASK-083: merchandise_item / billing_item — Update リクエスト ENUM フィールドの oneof バリデーション欠落

## 優先度

MEDIUM

---

## 概要

`merchandise_item_request.go` と `accounting_request.go`（billing_item）の Update リクエストで、
ENUM フィールドに `binding:"omitempty,oneof=..."` が付与されていない。

Create リクエストには正しく付与されているのに、Update リクエストで欠落している。
TASK-076（procedure/vaccine/cage）と同じパターン。

---

## 問題箇所

### merchandise_item_request.go

```go
// ❌ Create: 正しく実装済み
type createMerchandiseItemRequest struct {
    Name     string  `json:"name"     binding:"required"`
    Category string  `json:"category" binding:"required,oneof=food goods other"`
    TaxType  string  `json:"tax_type" binding:"required,oneof=included excluded exempt"`
    // ...
}

// ❌ Update: oneof バリデーションが欠落
type updateMerchandiseItemRequest struct {
    Name     *string `json:"name"`
    Category *string `json:"category"`  // ← binding なし
    TaxType  *string `json:"tax_type"`  // ← binding なし
    // ...
}
```

---

## 修正方針

```go
// ✅ 修正後
type updateMerchandiseItemRequest struct {
    Name     *string `json:"name"`
    Category *string `json:"category" binding:"omitempty,oneof=food goods other"`
    TaxType  *string `json:"tax_type" binding:"omitempty,oneof=included excluded exempt"`
    // ...
}
```

---

---

## 2. billing_item — accounting_request.go

### 問題箇所

```go
// ✅ Create: 正しく実装済み
type createBillingItemRequest struct {
    Category string `json:"category" binding:"omitempty,oneof=examination test procedure surgery medicine food goods other"`
    TaxType  string `json:"tax_type" binding:"omitempty,oneof=included excluded exempt"`
    // ...
}

// ❌ Update: TaxType に oneof バリデーションが欠落（L66）
type updateBillingItemRequest struct {
    TaxType *string `json:"tax_type"`  // ← binding なし
    // ...
}
```

### 修正後

```go
// ✅ 修正後
type updateBillingItemRequest struct {
    TaxType *string `json:"tax_type" binding:"omitempty,oneof=included excluded exempt"`
    // ...
}
```

---

## 修正ファイル

| ファイル | 修正内容 |
|---------|---------|
| `merchandise_item_request.go` | `updateMerchandiseItemRequest` の `Category`, `TaxType` に `binding:"omitempty,oneof=..."` 追加 |
| `accounting_request.go` | `updateBillingItemRequest` の `TaxType` に `binding:"omitempty,oneof=included excluded exempt"` 追加 |

---

## 関連タスク

- TASK-076: procedure_request / vaccine_request / cage_request の同種問題（Update ENUM oneof 欠落）
