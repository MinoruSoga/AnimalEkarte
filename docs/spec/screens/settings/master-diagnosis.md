# 診断・傷病名マスタ 仕様書 (Diagnosis Management)

## 概要
- **画面の目的**: 臨床現場で使用される標準的な診断名、および疾患カテゴリの体系的管理。
- **URLパターン**: `/settings/diagnosis`
- **アクセス権限**: 診療マスタ管理権限が必要（`ResourceMasterMedical`）

---

## 画面構成

### 1. タブ構成
- **診断分類/診断病名タブ**: `診断カテゴリ（例：皮膚科）` タブと `診断病名（例：アトピー性皮膚炎）` タブを `UnifiedTabs` で切り替え（`?tab=diagnosis_type` / `?tab=diagnosis_name`）。
- **並び替え**: 各タブともドラッグ操作（`dnd-kit`）で表示順を変更可能。

### 2. 詳細編集サイドパネル (`SidePeekPanel`)
- **診断カテゴリ**: 名称、有効/無効ステータス、備考。
- **診断病名**: 名称、有効/無効ステータス、所属カテゴリ（必須選択）、備考。

---

## 主要な機能

### 1. カルテ入力の高速化
ここで定義されたマスタは、カルテ詳細画面の「診断」タブにおいて、サジェスト機能（Combobox）の基礎データとして使用されます。

### 2. 統計分析への活用
疾患カテゴリごとに来院件数を集計することで、病院が得意とする診療分野や季節性の疾患トレンドの分析が可能になります。

---

## 技術仕様

### 使用コンポーネント
- **`DiagnosisSettings`**: メインページ（`UnifiedTabs` でカテゴリ/病名タブを切り替え）。
- **`DiagnosisSortableTable`**: `dnd-kit` によるドラッグ並び替え一覧。
- **`DiagnosisTypeSidePanel` / `DiagnosisNameSidePanel`**: 詳細編集パネル。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/masters/diagnosis-types` | 診断カテゴリ一覧の取得 | `master-medical` | `view` |
| GET | `/api/v1/masters/diagnosis-types/:id` | 特定の診断カテゴリ情報の取得 | `master-medical` | `view` |
| POST | `/api/v1/masters/diagnosis-types` | 診断カテゴリの作成 | `master-medical` | `create` |
| PATCH | `/api/v1/masters/diagnosis-types/:id` | 診断カテゴリ情報の更新 | `master-medical` | `edit` |
| DELETE | `/api/v1/masters/diagnosis-types/:id` | 診断カテゴリの削除 | `master-medical` | `delete` |
| PATCH | `/api/v1/masters/diagnosis-types/reorder` | 診断カテゴリ表示順の一括保存 | `master-medical` | `edit` |
| GET | `/api/v1/masters/diagnosis-names` | 診断名一覧の取得 | `master-medical` | `view` |
| GET | `/api/v1/masters/diagnosis-names/all` | 全ての診断名一覧（未ページネーション）の取得 | `master-medical` | `view` |
| GET | `/api/v1/masters/diagnosis-names/:id` | 特定の診断名情報の取得 | `master-medical` | `view` |
| POST | `/api/v1/masters/diagnosis-names` | 診断名の作成 | `master-medical` | `create` |
| PATCH | `/api/v1/masters/diagnosis-names/:id` | 診断名情報の更新 | `master-medical` | `edit` |
| DELETE | `/api/v1/masters/diagnosis-names/:id` | 診断名の削除 | `master-medical` | `delete` |
| PATCH | `/api/v1/masters/diagnosis-names/reorder` | 診断名表示順の一括保存 | `master-medical` | `edit` |


---
