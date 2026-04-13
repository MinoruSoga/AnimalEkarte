# 薬剤マスタ設定 仕様書

## 概要
- **画面の目的**: 院内で使用する薬剤（内服・外用・注射等）の名称、価格、剤形、在庫連携情報を一元管理する。
- **URLパターン**: `/settings/medicine`
- **コンポーネント**: `[R] MedicineSettings`
- **特徴**: カテゴリと薬剤の2階層構造。ドラッグ&ドロップによる並び替えとカテゴリ移動が可能。

## 画面構成
- **メインエリア**: `DataTable`（独自拡張）による階層表示。
  - **カテゴリ行**: 背景色 `bgPage30`。名称、アイテム数、Chevron（展開/折りたたみ）、インライン追加ボタンを表示。
  - **薬剤行**: カテゴリ配下またはルートに表示。名称、剤形、単価、ステータスを表示。
- **サイドパネル**: `MedicineSidePanel`（`SidePeekPanel` 拡張）による詳細編集。

## 表示・フォーム項目

### サイドパネル項目
| フィールド | 項目ID | 入力部品 | 必須 | 備考 |
|-----------|--------|---------|------|------|
| 薬品名 | `name` | `Input` | ✅ | タイトルエリア |
| 親カテゴリ | `parentId` | `Select` | - | 既存カテゴリから選択。カテゴリ自身は変更不可 |
| 単価(税込) | `price` | `input(number)` | - | カテゴリの場合は入力不可（0固定） |
| 課税区分 | `taxType` | `TaxTypeSelector` | - | `excluded` (外税) / `included` (内税) / `non_taxable` (非課税) |
| 税率 | `taxRate` | `TaxRateSelector` | - | 10% / 8% 等の選択。医院マスタの税率設定と連動 |
| ステータス | `isActive` | `NotionStatusPill` | - | クリックでトグル |
| 備考 | `description`| `PropertyInput` | - | |
| 剤形 | `dosageForm` | `Select` | - | 錠剤/液剤/注射剤/外用剤/散剤 |
| 単位 | `medicineUnit`| `Select` | - | 錠/ml/回/g あたり |

## 階層構造とDnD
- **カテゴリの定義**: `parentId` が空かつ `price` が 0 のアイテムをカテゴリとして扱う。
- **並び替え**: 同一親（またはルート）内でのドラッグ&ドロップにより並び順を更新。`PATCH /api/v1/masters/medicines/reorder` を呼び出し。
- **カテゴリ移動**: ドラッグした行を異なる親カテゴリの範囲へドロップすることで `parentId` を即座に更新。

## API連携
| メソッド | エンドポイント | 用途 | 状態 |
|---------|--------------|------|------|
| GET | `/api/v1/masters/medicines` | 全薬剤取得 | 実装済 |
| POST | `/api/v1/masters/medicines` | 薬剤作成 | 実装済 |
| PATCH | `/api/v1/masters/medicines/:id` | 薬剤/カテゴリ更新 | 実装済 |
| DELETE | `/api/v1/masters/medicines/:id` | 薬剤削除 | 実装済 |
| PATCH | `/api/v1/masters/medicines/reorder` | 並び順一括保存 | 実装済 |
