# Phase 4: 管理画面フロントエンド

> **状態**: ✅ 全タスク完了（2026-04-08〜09） → **Phase 8 で電カルマスタ設定に統合済み（2026-04-10）**
>
> **⚠️ 統合後の状態**: LINE予約管理の独立ページ5つ（カレンダー/コース/スタッフ/スケジュール/顧客管理）と
> 関連APIファイル14個は削除済み。機能は電カルのマスタ設定ページ（ServiceTypeSettings, StaffSettings, ShiftFormDialog, ReservationManagement, OwnerForm）に統合。
> LINE予約専用ページは「基本設定」「ページ編集」の2ページのみ残存。
> 詳細は [architecture.md](../../../line/architecture.md) セクション7を参照。

## TASK-RES-030: Feature scaffolding ✅

**実装済みファイル**:
```
frontend/src/features/line-reservation/
├── index.ts
├── api/
│   ├── get-reservation-settings.ts
│   ├── update-reservation-settings.ts
│   ├── get-reservation-courses.ts
│   ├── create-reservation-course.ts
│   ├── update-reservation-course.ts
│   ├── delete-reservation-course.ts
│   ├── get-reservation-staffs.ts
│   ├── create-reservation-staff.ts
│   ├── update-reservation-staff.ts
│   ├── delete-reservation-staff.ts
│   ├── get-staff-schedules.ts
│   ├── update-staff-schedule.ts
│   ├── delete-staff-schedule.ts
│   ├── get-reservation-admin.ts
│   ├── get-reservation-customers.ts
│   ├── update-owner-link.ts
│   └── types.ts
└── routes/
    ├── LineReservationSettings.tsx
    ├── LineReservationCourses.tsx
    ├── LineReservationStaffs.tsx
    ├── LineStaffSchedule.tsx
    ├── LineReservationCalendar.tsx
    ├── LineReservationPageEditor.tsx
    └── LineReservationCustomers.tsx
```

**完了条件**:
- [x] ルーティング登録（`frontend/src/app/router.tsx`）
- [x] サイドメニュー表示（「LINE予約管理」セクション）
- [x] RBAC: `reservation:read` / `reservation:write`

---

## TASK-RES-031: コース設定画面 ✅

**実装済みファイル**: `frontend/src/features/line-reservation/routes/LineReservationCourses.tsx`

**完了条件**:
- [x] CRUD全操作
- [x] `is_internal` フラグの表示・設定
- [x] 並び順変更

---

## TASK-RES-032: スタッフ設定画面 ✅

**実装済みファイル**: `frontend/src/features/line-reservation/routes/LineReservationStaffs.tsx`

**完了条件**:
- [x] CRUD全操作
- [x] `staff_type` (doctor/nurse/resource) 選択
- [x] `facility_name` 入力
- [x] 非対応コースのマルチセレクト

---

## TASK-RES-033: 基本設定画面 ✅

**実装済みファイル**: `frontend/src/features/line-reservation/routes/LineReservationSettings.tsx`

**完了条件**:
- [x] 全設定項目の表示・編集・保存
- [x] 追加入力フィールドの動的追加・削除・並び替え

---

## TASK-RES-034: 個人設定画面 ✅

**実装済みファイル**: `frontend/src/features/line-reservation/routes/LineStaffSchedule.tsx`

**完了条件**:
- [x] ガントチャート風表示（基本設定 + 個人上書きマージ）
- [x] 個人設定の作成・更新・削除
- [x] 中断時間の複数入力

---

## TASK-RES-035: 予約状況画面（★最も複雑な管理画面） ✅

**実装済みファイル**: `frontend/src/features/line-reservation/routes/LineReservationCalendar.tsx`

**完了条件**:
- [x] 月表示（予約サマリ付き）
- [x] 日表示（30分グリッド × スタッフ列）
- [x] 月→日遷移
- [x] 予約キャンセル
- [x] 手動予約入力（バリデーションなし）
- [x] 印刷対応

---

## TASK-RES-036: ページ編集画面 ✅

**実装済みファイル**: `frontend/src/features/line-reservation/routes/LineReservationPageEditor.tsx`

**完了条件**:
- [x] 全フィールドの編集・保存

---

## TASK-RES-037: LINE顧客管理画面（追加実装） ✅

> 仕様書に明示されていなかったが、TASK-RES-015（顧客管理API）のフロントエンドとして実装。

**実装済みファイル**:
- `frontend/src/features/line-reservation/routes/LineReservationCustomers.tsx`
- `frontend/src/app/pages/LineReservationCustomersPage.tsx`（cross-feature合成）

**機能**:
- LINE顧客一覧（名前フィルタ付き）
- オーナー紐付け・解除ダイアログ
- `owners` feature の `useGetOwners` を `app/pages/` 経由で dependency inversion

**完了条件**:
- [x] 顧客一覧表示・フィルタ
- [x] オーナー紐付け・解除
