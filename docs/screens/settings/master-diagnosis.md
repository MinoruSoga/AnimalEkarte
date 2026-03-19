# 診断マスタ設定 仕様書

## 概要
- **画面の目的**: カルテの「診察/治療プラン」タブで使用される診断カテゴリと病名を管理する。
- **URLパターン**: `/settings/diagnosis`
- **コンポーネント**: `[R] DiagnosisSettings`

## 画面構成とタブ
2つのタブでデータリレーションを管理します。
1. **診断病名カテゴリ** (`diagnosis_category`): 消化器系、循環器系などの大分類。
2. **診断病名** (`diagnosis_name`): 胃炎、心不全などの具体的な病名（カテゴリに属する）。

## 機能詳細

### 1. タブ間リレーション
- 診断病名は必ず1つのカテゴリに紐付きます。
- **削除ガード**: カテゴリを削除すると、紐付く診断病名も影響を受けるため、`ConfirmDialog` で明示的な警告を行います。

### 2. 並び順管理
- 両タブとも `DndContext` を用いたドラッグ＆ドロップによる並び替え（フラットソート）に対応しています。

## 表示・フォーム項目

### フォーム項目（カテゴリ）
| フィールド | 項目ID | 入力部品 | 必須 | 備考 |
|-----------|--------|---------|------|------|
| カテゴリ名| `name` | `Input` | ✅ | タイトルエリア |
| ステータス | `isActive` | `StatusToggleButton` | - | |
| 備考 | `description`| `PropInput` | - | |

### フォーム項目（診断病名）
| フィールド | 項目ID | 入力部品 | 必須 | 備考 |
|-----------|--------|---------|------|------|
| 病名 | `name` | `Input` | ✅ | タイトルエリア |
| ステータス | `isActive` | `StatusToggleButton` | - | |
| カテゴリ | `diagnosisCategoryId`| `Select` | ✅ | 登録済みのカテゴリから選択 |
| 備考 | `description`| `PropInput` | - | |

## API連携
| メソッド | エンドポイント | 用途 |
|---------|--------------|------|
| GET | `/api/v1/masters/diagnosis-categories` | カテゴリ一覧取得 |
| POST | `/api/v1/masters/diagnosis-names` | 病名作成 |
| PATCH | `/api/v1/masters/diagnosis-names/reorder`| 病名並び順一括保存 |
