# BUG-080: 同一スタッフ・同一時刻の予約を二重登録できる（重複チェックなし）

## 種類
バグ（バックエンド — 予約重複チェック未実装）

## 重要度
高

## 発見日
2026-03-29

## 再現手順

1. スタッフ A の予約を特定日時（例: 2026-03-28 10:00）で作成する（HTTP 201）
2. 同じ `staff_id` + 同じ日時で再度予約を作成する
3. `POST /api/v1/reservations` → HTTP 201 で2件目も作成成功する

## 期待動作

- 同一スタッフが同時刻に重複予約を持つ場合は HTTP 409 Conflict を返す
- フロントエンドに「この時間帯は既に予約が入っています」エラーを表示する

## 実際の動作

- 同一 `staff_id` + 同一 `start_time` で POST を2回送信すると両方 HTTP 201 で作成成功する（id=19, 20 等）
- サービス層に重複チェックが実装されていない
- 二重予約が DB に保存される

## 影響範囲

- 予約作成 API 全体（`POST /api/v1/reservations`）
- 同一スタッフへの同時刻二重予約によるスケジュール混乱

## 修正方針

`reservation_service.go` の Create 処理で、予約前に同一 `staff_id` + `clinic_id` + 時間帯が重複する予約が存在しないかチェックする。
重複がある場合は `errors.ErrConflict` を返し、handler で HTTP 409 を返す。

```go
// 追加するチェック
existing, _ := repo.FindByStaffAndTimeSlot(ctx, clinicID, staffID, startTime, endTime)
if existing != nil {
    return nil, fmt.Errorf("reservation conflict: %w", ErrConflict)
}
```

## 優先度
高（二重予約はスケジュール管理の根幹に関わる）

## 派生イシュー

| イシュー | 領域 | 内容 |
|---------|------|------|
| BUG-080（BE） | Backend | reservation_service.go に同一 staff_id + 時間帯重複チェックを追加（ErrConflict → 409） |
