# Phase 6: LINE連携

> **状態**: ✅ 全タスク完了（2026-04-08〜09）

## TASK-RES-060: LINE Messaging API連携 ✅

**実装済みファイル**: `backend/internal/service/line_messaging_service.go`

**実装内容**:
- `PushText(ctx, lineUserID, text string) error` メソッド
- `channelToken == ""` の場合は noop（環境変数未設定時はスキップ）
- `liff_service.go` 内の予約確定・キャンセル時に `ReservationNotifier` interface 経由で呼び出し

**完了条件**:
- [x] 予約確定時にLINE Push送信
- [x] キャンセル時にLINE Push送信
- [x] 送信失敗時のエラーハンドリング（予約自体は成功させる）

---

## TASK-RES-061: メール通知 ✅

**実装済みファイル**: `backend/internal/service/reservation_notification_service.go`

**実装内容**:
- `ReservationNotifier` interface（`NotifyCreated` / `NotifyCancelled`）
- fire-and-forget goroutine（15秒タイムアウト付き context）
- `net/smtp.SendMail` を使用したメール送信
- `backend/internal/config/config.go` に SMTP 設定を追加（環境変数）

**完了条件**:
- [x] 予約確定時にメール送信
- [x] キャンセル時にメール送信
- [x] 飼い主名・ペット情報・診察内容をメール本文に含む
