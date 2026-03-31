# 保険マスタ設定 仕様書

## 概要
- **画面の目的**: 会計精算時に使用するペット保険の名称、補償率、連絡先を一元管理する。
- **URLパターン**: `/settings/insurance`
- **コンポーネント**: `[R] InsuranceSettings`

## 画面構成
- **メインエリア**: `DataTable` による保険一覧。
- **サイドパネル**: `InsuranceSidePanel` による詳細編集。

## 表示・フォーム項目

### 一覧テーブル
| フィールド | 説明 | 備考 |
|-----------|------|------|
| 名称 | 保険会社の名称（例: アニコム, アイペット） | |
| 補償率 | 窓口精算時の標準補償率（%） | |
| 連絡先 | 保険会社の問い合わせ先電話番号 | |
| ステータス | 有効/無効 | |

### 編集項目（サイドパネル）
| フィールド | 項目ID | 入力部品 | 必須 | 備考 |
|-----------|--------|---------|------|------|
| 名称 | `name` | `Input` | ✅ | タイトルエリア |
| ステータス | `isActive` | `StatusToggleButton`| - | |
| 補償率(%) | `coverageRate`| `Input(number)`| - | 0〜100 |
| 連絡先 | `contactPhone`| `Input(tel)` | - | |
| 備考 | `description`| `PropInput` | - | |

## API連携
| メソッド | エンドポイント | 用途 |
|---------|--------------|------|
| GET | `/api/v1/masters/insurances` | 一覧取得 |
| POST | `/api/v1/masters/insurances` | 新規作成 |
| PATCH | `/api/v1/masters/insurances/:id` | 更新 |
| DELETE | `/api/v1/masters/insurances/:id` | 削除 |
