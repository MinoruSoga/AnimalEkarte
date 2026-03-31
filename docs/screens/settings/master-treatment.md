# 診療項目マスタ設定 仕様書

## 概要
- **画面の目的**: カルテや会計で使用される診療項目（診察・検査・処置・予防・健診）の名称と単価を管理する。
- **URLパターン**: `/settings/treatment-items`
- **コンポーネント**: `[R] TreatmentPlanMaster`

## 画面構成とタブ
Radix UI の Tabs を用いて、5つのカテゴリに分けて管理します。
1. **診察** (`consultation`)
2. **検査** (`examination`)
3. **処置** (`procedure`)
4. **予防接種** (`vaccine`)
5. **定期健診** (`checkup`)

各タブは共通の `TreatmentTabContent` コンポーネントを使用し、同じレイアウト・機能を提供します。

## 機能詳細

### 1. 階層構造とDnD
- **親子関係**: 各項目は任意で親項目（`parentId`）を持つことができ、テーブル上で階層（ツリー）表示されます。
- **並び替え**: `DndContext` と `useSortableList` を用い、ドラッグ＆ドロップで表示順を変更可能（親カテゴリを跨いだ移動もサポート）。
- **折りたたみ**: 子を持つ親項目は、行左端の `Chevron` アイコンで子項目の表示/非表示を切り替えられます。

### 2. インライン編集とサイドパネル
- 行クリックまたは「新規追加」ボタンで、右側から `MasterSidePanel` がスライドインします。

## 表示・フォーム項目

### フォーム項目（サイドパネル）
| フィールド | 項目ID | 入力部品 | 必須 | 備考 |
|-----------|--------|---------|------|------|
| 名称 | `name` | `Input` | ✅ | タイトルエリア |
| ステータス | `isActive` | `StatusToggleButton` | - | |
| 税区分 | `taxType` | `TaxTypeSelector` | - | 課税/非課税/軽減税率 |
| 税率 | `taxRate` | `TaxRateSelector` | - | 税区分に応じて自動選択 |
| 単価(税込) | `price` | `MoneyInput` | - | |
| 備考 | `description`| `PropInput` | - | |

## API連携
タブごとに異なる CRUD エンドポイントが割り当てられています。
| メソッド | エンドポイント | 用途 |
|---------|--------------|------|
| GET | `/api/v1/masters/consultations` | 一覧取得 |
| POST | `/api/v1/masters/consultations` | 作成 |
| PATCH | `/api/v1/masters/consultations/:id` | 更新 |
| DELETE | `/api/v1/masters/consultations/:id` | 削除 |
| PATCH | `/api/v1/masters/consultations/reorder` | 並び順一括保存 |

※ 他のタブも同様に `/api/v1/masters/examination-types`, `/api/v1/masters/procedures`, `/api/v1/masters/vaccines`, `/api/v1/masters/checkup-types` を使用。
