# BUG-042: 問診テンプレートマスタのカテゴリ列に英語コードが表示される

## 概要
問診テンプレートマスタの一覧でカテゴリ列に英語コード（`chief_complaint`、`history`、`current_medications`、`notes`）が日本語化されずそのまま表示されている。

## 期待する動作
- `chief_complaint` → 「主訴」
- `history` → 「病歴」
- `current_medications` → 「現在の薬」
- `notes` → 「備考」

## 実装場所
- `frontend/src/features/` の問診テンプレートマスタコンポーネント
- カテゴリコードを日本語ラベルにマッピングするロジック追加

## 優先度
Low

## 関連
- FUNCTIONAL_TEST_REPORT.md BUG-042
- テスト確認日: 2026-03-30
