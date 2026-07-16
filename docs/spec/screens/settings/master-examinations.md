# 検査項目定義マスタ 仕様書 (Examination Definitions)

## 概要
- **画面の目的**: 血液検査や生化学検査等、検査プラン（`exam_types`）の名称・価格・課税区分の定義。
- **URLパターン**: `/settings/treatment-items?tab=examination`
- **アクセス権限**: 診療マスタ管理権限が必要（`ResourceMasterMedical`）
- **注意**: 本画面が管理するのは検査プラン（`ExaminationType`）そのものであり、[master-treatment.md](./master-treatment.md) と同一の `TreatmentItemSidePanel` を共有する。個々の測定項目（GOT・CRE等）とその単位・基準値（`exam_type_fields` テーブル: `unit`, `normal_value`, `inspection_value`）を編集する管理画面は存在しない（backend/internal/model/examination_type.go:29-39）。

---

## 画面構成

### 1. 検査項目一覧
- **タブ構造**: 診察、検査、処置、予防、定期健診のタブの一つとして表示（[master-treatment.md](./master-treatment.md) 1.1参照）。
- **項目リスト**: 項目名、ステータス（親子ツリー表示、検索は項目名の部分一致）。

### 2. 詳細編集サイドパネル (`SidePeekPanel`)
[master-treatment.md](./master-treatment.md) 2. と同じ `TreatmentItemSidePanel` を共有するため、画面上は項目名、有効/無効ステータス、親カテゴリ、備考、単価、課税区分、税率、保険対象外が一律表示される。ただし `ExaminationType` モデル・API（`backend/internal/model/examination_type.go`、`backend/internal/handler/exam_type_request.go`）には `tax_type`/`tax_rate` 列が存在せず、課税区分・税率は画面上操作できても保存されない。実際に保存されるのは項目名、有効/無効ステータス、親カテゴリ、備考、単価、保険対象外（`is_non_insurance`）のみ。単位・基準値（Min/Max）を設定する項目は存在しない。

---

## 主要な機能

### 1. 検査結果入力との関係
カルテの検査結果入力画面（`/examinations`）が参照する測定項目・基準値（`exam_type_fields.normal_value` 等）は本画面では編集できない。異常値ハイライトのロジックは `frontend/src/features/examinations/` 側の実装を参照。

### 2. 表示順の柔軟なカスタマイズ
ドラッグ&ドロップ（`dnd-kit`、`reorder` API）により項目の表示順を変更できる。

---

## 技術仕様

### 使用コンポーネント
- **`TreatmentPlanMaster`**: メインページ（検査タブ）。[master-treatment.md](./master-treatment.md) と共通。
- **`TreatmentItemSidePanel`**: 詳細編集パネル。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/masters/examination-types` | 定義済み項目の一覧取得 | `master-medical` | `view` |
| GET | `/api/v1/masters/examination-types/:id` | 特定の検査項目情報の取得 | `master-medical` | `view` |
| POST | `/api/v1/masters/examination-types` | 新規検査項目の登録 | `master-medical` | `create` |
| PATCH | `/api/v1/masters/examination-types/:id` | 名称・単価・課税区分等の属性更新 | `master-medical` | `edit` |
| DELETE | `/api/v1/masters/examination-types/:id` | 検査項目の削除 | `master-medical` | `delete` |
| PATCH | `/api/v1/masters/examination-types/reorder` | 並び順の一括保存 | `master-medical` | `edit` |

---


