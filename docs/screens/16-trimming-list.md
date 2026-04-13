## 概要
- **画面の目的**: トリミング施術記録の検索・一覧管理。
- **URLパターン**: `/trimming`
- **アクセス権限**: 認証済ユーザー全員（`ResourceTrimming` 権限が必要）

## 画面構成
- **ヘッダー**: タイトル「トリミング管理」（Scissors アイコン）、新規登録ボタン
- **検索・フィルタ**: 
  - `NotionFilter`: 日時範囲（**サーバーサイド**）、種（動的生成）、担当（動的生成）、ステータス、キーワード検索（クライアントサイド）。
  - `FilteringIndicator`: `useDeferredValue` による計算中の視覚的フィードバック。
- **データテーブル**: `DataTable` による一覧表示。
  - 特徴: 担当スタッフが非アクティブな場合、氏名横に警告アイコンを表示。
- **ページネーション**: 10件/ページ。URLパラメータ `?page=N` と同期。

## 表示項目（テーブル）

| フィールド名 | 型 | 説明 | ソート | 備考 |
|------------|-----|------|--------|------|
| 診療日 | string | 施術日（等幅フォント） | ○ | `date` (YYYY-MM-DD) |
| 飼主名 | string | 飼い主氏名 | ○ | `ownerName` |
| ペット名 | string | ペット名（ペット番号併記） | ○ | `petName`, `petNumber` |
| 種 | string | 動物種 | ○ | `species` |
| 体重 | decimal | 当日の体重 | - | `weight` |
| スタイル希望 | string | カットスタイルの要望 | - | `styleRequest` |
| 担当 | string | 担当トリマー名 | ○ | `staff`（非アクティブ時警告あり） |
| ステータス | enum | 予約 / 進行中 / 完了 | ○ | `status` (getTrimmingStatusColor で配色) |
| 操作 | - | 編集・削除 | - | `RowActionDropdown` |

## ユーザーアクション
- **新規登録**: `/trimming/select-pet` 画面を経てトリミング登録フォームへ遷移。
- **編集**: 行クリックまたは操作ボタンから `/trimming/:id` へ遷移。
- **削除**: `ConfirmDialog` 後、論理削除を実行。

## API連携
| メソッド | エンドポイント | 用途 |
|---------|--------------|------|
| GET | `/api/v1/trimmings` | トリミング一覧取得（date フィルタ対応） |
| DELETE | `/api/v1/trimmings/:id` | トリミング記録削除 |
