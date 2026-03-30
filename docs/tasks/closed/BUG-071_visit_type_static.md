# BUG-071: 来院種別（visit_type）が変更できない

## 概要
カルテ編集画面のヘッダー部分に表示される「来院種別」（初診/再診 等）が
StaticText として表示されるのみで、UI から変更する手段がない。

## 再現手順

1. `/medical-records/:id` のカルテ編集画面を開く
2. ヘッダー部分の「来院種別」項目を確認する
3. 「再診」などの文字が表示されるが、クリックや編集が不可能

## 期待する動作

- 来院種別を UI から変更できる（例: ドロップダウンで「初診」「再診」「急患」などを選択）
- 変更時に PATCH `/api/v1/medical-records/:id` で `visit_type` が送信される

## 実際の動作

- `来院種別` ラベルの隣に「再診」が StaticText として表示されるのみ
- 編集・変更不可
- A11yツリー上: `uid=xxx StaticText "再診"`（button/combobox ではない）

## 影響範囲

- `frontend/src/features/medical-records/` のカルテ編集ヘッダーコンポーネント
- バックエンド: `PATCH /api/v1/medical-records/:id` の `visit_type` フィールド対応確認が必要

## 優先度

High（基本的な診察フローに影響）

## 関連

- FUNCTIONAL_TEST_REPORT.md BUG-071
- テスト確認日: 2026-03-30（ブラウザテストで NG 確認）
