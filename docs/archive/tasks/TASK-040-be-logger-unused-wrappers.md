# TASK-040: BE logger パッケージ未使用ラッパー関数削除

**作成日**: 2026-03-27
**ステータス**: Open
**優先度**: 低
**領域**: Backend

---

## 概要

`backend/internal/logger/logger.go` に5つのラッパー関数が定義されているが、production コードから一切呼ばれていない。

- `cmd/api/main.go` および `cmd/migrate/main.go` が使用しているのは `logger.Init`, `logger.Info`, `logger.Error`, `logger.Warn` のみ
- `internal/` パッケージ群は `slog.InfoContext()` 等を直接呼び出す（`logger.Init` が `slog.SetDefault` を呼ぶため動作する）

---

## 削除対象

```go
// logger/logger.go

// WithContext: InfoContext/ErrorContext からのみ呼ばれており、外部からは使用なし
func WithContext(ctx context.Context) *slog.Logger { ... }  // L57-60

// Debug: 呼び出し箇所なし
func Debug(msg string, args ...any) { ... }  // L68-70

// InfoContext: 呼び出し箇所なし
func InfoContext(ctx context.Context, msg string, args ...any) { ... }  // L82-84

// ErrorContext: 呼び出し箇所なし
func ErrorContext(ctx context.Context, msg string, args ...any) { ... }  // L87-90

// With: 呼び出し箇所なし
func With(args ...any) *slog.Logger { ... }  // L93-95
```

**残すもの**（production コードから使用中）:
- `Init`, `DefaultConfig`, `Default` — 初期化・起動
- `Info`, `Error`, `Warn` — cmd パッケージから使用

---

## 注意事項

- `context` パッケージのインポートは `WithContext`, `InfoContext`, `ErrorContext` 削除後に不要になるため、import も削除する
- 削除後に `go build ./...` でコンパイルエラーがないことを確認する

---

## 受入条件

- [ ] `logger.go` から5関数（`WithContext`, `Debug`, `InfoContext`, `ErrorContext`, `With`）が削除されている
- [ ] `logger.go` の `"context"` import が削除されている（他で使用していない場合）
- [ ] `docker compose exec backend go build ./...` 成功
- [ ] `docker compose exec backend go test ./...` 全テストパス
