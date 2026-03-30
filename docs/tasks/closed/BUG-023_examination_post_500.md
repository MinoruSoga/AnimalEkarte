# BUG-023: 検査登録 POST → 500 Internal Server Error

## 概要
`/examinations/new` から単独で検査を登録しようとすると、バックエンドが 500 エラーを返す。
`medical_record_id` が null で送信されるが、バックエンドがこれを必須扱いしているため。

## 再現手順
1. `/examinations/new?petId=1` にアクセス
2. 検査種別、担当医、ステータスを入力して「保存」
3. → POST `/api/v1/examinations` → HTTP 500

## リクエスト
```json
{
  "medical_record_id": null,
  "pet_id": 1,
  "exam_type_id": 1,
  "doctor_id": 1,
  "date": "2026-03-30T07:11:56.783Z"
}
```

## レスポンス
```json
{"error": "internal server error"}
```

## 期待する動作
- `medical_record_id` が null の場合でも検査登録は成功する（201）
- または、バックエンドが `medical_record_id` を必須とする場合は、400 Bad Request + 適切なエラーメッセージを返す

## 実装場所
- `backend/internal/handler/examination_handler.go` または `service/examination_service.go`
- `medical_record_id` が null の場合のハンドリングを追加

## 優先度
Medium

## 関連
- FUNCTIONAL_TEST_REPORT.md BUG-023
- テスト確認日: 2026-03-30
