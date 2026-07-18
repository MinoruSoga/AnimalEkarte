---
name: golang-gin-clean-arch
description: "軽量レイヤードアーキテクチャ（handler→service→repository→model）の設計・実装パターン。handler/*_request.go バインド・service.XxxInput 変換・handler/*_response.go 変換・PATCH ポインタ型・buildXxxUpdateFields・RespondError・service/validators.go・slog service層のみ・マルチテナント clinic_id を完全網羅。HTTPサーバ起動・middleware・routing・rate-limiting・graceful shutdownは golang-gin-api を参照。"
license: MIT
metadata:
  author: henriqueatila
  version: 1.1.1-animal-ekarte
---

# 軽量レイヤードアーキテクチャ — Go/Gin 実装ガイド

35テーブル規模・Gin WebAPI・最速実装を前提に、Clean Architectureの過剰な抽象化を避けつつ責務分離と依存方向の一貫性を確保する。

## When to Use

- Go/Gin バックエンドに新リソースを追加するとき
- handler/service/repository のいずれかをレビュー・修正するとき
- アーキテクチャ違反（依存方向・層の責務）をチェックするとき
- PATCH実装でGORMのゼロ値問題に対処するとき

## Core Rules（厳守）

1. **`handler` は `gin.Context` を扱う唯一の層** — service/repositoryに `*gin.Context` を渡さない
2. **`binding:` タグは `handler/*_request.go` のみ** — `service.XxxInput` に binding タグを書かない
3. **model をそのままレスポンスに使わない** — 必ず `toXxxResponse()` で `*_response.go` 型に変換する
4. **`main.go` が唯一の DI 配線場所** — `di/container.go` は作らない
5. **`validation/` パッケージを作らない** — 業務バリデーションは `service/validators.go` のみ
6. **`handler` から `repository` を直接呼ばない** — service 層を必ず経由する
7. **slog は `service` 層のみ** — handler・repository には書かない
8. **`tx`（トランザクション）内で Preload しない** — 完了後に `FindByID` で取得
9. **request 由来の clinic-scoped master FK は永続化前に所有権検証**: `masterRepo.FindByID(ctx, clinicID, id)`（別 clinic は NotFound）で検証してから書き込む。子テーブルが自前 clinic_id を持たず親経由でのみ隔離される場合、repository の clinicScope（P4）は効かず、service が検証しない限り越境 write が成立する（#124 exam_type / #125 vaccine の実発生源）。Create の必須 uint64 は `!= 0`、Update の *uint64 は `!= nil` で判定し、検証は tx を含む永続化より前に行う。
   （出典: memory cross_tenant_master_fk_write_audit_20260629、commit 03bf1cb5 / f4e7b7a7）

## Project Structure

```
backend/
├── cmd/api/main.go              # エントリーポイント + DI配線（唯一の汚い場所）
├── internal/
│   ├── handler/                 # HTTP受付層
│   │   ├── handler.go / context_helpers.go / response.go / bind_errors.go  # Handler struct・extractClinicID・RespondError・parseBindError
│   │   ├── owner_handler.go     # RegisterOwnerRoutes(rg) + CRUD handlers
│   │   ├── owner_request.go     # createOwnerRequest / updateOwnerRequest（binding struct）
│   │   ├── owner_response.go    # ownerResponse + toOwnerResponse()
│   │   └── ...
│   ├── service/                 # 業務処理層
│   │   ├── owner_service.go     # OwnerService interface + CreateOwnerInput + 実装
│   │   ├── validators.go        # validateDiscountRate / validateMembershipType 等
│   │   └── ...
│   ├── repository/              # データアクセス層
│   │   ├── owner_repository.go  # OwnerRepository interface + GORM実装
│   │   └── ...
│   ├── model/                   # GORMモデル
│   │   ├── owner.go
│   │   └── ...
│   ├── errors/                  # センチネルエラー
│   │   └── errors.go
│   ├── middleware/              # Ginミドルウェア
│   ├── logger/
│   ├── config/
│   └── db/
├── migrations/
├── tygo.yaml
└── go.mod
```

## Layer 1: model/（GORMモデル）

DBスキーマに対応するGORMモデル。APIレスポンスとして直接使わない。

```go
// internal/model/owner.go
package model

type MembershipType string
const (
    MembershipTypeNonMember   MembershipType = "non_member"
    MembershipTypeMember      MembershipType = "member"
)

type Owner struct {
    ID             uint64         `gorm:"primaryKey;autoIncrement"  json:"id"`
    ClinicID       uint64         `gorm:"not null"                   json:"clinic_id"`
    OwnerName      string         `gorm:"not null"                   json:"owner_name"`
    Email          string         `gorm:"default:''"                 json:"email"`
    DiscountRate   float64        `gorm:"default:0"                  json:"discount_rate"`
    MembershipType MembershipType `gorm:"type:membership_type"       json:"membership_type"`
    IsDangerous    bool           `gorm:"default:false"              json:"is_dangerous"`
    CreatedAt      time.Time      `gorm:"autoCreateTime"             json:"created_at"`
    UpdatedAt      time.Time      `gorm:"autoUpdateTime"             json:"updated_at"`
    DeletedAt      gorm.DeletedAt `                                  json:"-"`
    Pets           []Pet          `gorm:"foreignKey:OwnerID"         json:"pets,omitempty"`
}
func (Owner) TableName() string { return "owners" }
```

## Layer 2: repository/（データアクセス）

GORM操作のみ。業務判断禁止。DBエラーをセンチネルエラーに変換する。

```go
// internal/repository/owner_repository.go
type OwnerRepository interface {
    FindAll(ctx context.Context, clinicID uint64, page, limit int, search string) ([]model.Owner, int64, error)
    FindByID(ctx context.Context, clinicID, id uint64) (*model.Owner, error)
    CreateWithPets(ctx context.Context, owner *model.Owner, pets []model.Pet) error
    Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
    Delete(ctx context.Context, clinicID, id uint64) error
}

type ownerRepository struct{ db *gorm.DB }

func NewOwnerRepository(db *gorm.DB) OwnerRepository {
    return &ownerRepository{db: db}
}

func (r *ownerRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Owner, error) {
    var owner model.Owner
    err := r.db.WithContext(ctx).
        Preload("Pets").Preload("Pets.AnimalSpecies").
        First(&owner, "id = ? AND clinic_id = ?", id, clinicID).Error
    if err != nil {
        return nil, apperrors.FromGORM(err, "owner", strconv.FormatUint(id, 10))
    }
    return &owner, nil
}

func (r *ownerRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int, search string) ([]model.Owner, int64, error) {
    owners := make([]model.Owner, 0)  // nil返却禁止（JSON null になる）
    query := r.db.WithContext(ctx).Where("clinic_id = ?", clinicID)
    if search != "" {
        like := "%" + escapeLike(search) + "%"
        query = query.Where("owner_name ILIKE ? OR phone ILIKE ? OR email ILIKE ?", like, like, like)
    }
    var total int64
    query.Model(&model.Owner{}).Count(&total)
    // リスト取得では Preload しない（一覧表示に不要なデータ取得を避ける）
    err := query.Offset((page-1)*limit).Limit(limit).Order("created_at DESC").Find(&owners).Error
    return owners, total, err
}

// Update: map[string]any でゼロ値問題（false/0/""がスキップされる）を回避
func (r *ownerRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
    result := r.db.WithContext(ctx).
        Model(&model.Owner{}).
        Where("id = ? AND clinic_id = ?", id, clinicID).
        Updates(fields)  // ❌ Updates(&struct{}) はゼロ値をスキップする
    if result.Error != nil { return apperrors.Wrap(result.Error, "update owner") }
    if result.RowsAffected == 0 { return apperrors.WrapNotFound("owner", strconv.FormatUint(id, 10)) }
    return nil
}

// トランザクション内では Preload しない
func (r *ownerRepository) CreateWithPets(ctx context.Context, owner *model.Owner, pets []model.Pet) error {
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        if err := tx.Create(owner).Error; err != nil { return err }
        for i := range pets {
            pets[i].OwnerID = owner.ID
            pets[i].ClinicID = owner.ClinicID
            if err := tx.Create(&pets[i]).Error; err != nil { return err }
            // tx 内では Preload しない → service側でFindByIDを呼ぶ
        }
        return nil
    })
}

func escapeLike(s string) string {
    s = strings.ReplaceAll(s, `\`, `\\`)
    s = strings.ReplaceAll(s, `%`, `\%`)
    return strings.ReplaceAll(s, `_`, `\_`)
}
```

## Layer 3: service/（業務処理）

HTTPを知らない。binding:タグなし。slogはここだけ。

```go
// internal/service/owner_service.go

// Input DTO（binding:タグなし。serviceはHTTPを知らない）
type CreateOwnerInput struct {
    OwnerName      string
    Email          string
    DiscountRate   float64
    MembershipType model.MembershipType
    Pets           []CreatePetInput
}

type UpdateOwnerInput struct {
    OwnerName    *string               // nil = 未送信、非nil = 更新対象
    IsDangerous  *bool                 // false でも更新対象にできる（ゼロ値問題を回避）
    DiscountRate *float64
    MembershipType *model.MembershipType
}

type OwnerService interface {
    List(ctx context.Context, clinicID uint64, page, limit int, search string) ([]model.Owner, int64, error)
    GetByID(ctx context.Context, clinicID, id uint64) (*model.Owner, error)
    CreateWithPets(ctx context.Context, clinicID uint64, input *CreateOwnerInput) (*model.Owner, error)
    Update(ctx context.Context, clinicID, id uint64, input *UpdateOwnerInput) (*model.Owner, error)
    Delete(ctx context.Context, clinicID, id uint64) error
}

type ownerService struct{ repo repository.OwnerRepository }

func NewOwnerService(repo repository.OwnerRepository) OwnerService {
    return &ownerService{repo: repo}
}

func (s *ownerService) CreateWithPets(ctx context.Context, clinicID uint64, input *CreateOwnerInput) (*model.Owner, error) {
    // 業務バリデーション（service/validators.go に定義）
    if err := validateDiscountRate(input.DiscountRate); err != nil { return nil, err }
    if err := validateMembershipType(input.MembershipType); err != nil { return nil, err }

    owner := &model.Owner{
        ClinicID: clinicID, OwnerName: input.OwnerName,
        Email: input.Email, DiscountRate: input.DiscountRate,
    }
    pets := convertToPetModels(input.Pets, clinicID)

    if err := s.repo.CreateWithPets(ctx, owner, pets); err != nil { return nil, err }

    // slog は service 層のみ
    slog.InfoContext(ctx, "owner created",
        slog.Uint64("owner_id", owner.ID),
        slog.Uint64("clinic_id", clinicID),
        slog.Int("pets_count", len(pets)))

    // tx 完了後に FindByID で完全データを取得
    return s.repo.FindByID(ctx, clinicID, owner.ID)
}

func (s *ownerService) Update(ctx context.Context, clinicID, id uint64, input *UpdateOwnerInput) (*model.Owner, error) {
    if input.DiscountRate != nil {
        if err := validateDiscountRate(*input.DiscountRate); err != nil { return nil, err }
    }
    fields := buildOwnerUpdateFields(input)
    if len(fields) == 0 {
        return nil, apperrors.WrapInvalidInput("at least one field must be provided")
    }
    if err := s.repo.Update(ctx, clinicID, id, fields); err != nil { return nil, err }
    slog.InfoContext(ctx, "owner updated", slog.Uint64("owner_id", id))
    return s.repo.FindByID(ctx, clinicID, id)
}

// DB列名定数（typo防止）
const (
    colOwnerName         = "owner_name"
    colOwnerIsDangerous  = "is_dangerous"
    colOwnerDiscountRate = "discount_rate"
)

func buildOwnerUpdateFields(input *UpdateOwnerInput) map[string]any {
    fields := map[string]any{}
    if input.OwnerName    != nil { fields[colOwnerName]         = *input.OwnerName }
    if input.IsDangerous  != nil { fields[colOwnerIsDangerous]  = *input.IsDangerous }
    if input.DiscountRate != nil { fields[colOwnerDiscountRate] = *input.DiscountRate }
    return fields
}
```

```go
// internal/service/validators.go — 業務バリデーション専用（1ファイルに限定）
func validateDiscountRate(rate float64) error {
    if rate < 0 || rate > 100 {
        return apperrors.WrapInvalidInput("discount_rate must be between 0 and 100")
    }
    return nil
}
func validateMembershipType(t model.MembershipType) error {
    switch t {
    case model.MembershipTypeNonMember, model.MembershipTypeMember,
         model.MembershipTypeDeceased, model.MembershipTypeTransferred:
        return nil
    }
    return apperrors.WrapInvalidInput("invalid membership_type: " + string(t))
}
func validatePetGender(gender string) error { ... }
```

## Layer 4: handler/（HTTP受付）

gin.Context を扱うのはここだけ。

```go
// handler/owner_request.go — binding struct（handler専用。serviceに渡さない）
type createOwnerRequest struct {
    OwnerName      string  `json:"owner_name"      binding:"required"`
    Email          string  `json:"email"`
    DiscountRate   float64 `json:"discount_rate"`
    MembershipType string  `json:"membership_type"`
    Pets           []createPetRequest `json:"pets"`
}
type updateOwnerRequest struct {
    OwnerName    *string  `json:"owner_name"`
    IsDangerous  *bool    `json:"is_dangerous"`
    DiscountRate *float64 `json:"discount_rate"`
}

// handler/owner_response.go — model を直接返さない
type ownerResponse struct {
    ID             string `json:"id"`
    OwnerName      string `json:"owner_name"`
    Email          string `json:"email"`
    DiscountRate   float64 `json:"discount_rate"`
    MembershipType string `json:"membership_type"`
    IsDangerous    bool   `json:"is_dangerous"`
}
func toOwnerResponse(o *model.Owner) ownerResponse {
    return ownerResponse{
        ID:             strconv.FormatUint(o.ID, 10),
        OwnerName:      o.OwnerName,
        Email:          o.Email,
        DiscountRate:   o.DiscountRate,
        MembershipType: string(o.MembershipType),
        IsDangerous:    o.IsDangerous,
    }
}

// handler/owner_handler.go
func (h *Handler) CreateOwner(c *gin.Context) {
    clinicID, ok := extractClinicID(c)
    if !ok { return }
    var req createOwnerRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
        return
    }
    // request struct → service input に変換（HTTPを切り離す）
    input := service.CreateOwnerInput{
        OwnerName:      req.OwnerName,
        Email:          req.Email,
        DiscountRate:   req.DiscountRate,
        MembershipType: model.MembershipType(req.MembershipType),
    }
    owner, err := h.svc.Owner.CreateWithPets(c.Request.Context(), clinicID, &input)
    if err != nil {
        RespondError(c, err)  // 一元マッピング
        return
    }
    c.JSON(http.StatusCreated, toOwnerResponse(owner))  // response struct に変換
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
    if err != nil { RespondError(c, err); return }
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

## Error Flow

```go
// internal/apperrors/errors.go
var (
    ErrNotFound      = errors.New("resource not found")
    ErrInvalidInput  = errors.New("invalid input")
    ErrAlreadyExists = errors.New("resource already exists")
    ErrUnauthorized  = errors.New("unauthorized")
    ErrForbidden     = errors.New("forbidden")
)

func WrapNotFound(resource, id string) error {
    return &AppError{Code: "NOT_FOUND", Message: resource + " with id " + id + " not found", Err: ErrNotFound}
}
func WrapInvalidInput(msg string) error {
    return &AppError{Code: "INVALID_INPUT", Message: msg, Err: ErrInvalidInput}
}

// handler/handler.go — RespondError で一元マッピング
func RespondError(c *gin.Context, err error) {
    var appErr *apperrors.AppError
    if errors.As(err, &appErr) {
        switch {
        case apperrors.IsNotFound(err):
            c.JSON(http.StatusNotFound, gin.H{"error": appErr.Message})
        case apperrors.IsInvalidInput(err):
            c.JSON(http.StatusBadRequest, gin.H{"error": appErr.Message})
        // ... 他のエラータイプも同様にマッピング
        default:
            c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
        }
    } else {
        slog.ErrorContext(c.Request.Context(), "unhandled error", "error", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
    }
}
```

```
エラー伝播フロー:
Repository
  gorm error             → FromGORM()         → ErrNotFound 等 → 404 等
  その他DBエラー          → FromGORM() (default) → 500

Service
  バリデーション失敗     → WrapInvalidInput() → ErrInvalidInput → 400

Handler
  JWT検証失敗            → middleware が直接返す → 401
  ParseUint失敗          → 直接返す → 400
  ShouldBindJSON失敗     → parseBindError() → 400
```

## DI Wiring（main.go のみ）

```go
// cmd/api/main.go — 唯一の汚い場所
db := db.NewPostgres(cfg)

ownerRepo := repository.NewOwnerRepository(db)
ownerSvc  := service.NewOwnerService(ownerRepo)

petRepo := repository.NewPetRepository(db)
petSvc  := service.NewPetService(petRepo)

h := handler.New(cfg, &service.Services{
    Owner: ownerSvc,
    Pet:   petSvc,
})

r := gin.New()
api := r.Group("/api/v1")
h.RegisterOwnerRoutes(api)
h.RegisterPetRoutes(api)
```

## Quick Reference

| 疑問 | 答え |
|------|------|
| 新リソースのファイル構成? | `handler/xxx_handler.go` + `xxx_request.go` + `xxx_response.go` + `service/xxx_service.go` + `repository/xxx_repository.go` |
| binding タグはどこ? | `handler/*_request.go` のみ |
| service の Input 型は? | binding タグなし、HTTPを知らない plain struct |
| model をそのままレスポンスに使える? | 禁止。`toXxxResponse()` で変換する |
| PATCH で false/0/"" を更新したい? | `UpdateXxxInput` をポインタ型 + `buildXxxUpdateFields()` → `map[string]any` |
| slog はどこに書く? | `service/` 層のみ |
| 業務バリデーションはどこ? | `service/validators.go`（1ファイルに限定） |
| DI はどこ? | `cmd/api/main.go` のみ |
| tx 内で Preload? | 禁止。tx 完了後に FindByID で取得 |

## Anti-Patterns（禁止事項）

| 禁止 | 理由 | 正解 |
|------|------|------|
| `service` に `*gin.Context` を渡す | serviceがGinに依存する | `c.Request.Context()` だけ渡す |
| `service.XxxInput` に `binding:` タグ | serviceがHTTPを知ることになる | handler/*_request.go でバインドしてから変換 |
| `c.JSON(200, model)` 直接返却 | DBスキーマがAPI契約に漏れる | `toXxxResponse(model)` で変換 |
| `tx.Preload(...).First(...)` | tx内でN+1が発生、クエリが複雑化 | tx完了後にFindByIDを呼ぶ |
| `db.Updates(&struct{IsDangerous: false})` | GORMがfalsをスキップする | `db.Updates(map[string]any{"is_dangerous": false})` |
| `internal/validation/` パッケージを作る | HTTP/業務バリデーションが混ざる | `service/validators.go` のみ |
| `handler` から `repository` を直接呼ぶ | service層をバイパスする | 必ずservice層を経由する |

## Reference Files

> ⚠️ references/ は upstream 汎用版（usecase/domain 命名・SQLC・DI container・testcontainers 前提）。本体 Core Rules と矛盾する箇所は**本体が正**。特に dependency-injection.md の「DI Container Pattern」は本プロジェクトでは禁止（Core Rule 4: di/container.go は作らない）。

- **[references/layer-separation.md](references/layer-separation.md)** — 層の責務詳細・アンチパターン
- **[references/repository-pattern.md](references/repository-pattern.md)** — GORM、トランザクション、クエリパターン
- **[references/error-handling.md](references/error-handling.md)** — AppError、センチネルエラー、エラー伝播
- **[references/dependency-injection.md](references/dependency-injection.md)** — main.go DI配線パターン
- Go テスト戦略は `golang-testing` スキルを参照（手書き fn-field モックが正本。testing-by-layer.md は upstream 汎用版のため削除済み）
