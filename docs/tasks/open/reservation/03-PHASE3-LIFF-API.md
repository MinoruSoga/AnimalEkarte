# Phase 3: バックエンド公開API（LIFF用）

> **状態**: ✅ 全タスク完了（2026-04-08〜09）
>
> **設計方針**: 予約データは既存 `reservation_appointments` に直接INSERT。
> LINE予約確定 → reservation_appointments にレコード作成 → カルテシステムに自動反映。
> `source = 'line'` で手動予約（`source = 'manual'`）と区別する。

## TASK-RES-020: LIFF認証ミドルウェア ✅

**実装済みファイル**: `backend/internal/middleware/liff_auth.go`

**完了条件**:
- [x] 有効なLIFF ID Tokenで認証成功
- [x] 無効なトークンで401
- [x] 新規ユーザー自動作成

---

## TASK-RES-021: 公開予約フローAPI ✅

**実装済みファイル**:
- `backend/internal/handler/liff_handler.go`
- `backend/internal/handler/liff_request.go`
- `backend/internal/handler/liff_response.go`
- `backend/internal/service/liff_service.go`

**完了条件**:
- [x] 全エンドポイント実装（settings / profile / courses / staffs / available-dates / available-times / reservations / my-reservations）
- [x] LIFF認証ミドルウェアで保護

---

## TASK-RES-022: 時間枠生成エンジン（★核心ロジック） ✅

**実装済みファイル**:
- `backend/internal/service/timeslot_engine.go`
- `backend/internal/service/timeslot_engine_test.go`

**完了条件**:
- [x] 基本: 営業時間内で枠生成
- [x] 休憩時間をまたぐ枠が除外される
- [x] 既存予約と重複する枠が除外される
- [x] 個人設定で営業時間が変更された場合
- [x] 個人設定で休日の場合（空リスト）
- [x] allow_gaps モード: 指定間隔で生成
- [x] minimize_gaps モード: 最短コース考慮
- [x] 指名なし: 全有効スタッフの空き時間を統合
- [x] 60分コース（手術）の枠生成
- [x] 15分コース（一般診察）の枠生成

---

## TASK-RES-023: 空き日付計算 ✅

**実装済みファイル**: `backend/internal/service/available_dates.go`

**完了条件**:
- [x] 予約受付期間外の日付が除外される
- [x] 休業日・祝日が除外される（祝日ライブラリで判定）
- [x] 曜日オプション（土曜限定等）が機能する
- [x] スタッフ個人休日が反映される

---

## TASK-RES-024: 予約制限チェック ✅

**実装済みファイル**: `backend/internal/service/reservation_validators.go`

**完了条件**:
- [x] 各制限でエラーメッセージが返る
- [x] SELECT FOR UPDATEによる楽観ロック
- [x] 409 Conflict時にフロントエンドが時間選択に戻せるレスポンス形式

---

## TASK-RES-025: 指名なし委譲ロジック ✅

**実装済みファイル**: `backend/internal/service/liff_service.go` 内

**完了条件**:
- [x] first_available: 空きスタッフに自動割当
- [x] top_priority: 最上位スタッフに固定割当
- [x] `is_staff_delegated = true` がセットされる
