# Phase 2: バックエンド管理者API

> **設計方針**: 既存テーブル（staffs, service_types, reservation_appointments, shift_entries）を
> 予約APIから直接操作する。既存のrepository/serviceを拡張 or 予約専用のhandler/serviceを新設。
> 既存APIとの整合性を保つため、既存のエラーハンドリング（RespondError + apperrors）に準拠する。
>
> **テーブル名の対応**:
> - コース → `service_types`（既存）
> - スタッフ → `staffs`（既存）
> - 個人設定 → `shift_entries`（既存）+ `shift_entry_breaks`（新規）
> - 予約 → `reservation_appointments`（既存）
> - 非対応コース → `staff_excluded_service_types`（新規）

## TASK-RES-010: 基本設定 API

**エンドポイント**:
```
GET  /api/clinics/:clinicId/reservation-settings
PUT  /api/clinics/:clinicId/reservation-settings
```

**対象ファイル（すべて新規）**:
- `backend/internal/handler/reservation_setting_handler.go`
- `backend/internal/handler/reservation_setting_request.go`
- `backend/internal/handler/reservation_setting_response.go`
- `backend/internal/service/reservation_setting_service.go`
- `backend/internal/repository/reservation_setting_repository.go`

**バリデーション**:
- `business_hours.start` / `end`: HHMM形式、start < end
- `booking_window_min_days` < `booking_window_max_days`
- `daily_limit` >= 1（設定する場合）
- `time_slot_interval_minutes`: 15, 30, 60 のいずれか

**完了条件**:
- [ ] GET: 設定取得（未作成時は空レスポンス）
- [ ] PUT: 全フィールド更新
- [ ] バリデーションエラー時 400

---

## TASK-RES-011: コース CRUD API

**エンドポイント**:
```
GET    /api/clinics/:clinicId/reservation-courses
POST   /api/clinics/:clinicId/reservation-courses
PUT    /api/clinics/:clinicId/reservation-courses/:id
DELETE /api/clinics/:clinicId/reservation-courses/:id
PATCH  /api/clinics/:clinicId/reservation-courses/:id/status
PATCH  /api/clinics/:clinicId/reservation-courses/:id/sort-order
POST   /api/clinics/:clinicId/reservation-courses/:id/image
```

**対象ファイル（すべて新規）**:
- `backend/internal/handler/reservation_course_handler.go`
- `backend/internal/handler/reservation_course_request.go`
- `backend/internal/handler/reservation_course_response.go`
- `backend/internal/service/reservation_course_service.go`
- `backend/internal/repository/reservation_course_repository.go`

**注意点**:
- DELETE前に `reservations` で使用中かチェック → 使用中なら409
- `sort-order` は `{ "direction": "up" | "down" }` で隣接レコードとswap
- `is_internal` フラグの設定を POST/PUT で受け付ける

**完了条件**:
- [ ] 一覧取得（sort_order順）
- [ ] 新規作成（必須項目バリデーション）
- [ ] 更新
- [ ] 削除（使用中チェック付き）
- [ ] 有効/休止トグル
- [ ] 並び順上下入れ替え
- [ ] 画像アップロード（※v2スコープ。エンドポイントのみ定義し、v1では501 Not Implementedを返す）

---

## TASK-RES-012: スタッフ CRUD API

**エンドポイント**:
```
GET    /api/clinics/:clinicId/reservation-staffs
POST   /api/clinics/:clinicId/reservation-staffs
PUT    /api/clinics/:clinicId/reservation-staffs/:id
DELETE /api/clinics/:clinicId/reservation-staffs/:id
PATCH  /api/clinics/:clinicId/reservation-staffs/:id/status
PATCH  /api/clinics/:clinicId/reservation-staffs/:id/sort-order
POST   /api/clinics/:clinicId/reservation-staffs/:id/image
```

**対象ファイル（すべて新規）**:
- `backend/internal/handler/reservation_staff_handler.go`
- `backend/internal/handler/reservation_staff_request.go`
- `backend/internal/handler/reservation_staff_response.go`
- `backend/internal/service/reservation_staff_service.go`
- `backend/internal/repository/reservation_staff_repository.go`

**注意点**:
- POST/PUT で `excluded_course_ids: []uint64` を受け取り、`reservation_staff_excluded_courses` を差分更新
- `staff_type` (doctor/nurse/resource) を POST で設定
- `facility_name` は自由テキスト（「城東センター病院」等）
- GET レスポンスには `excluded_courses` をネスト

**完了条件**:
- [ ] 一覧取得（非対応コース含む）
- [ ] 新規作成（非対応コース同時設定）
- [ ] 更新（非対応コース差分更新）
- [ ] 削除（使用中チェック付き）
- [ ] 有効/休止トグル
- [ ] 並び順上下入れ替え
- [ ] 画像アップロード（※v2スコープ。エンドポイントのみ定義し、v1では501 Not Implementedを返す）

---

## TASK-RES-013: スタッフ個人設定 API

**エンドポイント**:
```
GET    /api/clinics/:clinicId/reservation-staffs/:staffId/schedules?month=2026-04
PUT    /api/clinics/:clinicId/reservation-staffs/:staffId/schedules/:date
DELETE /api/clinics/:clinicId/reservation-staffs/:staffId/schedules/:date
```

**対象ファイル（すべて新規）**:
- `backend/internal/handler/reservation_schedule_handler.go`
- `backend/internal/handler/reservation_schedule_request.go`
- `backend/internal/handler/reservation_schedule_response.go`
- `backend/internal/service/reservation_schedule_service.go`
- `backend/internal/repository/reservation_schedule_repository.go`

**PUT リクエスト例**:
```json
{
  "is_day_off": false,
  "work_start": "1100",
  "work_end": "1700",
  "breaks": [
    { "start": "1300", "end": "1330" },
    { "start": "1500", "end": "1530" }
  ]
}
```

**GET レスポンス**: 指定月の全日分（基本設定 + 個人上書きのマージ）。

**完了条件**:
- [ ] 月単位のスケジュール取得
- [ ] 日単位の個人設定作成/更新
- [ ] 個人設定の削除（基本設定に戻る）
- [ ] 中断時間の複数登録

---

## TASK-RES-014: 予約管理 API（管理者）

**エンドポイント**:
```
GET    /api/clinics/:clinicId/reservations?view=month&date=2026-04
GET    /api/clinics/:clinicId/reservations?view=day&date=2026-04-08
POST   /api/clinics/:clinicId/reservations
DELETE /api/clinics/:clinicId/reservations/:id
```

**対象ファイル（すべて新規）**:
- `backend/internal/handler/reservation_handler.go`
- `backend/internal/handler/reservation_request.go`
- `backend/internal/handler/reservation_response.go`
- `backend/internal/service/reservation_service.go`
- `backend/internal/repository/reservation_repository.go`

**月表示レスポンス**: 各日の予約サマリ（時間・顧客名・コース略称・スタッフ名・委譲フラグ・ソース）
**日表示レスポンス**: スタッフ列ごとの予約詳細（顧客フィールド含む）
**手動入力POST**: バリデーションなし（管理者責任）
**DELETE**: 論理削除（復元不可）

**完了条件**:
- [ ] 月表示取得
- [ ] 日表示取得
- [ ] 手動予約入力
- [ ] 予約キャンセル

---

## TASK-RES-015: 顧客管理 API

**エンドポイント**:
```
GET    /api/clinics/:clinicId/reservation-customers
PATCH  /api/clinics/:clinicId/reservation-customers/:id/link-owner
```

**完了条件**:
- [ ] 顧客一覧取得（紐付け状況含む）
- [ ] オーナー紐付け

---

## TASK-RES-016: ルーティング登録・DI配線

**対象ファイル**: `backend/cmd/api/main.go`（既存に追記）

**完了条件**:
- [ ] 全エンドポイントがルーティングに登録
- [ ] 既存の認証・RBACミドルウェアで保護
