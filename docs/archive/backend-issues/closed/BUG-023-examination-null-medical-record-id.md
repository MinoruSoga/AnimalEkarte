# BE: BUG-023 検査登録 medical_record_id=null で 500 エラー

## 概要

`POST /api/v1/examinations` で `medical_record_id: null` を送信すると HTTP 500 を返す。
カルテに紐づかない単独検査登録のユースケースが想定されているが、バックエンドが null を許容していない。

## 再現手順

```
POST /api/v1/examinations
{
  "medical_record_id": null,
  "pet_id": 1,
  "exam_type_id": 1,
  "doctor_id": 1,
  "date": "2026-03-30T07:11:56.783Z"
}
→ HTTP 500 Internal Server Error
```

## 期待する動作

- `medical_record_id` が null の場合でも検査登録が成功する（201）
- または、null を禁止とする場合は 400 Bad Request + 適切なエラーメッセージ（500 は不可）

## 実装場所

- `backend/internal/handler/examination_handler.go` または `service/examination_service.go`
- `medical_record_id` が null の場合のハンドリングを追加

## 優先度

Medium

## 関連

- `docs/tasks/open/crash/BUG-023_examination_post_500.md`
- FUNCTIONAL_TEST_REPORT.md BUG-023
