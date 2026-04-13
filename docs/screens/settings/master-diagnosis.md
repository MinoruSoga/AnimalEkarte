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

### フォーム項目（サイドパネル）
| フィールド | 項目ID | 入力部品 | 必須 | 備考 |
|-----------|--------|---------|------|------|
| 名称 | `name` | `Input` | ✅ | タイトルエリア（大文字表示） |
| ステータス | `isActive` | `NotionStatusPill` | - | クリックで有効/無効トグル |
| カテゴリ | `diagnosisCategoryId`| `Select` | ✅ | 病名タブのみ表示。カテゴリマスタ連動 |
| 備考 | `description`| `PropertyInput` | - | Notion スタイルのボーダーレス入力 |

## 特徴的なUI・機能
- **Notionスタイル**: `PropertyRow` と `PropertyInput` を使用したクリーンな編集体験。
- **ドラッグ&ドロップ**: 各タブの `DataTable` において `SortableDataTableRow` による並び替えが可能。
- **離脱防止**: `MasterSidePanel` が開いている間は `NavigationBlocker` により不意の離脱をガード。

## API連携
| メソッド | エンドポイント | 用途 |
|---------|--------------|------|
| GET | `/api/v1/masters/diagnosis-categories` | カテゴリ一覧取得 |
| POST | `/api/v1/masters/diagnosis-categories` | カテゴリ作成 |
| PATCH | `/api/v1/masters/diagnosis-categories/:id` | カテゴリ更新 |
| DELETE | `/api/v1/masters/diagnosis-categories/:id` | カテゴリ削除 |
| PATCH | `/api/v1/masters/diagnosis-categories/reorder` | カテゴリ並び順一括保存 |
| GET | `/api/v1/masters/diagnosis-names` | 病名一覧取得 |
| POST | `/api/v1/masters/diagnosis-names` | 病名作成 |
| PATCH | `/api/v1/masters/diagnosis-names/:id` | 病名更新 |
| DELETE | `/api/v1/masters/diagnosis-names/:id` | 病名削除 |
| PATCH | `/api/v1/masters/diagnosis-names/reorder` | 病名並び順一括保存 |
