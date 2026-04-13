## 概要
- **画面の目的**: 院内で販売するフードやサプリメント、グッズ等の名称、価格、税区分、税率を一元管理する。
- **URLパターン**: `/settings/merchandise-items`
- **コンポーネント**: `MerchandiseItemSettings`
- **アクセス権限**: `ResourceMasterMerchandise` 権限が必要

## 画面構成
- **メインエリア**: `DataTable` による品目一覧。
  - カラム: ドラッグハンドル、品目名、カテゴリ、単価(税込)、税率、ステータス、操作。
- **サイドパネル**: `MerchandiseSidePanel` による詳細編集。

## 表示・フォーム項目

### フォーム項目（サイドパネル）
| フィールド | 項目ID | 入力部品 | 必須 | 備考 |
|-----------|--------|---------|------|------|
| 品目名 | `name` | `Input` | ✅ | タイトルエリア |
| ステータス | `isActive` | `StatusToggleButton`| - | 有効/無効トグル |
| カテゴリ | `category` | `Select` | ✅ | フード / 物販 / その他 |
| 単価(税込) | `unitPrice` | `MoneyInput` | ✅ | |
| 課税区分 | `taxType` | `TaxTypeSelector` | ✅ | `excluded` (外税) / `included` (内税) / `non_taxable` (非課税) |
| 税率 | `taxRate` | `TaxRateSelector` | ✅ | 10% / 8% 等の選択。医院マスタの税率設定と連動 |

## 主要機能
- **ドラッグ&ドロップ**: `dnd-kit` による並び替え（フラットソート）。`PATCH /api/v1/masters/merchandise-items/reorder` で一括保存。
- **Notionスタイル**: `MasterSidePanel` を使用したクリーンな編集インターフェース。
- **権限管理**: 編集権限がないユーザーに対しては、ボタンの非表示やフォームのロック（`readOnly` モード）を適用。
