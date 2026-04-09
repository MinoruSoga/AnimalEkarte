# Phase 2: バックエンド管理者API

> **状態**: ✅ 全タスク完了（2026-04-08〜09） → **Phase 8 でマスタAPIに統合済み（2026-04-10）**
>
> **設計方針**: 既存テーブル（staffs, service_types, reservation_appointments, shift_entries）を
> 予約APIから直接操作する。既存のrepository/serviceを拡張 or 予約専用のhandler/serviceを新設。
> 既存APIとの整合性を保つため、既存のエラーハンドリング（RespondError + apperrors）に準拠する。
>
> **⚠️ Phase 8 統合**: コース（service_types）・スタッフ（staffs）のLINEフィールドは
> 既存マスタAPI（`/v1/masters/service-types`, `/v1/masters/staffs`）に統合。
> 本Phaseで作成した `/v1/clinics/:id/reservation-courses` 等の専用APIは互換性のため残存するが、
> 管理画面からは使用しない。シフト休憩時間は `/v1/shifts` APIに統合。
> 予約の source フィルタは `/v1/reservations?source=` で対応。

## TASK-RES-010: 基本設定 API ✅

**実装済みファイル**:
- `backend/internal/handler/reservation_setting_handler.go`
- `backend/internal/handler/reservation_setting_request.go`
- `backend/internal/handler/reservation_setting_response.go`
- `backend/internal/service/reservation_setting_service.go`
- `backend/internal/repository/reservation_setting_repository.go`

**完了条件**:
- [x] GET: 設定取得（未作成時は空レスポンス）
- [x] PUT: 全フィールド更新
- [x] バリデーションエラー時 400

---

## TASK-RES-011: コース CRUD API ✅

**実装済みファイル**:
- `backend/internal/handler/reservation_course_handler.go`
- `backend/internal/handler/reservation_course_request.go`
- `backend/internal/handler/reservation_course_response.go`
- `backend/internal/service/reservation_course_service.go`
- `backend/internal/repository/reservation_course_repository.go`

**完了条件**:
- [x] 一覧取得（sort_order順）
- [x] 新規作成（必須項目バリデーション）
- [x] 更新
- [x] 削除（使用中チェック付き）
- [x] 有効/休止トグル
- [x] 並び順上下入れ替え
- [x] 画像アップロード（v1: 501 Not Implemented）

---

## TASK-RES-012: スタッフ CRUD API ✅

**実装済みファイル**:
- `backend/internal/handler/reservation_staff_handler.go`
- `backend/internal/handler/reservation_staff_request.go`
- `backend/internal/handler/reservation_staff_response.go`
- `backend/internal/service/reservation_staff_service.go`
- `backend/internal/repository/reservation_staff_repository.go`

**完了条件**:
- [x] 一覧取得（非対応コース含む）
- [x] 新規作成（非対応コース同時設定）
- [x] 更新（非対応コース差分更新）
- [x] 削除（使用中チェック付き）
- [x] 有効/休止トグル
- [x] 並び順上下入れ替え
- [x] 画像アップロード（v1: 501 Not Implemented）

---

## TASK-RES-013: スタッフ個人設定 API ✅

**実装済みファイル**:
- `backend/internal/handler/reservation_schedule_handler.go`
- `backend/internal/handler/reservation_schedule_request.go`
- `backend/internal/handler/reservation_schedule_response.go`
- `backend/internal/service/reservation_schedule_service.go`
- `backend/internal/repository/reservation_schedule_repository.go`

**補足**: レスポンスのフィールド名を `work_start`/`work_end` に統一済み（BUG修正）

**完了条件**:
- [x] 月単位のスケジュール取得
- [x] 日単位の個人設定作成/更新
- [x] 個人設定の削除（基本設定に戻る）
- [x] 中断時間の複数登録

---

## TASK-RES-014: 予約管理 API（管理者） ✅

**実装済みファイル**:
- `backend/internal/handler/reservation_admin_handler.go`
- `backend/internal/handler/reservation_admin_request.go`
- `backend/internal/handler/reservation_admin_response.go`
- `backend/internal/service/reservation_admin_service.go`
- `backend/internal/repository/reservation_admin_repository.go`
- `backend/internal/handler/reservation_handler.go`（既存拡張）

**完了条件**:
- [x] 月表示取得
- [x] 日表示取得
- [x] 手動予約入力
- [x] 予約キャンセル

---

## TASK-RES-015: 顧客管理 API ✅

**実装済みファイル**:
- `backend/internal/handler/reservation_customer_handler.go`
- `backend/internal/handler/reservation_customer_request.go`
- `backend/internal/handler/reservation_customer_response.go`
- `backend/internal/service/reservation_customer_service.go`
- `backend/internal/repository/reservation_customer_repository.go`

**完了条件**:
- [x] 顧客一覧取得（紐付け状況含む）
- [x] オーナー紐付け

---

## TASK-RES-016: ルーティング登録・DI配線 ✅

**実装済みファイル**:
- `backend/internal/handler/reservation_line_routes.go`（新規）
- `backend/cmd/api/main.go`（既存に追記）
- `backend/internal/service/service.go`（DI配線追記）

**完了条件**:
- [x] 全エンドポイントがルーティングに登録
- [x] 既存の認証・RBACミドルウェアで保護
