# BUG-419: consultation_service のエラーメッセージが英語（全サービス日本語統一に違反）

## 概要

`consultation_service.go` のバリデーションエラーメッセージが英語で記述されており、
他の全マスタサービスが日本語で統一しているプロジェクト規約に違反している。
また定数化もされておらず、`validators.go` の共通定数が使用されていない。

## 問題箇所

```go
// consultation_service.go:105-106
return nil, apperrors.WrapInvalidInput("input must not be nil")
// 英語 ← 違反

// consultation_service.go:113
return nil, apperrors.WrapInvalidInput("at least one field must be provided")
// 英語 ← 違反

// consultation_service.go:139
return nil, apperrors.WrapInvalidInput("ids must not be empty")
// 英語 ← 違反
```

## 他サービスとの比較

```go
// animal_species_service.go（標準パターン）
return nil, apperrors.WrapInvalidInput("少なくとも1つのフィールドを指定してください")

// medicine_service.go
return nil, apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)   // 定数使用

// vaccine_service.go
return nil, apperrors.WrapInvalidInput(ErrMsgIDsNotEmpty)
```

## 修正方針

`validators.go` に定義済みの共通定数を使用し、日本語に統一する。

```go
// consultation_service.go 修正後

// 行 106（input nil チェック）
return nil, apperrors.WrapInvalidInput("入力データが不正です")

// 行 113（少なくとも1フィールド）
return nil, apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)   // 定数使用

// 行 139（IDs 空）
return nil, apperrors.WrapInvalidInput(ErrMsgIDsNotEmpty)       // 定数使用
```

## 影響ファイル

- `backend/internal/service/consultation_service.go` — 行 105-106, 113, 139

## 優先度

**Medium** — ユーザー向けエラーメッセージが英語で表示される。API クライアント・フロントエンドへの影響あり。

## 関連規約

- `.claude/CLAUDE.md` — 型安全性最優先・統一された実装方法
- `backend/internal/service/validators.go` — 共通エラーメッセージ定数
