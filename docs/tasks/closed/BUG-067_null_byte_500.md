# BUG-067: NULLバイト（\0）を含む入力でバックエンドが500エラー

## 概要
飼主名等のフィールドにNULLバイト（\u0000）を含む文字列を入力してPOSTすると、
バックエンドが500 Internal Server Errorを返す。
フロントエンドはサニタイズなし、バックエンドも適切なバリデーション（400）を返せていない。

## 再現手順
1. `POST /api/v1/owners` に `{"name":"テスト\u0000ヌル", ...}` を送信
2. → HTTP 500 Internal Server Error

## 期待する動作
- フロントエンド: 入力文字列からNULLバイト・制御文字をサニタイズ
- バックエンド: 不正文字を含む場合は 400 Bad Request を返す

## 実装場所
- フロントエンド: 入力フィールドの sanitize ロジック
- バックエンド: `internal/handler/` または `service/` のバリデーション処理

## 優先度
High（意図しないクラッシュ）

## 関連
- FUNCTIONAL_TEST_REPORT.md BUG-067
- テスト確認日: 2026-03-30
