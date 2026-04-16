# BUG-245: Gin ルート登録の起動時 panic を検出するテストが存在しない

## 概要
`handler.RegisterRoutes` のルート登録において、Gin の同一パスセグメント内でのワイルドカード名衝突は**コンパイル時に検出されず、実行時 panic** する。
現在テストが存在しないため、`make up` で実際に起動するまでバグが発見されない。
BUG-246（`reservation_line_routes.go` の `:clinicId`/`:staffId` 衝突、2026-04-09 修正済み）はこのテスト不在が直接の原因。

## 再現手順
1. `reservation_line_routes.go` を修正前の状態に戻す（`:clinicId` を使用）
2. `docker compose up` を実行
3. **結果**: backend コンテナが unhealthy になり起動失敗

`go test ./...` では検出されない（テストが存在しないため）。

## 期待する動作
- `go test ./...` でルート登録の panic が検出される
- 新規ルート追加時に CI が失敗し、衝突を即時発見できる

## 現状コード

### テスト不在の確認
```bash
grep -r "RegisterRoutes\|RegisterLineReservation\|NoPanic" backend/internal/handler/*_test.go
# → 結果なし
```

### 修正前に発生した panic（参考）
```
panic: ':clinicId' in new path '/api/v1/clinics/:clinicId/reservation-settings'
conflicts with existing wildcard ':id' in existing prefix '/api/v1/clinics/:id'
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `backend/internal/handler/` | ルート登録テストなし | 未修正 |
| `backend/internal/handler/reservation_line_routes.go` | BUG-246 として修正済み | 修正済み |

## 修正方針

### 1. ルート登録テストの追加 — `backend/internal/handler/handler_routes_test.go`（新規作成）

```go
package handler_test

import (
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"

    "github.com/animal-ekarte/backend/internal/handler"
)

// TestRegisterRoutes_NoPanic はルート登録時に panic が発生しないことを保証する。
// Gin のワイルドカード名衝突は実行時 panic になるため、このテストで早期検出する。
func TestRegisterRoutes_NoPanic(t *testing.T) {
    gin.SetMode(gin.TestMode)
    r := gin.New()

    h := handler.NewTestHandler() // テスト用の最小ハンドラ（下記参照）

    assert.NotPanics(t, func() {
        h.RegisterRoutes(nil, r)
    })
}
```

### 2. テスト用ハンドラの追加 — `backend/internal/handler/handler_test_helper_test.go`（新規作成）

```go
package handler_test

import "github.com/animal-ekarte/backend/internal/handler"

// NewTestHandler はルート登録テスト用の最小ハンドラを生成する。
// DB 接続や実際のサービスは不要。
func NewTestHandler() *handler.Handler {
    return &handler.Handler{} // ゼロ値で可（ルート登録のみ検証）
}
```

注意: `handler.Handler` の構造体フィールドに応じて、nil pointer panic を避けるため
`handler.Handler` の `RegisterRoutes` がサービス/リポジトリを参照しない部分のみをテストするか、
または DI でモック注入する設計に変更する必要がある場合がある。

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/testing.md` — テスト構造
> - Test files: `*_test.go` in same package
> - New features: Minimum 80% coverage

### プロジェクト内参照実装
- `backend/internal/handler/reservation_handler_test.go` — 既存のハンドラテスト（参考）

## 優先度
**High** — 同種の起動時 panic が再発するリスクを防ぐ。新規ルート追加のたびに `make up` で確認する運用は持続不可能。

## 関連チケット
- BUG-246（修正済み）: `reservation_line_routes.go` の Gin ワイルドカード名衝突（2026-04-09 修正）

## 関連ファイル
- `backend/internal/handler/reservation_line_routes.go` — 今回の衝突が発生したファイル
- `backend/internal/handler/handler.go:38-95` — `RegisterRoutes` 関数（テスト対象）
