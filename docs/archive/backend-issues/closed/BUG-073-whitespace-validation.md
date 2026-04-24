# BE: BUG-073 空白のみ入力のバリデーション未実装

## 概要

飼主名（owner_name）等の文字列フィールドに空白のみ（`"   "`）を送信すると、HTTP 201 で登録成功してしまう。
`strings.TrimSpace()` 後の空チェックが行われていない。

## 再現手順

```
POST /v1/owners
{ "owner_name": "   ", "phone": "" }
→ HTTP 201 Created（期待: 400 Bad Request）
```

## 期待する動作

- 全ての必須文字列フィールドで `strings.TrimSpace(v) == ""` の場合は `400 Bad Request`
- エラーメッセージ例: `"owner_name は必須です"`

## 実装場所

- `backend/internal/service/` の各 validators.go
- `owner_service.go`、`pet_service.go`、`medical_record_service.go` 等
- または `backend/internal/handler/` のリクエストバリデーション共通処理

## 追加: FE/BE 必須フィールド定義の乖離

- フロントエンドは `phone` を必須として扱う
- バックエンドは `phone=""` で 201 成功
- どちらかに合わせて統一する（フロントに合わせて BE でも phone を必須化することを推奨）

## 優先度

Medium

## 関連

- `docs/tasks/open/validation/BUG-073_whitespace_validation.md`
- FUNCTIONAL_TEST_REPORT.md BUG-073
