# Backend Coding Rules

## 概要

Go 1.22+ / Gin / GORM のバックエンド開発規約。
クリーンアーキテクチャに準拠する。

---

## 1. アーキテクチャ

### 1.1 ディレクトリ構成

```
backend/
├── cmd/
│   └── api/
│       └── main.go             # エントリーポイント
│
├── internal/                   # 内部パッケージ（外部からimport不可）
│   ├── config/
│   │   └── config.go           # 環境変数・設定読み込み
│   │
│   ├── errors/
│   │   └── errors.go           # センチネルエラー定義
│   │
│   ├── handler/                # HTTPハンドラ（プレゼンテーション層）
│   │   ├── owner_handler.go
│   │   ├── pet_handler.go
│   │   └── ...
│   │
│   ├── service/                # ビジネスロジック（ユースケース層）
│   │   ├── owner_service.go
│   │   ├── pet_service.go
│   │   └── ...
│   │
│   ├── repository/             # データアクセス（インフラ層）
│   │   ├── owner_repository.go
│   │   ├── pet_repository.go
│   │   └── ...
│   │
│   ├── model/                  # ドメインモデル
│   │   ├── owner.go
│   │   ├── pet.go
│   │   └── ...
│   │
│   ├── middleware/             # ミドルウェア
│   │   ├── auth.go
│   │   ├── cors.go
│   │   └── logging.go
│   │
│   └── logger/                 # ロガー設定
│       └── logger.go
│
├── migrations/                 # DBマイグレーション
│   ├── 001_create_owners.sql
│   ├── 002_create_pets.sql
│   └── ...
│
├── docs/                       # Swagger
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
│
├── .golangci.yml               # Linter設定
├── go.mod
├── go.sum
└── Dockerfile.dev
```

### 1.2 依存関係の方向

```
┌─────────────────────────────────────────────────┐
│                   cmd/api                        │
│              (エントリーポイント)                  │
└─────────────────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────┐
│                   handler                        │
│              (プレゼンテーション層)                │
│         HTTP リクエスト/レスポンス処理            │
└─────────────────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────┐
│                   service                        │
│              (ビジネスロジック層)                 │
│              ユースケース実装                    │
└─────────────────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────┐
│                  repository                      │
│              (データアクセス層)                   │
│              DB操作の実装                        │
└─────────────────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────┐
│                    model                         │
│              (ドメインモデル)                     │
│              エンティティ定義                    │
└─────────────────────────────────────────────────┘
```

**ルール:**
- 上位層は下位層にのみ依存
- 逆方向の依存は禁止
- 各層はインターフェースを通じて疎結合

### 1.3 インターフェース定義

```go
// repository/owner_repository.go
type OwnerRepository interface {
    FindByID(ctx context.Context, id uuid.UUID) (*model.Owner, error)
    FindAll(ctx context.Context, filter OwnerFilter) ([]model.Owner, error)
    Create(ctx context.Context, owner *model.Owner) error
    Update(ctx context.Context, owner *model.Owner) error
    Delete(ctx context.Context, id uuid.UUID) error
}

// service/owner_service.go
type OwnerService interface {
    GetOwner(ctx context.Context, id string) (*model.Owner, error)
    ListOwners(ctx context.Context, filter OwnerFilter) ([]model.Owner, error)
    CreateOwner(ctx context.Context, input CreateOwnerInput) (*model.Owner, error)
    UpdateOwner(ctx context.Context, id string, input UpdateOwnerInput) (*model.Owner, error)
    DeleteOwner(ctx context.Context, id string) error
}
```

---

## 2. Context伝播

### 2.1 必須ルール

**全ての関数・メソッドの第一引数に `context.Context` を渡す。**

```go
// ✅ 正しい: 全レイヤーでContextを伝播
func (h *OwnerHandler) GetOwner(c *gin.Context) {
    ctx := c.Request.Context()
    owner, err := h.service.GetOwner(ctx, id)
    // ...
}

func (s *ownerService) GetOwner(ctx context.Context, id string) (*model.Owner, error) {
    return s.repo.FindByID(ctx, uuid.MustParse(id))
}

func (r *ownerRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Owner, error) {
    var owner model.Owner
    if err := r.db.WithContext(ctx).First(&owner, "id = ?", id).Error; err != nil {
        return nil, err
    }
    return &owner, nil
}

// ❌ 禁止: Contextを省略
func (s *ownerService) GetOwner(id string) (*model.Owner, error) {
    return s.repo.FindByID(uuid.MustParse(id))
}
```

### 2.2 Context活用

```go
// タイムアウト設定
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()

result, err := doSomething(ctx)
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        // タイムアウト処理
    }
    return err
}

// リクエストID伝播
type contextKey string
const requestIDKey contextKey = "request_id"

func SetRequestID(ctx context.Context, id string) context.Context {
    return context.WithValue(ctx, requestIDKey, id)
}

func GetRequestID(ctx context.Context) string {
    if id, ok := ctx.Value(requestIDKey).(string); ok {
        return id
    }
    return ""
}
```

---

## 3. エラーハンドリング

### 3.1 センチネルエラー

```go
// internal/errors/errors.go
package errors

import (
    "errors"
    "fmt"
)

// センチネルエラー（パッケージレベルで定義）
var (
    ErrNotFound      = errors.New("resource not found")
    ErrInvalidInput  = errors.New("invalid input")
    ErrUnauthorized  = errors.New("unauthorized")
    ErrForbidden     = errors.New("forbidden")
    ErrConflict      = errors.New("resource conflict")
    ErrInternal      = errors.New("internal server error")
)

// エラーラッピング
func Wrap(err error, message string) error {
    if err == nil {
        return nil
    }
    return fmt.Errorf("%s: %w", message, err)
}

// エラーラッピング（フォーマット付き）
func Wrapf(err error, format string, args ...interface{}) error {
    if err == nil {
        return nil
    }
    return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err)
}
```

### 3.2 エラー判定

```go
// errors.Is() でセンチネルエラーを判定
func (h *OwnerHandler) GetOwner(c *gin.Context) {
    ctx := c.Request.Context()
    id := c.Param("id")

    owner, err := h.service.GetOwner(ctx, id)
    if err != nil {
        switch {
        case errors.Is(err, apperrors.ErrNotFound):
            c.JSON(http.StatusNotFound, gin.H{
                "error": "飼主が見つかりません",
                "code":  "OWNER_NOT_FOUND",
            })
        case errors.Is(err, apperrors.ErrInvalidInput):
            c.JSON(http.StatusBadRequest, gin.H{
                "error": "無効なリクエストです",
                "code":  "INVALID_INPUT",
            })
        default:
            slog.ErrorContext(ctx, "failed to get owner",
                slog.String("owner_id", id),
                slog.String("error", err.Error()))
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": "サーバーエラーが発生しました",
                "code":  "INTERNAL_ERROR",
            })
        }
        return
    }

    c.JSON(http.StatusOK, owner)
}
```

### 3.3 エラーラッピングパターン

```go
// Repository層: DBエラーをセンチネルエラーに変換
func (r *ownerRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Owner, error) {
    var owner model.Owner
    if err := r.db.WithContext(ctx).First(&owner, "id = ?", id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, apperrors.ErrNotFound
        }
        return nil, apperrors.Wrap(err, "failed to find owner")
    }
    return &owner, nil
}

// Service層: コンテキストを追加してラップ
func (s *ownerService) GetOwner(ctx context.Context, id string) (*model.Owner, error) {
    ownerID, err := uuid.Parse(id)
    if err != nil {
        return nil, apperrors.Wrapf(apperrors.ErrInvalidInput, "invalid owner id: %s", id)
    }

    owner, err := s.repo.FindByID(ctx, ownerID)
    if err != nil {
        return nil, apperrors.Wrapf(err, "owner_id=%s", id)
    }

    return owner, nil
}
```

### 3.4 禁止パターン

```go
// ❌ 禁止: エラー握りつぶし
result, _ := doSomething()

// ❌ 禁止: panicの乱用
if err != nil {
    panic(err)
}

// ❌ 禁止: 文字列でエラー判定
if err.Error() == "not found" {
    // ...
}

// ❌ 禁止: エラーログだけ出して握りつぶし
if err != nil {
    log.Println(err)
    return nil  // エラーを返さない
}
```

---

## 4. ロギング (slog)

### 4.1 基本的な使い方

```go
import "log/slog"

// 情報ログ
slog.InfoContext(ctx, "owner created",
    slog.String("owner_id", owner.ID.String()),
    slog.String("name", owner.Name))

// エラーログ
slog.ErrorContext(ctx, "failed to create owner",
    slog.String("error", err.Error()),
    slog.String("input", fmt.Sprintf("%+v", input)))

// 警告ログ
slog.WarnContext(ctx, "deprecated endpoint called",
    slog.String("endpoint", "/api/v1/legacy"))

// デバッグログ
slog.DebugContext(ctx, "query executed",
    slog.String("sql", query),
    slog.Duration("duration", duration))
```

### 4.2 構造化ログ属性

```go
// ✅ 構造化された属性を使用
slog.InfoContext(ctx, "request completed",
    slog.String("method", c.Request.Method),
    slog.String("path", c.Request.URL.Path),
    slog.Int("status", c.Writer.Status()),
    slog.Duration("duration", time.Since(start)),
    slog.String("request_id", GetRequestID(ctx)))

// ❌ 禁止: 文字列結合
slog.Info("request completed: " + method + " " + path)

// ❌ 禁止: fmt.Sprintf での構造化
slog.Info(fmt.Sprintf("request: method=%s path=%s", method, path))
```

### 4.3 ログレベル指針

| レベル | 用途 | 例 |
|--------|------|-----|
| Debug | 開発時のデバッグ情報 | SQL、詳細なリクエスト情報 |
| Info | 通常の操作ログ | リクエスト完了、作成/更新成功 |
| Warn | 注意が必要な状態 | 非推奨API使用、リトライ発生 |
| Error | エラー発生 | DB接続失敗、外部API失敗 |

### 4.4 ログ出力禁止事項

```go
// ❌ 禁止: パスワード、トークン等の機密情報
slog.Info("user login",
    slog.String("password", password))  // ❌

// ❌ 禁止: 個人情報の過剰出力
slog.Info("owner created",
    slog.String("credit_card", owner.CreditCard))  // ❌

// ✅ 機密情報はマスク or 出力しない
slog.Info("user login",
    slog.String("email", email))  // ✅
```

---

## 5. データベース (GORM)

### 5.1 基本的なCRUD

```go
// Create
func (r *ownerRepository) Create(ctx context.Context, owner *model.Owner) error {
    if err := r.db.WithContext(ctx).Create(owner).Error; err != nil {
        return apperrors.Wrap(err, "failed to create owner")
    }
    return nil
}

// Read (単一)
func (r *ownerRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Owner, error) {
    var owner model.Owner
    if err := r.db.WithContext(ctx).First(&owner, "id = ?", id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, apperrors.ErrNotFound
        }
        return nil, apperrors.Wrap(err, "failed to find owner")
    }
    return &owner, nil
}

// Read (複数 + フィルタ)
func (r *ownerRepository) FindAll(ctx context.Context, filter OwnerFilter) ([]model.Owner, error) {
    var owners []model.Owner
    query := r.db.WithContext(ctx)

    if filter.Name != "" {
        query = query.Where("name LIKE ?", "%"+filter.Name+"%")
    }
    if filter.Phone != "" {
        query = query.Where("phone = ?", filter.Phone)
    }

    if err := query.Find(&owners).Error; err != nil {
        return nil, apperrors.Wrap(err, "failed to find owners")
    }
    return owners, nil
}

// Update
func (r *ownerRepository) Update(ctx context.Context, owner *model.Owner) error {
    result := r.db.WithContext(ctx).Save(owner)
    if result.Error != nil {
        return apperrors.Wrap(result.Error, "failed to update owner")
    }
    if result.RowsAffected == 0 {
        return apperrors.ErrNotFound
    }
    return nil
}

// Delete (ソフトデリート)
func (r *ownerRepository) Delete(ctx context.Context, id uuid.UUID) error {
    result := r.db.WithContext(ctx).Delete(&model.Owner{}, "id = ?", id)
    if result.Error != nil {
        return apperrors.Wrap(result.Error, "failed to delete owner")
    }
    if result.RowsAffected == 0 {
        return apperrors.ErrNotFound
    }
    return nil
}
```

### 5.2 N+1問題の回避

```go
// ❌ N+1問題
func (r *ownerRepository) FindAllWithPets(ctx context.Context) ([]model.Owner, error) {
    var owners []model.Owner
    r.db.WithContext(ctx).Find(&owners)

    for i := range owners {
        // N回のクエリが発生
        r.db.WithContext(ctx).Where("owner_id = ?", owners[i].ID).Find(&owners[i].Pets)
    }
    return owners, nil
}

// ✅ Preloadで解決
func (r *ownerRepository) FindAllWithPets(ctx context.Context) ([]model.Owner, error) {
    var owners []model.Owner
    if err := r.db.WithContext(ctx).Preload("Pets").Find(&owners).Error; err != nil {
        return nil, apperrors.Wrap(err, "failed to find owners with pets")
    }
    return owners, nil
}

// ✅ 条件付きPreload
func (r *ownerRepository) FindAllWithActivePets(ctx context.Context) ([]model.Owner, error) {
    var owners []model.Owner
    if err := r.db.WithContext(ctx).
        Preload("Pets", "status = ?", "active").
        Find(&owners).Error; err != nil {
        return nil, apperrors.Wrap(err, "failed to find owners")
    }
    return owners, nil
}
```

### 5.3 トランザクション

```go
func (r *ownerRepository) CreateWithPets(ctx context.Context, owner *model.Owner) error {
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        // Owner作成
        if err := tx.Create(owner).Error; err != nil {
            return apperrors.Wrap(err, "failed to create owner")
        }

        // Pets作成
        for _, pet := range owner.Pets {
            pet.OwnerID = owner.ID
            if err := tx.Create(&pet).Error; err != nil {
                return apperrors.Wrap(err, "failed to create pet")
            }
        }

        return nil  // コミット
    })
}
```

### 5.4 SQLインジェクション対策

```go
// ❌ 禁止: 文字列結合
r.db.Raw("SELECT * FROM owners WHERE name = '" + name + "'")

// ❌ 禁止: fmt.Sprintfでの埋め込み
r.db.Raw(fmt.Sprintf("SELECT * FROM owners WHERE name = '%s'", name))

// ✅ プレースホルダを使用
r.db.Raw("SELECT * FROM owners WHERE name = ?", name)

// ✅ Where句
r.db.Where("name = ?", name).Find(&owners)

// ✅ IN句
r.db.Where("id IN ?", ids).Find(&owners)
```

---

## 6. HTTPハンドラ (Gin)

### 6.1 基本構造

```go
// internal/handler/owner_handler.go
package handler

type OwnerHandler struct {
    service service.OwnerService
}

func NewOwnerHandler(s service.OwnerService) *OwnerHandler {
    return &OwnerHandler{service: s}
}

// GET /api/owners/:id
func (h *OwnerHandler) GetOwner(c *gin.Context) {
    ctx := c.Request.Context()
    id := c.Param("id")

    owner, err := h.service.GetOwner(ctx, id)
    if err != nil {
        h.handleError(c, err)
        return
    }

    c.JSON(http.StatusOK, owner)
}

// POST /api/owners
func (h *OwnerHandler) CreateOwner(c *gin.Context) {
    ctx := c.Request.Context()

    var input CreateOwnerInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "無効なリクエストです",
            "code":  "INVALID_REQUEST",
        })
        return
    }

    owner, err := h.service.CreateOwner(ctx, input)
    if err != nil {
        h.handleError(c, err)
        return
    }

    c.JSON(http.StatusCreated, owner)
}
```

### 6.2 リクエストバリデーション

```go
// 入力構造体にバリデーションタグ
type CreateOwnerInput struct {
    Name     string `json:"name" binding:"required,min=1,max=100"`
    Email    string `json:"email" binding:"required,email"`
    Phone    string `json:"phone" binding:"required,len=11"`
    Address  string `json:"address" binding:"max=200"`
}

// カスタムバリデーション
func init() {
    if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
        v.RegisterValidation("phone_jp", validateJapanesePhone)
    }
}

func validateJapanesePhone(fl validator.FieldLevel) bool {
    phone := fl.Field().String()
    matched, _ := regexp.MatchString(`^0\d{9,10}$`, phone)
    return matched
}
```

### 6.3 レスポンス形式

```go
// 成功レスポンス
c.JSON(http.StatusOK, gin.H{
    "data": owner,
})

// ページネーションレスポンス
c.JSON(http.StatusOK, gin.H{
    "data": owners,
    "pagination": gin.H{
        "total":     total,
        "page":      page,
        "page_size": pageSize,
    },
})

// エラーレスポンス
c.JSON(http.StatusBadRequest, gin.H{
    "error":   "無効なリクエストです",
    "code":    "INVALID_REQUEST",
    "details": validationErrors,
})
```

### 6.4 Swagger アノテーション

```go
// GetOwner godoc
// @Summary      飼主取得
// @Description  IDを指定して飼主情報を取得
// @Tags         owners
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "飼主ID"
// @Success      200  {object}  model.Owner
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /owners/{id} [get]
func (h *OwnerHandler) GetOwner(c *gin.Context) {
    // ...
}
```

---

## 7. モデル定義

### 7.1 基本構造

```go
// internal/model/owner.go
package model

import (
    "time"

    "github.com/google/uuid"
    "gorm.io/gorm"
)

type Owner struct {
    ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
    Name      string         `gorm:"type:varchar(100);not null" json:"name"`
    NameKana  string         `gorm:"type:varchar(100)" json:"nameKana"`
    Email     string         `gorm:"type:varchar(255);uniqueIndex" json:"email"`
    Phone     string         `gorm:"type:varchar(20)" json:"phone"`
    Address   string         `gorm:"type:text" json:"address"`
    Pets      []Pet          `gorm:"foreignKey:OwnerID" json:"pets,omitempty"`
    CreatedAt time.Time      `json:"createdAt"`
    UpdatedAt time.Time      `json:"updatedAt"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName テーブル名を明示的に指定
func (Owner) TableName() string {
    return "owners"
}

// BeforeCreate UUID自動生成
func (o *Owner) BeforeCreate(tx *gorm.DB) error {
    if o.ID == uuid.Nil {
        o.ID = uuid.New()
    }
    return nil
}
```

### 7.2 リレーション

```go
// 1対多: Owner has many Pets
type Owner struct {
    ID   uuid.UUID `gorm:"type:uuid;primary_key"`
    Name string
    Pets []Pet `gorm:"foreignKey:OwnerID"`
}

type Pet struct {
    ID      uuid.UUID `gorm:"type:uuid;primary_key"`
    Name    string
    OwnerID uuid.UUID `gorm:"type:uuid;not null"`
    Owner   Owner     `gorm:"foreignKey:OwnerID"`
}

// 多対多: Pet has many MedicalRecords, MedicalRecord has many Pets
type Pet struct {
    ID             uuid.UUID        `gorm:"type:uuid;primary_key"`
    MedicalRecords []MedicalRecord  `gorm:"many2many:pet_medical_records;"`
}
```

### 7.3 JSONタグ

```go
type Owner struct {
    ID        uuid.UUID `json:"id"`                    // キャメルケース
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    Password  string    `json:"-"`                     // レスポンスに含めない
    CreatedAt time.Time `json:"createdAt"`
    DeletedAt time.Time `json:"deletedAt,omitempty"`   // 空なら省略
}
```

---

## 8. 命名規則

### 8.1 パッケージ名

```go
// ✅ lowercase, 短く、説明的
package handler
package service
package repository
package model
package middleware

// ❌ 禁止
package ownerHandler    // キャメルケース禁止
package owner_handler   // スネークケース禁止
package handlers        // 複数形は避ける
```

### 8.2 ファイル名

```go
// ✅ snake_case
owner_handler.go
pet_service.go
medical_record_repository.go

// ❌ 禁止
ownerHandler.go         // キャメルケース禁止
owner-handler.go        // ケバブケース禁止
```

### 8.3 変数・関数名

```go
// エクスポート（公開）: PascalCase
func GetOwner() {}
type OwnerService struct {}
var MaxRetryCount = 3

// 非エクスポート（非公開）: camelCase
func validateInput() {}
type ownerRepository struct {}
var defaultTimeout = 5 * time.Second

// 定数: PascalCase or UPPER_SNAKE_CASE
const MaxPageSize = 100
const DB_TIMEOUT = 30 * time.Second

// インターフェース: 動詞+er or 名詞
type Reader interface {}
type OwnerRepository interface {}
type OwnerService interface {}
```

### 8.4 レシーバ名

```go
// ✅ 短い1-2文字
func (h *OwnerHandler) GetOwner() {}
func (s *ownerService) GetOwner() {}
func (r *ownerRepository) FindByID() {}
func (o *Owner) Validate() error {}

// ❌ 禁止
func (handler *OwnerHandler) GetOwner() {}  // 長い
func (self *OwnerHandler) GetOwner() {}     // self禁止
func (this *OwnerHandler) GetOwner() {}     // this禁止
```

---

## 9. テスト

### 9.1 ファイル配置

```
internal/
├── service/
│   ├── owner_service.go
│   └── owner_service_test.go    # 同パッケージ内
├── repository/
│   ├── owner_repository.go
│   └── owner_repository_test.go
```

### 9.2 テスト命名

```go
// 関数名: Test + 関数名 + 条件
func TestOwnerService_GetOwner(t *testing.T) {}
func TestOwnerService_GetOwner_NotFound(t *testing.T) {}
func TestOwnerService_CreateOwner_ValidationError(t *testing.T) {}

// テーブル駆動テスト
func TestOwnerService_GetOwner(t *testing.T) {
    tests := []struct {
        name      string
        id        string
        want      *model.Owner
        wantErr   error
    }{
        {
            name:    "returns owner when found",
            id:      "valid-uuid",
            want:    &model.Owner{Name: "Test"},
            wantErr: nil,
        },
        {
            name:    "returns error when not found",
            id:      "invalid-uuid",
            want:    nil,
            wantErr: apperrors.ErrNotFound,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // setup
            // execute
            // assert
        })
    }
}
```

### 9.3 モック

```go
// モックの定義
type mockOwnerRepository struct {
    findByIDFunc func(ctx context.Context, id uuid.UUID) (*model.Owner, error)
}

func (m *mockOwnerRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Owner, error) {
    return m.findByIDFunc(ctx, id)
}

// テストでの使用
func TestOwnerService_GetOwner(t *testing.T) {
    mockRepo := &mockOwnerRepository{
        findByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.Owner, error) {
            return &model.Owner{ID: id, Name: "Test"}, nil
        },
    }

    service := NewOwnerService(mockRepo)
    owner, err := service.GetOwner(context.Background(), "valid-uuid")

    assert.NoError(t, err)
    assert.Equal(t, "Test", owner.Name)
}
```

---

## 10. 禁止事項一覧

| 禁止 | 理由 | 代替 |
|------|------|------|
| `_ = err` | エラー握りつぶし | 適切にハンドリング |
| `panic` の乱用 | 予期せぬクラッシュ | error返却 |
| グローバル変数 | 状態管理の複雑化 | 依存性注入 |
| init() の乱用 | テスト困難 | 明示的初期化 |
| Context省略 | キャンセル・タイムアウト不可 | 第一引数に常に渡す |
| SQL文字列結合 | SQLインジェクション | プレースホルダ |
| 機密情報ログ出力 | セキュリティリスク | マスクまたは出力しない |

---

## 11. チェックリスト

### 新規機能開発時
- [ ] Context伝播している
- [ ] エラーをセンチネルエラーでラップしている
- [ ] slogで構造化ログ出力している
- [ ] SQLインジェクション対策している
- [ ] N+1問題が発生していない
- [ ] テストを書いている

### PR作成時
- [ ] `golangci-lint run ./...` がパス
- [ ] `go test ./... -v` がパス
- [ ] Swagger更新済み
- [ ] 不要なコメントアウトがない
- [ ] デバッグ用コードが残っていない

---

## 12. 参照

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [Gin Web Framework](https://gin-gonic.com/docs/)
- [GORM](https://gorm.io/docs/)
