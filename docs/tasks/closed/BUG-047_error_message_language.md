# BUG-047: フロントエンドのAPIエラーメッセージが英語で表示される

## 概要
401 Unauthorized 時などにフロントエンドが「Request failed with status code 401」という英語メッセージを表示する。
またFE/BEのエラーメッセージ言語・内容が不一致（FE: 「飼主名を入力してください」、BE: "owner_name is required"）。

## 再現手順
1. 無効なトークンで API リクエストを送信
2. → 「Request failed with status code 401」英語メッセージが表示される

## 期待する動作
- すべてのエラーメッセージを日本語で表示
- `handleApiError` で HTTP ステータスコードごとに日本語メッセージを返す

## 実装場所
- `frontend/src/lib/` の `handleApiError` 関数
- 401: 「セッションが切れました。再度ログインしてください」
- 403: 「この操作を行う権限がありません」
- 500: 「サーバーエラーが発生しました。しばらく経ってから再度お試しください」

## 優先度
Medium（UX）

## 関連
- FUNCTIONAL_TEST_REPORT.md BUG-047
- テスト確認日: 2026-03-30
