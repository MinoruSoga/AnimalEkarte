# 入院登録/編集 仕様書

## 概要
- **画面の目的**: 入院・ペットホテルの新規登録および既存データの編集。
- **URLパターン**:
  - 新規登録: `/hospitalization/new?petId=xxx`
  - 編集: `/hospitalization/:id/edit`
- **アクセス権限**: 認証済ユーザー全員

## 画面構成
- **ヘッダー**: `PatientInfoCard` (ペット・飼主情報)、保存ボタン、削除ボタン（編集時のみ）、デイリーカルテへのリンク
- **メインフォーム**: 3カラム構成
  - 基本情報（`HospitalizationBasicInfo`）: 入院開始/終了、タイプ、ケージ選択
  - 飼主リクエスト（`HospitalizationNoteCard`）
  - スタッフ連絡事項（`HospitalizationNoteCard`）
- **治療プラン**: 治療プランテーブル（マスタ連携）
- **費用計算**: 金額サマリ（小計、税、値引、合計）

## 主な機能
- **ケージ管理**: ケージマスタから選択し、入院中のステータスを管理。
- **費用自動計算**: 治療プランの変更に基づき、入院費用の合計を算出。
- **カルテ連携**: 登録後、`/hospitalization/:id` の詳細画面（デイリーカルテ）へ遷移。

## API連携
| メソッド | エンドポイント | 用途 |
|---------|--------------|------|
| POST | `/api/v1/hospitalizations` | 入院作成 |
| PATCH | `/api/v1/hospitalizations/:id` | 入院更新 |
| DELETE | `/api/v1/hospitalizations/:id` | 入院削除 |
