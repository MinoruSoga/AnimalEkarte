---
name: golang-gin-api
description: "軽量レイヤードアーキテクチャ（handler→service→repository→model）でGin REST APIを構築。*_request.go バインド・service.XxxInput 変換・*_response.go 変換・PATCH map[string]any・RespondError 一元エラー処理・slog service層のみ・マルチテナント clinic_id を網羅。"
license: MIT
metadata:
  author: henriqueatila
  version: 1.0.5-animal-ekarte
---

# golang-gin-api — Core REST API Development

Build production-grade REST APIs with Go and Gin. This skill covers the 80% of patterns you need daily: server setup, routing, request binding, response formatting, and error handling.

## When to Use

- Creating a new Go REST API or HTTP server
- Adding routes, handlers, or middleware to a Gin app
- Binding and validating incoming JSON/query/URI parameters
- Structuring a Go project with a layered project structure
- Wiring handlers → services → repositories in main.go
- Returning consistent JSON error responses

## Project Structure

```
myapp/
├── cmd/
│   └── api/
│       └── main.go          # Entry point, DI wiring only
├── internal/
│   ├── handler/             # HTTP handlers, *_request.go, *_response.go
│   ├── service/             # Business logic, Input DTOs, validators.go
│   ├── repository/          # Data access (GORM)
│   ├── model/               # GORM models (DB schema)
│   ├── errors/              # Sentinel errors
│   ├── middleware/          # Auth, CORS, logging
│   ├── config/              # Configuration
│   ├── logger/              # slog structured logging
│   └── db/                  # DB connection management
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

## GORM Model

GORM models live in `internal/model/` and map directly to DB schema. They carry both `gorm:` and `json:` tags. Never return model structs directly from handlers — use response structs.

```go
// internal/model/owner.go
package model

import (
    "time"
    "gorm.io/gorm"
)

type Owner struct {
    ID           uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
    ClinicID     uint64         `gorm:"not null"                 json:"clinic_id"`
    OwnerName    string         `gorm:"not null"                 json:"owner_name"`
    Email        string         `gorm:"default:''"              json:"email"`
    IsDangerous  bool           `gorm:"default:false"           json:"is_dangerous"`
    DiscountRate float64        `gorm:"default:0"               json:"discount_rate"`
    CreatedAt    time.Time      `gorm:"autoCreateTime"          json:"created_at"`
    UpdatedAt    time.Time      `gorm:"autoUpdateTime"          json:"updated_at"`
    DeletedAt    gorm.DeletedAt `                               json:"-"`
    Pets         []Pet          `gorm:"foreignKey:OwnerID"      json:"pets,omitempty"`
}

func (Owner) TableName() string { return "owners" }
```

## Handler Layer

The handler layer is the only layer that touches `*gin.Context`. It has three responsibilities:

1. Extract `clinic_id` from context (set by auth middleware)
2. Bind and parse the HTTP request
3. Call the service and format the response

### handler/owner_request.go — binding structs

```go
// handler/owner_request.go
package handler

// createOwnerRequest はリクエストボディのバインド専用。
// binding: タグはここにのみ置く。
type createOwnerRequest struct {
    OwnerName    string  `json:"owner_name"    binding:"required"`
    Email        string  `json:"email"`
    DiscountRate float64 `json:"discount_rate"`
}

// updateOwnerRequest は PATCH 用。
// ポインタ型にすることで nil（未送信）と false/0/""（明示的ゼロ値）を区別する。
type updateOwnerRequest struct {
    OwnerName    *string  `json:"owner_name"`
    IsDangerous  *bool    `json:"is_dangerous"`
    DiscountRate *float64 `json:"discount_rate"`
}
```

### handler/owner_response.go — response structs

```go
// handler/owner_response.go
package handler

import (
    "strconv"
    "myapp/internal/model"
)

// ownerResponse はAPIレスポンス専用の型。model.Owner を直返しにしない。
// ID は uint64 → string 変換（フロントエンドの number 精度問題を回避）。
type ownerResponse struct {
    ID           string  `json:"id"`
    OwnerName    string  `json:"owner_name"`
    Email        string  `json:"email"`
    IsDangerous  bool    `json:"is_dangerous"`
    DiscountRate float64 `json:"discount_rate"`
}

func toOwnerResponse(o *model.Owner) ownerResponse {
    return ownerResponse{
        ID:           strconv.FormatUint(o.ID, 10),
        OwnerName:    o.OwnerName,
        Email:        o.Email,
        IsDangerous:  o.IsDangerous,
        DiscountRate: o.DiscountRate,
    }
}
```

### handler/owner_handler.go

```go
// handler/owner_handler.go
package handler

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "myapp/internal/service"
)

func (h *Handler) CreateOwner(c *gin.Context) {
    clinicID, ok := extractClinicID(c)
    if !ok { return }

    var req createOwnerRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
        return
    }

    // HTTP非依存の型に変換してからserviceに渡す
    input := service.CreateOwnerInput{
        OwnerName:    req.OwnerName,
        Email:        req.Email,
        DiscountRate: req.DiscountRate,
    }
    owner, err := h.svc.Owner.CreateWithPets(c.Request.Context(), clinicID, &input)
    if err != nil {
        RespondError(c, err)  // ErrNotFound→404, ErrInvalidInput→400 etc.
        return
    }
    c.JSON(http.StatusCreated, toOwnerResponse(owner))
}

func (h *Handler) GetOwner(c *gin.Context) {
    clinicID, ok := extractClinicID(c)
    if !ok { return }

    id, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }

    owner, err := h.svc.Owner.GetByID(c.Request.Context(), clinicID, id)
    if err != nil {
        RespondError(c, err)
        return
    }
    c.JSON(http.StatusOK, toOwnerResponse(owner))
}

func (h *Handler) UpdateOwner(c *gin.Context) {
    clinicID, ok := extractClinicID(c)
    if !ok { return }

    id, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }

    var req updateOwnerRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
        return
    }

    input := service.UpdateOwnerInput{
        OwnerName:    req.OwnerName,
        IsDangerous:  req.IsDangerous,
        DiscountRate: req.DiscountRate,
    }
    owner, err := h.svc.Owner.Update(c.Request.Context(), clinicID, id, &input)
    if err != nil {
        RespondError(c, err)
        return
    }
    c.JSON(http.StatusOK, toOwnerResponse(owner))
}

func (h *Handler) RegisterOwnerRoutes(rg *gin.RouterGroup) {
    owners := rg.Group("/owners")
    owners.GET("",        h.ListOwners)
    owners.POST("",       h.CreateOwner)
    owners.GET("/:id",    h.GetOwner)
    owners.PATCH("/:id",  h.UpdateOwner)
    owners.DELETE("/:id", h.DeleteOwner)
}
```

## Service Layer — Input DTO

The service layer knows nothing about HTTP. No `binding:` tags, no `*gin.Context`.

```go
// service/owner_service.go — Input DTO
package service

import "context"
import "myapp/internal/model"

// CreateOwnerInput は HTTP を知らない純粋な入力型。
// binding: タグは持たない（handler の *_request.go に置く）。
type CreateOwnerInput struct {
    OwnerName    string
    Email        string
    DiscountRate float64
}

// UpdateOwnerInput は PATCH 用。ポインタ型で nil（未送信）を表現する。
type UpdateOwnerInput struct {
    OwnerName    *string  // nil = 未送信、"" = 空文字に更新
    IsDangerous  *bool    // nil = 未送信、false = false に更新
    DiscountRate *float64
}

type OwnerService interface {
    List(ctx context.Context, clinicID uint64, page, limit int, search string) ([]model.Owner, int64, error)
    GetByID(ctx context.Context, clinicID, id uint64) (*model.Owner, error)
    CreateWithPets(ctx context.Context, clinicID uint64, input *CreateOwnerInput) (*model.Owner, error)
    Update(ctx context.Context, clinicID, id uint64, input *UpdateOwnerInput) (*model.Owner, error)
    Delete(ctx context.Context, clinicID, id uint64) error
}
```

## PATCH with map[string]any

GORM の `Updates(struct)` はゼロ値フィールド（`false`, `0`, `""`）をスキップする。
PATCH では `map[string]any` を使うことでこの問題を回避する。

```go
// service/owner_service.go — PATCH 実装
func (s *ownerService) Update(ctx context.Context, clinicID, id uint64, input *UpdateOwnerInput) (*model.Owner, error) {
    // 業務バリデーション
    if input.DiscountRate != nil {
        if err := validateDiscountRate(*input.DiscountRate); err != nil {
            return nil, err
        }
    }

    fields := buildOwnerUpdateFields(input)
    if len(fields) == 0 {
        return nil, apperrors.WrapInvalidInput("at least one field must be provided")
    }

    if err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
        return nil, err
    }

    slog.InfoContext(ctx, "owner updated", slog.Uint64("owner_id", id))
    return s.repo.FindByID(ctx, clinicID, id)
}

// buildOwnerUpdateFields はポインタが非 nil のフィールドだけを map に詰める。
// false/0/"" もポインタ経由で明示的に送られた場合は更新対象になる。
func buildOwnerUpdateFields(input *UpdateOwnerInput) map[string]any {
    fields := map[string]any{}
    if input.OwnerName    != nil { fields["owner_name"]    = *input.OwnerName }
    if input.IsDangerous  != nil { fields["is_dangerous"]  = *input.IsDangerous }
    if input.DiscountRate != nil { fields["discount_rate"] = *input.DiscountRate }
    return fields
}
```

```go
// repository/owner_repository.go — map[string]any で更新
func (r *ownerRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
    result := r.db.WithContext(ctx).
        Model(&model.Owner{}).
        Where("id = ? AND clinic_id = ?", id, clinicID).
        Updates(fields)  // struct 渡しだと false/0/"" がスキップされる

    if result.RowsAffected == 0 {
        return apperrors.WrapNotFound("owner", strconv.FormatUint(id, 10))
    }
    return result.Error
}
```

## Centralized Error Handling

### internal/errors/errors.go — sentinel errors

```go
// internal/errors/errors.go
package apperrors

import (
    "errors"
    "fmt"
)

var (
    ErrNotFound      = errors.New("resource not found")
    ErrInvalidInput  = errors.New("invalid input")
    ErrAlreadyExists = errors.New("resource already exists")
    ErrUnauthorized  = errors.New("unauthorized")
    ErrForbidden     = errors.New("forbidden")
)

type AppError struct {
    Code    string
    Message string
    Err     error
}

func (e *AppError) Error() string {
    if e.Err != nil { return e.Message + ": " + e.Err.Error() }
    return e.Message
}
func (e *AppError) Unwrap() error { return e.Err }

func WrapNotFound(resource, id string) error {
    return &AppError{
        Code:    "NOT_FOUND",
        Message: fmt.Sprintf("%s not found: %s", resource, id),
        Err:     ErrNotFound,
    }
}

func WrapInvalidInput(msg string) error {
    return &AppError{
        Code:    "INVALID_INPUT",
        Message: msg,
        Err:     ErrInvalidInput,
    }
}

func WrapAlreadyExists(resource, detail string) error {
    return &AppError{
        Code:    "ALREADY_EXISTS",
        Message: fmt.Sprintf("%s already exists: %s", resource, detail),
        Err:     ErrAlreadyExists,
    }
}
```

### handler/handler.go — RespondError

```go
// handler/handler.go — RespondError で一元マッピング
package handler

import (
    "errors"
    "log/slog"
    "net/http"

    "github.com/gin-gonic/gin"
    apperrors "myapp/internal/errors"
)

// RespondError はすべてのハンドラが使う一元エラーレスポンス関数。
// 個別ハンドラに switch/if を書かない。
func RespondError(c *gin.Context, err error) {
    switch {
    case errors.Is(err, apperrors.ErrNotFound):
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
    case errors.Is(err, apperrors.ErrInvalidInput):
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    case errors.Is(err, apperrors.ErrAlreadyExists):
        c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
    case errors.Is(err, apperrors.ErrUnauthorized):
        c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
    case errors.Is(err, apperrors.ErrForbidden):
        c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
    default:
        slog.ErrorContext(c.Request.Context(), "unhandled error",
            slog.String("error", err.Error()),
            slog.String("path", c.FullPath()))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
    }
}

// extractClinicID は認証ミドルウェアがセットした clinic_id を取り出す。
// 失敗時は 401 を返して false を返すため、呼び出し側は if !ok { return } で早期リターンできる。
func extractClinicID(c *gin.Context) (uint64, bool) {
    v, exists := c.Get("clinic_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "clinic_id not found in context"})
        return 0, false
    }
    id, ok := v.(uint64)
    if !ok {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid clinic_id type"})
        return 0, false
    }
    return id, true
}
```

## slog Rules

slog はサービス層のみに書く。handler と repository には書かない。

```go
// ✅ service 層のみ — 操作の結果をログに残す
func (s *ownerService) Create(ctx context.Context, clinicID uint64, input *CreateOwnerInput) (*model.Owner, error) {
    // ... 処理 ...
    slog.InfoContext(ctx, "owner created",
        slog.Uint64("owner_id", owner.ID),
        slog.Uint64("clinic_id", clinicID))
    return owner, nil
}

func (s *ownerService) Delete(ctx context.Context, clinicID, id uint64) error {
    if err := s.repo.Delete(ctx, clinicID, id); err != nil {
        return err  // ❌ ここで slog を書かない。RespondError の default ブランチが 500 をログする
    }
    slog.InfoContext(ctx, "owner deleted",
        slog.Uint64("owner_id", id),
        slog.Uint64("clinic_id", clinicID))
    return nil
}

// ❌ handler に slog を書かない（RespondError の default 以外）
// ❌ repository に slog を書かない
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

## Request Binding Patterns

```go
// JSON body
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
| 新モデル? | `internal/model/` |
| HTTPバインド? | `handler/*_request.go`（binding struct） |
| レスポンス変換? | `handler/*_response.go`（toXxxResponse） |
| ビジネスロジック? | `internal/service/` |
| 業務バリデーション? | `service/validators.go` |
| DB操作? | `internal/repository/` |
| センチネルエラー? | `internal/errors/errors.go` |
| DI配線? | `cmd/api/main.go` のみ |
| ミドルウェア? | `internal/middleware/`（`pkg/` は使わない） |
| slogを書く場所? | service層のみ |

## Reference Files

Load these when you need deeper detail:

- **[references/routing.md](references/routing.md)** — Route groups, API versioning, path parameters, pagination, wildcard routes, file uploads, custom validators, request size limits
- **[references/middleware.md](references/middleware.md)** — CORS, security headers, request logging with slog, rate limiting, request ID, timeout, recovery, custom middleware template
- **[references/error-handling.md](references/error-handling.md)** — Full AppError system, sentinel errors, validation error formatting, panic recovery, consistent JSON error format
- **[references/websocket.md](references/websocket.md)** — WebSocket with gorilla/websocket: upgrade handler, hub pattern, auth before upgrade, ping/pong keepalive, graceful shutdown, JSON messages, testing
- **[references/rate-limiting.md](references/rate-limiting.md)** — Deep-dive rate limiting: token bucket, sliding window, Redis distributed, per-user/API-key quotas, tiered limits, response headers, graceful degradation

## Cross-Skill References

- For JWT middleware to protect routes: see the **golang-gin-auth** skill
- For wiring repositories into services and handlers: see the **golang-gin-database** skill
- For testing handlers and services: see the **golang-gin-testing** skill
- For Dockerizing this project structure: see the **golang-gin-deploy** skill

## Official Docs

If this skill doesn't cover your use case, consult the [Gin documentation](https://gin-gonic.com/docs/) or [Gin GoDoc](https://pkg.go.dev/github.com/gin-gonic/gin).
