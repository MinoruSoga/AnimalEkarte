# 検査項目定義マスタ 仕様書 (Examination Definitions)

## 概要
- **画面の目的**: 血液検査や生化学検査等、検査プラン（`exam_types`）の名称・価格の定義、および測定項目（`exam_type_fields`）と基準値の管理。
- **URLパターン**: `/settings/treatment-items?tab=examination`
- **アクセス権限**: 診療マスタ管理権限が必要（`ResourceMasterMedical`）
- **注意**: 検査プラン本体は [master-treatment.md](./master-treatment.md) と同一の `TreatmentPlanMaster` / サイドパネル UI を共有する。測定項目・単位・基準値はサイドパネル内の **`ExamTypeFieldsEditor`** で編集する（「管理 UI が存在しない」は誤り）。

---

## 画面構成

### 1. 検査項目一覧
- **タブ構造**: 診察、検査、処置、予防接種、定期健診のタブの一つとして表示。
- **項目リスト**: 項目名、単価、ステータス（親子ツリー表示）。

### 2. 詳細編集サイドパネル
- **プラン本体**: 項目名、有効/無効、親カテゴリ、備考、単価、保険対象外等。
  - `ExaminationType` に `tax_type`/`tax_rate` 列は無く、課税区分 UI が出ても保存されない場合がある（master-treatment と同趣旨）。
- **測定項目エディタ (`ExamTypeFieldsEditor`)**: 検査プラン選択時にサイドパネル詳細へ表示。
  - フィールド名、単位（`unit`）
  - 基準値（reference ranges: Min/Max 等）
  - 追加・更新・削除・DnD 並び替え

---

## 主要な機能

### 1. 検査結果入力との関係
カルテの検査結果入力（`/examinations`）は、本画面で定義した `exam_type_fields` / 基準値を参照して HIGH/LOW ハイライト等を行う。

### 2. 表示順
プラン一覧の reorder API に加え、フィールド単位の reorder も `ExamTypeFieldsEditor` から実行できる。

---

## 技術仕様

### 使用コンポーネント
- **`TreatmentPlanMaster`**: メインページ（検査タブ）。
- **`TreatmentPlanSidePanelHost`**: サイドパネル host。`examinationType` があるとき `ExamTypeFieldsEditor` をマウント。
- **`ExamTypeFieldsEditor`**: 測定項目・単位・基準値・並び替え。
- **`TreatmentItemSidePanel` / 共通 property UI**: プラン本体フィールド。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET/POST/PATCH/DELETE | `/api/v1/masters/examination-types` 系 | 検査プラン CRUD | `master-medical` | view/create/edit/delete |
| POST | `.../examination-types/:id/fields` | 測定項目作成 | `master-medical` | create |
| PATCH | `.../fields/:id` | 測定項目更新 | `master-medical` | edit |
| DELETE | `.../fields/:id` | 測定項目削除 | `master-medical` | delete |
| PATCH | `.../fields/reorder` | 測定項目並び替え | `master-medical` | edit |
| PUT | `.../fields/:id/reference-ranges` | 基準値更新 | `master-medical` | edit |

（正確な path 接頭辞は `frontend/src/features/master/api/exam-types-master.ts` を正とする。）
