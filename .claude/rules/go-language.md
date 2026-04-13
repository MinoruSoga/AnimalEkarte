---
description: Go言語コーディング規約（エラー処理、Context伝播、並行性）
alwaysApply: true
globs: ["backend/**/*.go"]
---

# Go Language Rules

Go 1.25 プロジェクト標準ルール。

## 核心ルール

### 1. Context伝播（必須）

全関数の第一引数に `context.Context` を配置。

```go
// ✅ 正しい
func (r *OwnerRepository) GetByID(ctx context.Context, id uint64) (*Owner, error) {
  return r.db.WithContext(ctx).First(&owner, id).Error
}

// ❌ 禁止: Context なし
func (r *OwnerRepository) GetByID(id uint64) (*Owner, error) {
  return r.db.First(&owner, id).Error
}
```

### 2. エラーハンドリング（Sentinel + Wrap）

```go
// errors/errors.go で Sentinel 定義
var (
  ErrNotFound = errors.New("not found")
  ErrConflict = errors.New("conflict")
)

// service層で Wrap
if err := repo.Create(ctx, owner); err != nil {
  if errors.Is(err, ErrConflict) {
    return fmt.Errorf("owner already exists: %w", err)
  }
  return fmt.Errorf("failed to create owner: %w", err)
}

// handler層で RespondError()
RespondError(c, err)
```

### 3. インターフェース設計

```go
// 最小化: 必要な操作のみ
type OwnerRepository interface {
  GetByID(ctx context.Context, id uint64) (*Owner, error)
  Create(ctx context.Context, owner *Owner) error
  Update(ctx context.Context, owner *Owner) error
  Delete(ctx context.Context, id uint64) error
}

// ❌ 禁止: 巨大インターフェース
type Repository interface {
  GetByID() (*Owner, error)
  Create() error
  Update() error
  Delete() error
  GetAll() ([]Owner, error)
  Count() (int64, error)
  // ... 20個のメソッド
}
```

### 4. GORM PATCH（ポインタ型 + buildUpdateFields）

```go
// service/owner_service.go
type UpdateOwnerInput struct {
  Name  *string `json:"name"`
  Email *string `json:"email"`
}

func (s *OwnerService) Update(ctx context.Context, id uint64, input UpdateOwnerInput) (*Owner, error) {
  // ゼロ値問題を回避
  fields := buildOwnerUpdateFields(input)
  return s.repo.UpdateFields(ctx, id, fields)
}

// repository/owner_repository.go
func (r *OwnerRepository) UpdateFields(ctx context.Context, id uint64, fields map[string]any) (*Owner, error) {
  var owner Owner
  return &owner, r.db.WithContext(ctx).Model(&Owner{}).Where("id = ?", id).Updates(fields).First(&owner).Error
}

func buildOwnerUpdateFields(input UpdateOwnerInput) map[string]any {
  fields := make(map[string]any)
  if input.Name != nil {
    fields["name"] = *input.Name
  }
  if input.Email != nil {
    fields["email"] = *input.Email
  }
  return fields
}
```

### 5. 並行処理（errgroup）

```go
import "golang.org/x/sync/errgroup"

// 複数の独立した操作を並列実行
g, ctx := errgroup.WithContext(ctx)

g.Go(func() error {
  return r.saveOwner(ctx, owner)
})

g.Go(func() error {
  return r.savePets(ctx, owner.ID)
})

if err := g.Wait(); err != nil {
  return fmt.Errorf("failed to save data: %w", err)
}
```

### 6. ログ（slog構造化ログ）

原則 service 層のみ。handler・repository には記述しない。

**例外（handler 層で許容）:**
- `response.go` の `RespondError` 内部サーバーエラーログ（HTTP インフラ関心事）
- 監査ログ書き込み失敗時の best-effort エラーログ（`auth_handler.go`, `permission_group_handler.go`）
- ファイルクリーンアップ失敗時の警告ログ（`medical_record_image_handler.go`）

```go
import "log/slog"

func (s *OwnerService) Create(ctx context.Context, input CreateOwnerInput) (*Owner, error) {
  slog.InfoContext(ctx, "creating owner", "name", input.Name, "email", input.Email)

  owner, err := s.repo.Create(ctx, &Owner{Name: input.Name, Email: input.Email})
  if err != nil {
    slog.ErrorContext(ctx, "failed to create owner", "error", err)
    return nil, fmt.Errorf("failed to create owner: %w", err)
  }

  slog.InfoContext(ctx, "owner created", "id", owner.ID)
  return owner, nil
}
```

### 7. 命名規則

| 対象 | 規則 | 例 |
|------|------|-----|
| パッケージ | lowercase（1単語） | `handler`, `repository` |
| エクスポート | PascalCase | `GetOwner`, `OwnerService` |
| 非エクスポート | camelCase | `getOwnerByID` |
| ファイル | snake_case | `owner_handler.go` |
| インターフェース | PascalCase + er | `OwnerRepository` |
| テーブル | snake_case（複数形） | `owners`, `medical_records` |

## チェックリスト

- [ ] すべての関数に `ctx context.Context` 引数
- [ ] Sentinel エラー + fmt.Errorf("...: %w", err) で Wrap
- [ ] service層のみ slog 記述
- [ ] PATCH は ポインタ型 + buildUpdateFields()
- [ ] errgroup で並列処理
- [ ] インターフェース最小化（3-5メソッド）
