# BUG-065: 飼主登録完了後に詳細ページでなく一覧ページにリダイレクト

## 概要
飼主登録フォームで登録を完了すると、作成された飼主の詳細ページ（`/owners/:id`）ではなく
一覧ページ（`/owners`）にリダイレクトされる。

## 期待する動作
- 登録完了後は作成された飼主の詳細ページ（`/owners/:newId`）にリダイレクト

## 実装場所
- `frontend/src/features/owners/` の飼主登録フォームの送信後処理
- レスポンスの `id` を使って `/owners/:id` に navigate する

## 優先度
Low

## 関連
- FUNCTIONAL_TEST_REPORT.md BUG-065
- テスト確認日: 2026-03-30
