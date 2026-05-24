# TASK-002: 統一予約基盤 — トリミング予約を appointments に統合

**作成日**: 2026-04-16
**ステータス**: Closed（2026-05-19 実装済み確認）
**依頼元**: アーキテクチャ改善提案（設計議論より）

---

## 概要

現状、トリミング記録は `trimming_records` テーブルで独立管理されており、
LINE予約の `appointments` テーブルと完全に分離している。
この分離により「LINE経由でトリミング予約を受け付ける」ことが構造上不可能になっている。

本タスクでは `trimming_records` を廃止し、すべての予約を `appointments` で一元管理する
**統一予約基盤** を構築する。トリミング固有の臨床情報は 1:1 拡張テーブル
`appointment_trimming_details` に分離して格納する。

## 背景（アーキテクチャ議論の要約）

- `trimming_records.date DATE` — 日付のみでタイムスロットを持たない
- `appointments` には `start_time / end_time TIMESTAMPTZ` があり、時刻管理が可能
- LIFF予約フローは `reservation_types` → スタッフ → 日時選択で完結している
- トリミングを `reservation_types.category = 'trimming'` で識別すれば
  LIFF が追加ステップ（コース・オプション選択）を表示できる
- `trimming_courses` / `trimming_options` マスタテーブルは継続活用する

## サブタスク分解

| # | サブタスク | 領域 | イシュー | 依存 | 完了 |
|---|----------|------|---------|------|------|
| 1 | DBスキーマ変更 + Go モデル + make codegen | DB/BE | BE-118 | - | [x] |
| 2 | カルテ側トリミング管理API変更 | BE | BE-119 | #1 | [x] |
| 3 | LIFF トリミング予約フロー拡張 | BE | BE-120 | #1 | [x] |
| 4 | カルテ側トリミング機能フロントエンド対応 | FE | FE-253 | #2 | [x] |
| 5 | LIFF トリミング予約UI拡張 | FE | FE-254 | #3 | [x] |

## 受入条件（Acceptance Criteria）

- [x] AC-1: `reservation_types` にカテゴリ `general` / `trimming` を設定できる
- [x] AC-2: `category = 'trimming'` の予約区分はカルテのトリミングリストに表示される
- [x] AC-3: カルテ側トリミングフォームでコース・オプション・臨床情報（体重・体温等）を記録できる
- [x] AC-4: LINE予約でトリミング区分を選択すると「コース選択」「オプション選択」ステップが追加表示される
- [x] AC-5: LINE経由で作成されたトリミング予約がカルテのトリミングリストに表示される
- [x] AC-6: `trimming_records` テーブルが廃止されており、新規レコードは `appointments` に作成される
- [x] AC-7: 職種ガード・予約不可時間（TASK-001）はトリミング区分にも適用される

## 技術的判断

| 判断事項 | 採用案 | 理由 | 却下案 |
|---------|--------|------|--------|
| トリミング識別方法 | `reservation_types.category ENUM` | 既存の予約区分マスタを拡張するだけで済む。FK不要 | 専用テーブル `trimming_reservation_types` |
| 臨床情報の格納 | 1:1拡張テーブル `appointment_trimming_details` | `appointments` を汚染せず、NULL列を増やさない。ON DELETE CASCADE で連動削除 | `appointments` にカラム追加 |
| オプション格納 | 中間テーブル `appointment_trimming_options` | M:N。既存 `trimming_record_options` と同パターン | JSONB配列 |
| `trimming_records` 廃止方式 | `001_init.sql` から削除（DB リセット運用） | プロジェクトはリリース前のDBリセットポリシー。incremental migrationは不要 | ALTER TABLE で段階移行 |
| `trimming_status` ENUM | 廃止。`reservation_status` を流用 | `appointments.status` に統一。フロント表示名はUIレイヤーで制御 | 新ENUM追加 |

## 影響範囲

### DB（001_init.sql）
- **廃止**: `CREATE TYPE trimming_status AS ENUM (...)` (line 70)
- **廃止**: `CREATE TABLE trimming_records (...)` (lines 718-740)
- **廃止**: `CREATE TABLE trimming_record_options (...)` (lines 1145-1154)
- **廃止**: 関連インデックス（`idx_trimming_records_*`, `idx_trimming_record_options_unique`）
- **追加**: `CREATE TYPE reservation_type_category AS ENUM ('general', 'trimming')`
- **追加**: `reservation_types.category reservation_type_category NOT NULL DEFAULT 'general'`
- **追加**: `CREATE TABLE appointment_trimming_details (...)`
- **追加**: `CREATE TABLE appointment_trimming_options (...)`

### Backend
- `backend/migrations/001_init.sql` — 上記DB変更
- `backend/internal/model/trimming.go` — `TrimmingRecord`, `TrimmingRecordOption`, `TrimmingStatus` 削除。`AppointmentTrimmingDetail`, `AppointmentTrimmingOption` 追加
- `backend/internal/model/reservation_type.go` — `ReservationTypeCategory` 型・定数追加、`Category` フィールド追加
- `backend/internal/model/reservation.go` — `TrimmingDetail *AppointmentTrimmingDetail` リレーション追加
- `backend/internal/handler/trimming_handler.go` — appointments ベース API に全面書き換え
- `backend/internal/service/trimming_service.go` — appointments ベースに全面書き換え
- `backend/internal/repository/trimming_repository.go` — 廃止。`appointment_trimming_detail_repository.go` 新規作成
- `backend/internal/service/liff_service.go` — `GetTrimmingCourses`, `GetTrimmingOptions` 追加。`CreateReservation` でトリミング詳細作成
- `backend/internal/handler/reservation_line_handler.go` — LIFF新エンドポイント追加

### Frontend
- `frontend/src/features/trimming/api/*.ts` — 新 appointments API に対応
- `frontend/src/features/trimming/routes/TrimmingList.tsx` — 表示データ構造変更対応
- `frontend/src/features/trimming/routes/TrimmingForm.tsx` — フィールド構造変更対応
- `frontend/src/types/generated/models.ts` — `make codegen` で自動更新
- LIFF側: 新コース・オプション選択UIの追加（FE-254）

## 参照実装

- `backend/internal/handler/trimming_handler.go` — 既存のトリミングCRUD（移行前）
- `backend/internal/service/liff_service.go:CreateReservation()` — 既存の予約作成フロー
- `backend/internal/repository/trimming_repository.go:SetOptions()` — オプションM:N管理パターン
- `features/owners/` — React 19 Action パターン（useActionState + SubmitButton）

## リスク・懸念事項

| リスク | 影響度 | 対策 |
|--------|--------|------|
| TASK-001(BE-115)との`001_init.sql`競合 | 高 | BE-118はBE-115完了後に着手。または同一コミットで両方のスキーマ変更をまとめる |
| トリミングリストの日付フィルタ変更 | 中 | 既存は`date DATE`、新は`start_time TIMESTAMPTZ`。リポジトリの日付クエリをJST変換に変更 |
| `trimming_status`廃止による管理画面影響 | 中 | フロント側でステータス表示名を`reservation_status`値に対してマッピング変更 |
| フロントendのtrimming feature全面書き換え | 中 | BE-119完了後にFE-253で対応。API契約をBE-119で確定してから着手 |

## 未解決事項

- なし（設計議論ですべて解決済み）

## 実装順序

1. BE-118: DBスキーマ変更 + Go モデル + make codegen（TASK-001 BE-115完了後）
2. BE-119: カルテ側トリミング管理API変更（BE-118完了後）
3. BE-120: LIFF トリミング予約フロー拡張（BE-118完了後、BE-119と並行可）
4. FE-253: カルテ側トリミング機能フロントエンド対応（BE-119完了後）
5. FE-254: LIFF トリミング予約UI拡張（BE-120完了後）

## 関連イシュー

- [BE-118](../../backend/issues/open/BE-118-unified-reservation-schema-and-models.md)
- [BE-119](../../backend/issues/open/BE-119-trimming-management-api-refactoring.md)
- [BE-120](../../backend/issues/open/BE-120-liff-trimming-reservation-flow.md)
- [FE-253](../../frontend/issues/open/FE-253-trimming-feature-api-migration.md)
- [FE-254](../../frontend/issues/open/FE-254-liff-trimming-reservation-ui.md)
- [TASK-001](TASK-001-reservation-type-unavailable-times-and-occupation-guard.md) — 依存元（予約区分マスタ拡張）
