# TASK-010: マスタテーブルの論理削除方針を統一（deleted_at vs is_active）

## 概要

マスタテーブルの論理削除実装が `is_active` フラグと `deleted_at` カラムに二分されており、一貫性がない。`checkup_types` テーブルのみが `deleted_at` を持ち、他のマスタ（exam_types, reservation_types, vaccines 等）は `is_active` で運用している。

## 優先度

MEDIUM（設計方針の決定が必要）

## 現状

| 方針 | 適用テーブル |
|------|------------|
| `deleted_at` + GORM soft delete | `checkup_types`（のみ） |
| `is_active = false` で無効化 | `exam_types`, `reservation_types`, `vaccines`, `medicines`, `trimming_courses`, `trimming_options`, `chief_complaint_types`, `diagnosis_types` 等 |
| 物理削除のみ | `shift_entries`, `clinic_holidays`, `payments`, `billing_refunds` 等 |

## 規約違反

`.claude/rules/database-design.md`:
> 全テーブルに `deleted_at` が望ましい（ただし append-only の監査ログ等は除く）

## 意思決定事項

以下の2択からプロジェクト方針を決定すること:

### 選択肢 A: 全マスタを `deleted_at` に統一（推奨）
- GORM の soft delete が自動的に `WHERE deleted_at IS NULL` を付与する
- `checkup_types` が正しい実装例となる
- 移行コスト: 全マスタテーブルに `deleted_at` カラムを追加、`is_active` は残す（既存の `is_active` 用途は「一時的に非表示」として継続）

### 選択肢 B: 全マスタを `is_active` に統一
- `checkup_types` から `deleted_at` を削除
- 「削除」は `is_active = false` に統一（物理削除はしない）
- FK 依存チェックは `CountUsage` で対応（現行通り）

## 推奨

選択肢 A（`deleted_at` 統一）を推奨。GORM のソフトデリートスコープが自動的に適用されるため、JOIN クエリでの `deleted_at IS NULL` 漏れリスクが減る。

## 影響ファイル

- `backend/migrations/001_init.sql`（DDL に `deleted_at TIMESTAMPTZ` 追加）
- 対象モデル: `exam_types`, `reservation_types`, `vaccines`, `medicines`, `trimming_courses`, `trimming_options`, `chief_complaint_types`, `diagnosis_types`, `diagnosis_names`, `occupations`, `cages`, `insurances`, `procedures`, `consultations`, `inquiry_templates`
