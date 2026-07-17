---
name: golang-gin-api
description: "Gin のサーバ起動・middleware・routing・rate-limiting・graceful shutdown・c.Copy() ゴルーチン安全性・input sanitization を扱う。レイヤードアーキテクチャ（handler→service→repository）の設計パターン、*_request.go/*_response.go 変換、PATCH map[string]any、RespondError 一元エラー処理、slog ルールは golang-gin-clean-arch を参照。"
license: MIT
metadata:
  author: henriqueatila
  version: 2.0.0-animal-ekarte
---

# golang-gin-api — HTTP Infrastructure for Gin

This skill covers Gin **server-level infrastructure**: process lifecycle, middleware, routing, rate limiting, and goroutine safety.

For **handler → service → repository layering**, request/response DTO変換、PATCH のゼロ値問題、エラーマッピングなどのアーキテクチャ設計パターンは **golang-gin-clean-arch** スキルを参照すること。このスキルはそれらを重複して扱わない。

## When to Use

- Gin サーバのセットアップ（`gin.New()`、trusted proxies、タイムアウト設定）
- Graceful shutdown の実装
- ミドルウェアの作成・配線（auth, CORS, logging, recovery, rate limiting）
- ルート登録・ルートグループ・APIバージョニング
- `*gin.Context` をゴルーチンに安全に渡す（`c.Copy()`）
- バインド後の入力値サニタイズ

## Project Structure

```
myapp/
├── cmd/api/main.go          # Entry point, DI wiring only
├── internal/
│   ├── handler/             # HTTP handlers（レイヤー設計は golang-gin-clean-arch 参照）
│   ├── middleware/          # Auth, CORS, logging, rate limiting
│   ├── config/              # Configuration
│   ├── logger/              # slog structured logging
│   └── dbconn/              # DB connection management（実プロジェクトのディレクトリ名）
└── go.mod
```

All code lives under `internal/`. There is no `pkg/` — middleware is in `internal/middleware/`.

## Server Setup with Graceful Shutdown

```go
package main

import (
    "context"
    "log/slog"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    "myapp/internal/handler"
    "myapp/internal/service"
    "myapp/internal/repository"
    "myapp/internal/middleware"
)

func main() {
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

    // Production: gin.New() + explicit middleware (NOT gin.Default())
    r := gin.New()
    r.SetTrustedProxies([]string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"})
    r.Use(middleware.Logger(logger))
    r.Use(middleware.Recovery(logger))

    // Dependency injection (db initialized via db package)
    var db *gorm.DB
    ownerRepo := repository.NewOwnerRepository(db)
    ownerSvc  := service.NewOwnerService(ownerRepo)
    h         := handler.NewHandler(ownerSvc)

    h.RegisterOwnerRoutes(r.Group("/api/v1"))

    srv := &http.Server{
        Addr:              ":" + os.Getenv("PORT"),
        Handler:           r,
        ReadHeaderTimeout: 10 * time.Second,  // guards against Slowloris (CWE-400)
        ReadTimeout:       30 * time.Second,
        WriteTimeout:      30 * time.Second,
        IdleTimeout:       120 * time.Second,
    }

    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Error("server failed", "error", err)
            os.Exit(1)
        }
    }()

    // Buffered channel — unbuffered misses signals
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        logger.Error("graceful shutdown failed", "error", err)
    }
}
```

## Route Registration

```go
func registerRoutes(r *gin.Engine, h *handler.Handler) {
    // Health check — no auth required
    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
    })

    api := r.Group("/api/v1")

    // Public routes
    public := api.Group("")
    {
        public.POST("/auth/login", h.Login)
    }

    // Protected routes — clinic_id は auth middleware がセット
    protected := api.Group("")
    protected.Use(middleware.Auth())
    {
        h.RegisterOwnerRoutes(protected)
        h.RegisterPetRoutes(protected)
    }
}
```

ハンドラ内部の実装パターン（`*_request.go` バインド、`service.XxxInput` 変換、`*_response.go` 変換）は golang-gin-clean-arch を参照。

## Request Binding Patterns

```go
// JSON body（bind struct の設計自体は golang-gin-clean-arch 参照）
var req createOwnerRequest
if err := c.ShouldBindJSON(&req); err != nil { ... }

// Query string: GET /owners?page=1&limit=20&search=yamada
type listOwnerQuery struct {
    Page   int    `form:"page"   binding:"min=1"`
    Limit  int    `form:"limit"  binding:"min=1,max=100"`
    Search string `form:"search"`
}
var q listOwnerQuery
if err := c.ShouldBindQuery(&q); err != nil { ... }

// URI parameters: GET /owners/:id
id, err := strconv.ParseUint(c.Param("id"), 10, 64)
if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
    return
}
```

**Critical:** Always use `ShouldBind*` — `Bind*` auto-aborts with 400 and prevents custom error responses.

## Input Sanitization

Sanitize string fields after binding to neutralize injection payloads before they reach services or storage.

```go
func (h *Handler) CreateOwner(c *gin.Context) {
    var req createOwnerRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
        return
    }

    // Sanitize after bind: trim whitespace
    req.OwnerName = strings.TrimSpace(req.OwnerName)
    req.Email     = strings.ToLower(strings.TrimSpace(req.Email))

    input := service.CreateOwnerInput{
        OwnerName: req.OwnerName,
        Email:     req.Email,
    }
    // ...
}
```

For file uploads, always strip directory components from client-supplied filenames:

```go
safeName := filepath.Base(file.Filename)
dst := filepath.Join("uploads", safeName)
```

## Goroutine Safety

**Always call `c.Copy()` before passing `*gin.Context` to a goroutine.** The original context is reused by the pool after the request ends.

```go
func (h *Handler) CreateOwnerWithNotification(c *gin.Context) {
    // ... bind & call service ...

    // c.Copy() — safe to use in goroutine, original c is NOT
    cCopy := c.Copy()
    go func() {
        h.notifier.SendWelcome(cCopy.Request.Context(), owner)
    }()

    c.JSON(http.StatusCreated, toOwnerResponse(owner))
}
```

## Quick Reference

| Question | Answer |
|----------|--------|
| サーバ起動・graceful shutdown? | このスキル |
| middleware（auth/CORS/rate limit）? | このスキル / `internal/middleware/` |
| ルート登録・APIバージョニング? | このスキル |
| ゴルーチンへの `*gin.Context` 受け渡し? | `c.Copy()`（このスキル） |
| handler/service/repository の設計・DTO変換・PATCH・エラーマッピング? | **golang-gin-clean-arch** |

## Reference Files

Load these when you need deeper detail:

- **[references/routing.md](references/routing.md)** — Route groups, API versioning, path parameters, pagination, wildcard routes, file uploads, custom validators, request size limits
- **[references/middleware.md](references/middleware.md)** — CORS, security headers, request logging with slog, rate limiting, request ID, timeout, recovery, custom middleware template
- **[references/error-handling.md](references/error-handling.md)** — panic recovery とHTTPレベルの一貫したJSONエラーフォーマット（AppError/センチネルエラー本体は golang-gin-clean-arch 参照）
- Rate limiting の実装例は `backend/internal/middleware/rate_limit.go`（インメモリ `golang.org/x/time/rate`。本プロジェクトに Redis は不採用 — 分散レートリミットの references/rate-limiting.md は upstream 汎用版のため削除済み）

## Cross-Skill References

- For handler → service → repository layering, request/response DTO変換、PATCH、エラーマッピング: see **golang-gin-clean-arch**
- For testing handlers and services: see the **golang-testing** skill

## Official Docs

If this skill doesn't cover your use case, consult the [Gin documentation](https://gin-gonic.com/docs/) or [Gin GoDoc](https://pkg.go.dev/github.com/gin-gonic/gin).
