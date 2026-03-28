# BUG-034: 予約登録で end_time < start_time を指定すると HTTP 500 が返る

## 種類
バグ（バックエンド — バリデーション未実装）

## 重要度
中

## 発見日
2026-03-28

## 再現手順
1. `POST /api/v1/reservations` に end_time が start_time より前の値を指定して送信
   ```json
   {
     "pet_id": 1,
     "staff_id": 1,
     "service_type_id": 1,
     "start_time": "2026-04-01T14:00:00Z",
     "end_time": "2026-04-01T13:00:00Z",
     "status": "confirmed"
   }
   ```

## 期待動作
- HTTP 400 Bad Request が返る
- エラーメッセージ: 「終了時刻は開始時刻より後に設定してください」等

## 実際の動作
- HTTP 500 Internal Server Error が返る（`{"error":"internal server error"}`）

## 影響
- 不正な予約データ（終了前に開始する予約）がDBに入る可能性
- バリデーションエラーとして適切に処理されない

## 修正方針
### バックエンド
- `backend/internal/service/` の CreateReservation 処理で end_time > start_time の検証を追加
- 違反時は `errors.ErrInvalidInput` を返して HTTP 400 にする

## 対象ファイル（推定）
- `backend/internal/service/reservation_service.go`
