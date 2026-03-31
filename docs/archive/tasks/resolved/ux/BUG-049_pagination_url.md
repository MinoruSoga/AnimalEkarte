# BUG-049: 一覧ページのページネーションがURLに反映されない

## 概要
飼主一覧等のページネーションで2ページ目に遷移してもURLが `/owners` のまま変わらない（`?page=2` が付かない）。
ブラウザリロード後にページ1に戻ってしまう。

## 期待する動作
- ページ遷移時にURLに `?page=N` が付く（例: `/owners?page=2`）
- ブラウザリロード・共有リンクでも同じページが表示される

## 実装場所
- `frontend/src/features/owners/` の一覧コンポーネント
- `useSearchParams` を使ってページ番号をURLパラメータで管理

## 優先度
Low

## 関連
- FUNCTIONAL_TEST_REPORT.md BUG-049
- テスト確認日: 2026-03-30
