# Phase 1: DB・モデル基盤（既存テーブル統合方式）

> **状態**: ✅ 全タスク完了（2026-04-08〜09）
>
> **設計方針**: 既存テーブル（staffs, service_types, reservation_appointments, shift_entries）を拡張し、
> 新規テーブルは4つのみ（reservation_settings, reservation_customers, staff_excluded_service_types, shift_entry_breaks）。

## TASK-RES-001: マイグレーションSQL作成 ✅

**実装済みファイル**: `backend/migrations/001_init.sql`（統合済み）

**完了条件**:
- [x] ALTER TABLE が既存データを壊さない（全カラムにデフォルト値あり）
- [x] 新規4テーブル作成確認
- [x] FK制約・UNIQUE制約が正しく機能
- [x] 既存の reservation_appointments, staffs, service_types のデータが保持される

---

## TASK-RES-002: Goモデル拡張・新規定義 ✅

**実装済みファイル**:
- `backend/internal/model/service_type.go`（拡張）
- `backend/internal/model/staff.go`（拡張）
- `backend/internal/model/reservation.go`（拡張）
- `backend/internal/model/reservation_setting.go`（新規）
- `backend/internal/model/reservation_customer.go`（新規）
- `backend/internal/model/staff_excluded_service_type.go`（新規）
- `backend/internal/model/shift_entry_break.go`（新規）

**完了条件**:
- [x] `make codegen` で `models.ts` に新フィールドが反映される
- [x] 既存のテスト・コードが壊れない（追加フィールドは全てデフォルト値あり）
- [x] 既存の reservation, staff, service_type のAPI応答に新フィールドが含まれる

---

## TASK-RES-003: シードデータ作成 ✅

**実装済みファイル**: `backend/migrations/002_seed.sql`（統合済み）

**完了条件**:
- [x] 既存データが正しく更新される
- [x] reservation_settings に基本設定1件が投入される
- [x] staff_excluded_service_types に非対応コース紐付けが投入される
