# BE: BUG-064 電話番号の重複チェックなし（オーナー登録）

## 概要

オーナー新規登録時、メールアドレス重複は 409 を返すが電話番号重複はチェックされない。
同一電話番号・別メールアドレスで重複登録が成功してしまう。

## 再現手順

```
POST /api/v1/owners
{"email": "new@example.com", "phone": "080-7788-9900"}  ← 既存と同じ電話番号
→ HTTP 201 Created（期待: 409 Conflict）
```

## 期待する動作

- 同一電話番号が既存の場合: 409 Conflict
- エラーメッセージ: `"この電話番号はすでに登録されています"`
- 409 レスポンスボディに既存オーナーの ID を含めると FE での UX 改善に使える

```json
{
  "error": "この電話番号はすでに登録されています",
  "existing_owner_id": 42
}
```

## 実装場所

- `backend/internal/repository/owner_repository.go` の Create メソッド
- DB: `owners.phone` に UNIQUE 制約追加（既存レコードへの影響を確認すること）

## 優先度

High（データ品質・業務オペレーション上の重大リスク）

## 関連

- `docs/tasks/open/validation/BUG-064_duplicate_owner_check.md`
- FUNCTIONAL_TEST_REPORT.md BUG-064
