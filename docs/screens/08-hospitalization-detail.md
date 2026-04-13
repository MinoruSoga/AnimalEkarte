# 入院詳細・デイリーカルテ 仕様書

## 概要
- **画面の目的**: 入院患者のケアプラン管理、デイリーログ（バイタル・ケアログ・スタッフメモ）の記録、入院サマリーの管理。
- **URLパターン**: `/hospitalization/:id`
- **アクセス権限**: 認証済ユーザー全員（操作権限は `usePermission` で制御）

## 画面構成
- **ヘッダー**: タイトル「入院詳細・カルテ」、戻るボタン、退院ボタン、入院サマリー印刷。
- **デスクトップ表示 (`HospitalizationExpandedView`)**:
  - **左エリア**: 患者情報カード、ケアプラン（投薬・給餌・処置タスク）。
  - **右エリア**: デイリーカルテ（タイムライン形式）。
- **モバイル表示 (`HospitalizationTabbedView`)**:
  - タブ形式でプランとログを切り替え。

## 主要機能
- **時系列タスク管理**: 朝・昼・夜の時間帯ごとにタスク実施状況を表示。完了時に `TaskCompleteDialog` で詳細（実施者、メモ、バイタル値等）を追記。
- **退院フロー**: `DischargeAlertDialog` 承認後、ステータスを退院に変更。
  - 会計統合: 紐付く未精算会計があればその詳細ページへ、なければ `?petId=xxx` を付与して会計作成ページへ自動遷移する。
- **入院サマリー印刷**: 入院期間中の全経過を A4 帳票形式で出力。


## UI コンポーネント
| コンポーネント | 種別 | 説明 |
|---|---|---|
| `HospitalizationDetail` | `[R]` | メインページ。 |
| `HospitalizationExpandedView`| `[C]` | デスクトップ向けの左右分割レイアウト。 |
| `HospitalizationTabbedView` | `[C]` | モバイル向けのタブ形式レイアウト。 |
| `DailyRecordSection` | `[C]` | 時系列のデイリーログ管理。 |
| `TaskCompleteDialog` | `[C][M]` | タスク完了時のメモ入力ダイアログ。 |

## API連携
| メソッド | エンドポイント | 用途 | 状態 |
|---------|--------------|------|------|
| GET | `/api/v1/hospitalizations/:id` | 入院詳細取得 | 実装済 |
| PATCH | `/api/v1/hospitalizations/:id` | ステータス更新（退院等） | 実装済 |
| POST | `/api/v1/hospitalizations/:id/daily-records` | 日次記録作成 | 実装済 |
| GET | `/api/v1/hospitalizations/:id/daily-records` | 日次記録一覧 | 実装済 |
