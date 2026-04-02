# 薬剤マスタ設定 仕様書

## 概要
- **画面の目的**: 院内で使用する薬剤（内服・外用・注射等）の名称、価格、剤形、在庫連携情報を一元管理する。
- **URLパターン**: `/settings/medicine`
- **コンポーネント**: `[R] MedicineSettings`
- **特徴**: カテゴリと薬剤の2階層構造。ドラッグ&ドロップによる並び替えとカテゴリ移動が可能。

## 画面構成
- **ヘッダー**: ページタイトル、新規登録ボタン
- **検索バー**: `NotionFilter`（薬品名でフィルタ）
- **テーブル**: 薬剤一覧（階層表示、D&Dハンドル付き）
- **サイドパネル**: `MedicineSidePanel`（Notion風のプロパティ編集エリア）

## 表示・フォーム項目

### 共通項目（サイドパネル）
| フィールド | 項目ID | 入力部品 | 必須 | 備考 |
|-----------|--------|---------|------|------|
| 薬品名 | `name` | `Input` | ✅ | タイトルエリアで編集 |
| 親カテゴリ | `parentId` | `Select` | - | 既存のカテゴリ薬（単価0円のもの）から選択 |
| 税区分 | `taxType` | `TaxTypeSelector` | - | 課税/非課税/軽減税率 |
| 税率 | `taxRate` | `TaxRateSelector` | - | 税区分に応じて自動選択 |
| 単価(税込) | `price` | `MoneyInput`| - | カテゴリの場合は入力不可 |
| ステータス | `isActive` | `StatusToggleButton`| - | 有効/無効の切り替え |
| 備考 | `description`| `PropInput` | - | |

### 薬剤詳細セクション
| フィールド | 項目ID | 入力部品 | 選択肢 |
|-----------|--------|----------|-------|
| 剤形 | `dosageForm` | `Select` | 錠剤 / 液剤 / 注射剤 / 外用剤 / 散剤 |
| 単位 | `medicineUnit`| `Select` | 1錠 / 1ml / 1回 / 1g あたり |

## 階層構造とDnD
- **カテゴリ**: `parentId` が空かつ `price` が 0 のアイテム。
- **薬剤**: カテゴリに属する（`parentId` あり）、または独立したアイテム。
- **並び替え**: `useSortableList` (dnd-kit) を使用。
- **カテゴリ移動**: ドラッグした行を別のカテゴリ行の範囲にドロップすることで `parentId` を更新。

## API連携
| メソッド | エンドポイント | 用途 | 状態 |
|---------|--------------|------|------|
| GET | `/api/v1/masters/medicines` | 全薬剤取得 | 実装済 |
| POST | `/api/v1/masters/medicines` | 薬剤作成 | 実装済 |
| PATCH | `/api/v1/masters/medicines/:id` | 薬剤/カテゴリ更新 | 実装済 |
| DELETE | `/api/v1/masters/medicines/:id` | 薬剤削除 | 実装済 |
| PATCH | `/api/v1/masters/medicines/reorder` | 並び順一括保存 | 実装済 |
