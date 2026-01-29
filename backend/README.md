# Animal Ekarte Backend

動物病院 電子カルテシステムのバックエンドAPI

## 技術スタック

- **言語**: Go 1.22+
- **フレームワーク**: Gin v1.10
- **ORM**: GORM v1.30
- **データベース**: PostgreSQL 18
- **API仕様**: OpenAPI/Swagger
- **ホットリロード**: Air
- **ロギング**: slog（構造化ログ）
- **リンター**: golangci-lint

## ディレクトリ構成

```
backend/
├── cmd/
│   └── api/
│       └── main.go          # エントリーポイント
├── internal/
│   ├── config/
│   │   └── config.go        # 環境設定
│   ├── errors/
│   │   └── errors.go        # エラー定義（センチネルエラー）
│   ├── handler/
│   │   ├── handler.go       # ルーティング・共通ハンドラー
│   │   └── pet.go           # ペットCRUDハンドラー
│   ├── logger/
│   │   └── logger.go        # slog構造化ロガー
│   ├── middleware/
│   │   └── *.go             # ミドルウェア（認証、CORS等）
│   ├── model/
│   │   └── pet.go           # データモデル・リクエスト型
│   ├── repository/
│   │   ├── db.go            # DB接続
│   │   ├── repository.go    # リポジトリ基底
│   │   └── pet.go           # ペットリポジトリ
│   ├── service/
│   │   ├── service.go       # サービス基底
│   │   └── pet.go           # ペットビジネスロジック
│   └── validation/
│       └── *.go             # バリデーション
├── migrations/              # DBマイグレーション
├── docs/                    # Swagger生成ドキュメント（自動生成）
├── .golangci.yml            # リンター設定
├── Dockerfile               # 本番用
├── Dockerfile.dev           # 開発用
├── entrypoint.sh            # コンテナ起動スクリプト
├── .air.toml                # ホットリロード設定
├── go.mod                   # 依存関係
└── go.sum                   # 依存関係ロック
```

## 開発環境セットアップ

### 前提条件

- Go 1.22+以上
- PostgreSQL 18（またはDocker）

### ローカル開発（Docker使用）

プロジェクトルートから:

```bash
# コンテナ起動
make build

# ログ確認
make logs-api

# API再起動
make restart-api
```

### ローカル開発（Docker不使用）

```bash
# 依存関係インストール
go mod download

# Swagger CLIインストール
go install github.com/swaggo/swag/cmd/swag@latest

# Swaggerドキュメント生成
swag init -g cmd/api/main.go -o docs

# 環境変数設定
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=ekarte_user
export DB_PASSWORD=ekarte_password
export DB_NAME=ekarte_db

# 実行
go run cmd/api/main.go

# または Air でホットリロード
go install github.com/air-verse/air@latest
air -c .air.toml
```

## API エンドポイント

| メソッド | パス | 説明 |
|---------|------|------|
| GET | /health | ヘルスチェック |
| GET | /api/v1/ | ウェルカムメッセージ |
| GET | /api/v1/pets | ペット一覧取得 |
| GET | /api/v1/pets/:id | ペット詳細取得 |
| POST | /api/v1/pets | ペット新規登録 |
| PUT | /api/v1/pets/:id | ペット情報更新 |
| DELETE | /api/v1/pets/:id | ペット削除 |

## Swagger UI

開発サーバー起動後、以下のURLでAPIドキュメントを確認できます:

```
http://localhost:8080/swagger/index.html
```

## 新しいCRUD機能の追加手順

### 1. モデル作成

`internal/model/` に新しいモデルファイルを作成:

```go
// internal/model/owner.go
package model

import (
    "time"
    "github.com/google/uuid"
)

type Owner struct {
    ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
    Name      string    `json:"name" gorm:"size:100;not null"`
    Phone     string    `json:"phone" gorm:"size:20"`
    Email     string    `json:"email" gorm:"size:255"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type CreateOwnerRequest struct {
    Name  string `json:"name" binding:"required"`
    Phone string `json:"phone"`
    Email string `json:"email"`
}
```

### 2. リポジトリ作成

`internal/repository/` にリポジトリを作成:

```go
// internal/repository/owner.go
package repository

import (
    "github.com/animal-ekarte/backend/internal/model"
    "github.com/google/uuid"
)

func (r *Repository) GetAllOwners() ([]model.Owner, error) {
    var owners []model.Owner
    result := r.db.Find(&owners)
    return owners, result.Error
}

func (r *Repository) CreateOwner(owner *model.Owner) error {
    return r.db.Create(owner).Error
}
// ... 他のCRUDメソッド
```

### 3. サービス作成

`internal/service/` にビジネスロジックを作成:

```go
// internal/service/owner.go
package service

import "github.com/animal-ekarte/backend/internal/model"

func (s *Service) GetAllOwners() ([]model.Owner, error) {
    return s.repo.GetAllOwners()
}

func (s *Service) CreateOwner(req *model.CreateOwnerRequest) (*model.Owner, error) {
    owner := &model.Owner{
        Name:  req.Name,
        Phone: req.Phone,
        Email: req.Email,
    }
    err := s.repo.CreateOwner(owner)
    return owner, err
}
```

### 4. ハンドラー作成

`internal/handler/` にハンドラーを作成（Swaggerアノテーション付き）:

```go
// internal/handler/owner.go
package handler

import (
    "net/http"
    "github.com/animal-ekarte/backend/internal/model"
    "github.com/gin-gonic/gin"
)

// GetOwners godoc
// @Summary 飼い主一覧取得
// @Description 登録されている飼い主の一覧を取得します
// @Tags owners
// @Produce json
// @Success 200 {array} model.Owner
// @Router /owners [get]
func (h *Handler) GetOwners(c *gin.Context) {
    owners, err := h.svc.GetAllOwners()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, owners)
}
```

### 5. ルート登録

`internal/handler/handler.go` の `RegisterRoutes` にルートを追加:

```go
// Owners CRUD
v1.GET("/owners", h.GetOwners)
v1.POST("/owners", h.CreateOwner)
```

### 6. マイグレーション

`cmd/api/main.go` の `AutoMigrate` にモデルを追加:

```go
db.AutoMigrate(&model.Pet{}, &model.Owner{})
```

### 7. Swagger再生成

```bash
swag init -g cmd/api/main.go -o docs
```

## テスト

```bash
# 全テスト実行
go test ./...

# カバレッジ付き
go test -cover ./...

# 特定パッケージ
go test ./internal/service/...
```

## 本番ビルド

```bash
# Dockerイメージビルド
docker build -t animal-ekarte-api:latest .

# バイナリビルド
CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api
```

## 環境変数

| 変数名 | 説明 | デフォルト値 |
|-------|------|-------------|
| PORT | APIポート | 8080 |
| DB_HOST | DBホスト | localhost |
| DB_PORT | DBポート | 5432 |
| DB_USER | DBユーザー | ekarte_user |
| DB_PASSWORD | DBパスワード | ekarte_password |
| DB_NAME | DB名 | ekarte_db |
| GIN_MODE | Ginモード | debug |
| LOG_LEVEL | ログレベル | info |

## コーディングパターン

### エラーハンドリング

センチネルエラーとエラーラッピングを使用:

```go
import apperrors "github.com/animal-ekarte/backend/internal/errors"

// エラー発生時はラッピング
if err != nil {
    return nil, apperrors.Wrap(err, "failed to get pet")
}

// NotFound エラー
return nil, apperrors.WrapNotFound("pet", id.String())

// バリデーションエラー
return nil, apperrors.WrapInvalidInput("pet name is required")

// エラー判定
if apperrors.IsNotFound(err) {
    c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
}
```

### 構造化ログ（slog）

```go
import "log/slog"

// 基本的なログ
slog.Info("server starting", slog.String("port", cfg.Port))

// コンテキスト付きログ
slog.InfoContext(ctx, "pet created", slog.String("pet_id", pet.ID.String()))
slog.ErrorContext(ctx, "failed to create pet", slog.String("error", err.Error()))
```

### Context伝播

全てのレイヤーで `context.Context` を第一引数に:

```go
// Service層
func (s *Service) GetPetByID(ctx context.Context, id string) (*model.Pet, error)

// Repository層
func (r *Repository) GetPetByID(ctx context.Context, id uuid.UUID) (*model.Pet, error) {
    var pet model.Pet
    result := r.db.WithContext(ctx).First(&pet, "id = ?", id)
    return &pet, result.Error
}
```

## リンター

golangci-lint でコード品質をチェック:

```bash
# インストール
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 実行
golangci-lint run

# 自動修正
golangci-lint run --fix
```
