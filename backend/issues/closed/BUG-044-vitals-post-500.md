# BE: BUG-044 バイタル保存 POST /vitals → 500 エラー

## 概要

カルテのバイタルタブでバイタル値を保存すると `POST /api/v1/vitals` が HTTP 500 を返す。
エラーログを確認してルートコーズを特定・修正する。

## 再現手順

1. カルテ詳細画面 → バイタルタブを開く
2. 体重・体温などを入力して保存
3. → `POST /api/v1/vitals` → 500 Internal Server Error

## 期待する動作

- バイタル保存成功（201）
- 保存後に画面に値が反映される
- 必須フィールド未入力時は 400 バリデーションエラー

## 実装場所

- `backend/internal/handler/vitals_handler.go` または対応するサービス/リポジトリ
- Docker ログで具体的な panic/エラー内容を確認してから修正

## 優先度

High

## 関連

- `docs/tasks/open/crash/BUG-044_vitals_post_500.md`
- FUNCTIONAL_TEST_REPORT.md BUG-044
