# BUG-021: 会計処理 POST → 400エラー（日付フォーマット不一致）

## 概要
会計待ち患者の会計処理完了時、POST /api/v1/accountings が 400 エラーを返す。
フロントエンドが `YYYY-MM-DD` 形式で `scheduled_date` を送信しているが、バックエンドが RFC3339 形式を要求している。

## 再現手順
1. ダッシュボードの「会計待ち」カラムにいる患者の会計処理を実行
2. → POST /api/v1/accountings → 400 エラー
3. → カンバンから消えない

## 期待する動作
- 会計処理が成功（201）
- 「会計待ち」カンバンから患者が消える

## 実装場所
選択肢A: フロントエンドで `scheduled_date` を RFC3339 形式（`2026-03-30T00:00:00Z`）に変換
選択肢B: バックエンドで `YYYY-MM-DD` 形式も受け付けるようにパース処理を追加

推奨: フロントエンドで変換（`frontend/src/features/accounting/` の API 呼び出し箇所）

## 優先度
High

## 関連
- FUNCTIONAL_TEST_REPORT.md BUG-021
- テスト確認日: 2026-03-30
