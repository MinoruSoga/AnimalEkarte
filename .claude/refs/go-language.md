---
description: Go 1.25 coding standards (error handling, context propagation, concurrency)
alwaysApply: true
globs: ["backend/**/*.go"]
---

# Go Language Rules

Go 1.25 project standard rules.

## Core Rules

### 1. Context Propagation (Required)

Place `context.Context` as first argument in all functions.

```go
// ✅ Correct
func (r *OwnerRepository) GetByID(ctx context.Context, id uint64) (*Owner, error) {
  return r.db.WithContext(ctx).First(&owner, id).Error
}

// ❌ Prohibited: No Context
func (r *OwnerRepository) GetByID(id uint64) (*Owner, error) {
  return r.db.First(&owner, id).Error
}
```

### 2. Error Handling (Sentinel + Wrap)

```go
// Define Sentinel errors in errors/errors.go
var (
  ErrNotFound = errors.New("not found")
  ErrConflict = errors.New("conflict")
)

// Wrap in service layer
if err := repo.Create(ctx, owner); err != nil {
  if errors.Is(err, ErrConflict) {
    return fmt.Errorf("owner already exists: %w", err)
  }
  return fmt.Errorf("failed to create owner: %w", err)
}

// Call RespondError() in handler layer
RespondError(c, err)
```

### 3. Interface Design

```go
// Minimize: only necessary operations
type OwnerRepository interface {
  GetByID(ctx context.Context, id uint64) (*Owner, error)
  Create(ctx context.Context, owner *Owner) error
  Update(ctx context.Context, owner *Owner) error
  Delete(ctx context.Context, id uint64) error
}

// ❌ Prohibited: Massive interface
type Repository interface {
  GetByID() (*Owner, error)
  Create() error
  Update() error
  Delete() error
  GetAll() ([]Owner, error)
  Count() (int64, error)
  // ... 20 more methods
}
```

### 4. GORM PATCH (Pointer types + buildUpdateFields)

```go
// service/owner_service.go
type UpdateOwnerInput struct {
  Name  *string `json:"name"`
  Email *string `json:"email"`
}

func (s *OwnerService) Update(ctx context.Context, id uint64, input UpdateOwnerInput) (*Owner, error) {
  // Avoid zero value problem
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

### 5. Concurrency (errgroup)

```go
import "golang.org/x/sync/errgroup"

// Run multiple independent operations in parallel
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

### 6. Logging (slog structured logging)

Logging only in service layer. No handler or repository logging.

**Exceptions (allowed in handler layer):**
- `response.go` `RespondError` internal server error logging (HTTP infrastructure concern)
- Audit log write failures with best-effort error logging (`auth_handler.go`, `permission_group_handler.go`)
- File cleanup failure warnings (`medical_record_image_handler.go`)

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

### 7. Naming Conventions

| Target | Rule | Example |
|--------|------|---------|
| Package | lowercase (single word) | `handler`, `repository` |
| Exported | PascalCase | `GetOwner`, `OwnerService` |
| Unexported | camelCase | `getOwnerByID` |
| File | snake_case | `owner_handler.go` |
| Interface | PascalCase + er | `OwnerRepository` |
| Table | snake_case (plural) | `owners`, `medical_records` |

### 8. Package Layout（BE8 規約・2026-07-17 決定 — 正本 = `.claude/skills/go-package-conventions/SKILL.md`（本節は要約））

**目標構成**（層優先 × ドメインサブパッケージ）:

```
backend/internal/
  handler/                 # 現状フラット維持（分割は BE-refactor.md BE8-7 で service 完了後に判断）
  service/
    <domain>/              # 例: service/reservation/ — 新規ドメインはここに作る
  repository/
    repohelpers/           # 共有 clinic-scope / DBOrTx ヘルパ
    <domain>/              # 例: repository/paymentmethod/（先行 14 ドメイン分割済み）
  model/                   # 単一パッケージ維持（FK 相互参照のため分割しない — 決定事項）
  middleware/ infra/ config/ ...  # 健全・変更不要
```

出典: [Go 公式 module layout](https://go.dev/doc/modules/layout)（server logic は internal/ 配下のドメイン名パッケージ）・[Google Go Style: Best Practices](https://google.github.io/styleguide/go/best-practices)（util 禁止・stutter 禁止）。採用理由と不採用案（ドメイン優先 Option A・pkg/ 新設）は `BE-refactor.md` §2-3（対応後削除・記録は git 履歴）。

- **層優先 × ドメインサブパッケージ**: 新規ドメインの repository / service 実装はフラット直下に置かず、`internal/<layer>/<domain>/` のサブパッケージで作る（先例: `repository/paymentmethod`）。
- パッケージ名 = 単数形・全小文字・アンダースコアなしのドメイン名。`util` / `common` / `helpers` 単独名は禁止（`repohelpers` は既存例外）。
- **stutter 禁止**: 新規は `reservation.NewRepository` 形。既存型の公開リネームは移動と同時にやらない（別コミット）。
- service ↔ service のドメイン間参照は **consumer 側で小文字ローカル interface** を定義して受ける（先例: `reservation_service.go` の `reservationTypeFinder`）。import cycle は interface 抽出で解決し、移動の巻き戻しはしない。
- **サブパッケージ内にさらにディレクトリを掘らない** — repository の走査 lint 群は `go:embed *.go */*.go` の 1 階層走査で、2 階層目は不可視（サイレント緑）になる。
- 既存フラットファイルは実装変更で触るときに移す（strangler）。一斉移動は禁止。

## Checklist

- [ ] All functions have `ctx context.Context` argument
- [ ] Sentinel errors + `fmt.Errorf("...: %w", err)` wrapping
- [ ] Logging only in service layer
- [ ] PATCH uses pointer types + buildUpdateFields()
- [ ] errgroup for parallel processing
- [ ] Interfaces minimized (3-5 methods)
- [ ] 新規ドメインの repository/service はサブパッケージで作成（§8）

## Architecture Compliance

Layer-specific rules (P1–P18) are defined in `gin-architecture-compliance.md`.
Run compliance checks when implementing or reviewing handler/service/repository code.
