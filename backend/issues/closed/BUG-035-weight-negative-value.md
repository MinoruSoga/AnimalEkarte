# BE: BUG-035 体重フィールドの負値をバックエンドが受け入れる

## 概要

ペット情報の体重フィールドで負の値を API に直接送信すると保存される。
フロントエンドバリデーションは修正済みだが、バックエンドに検証がない。

## 再現手順

```
PATCH /api/v1/pets/:id
{"weight": -5}
→ HTTP 200 で保存成功（期待: 400 Bad Request）
```

## 期待する動作

- `weight < 0` の場合は 400 Bad Request を返す
- エラーメッセージ: `"weight must be greater than or equal to 0"`

## 実装場所

- `backend/internal/service/pet_service.go` または `handler/pet_handler.go`
- `UpdatePetInput` の weight フィールドにバリデーション追加

## 優先度

Medium

## 関連

- `docs/tasks/open/validation/BUG-035_weight_backend_negative.md`
- FUNCTIONAL_TEST_REPORT.md BUG-035
