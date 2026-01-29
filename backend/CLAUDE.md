# Backend - Go 1.22+ / Gin / GORM（クリーンアーキテクチャ）

## ⚠️ コマンド実行ルール

**goコマンドはローカル実行禁止。必ずDocker経由で実行する。**

```bash
# ❌ NG
go test ./...

# ✅ OK
docker compose exec backend go test ./...
```

## コマンド一覧

| タスク | コマンド |
|--------|---------|
| テスト | `docker compose exec backend go test ./... -v` |
| カバレッジ | `docker compose exec backend go test ./... -cover` |
| Lint | `docker compose exec backend golangci-lint run ./...` |
| モジュール更新 | `docker compose exec backend go mod tidy` |
| Swagger生成 | `docker compose exec backend swag init -g cmd/api/main.go` |

---

## 技術スタック

| 技術 | 用途 |
|------|------|
| Go 1.22+ | 言語 |
| Gin | Webフレームワーク |
| GORM | ORM |
| slog | 構造化ログ |
| Air | ホットリロード |
| PostgreSQL 18 | データベース |

---

## ディレクトリ構造（クリーンアーキテクチャ）

```
backend/
├── cmd/
│   └── api/
│       └── main.go             # エントリーポイント
│
├── internal/                   # 内部パッケージ
│   ├── config/
│   │   └── config.go           # 環境変数・設定
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
│   ├── logger/                 # slogロガー設定
│   │   └── logger.go
│   │
│   └── validation/             # バリデーション
│       └── *.go
│
├── migrations/                 # DBマイグレーション
│   ├── 001_create_owners.sql
│   └── ...
│
└── docs/                       # Swagger
    ├── docs.go
    └── swagger.yaml
```

---

## 依存関係の方向

```
handler (プレゼンテーション層)
    ↓
service (ビジネスロジック層)
    ↓
repository (データアクセス層)
    ↓
model (ドメインモデル)
```

**ルール**: 上位層は下位層にのみ依存。逆方向禁止。

---

## Context伝播（必須）

**全ての関数・メソッドの第一引数に `context.Context` を渡す。**

```go
// Handler
func (h *OwnerHandler) GetOwner(c *gin.Context) {
    ctx := c.Request.Context()
    owner, err := h.service.GetOwner(ctx, id)
}

// Service
func (s *ownerService) GetOwner(ctx context.Context, id string) (*model.Owner, error) {
    return s.repo.FindByID(ctx, uuid.MustParse(id))
}

// Repository
func (r *ownerRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Owner, error) {
    var owner model.Owner
    if err := r.db.WithContext(ctx).First(&owner, "id = ?", id).Error; err != nil {
        // ...
    }
    return &owner, nil
}
```

---

## エラーハンドリング

### センチネルエラー定義

```go
// internal/errors/errors.go
var (
    ErrNotFound      = errors.New("resource not found")
    ErrInvalidInput  = errors.New("invalid input")
    ErrUnauthorized  = errors.New("unauthorized")
    ErrForbidden     = errors.New("forbidden")
    ErrConflict      = errors.New("resource conflict")
)
```

### エラーラッピング

```go
func Wrap(err error, message string) error {
    if err == nil {
        return nil
    }
    return fmt.Errorf("%s: %w", message, err)
}

// 使用例
if err := r.db.Create(&owner).Error; err != nil {
    return apperrors.Wrap(err, "failed to create owner")
}
```

### エラー判定

```go
if errors.Is(err, apperrors.ErrNotFound) {
    c.JSON(http.StatusNotFound, gin.H{"error": "飼主が見つかりません"})
    return
}
```

---

## slog構造化ログ

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
    slog.String("endpoint", path))
```

### ログ禁止事項

```go
// ❌ パスワード、トークン等の機密情報をログ出力しない
slog.Info("login", slog.String("password", password))  // ❌

// ✅ 機密情報は出力しない
slog.Info("login", slog.String("email", email))  // ✅
```

---

## GORM

### N+1問題回避

```go
// ❌ N+1問題
for _, owner := range owners {
    r.db.Where("owner_id = ?", owner.ID).Find(&owner.Pets)
}

// ✅ Preloadで解決
r.db.WithContext(ctx).Preload("Pets").Find(&owners)
```

### トランザクション

```go
return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&owner).Error; err != nil {
        return apperrors.Wrap(err, "failed to create owner")
    }
    for _, pet := range owner.Pets {
        if err := tx.Create(&pet).Error; err != nil {
            return apperrors.Wrap(err, "failed to create pet")
        }
    }
    return nil  // コミット
})
```

### SQLインジェクション対策

```go
// ❌ 禁止: 文字列結合
r.db.Raw("SELECT * FROM owners WHERE name = '" + name + "'")

// ✅ 必須: プレースホルダ
r.db.Where("name = ?", name).Find(&owners)
```

---

## 命名規則

| 対象 | 規則 | 例 |
|------|------|-----|
| パッケージ | lowercase | `handler`, `repository` |
| エクスポート | PascalCase | `GetOwner`, `OwnerService` |
| 非エクスポート | camelCase | `validateInput` |
| ファイル | snake_case | `owner_handler.go` |
| インターフェース | PascalCase + er | `OwnerRepository` |
| レシーバ | 1-2文字 | `h`, `s`, `r`, `o` |

---

## 禁止事項

| 禁止 | 理由 | 代替 |
|------|------|------|
| `_ = err` | エラー握りつぶし | 適切にハンドリング |
| `panic` 乱用 | 予期せぬクラッシュ | error返却 |
| グローバル変数 | 状態管理の複雑化 | 依存性注入 |
| `init()` 乱用 | テスト困難 | 明示的初期化 |
| Context省略 | キャンセル不可 | 第一引数に常に渡す |
| SQL文字列結合 | SQLインジェクション | プレースホルダ |

---

## テスト

### ファイル配置

```
internal/
├── service/
│   ├── owner_service.go
│   └── owner_service_test.go    # 同パッケージ内
```

### テスト命名

```go
func TestOwnerService_GetOwner(t *testing.T) {}
func TestOwnerService_GetOwner_NotFound(t *testing.T) {}
```

### テーブル駆動テスト

```go
func TestOwnerService_GetOwner(t *testing.T) {
    tests := []struct {
        name    string
        id      string
        want    *model.Owner
        wantErr error
    }{
        {
            name:    "returns owner when found",
            id:      "valid-uuid",
            want:    &model.Owner{Name: "Test"},
            wantErr: nil,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // setup, execute, assert
        })
    }
}
```

---

## 参照

| ドキュメント | 説明 |
|-------------|------|
| [詳細コーディング規約](./CODING_RULES.md) | 完全なルール集 |
| [プロジェクト全体](../CODING_RULES.md) | 共通ルール |
| [Effective Go](https://go.dev/doc/effective_go) | Go公式ガイド |
| [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md) | スタイルガイド |
| [Gin](https://gin-gonic.com/docs/) | Webフレームワーク |
| [GORM](https://gorm.io/docs/) | ORM |
