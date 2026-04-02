# 物販マスタ設定 仕様書

## 概要
- **画面の目的**: 院内で販売するフードやサプリメント、グッズ等の名称、価格、税率を一元管理する。
- **URLパターン**: `/settings/merchandise-items`
- **コンポーネント**: `[R] MerchandiseItemSettings`
- **特徴**: 並び替え（D&D）に対応。

## 画面構成
- **メインエリア**: `DataTable` による品目一覧（D&Dハンドル付き）。
- **サイドパネル**: `MerchandiseSidePanel` による詳細編集。

## 表示・フォーム項目

### 一覧テーブル
| フィールド | 説明 | 備考 |
|-----------|------|------|
| 品目名 | 商品の名称 | |
| カテゴリ | フード / 物販 / その他 | |
| 単価(税込) | 販売単価 | |
| 税率 | 10% / 8% / 非課税 | |
| ステータス | 有効/無効 | |

### 編集項目（サイドパネル）
| フィールド | 項目ID | 入力部品 | 必須 | 備考 |
|-----------|--------|---------|------|------|
| 品目名 | `name` | `Input` | ✅ | タイトルエリア |
| カテゴリ | `category` | `Select` | ✅ | フード / 物販 / その他 |
| 単価(税込) | `unitPrice` | `MoneyInput` | ✅ | |
| 税率 | `taxRate` | `Select` | ✅ | 0.1 / 0.08 / 0 |
| ステータス | `isActive` | `StatusToggleButton`| - | |

## API連携
| メソッド | エンドポイント | 用途 |
|---------|--------------|------|
| GET | `/api/v1/masters/merchandise-items` | 一覧取得 |
| POST | `/api/v1/masters/merchandise-items` | 新規作成 |
| PATCH | `/api/v1/masters/merchandise-items/:id` | 更新 |
| DELETE | `/api/v1/masters/merchandise-items/:id` | 削除 |
| PATCH | `/api/v1/masters/merchandise-items/reorder` | 並び順保存 |
