# CODE-QUALITY-236: billing_item_service の英語エラーメッセージ

## 概要

`billing_item_service.go` に英語のエラーメッセージが存在する。
プロジェクト全体では日本語エラーメッセージに統一されているため不統一。

---

## 該当箇所

**ファイル:** `backend/internal/service/billing_item_service.go`

```go
// 約 line 90
return nil, apperrors.WrapInvalidInput("billing_id is required")  // ❌ 英語

// 約 line 93
return nil, apperrors.WrapInvalidInput("name is required")  // ❌ 英語
```

---

## 問題

- バックエンドの全サービスは日本語エラーメッセージを使用している
- `billing_item_service.go` のみ英語のメッセージが残っている
- フロントエンドにそのまま表示されると UX が一貫しない

---

## 修正案

```go
// billing_id is required → 日本語 + 定数化
return nil, apperrors.WrapInvalidInput("請求IDは必須です")

// name is required → validateRequiredName を使用
if err := validateRequiredName(input.Name); err != nil {
    return nil, err
}
```

または `validators.go` に定数を追加:

```go
const ErrMsgBillingIDRequired = "請求IDは必須です"
```

---

## 優先度

LOW — 機能上の問題はなくユーザー体験上の一貫性の問題。
billing_item は handler 側の ShouldBindJSON バリデーションで
`billing_id` や `name` の必須チェックが先に行われる可能性があるため
実際にこのメッセージが表示されるケースは限定的。
