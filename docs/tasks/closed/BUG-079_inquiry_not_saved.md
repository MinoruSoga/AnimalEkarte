# BUG-079: 問診フィールド（主訴・メモ）がDBに保存されないGORMバグ

## 概要
PATCH `/api/v1/medical-records/:id/inquiries` で主訴・メモを送信するとサービス層で "inquiry upserted" ログが出るが、
DBの値が実際には変わらない。GORMのFirstOrCreate+Assignに同一ポインタを渡すバグが原因。
また上限チェック（1000文字）も未実装。

## 再現手順
1. カルテの問診タブで主訴を入力して保存
2. ページをリロード
3. → 入力した内容が消えている（DBに保存されていない）

## 原因
`backend/internal/service/` の inquiry upsert 処理で GORM の `FirstOrCreate` + `Assign` に同一ポインタを渡しているため、
実際のUPDATEが発生していない。

## 期待する動作
- 主訴・メモの変更がDBに正しく保存される
- 主訴フィールドの上限（1000文字）を超えた場合は 400 Bad Request を返す

## 実装場所
- `backend/internal/service/` の inquiry/medical record 関連サービス
- GORMの upsert ロジックを修正（`Save` または `Updates` を使用）

## 優先度
Critical（データが保存されない）

## 関連
- FUNCTIONAL_TEST_REPORT.md BUG-079
- テスト確認日: 2026-03-30
