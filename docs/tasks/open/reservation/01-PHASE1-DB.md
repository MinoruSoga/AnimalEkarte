# Phase 1: DB・モデル基盤（既存テーブル統合方式）

> **設計方針**: 既存テーブル（staffs, service_types, reservation_appointments, shift_entries）を拡張し、
> 新規テーブルは4つのみ（reservation_settings, reservation_customers, staff_excluded_service_types, shift_entry_breaks）。

## TASK-RES-001: マイグレーションSQL作成

**概要**: 既存テーブルへのALTER TABLE + 新規4テーブル作成。

**対象ファイル**: `backend/migrations/003_line_reservation.sql`（新規）

**A. 既存テーブル拡張（ALTER TABLE）**:

| テーブル | 追加カラム | 型 | デフォルト |
|---|---|---|---|
| `service_types` | `duration_minutes` | INT NOT NULL | 15 |
| `service_types` | `short_name` | TEXT NOT NULL | '' |
| `service_types` | `show_short_name` | BOOLEAN NOT NULL | false |
| `service_types` | `reservation_visible` | BOOLEAN NOT NULL | true |
| `service_types` | `reservation_comment` | TEXT NOT NULL | '' |
| `service_types` | `reservation_image_url` | TEXT NOT NULL | '' |
| `service_types` | `reservation_day_option` | TEXT NOT NULL | 'none' |
| `service_types` | `is_internal` | BOOLEAN NOT NULL | false |
| `staffs` | `staff_type` | staff_type ENUM (doctor/nurse/resource) | 'doctor' |
| `staffs` | `reservation_visible` | BOOLEAN NOT NULL | true |
| `staffs` | `reservation_comment` | TEXT NOT NULL | '' |
| `staffs` | `reservation_image_url` | TEXT NOT NULL | '' |
| `reservation_appointments` | `source` | reservation_source ENUM (manual/line) | 'manual' |
| `reservation_appointments` | `line_customer_id` | BIGINT FK | NULL |
| `reservation_appointments` | `is_staff_delegated` | BOOLEAN NOT NULL | false |
| `reservation_appointments` | `customer_fields` | JSONB NOT NULL | '{}' |

**B. 新規テーブル（4つ）**:
1. `reservation_settings` — LINE予約設定（1 per clinic）
2. `reservation_customers` — LINE顧客（line_user_id管理）
3. `staff_excluded_service_types` — スタッフ非対応サービスM:N
4. `shift_entry_breaks` — シフト中断時間

**DDL詳細**: `docs/line-reseavation.md` セクション15.7

**完了条件**:
- [ ] ALTER TABLE が既存データを壊さない（全カラムにデフォルト値あり）
- [ ] 新規4テーブル作成確認
- [ ] FK制約・UNIQUE制約が正しく機能
- [ ] 既存の reservation_appointments, staffs, service_types のデータが保持される

---

## TASK-RES-002: Goモデル拡張・新規定義

**概要**: 既存モデルに予約用フィールドを追加 + 新規モデル4つを定義。

**A. 既存モデルの拡張**:

| ファイル | 追加フィールド |
|---|---|
| `model/service_type.go` | DurationMinutes, ShortName, ShowShortName, ReservationVisible, ReservationComment, ReservationImageURL, ReservationDayOption, IsInternal |
| `model/staff.go` | ReservationVisible, ReservationComment, ReservationImageURL |
| `model/reservation.go` | Source, LineCustomerID, IsStaffDelegated, CustomerFields |

**B. 新規モデル（4ファイル）**:
- `model/reservation_setting.go`（新規）
- `model/reservation_customer.go`（新規）
- `model/staff_excluded_service_type.go`（新規）
- `model/shift_entry_break.go`（新規）

**型の注意点**:
- `CustomerFields` (JSONB) → `datatypes.JSON` or カスタム型
- `Source` → string + const（`SourceLine = "line"`, `SourceManual = "manual"`）
- `ReservationDayOption` → string + const（`DayOptionNone`, `DayOptionSaturday`, etc.）
- shift_entry_breaks の `break_start/end` → `time.Time` 型（PostgreSQL TIME型に対応）

**完了条件**:
- [ ] `make codegen` で `models.ts` に新フィールドが反映される
- [ ] 既存のテスト・コードが壊れない（追加フィールドは全てデフォルト値あり）
- [ ] 既存の reservation, staff, service_type のAPI応答に新フィールドが含まれる

---

## TASK-RES-003: シードデータ作成

**概要**: 既存 service_types / staffs の予約用カラムに初期値を設定。

**対象ファイル**: `backend/migrations/003_line_reservation_seed.sql`（新規）

**データ内容**:

A. **service_types の予約用カラム更新**（既存レコードに対するUPDATE）:
```sql
-- 一般診察: 15分, 略称=診察, 予約表示=true
UPDATE service_types SET duration_minutes=15, short_name='診察',
  reservation_visible=true WHERE name='一般診察' AND clinic_id=?;

-- 休憩枠: 60分, is_internal=true, reservation_visible=false
UPDATE service_types SET duration_minutes=60, short_name='休憩枠',
  is_internal=true, reservation_visible=false WHERE name='休憩枠' AND clinic_id=?;
-- ... 全22コース分
```

B. **staffs の予約用カラム更新**:
```sql
-- 獣医師: reservation_visible=true
UPDATE staffs SET reservation_visible=true WHERE name='林 文明';
-- クイックシャンプー枠等のリソース: 既存staffsにない場合はINSERT
```

C. **staff_excluded_service_types の初期データINSERT**

D. **reservation_settings の初期データINSERT**（1件）

**注意**: 既存 service_types / staffs にデータがない場合（新規セットアップ時）は INSERT を使用。

**完了条件**:
- [ ] 既存データが正しく更新される
- [ ] reservation_settings に基本設定1件が投入される
- [ ] staff_excluded_service_types に非対応コース紐付けが投入される
