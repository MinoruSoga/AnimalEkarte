# TASK-047: BE enum バリデーション統一

**作成日**: 2026-03-27
**ステータス**: Open
**優先度**: Medium
**領域**: Backend

---

## 概要

`handler/billing_item_handler.go` 等で Go の型キャスト（`model.TaxType(req.TaxType)`）を使用しているが、
リクエスト値が Go enum に存在しない文字列でも型変換が通過し、DB の ENUM 制約違反が内部エラーとしてクライアントに返る。

---

## 対象箇所

```go
// handler/billing_item_handler.go:22-33
model.TaxType(req.TaxType)         // バリデーションなし
model.ItemCategory(req.Category)   // バリデーションなし

// handler/treatment_handler.go:60
model.TreatmentItemType(req.ItemType)  // バリデーションなし
```

---

## 修正方針

`service/validators.go` に `validateEnum` 関数を追加し、handler 層でバリデーションする。

```go
// service/validators.go
func validateTaxType(v string) (model.TaxType, error) {
    switch model.TaxType(v) {
    case model.TaxTypeStandard, model.TaxTypeReduced, model.TaxTypeExempt:
        return model.TaxType(v), nil
    default:
        return "", apperrors.WrapInvalidInput(fmt.Sprintf("invalid tax_type: %s", v))
    }
}
```

または `ShouldBindJSON` の `binding:"oneof=standard reduced exempt"` タグで解決する（`binding:"oneof"` は Go Validator v10 でサポート済み）。

---

## 受入条件

- [ ] 無効な enum 値を送信した場合 400 Bad Request が返る（500 でなく）
- [ ] `billing_item_handler.go` と `treatment_handler.go` でバリデーション実装済み
- [ ] `docker compose exec backend go test ./...` 全テストパス
