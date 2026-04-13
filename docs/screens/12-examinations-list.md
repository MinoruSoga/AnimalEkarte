## 概要
- **画面の目的**: 検査オーダー・結果の一覧管理、および進捗状況（ステータス）の把握。
- **URLパターン**: `/examinations`
- **アクセス権限**: 認証済ユーザー全員（`ResourceExaminations` 権限が必要）

## 画面構成
- **ヘッダー**: タイトル「検査管理」、新規検査登録ボタン
- **検索・フィルタ**: 
  - `NotionFilter`: 日時範囲（**サーバーサイド**）、ステータス、検査種別（動的）、担当医（動的）、キーワード検索（クライアントサイド）。
- **データテーブル**: `DataTable` による一覧表示。20件ごとのページネーション。

## 表示項目（テーブル）

| フィールド名 | 型 | 説明 | ソート | 備考 |
|------------|-----|------|--------|------|
| 日時 | string | 検査実施日時 | ○ | `date` (YYYY-MM-DD HH:mm) |
| 飼主名 | string | 飼い主氏名 | ○ | `ownerName` |
| ペット名 | string | ペット名 | ○ | `petName` |
| 検査種別 | string | 検査の種類 | ○ | `testType` |
| 結果概要 | string | 結果のテキストサマリ | - | `resultSummary` |
| 担当医 | string | 担当獣医師名 | ○ | `doctor` |
| ステータス | enum | 依頼中 / 検査中 / 完了 | ○ | `status` (getExaminationStatusColor で配色) |
| 操作 | - | 詳細（検査フォームへ遷移） | - | `RowActionButton` |

## ユーザーアクション
- **新規検査登録**: `/examinations/select-pet` を経て `/examinations/new` へ遷移。
- **詳細表示**: 行クリックまたは操作ボタンから `/examinations/:id` 編集フォームへ遷移。


## API連携
| メソッド | エンドポイント | 用途 |
|---------|--------------|------|
| GET | `/api/v1/examinations` | 検査一覧取得 |
| DELETE | `/api/v1/examinations/:id` | 検査削除 |
