# BUG-082: バックエンドの内部エラー文字列（Go パース情報）が API レスポンスに露出

## 種類
バグ（バックエンド — エラーレスポンスの情報漏洩）

## 重要度
中（セキュリティ：内部実装詳細の漏洩）

## 発見日
2026-03-29

## 再現手順

1. 飼主登録 API に不正な日付を送信する:
   ```
   POST /api/v1/owners
   { "birth_date": "2023-02-29", ... }
   ```

## 期待動作

- HTTP 400 で `{"error": "生年月日の形式が正しくありません"}` のような日本語エラーを返す

## 実際の動作

- HTTP 400 で Go の内部パース文字列が露出する:
  ```json
  {"error": "parsing time \"2023-02-29\" as \"2006-01-02T15:04:05Z07:00\": cannot parse ..."}
  ```
- Go の `time.Parse` のエラー文字列がそのままレスポンスに含まれる
- バックエンドの実装詳細（Go の時刻フォーマット文字列 `2006-01-02T15:04:05Z07:00`）が外部に漏洩する

## 影響範囲

- 不正な日付形式を受け取るエンドポイント全体
- セキュリティ的に内部情報が外部から確認できる状態

## 修正方針

バックエンドの日付パースエラーを catch し、ユーザー向けの汎用メッセージに変換する。
Go の `time.Parse` エラーは `ErrInvalidInput` にラップして返す。

```go
// ❌ 現状（Go内部エラーをそのまま返す）
if err := parseDate(input.BirthDate); err != nil {
    return fmt.Errorf("%w: %v", ErrInvalidInput, err)
}

// ✅ 修正後（汎用メッセージに変換）
if err := parseDate(input.BirthDate); err != nil {
    return fmt.Errorf("birth_date format is invalid: %w", ErrInvalidInput)
}
```

## 優先度
中（直接的なセキュリティ被害はないが、情報漏洩のリスクあり）

## 派生イシュー

| イシュー | 領域 | 内容 |
|---------|------|------|
| BUG-082（BE） | Backend | 日付パースエラーの Go 内部文字列を汎用メッセージに変換（ErrInvalidInput ラップ） |
