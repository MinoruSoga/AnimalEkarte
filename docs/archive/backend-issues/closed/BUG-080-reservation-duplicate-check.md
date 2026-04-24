# BE: BUG-080 同一医師・同一時刻の予約が重複作成できる

## 概要

同一 `staff_id` + 同一時刻で予約を2回送信すると、両方とも 201 で作成成功してしまう。
サービス層に時間帯重複チェックがない。

## 再現手順

```
POST /api/v1/reservations（同一 staff_id + 同一 scheduled_at を2回送信）
→ 両方 HTTP 201 Created（id が異なる2件が作成される）
```

## 期待する動作

- 同一医師・同一時刻の予約が既存の場合は 409 Conflict を返す
- エラーメッセージ: 「指定された時刻にはすでに予約が入っています」

## 実装場所

- `backend/internal/service/reservation_service.go` の Create メソッド
- 予約作成前に同一 `staff_id` + `scheduled_at` の既存予約を確認するロジックを追加

```go
// 重複チェック例
var count int64
s.repo.db.Model(&Reservation{}).
    Where("staff_id = ? AND scheduled_at = ? AND deleted_at IS NULL", staffID, scheduledAt).
    Count(&count)
if count > 0 {
    return nil, apperrors.ErrConflict
}
```

## 優先度

High（データ整合性）

## 関連

- `docs/tasks/open/crash/BUG-080_reservation_duplicate.md`
- FUNCTIONAL_TEST_REPORT.md BUG-080
