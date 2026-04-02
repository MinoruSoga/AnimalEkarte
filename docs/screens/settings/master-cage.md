# ケージマスタ設定 仕様書

## 概要
- **画面の目的**: 入院管理で使用するケージ（犬舎・猫舎・ICU等）の名称、エリア、サイズ、単価を管理する。
- **URLパターン**: `/settings/cage`
- **コンポーネント**: `[R] CageSettings`
- **特徴**: 並び替え（D&D）に対応。

## 画面構成
- **メインエリア**: `DataTable` によるケージ一覧（D&Dハンドル付き）。
- **サイドパネル**: `CageSidePanel` による詳細編集。

## 表示・フォーム項目

### 一覧テーブル
| フィールド | 説明 | 備考 |
|-----------|------|------|
| ケージ名 | ケージの名称（例: 犬舎 A-1） | |
| エリア | ICU / 犬舎 / 猫舎 / 汎用 | |
| サイズ | 小型 / 中型 / 大型 | |
| 単価(税込) | 1日あたりの使用料（オプション） | |
| ステータス | 有効/無効 | |

### 編集項目（サイドパネル）
| フィールド | 項目ID | 入力部品 | 必須 | 備考 |
|-----------|--------|---------|------|------|
| ケージ名 | `name` | `Input` | ✅ | タイトルエリア |
| ステータス | `isActive` | `StatusToggleButton`| - | |
| エリア | `cageType` | `Select` | ✅ | icu / dog / cat / general |
| サイズ | `cageSize` | `Select` | ✅ | small / medium / large |
| 単価(税込) | `price` | `MoneyInput` | - | |
| 備考 | `description`| `PropInput` | - | |

## API連携
| メソッド | エンドポイント | 用途 |
|---------|--------------|------|
| GET | `/api/v1/masters/cages` | 一覧取得 |
| POST | `/api/v1/masters/cages` | 新規作成 |
| PATCH | `/api/v1/masters/cages/:id` | 更新 |
| DELETE | `/api/v1/masters/cages/:id` | 削除 |
| PATCH | `/api/v1/masters/cages/reorder` | 並び順保存 |
