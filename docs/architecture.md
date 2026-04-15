# Backend Architecture

## 採用方針

**軽量レイヤードアーキテクチャ**

67テーブル規模・Gin WebAPI・最速実装を前提に、Clean Architecture の過剰な抽象化を避けつつ、責務分離と依存方向の一貫性を確保する。

---

## フォルダ構成

```
backend/
├── cmd/
│   └── api/
│       └── main.go              # エントリーポイント + DI配線
├── internal/
│   ├── handler/                 # HTTP受付層（binding, response 定義含む）
│   ├── service/                 # 業務処理層（インターフェース、バリデータ含む）
│   ├── repository/              # データアクセス層（GORM）
│   ├── model/                   # GORMモデル（単一の真実）
│   ├── errors/                  # センチネルエラーと AppError 型（handler では apperrors と別名定義）
│   ├── middleware/              # 認証・ログ・CORS
│   └── infra/                   # 外部連携（LINE, Storage 等）
├── migrations/                  # SQLマイグレーション
├── go.mod
└── go.sum
```

---

## 各層の責務

### handler/

**責務**: HTTP の受け取りと返却のみ。

- `*gin.Context` を扱うのはこの層だけ
- `RespondError(c, err)` を使用した統一エラーレスポンス（`internal/errors` パッケージ連動）
- `ShouldBindJSON` エラー時は `RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))` を徹底（`apperrors` は `internal/errors` の別名）
- `RegisterXxxRoutes(rg *gin.RouterGroup)` でルートを自己登録

**禁止事項**:
- SQL / GORM を書かない
- `repository` を直接呼ばない

### service/

**責務**: 業務処理の中核。

- Service インターフェースと DTO (Input) をこの層で定義
- 業務バリデーションを `validators.go` に集約
- 削除時は `CountUsageByXxxID()` 等で外部キー整合性をチェックし、使用中の場合は `apperrors.WrapConflict` を返却
- slog による構造化ログ

### repository/

**責務**: DB アクセスの永続化処理のみ。

- Repository インターフェースをこの層で定義
- DB エラーをセンチネルエラーに変換（`gorm.ErrRecordNotFound` → `apperrors.WrapNotFound`）

### model/

**責務**: GORM モデル（DB スキーマ対応）。

- **Single Source of Truth**: フロントエンドの型定義（`models.ts`）の生成元
- `json:"-"` で内部フィールドを非公開

---

## フロントエンド・アーキテクチャ

React 19 + TypeScript + Vite 6 + Tailwind 4。

- **Feature-Based 構造**: `src/features/[feature]/` にカプセル化（api, components, hooks, routes, types）
- **Dependency Inversion**: Feature 間 import は禁止。`src/app/pages/` でのみ合成。
- **React 19 パターン**: `useActionState` (Form), `useTransition` (Pending), `useDeferredValue` (Performance)。
- **中央ルーター**: React Router 7 Data Mode (`createBrowserRouter`)。

```go
// handler/owner_handler.go
func (h *Handler) CreateOwner(c *gin.Context) {
    clinicID, ok := extractClinicID(c)
    if !ok {
        return
    }

    var req createOwnerRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
        return
    }

    input := service.CreateOwnerInput{
        OwnerName: req.OwnerName,
        Email:     req.Email,
        // ... 他のフィールド
    }

    owner, err := h.svc.Owner.CreateWithPets(c.Request.Context(), clinicID, &input)
    if err != nil {
        RespondError(c, err)
        return
    }

    c.JSON(http.StatusCreated, toOwnerResponse(owner))
}
```
func (h *Handler) RegisterOwnerRoutes(rg *gin.RouterGroup) {
    owners := rg.Group("/owners")
    owners.GET("", h.ListOwners)
    owners.POST("", h.CreateOwner)
    owners.GET("/:id", h.GetOwner)
    owners.PATCH("/:id", h.UpdateOwner)
    owners.DELETE("/:id", h.DeleteOwner)
}
```

### handler/*_request.go

**責務**: HTTP 入力の binding 定義。handler 専用の型。

- `binding:"required"` 等の構造的バリデーションタグを持つ
- 別パッケージに切り出さない（handler の都合に閉じる）

```go
// handler/owner_request.go
type createOwnerRequest struct {
    OwnerName string `json:"owner_name" binding:"required"`
    Email     string `json:"email"`
    // ...
}

type updateOwnerRequest struct {
    OwnerName *string `json:"owner_name"`
    Email     *string `json:"email"`
    // ...
}
```

### handler/*_response.go

**責務**: API レスポンス型の定義と model からの変換。

- `model.Owner` を直接 JSON 出力しない（DBスキーマとAPI契約を分離）
- `json:"-"` フィールドの露出を防ぐ

```go
// handler/owner_response.go
type ownerResponse struct {
    ID        string `json:"id"`
    OwnerName string `json:"owner_name"`
    Email     string `json:"email"`
    // ...
}

func toOwnerResponse(o *model.Owner) ownerResponse {
    return ownerResponse{
        ID:        strconv.FormatUint(o.ID, 10),
        OwnerName: o.OwnerName,
        Email:     o.Email,
    }
}
```

### service/

**責務**: 業務処理の中核。

- HTTP を知らない（`*gin.Context` を受け取らない）
- `binding:` タグを持たない
- Service インターフェースをこの層で定義
- Service 専用の Input 型（DTO）をこの層で定義
- 業務バリデーションを `validators.go` に集約
- slog による構造化ログ

```go
// service/owner_service.go
type OwnerService interface {
    List(ctx context.Context, clinicID uint64, page, limit int, search string) ([]model.Owner, int64, error)
    GetByID(ctx context.Context, clinicID, id uint64) (*model.Owner, error)
    CreateWithPets(ctx context.Context, clinicID uint64, input *CreateOwnerInput) (*model.Owner, error)
    Update(ctx context.Context, clinicID, id uint64, input *UpdateOwnerInput) (*model.Owner, error)
    Delete(ctx context.Context, clinicID, id uint64) error
}

type CreateOwnerInput struct {
    OwnerName string  // binding タグなし
    Email     string
    // ...
}
```

### service/validators.go

**責務**: 業務バリデーション関数の集約。

- HTTP 入力チェック（binding タグ）とは分離
- 業務ルール・複数フィールド整合性・状態遷移チェック
- `validators.go` 1ファイルに限定（共通ゴミ箱化を防ぐ）

```go
// service/validators.go
func validateDiscountRate(rate float64) error {
    if rate < 0 || rate > 100 {
        return apperrors.WrapInvalidInput("discount_rate must be between 0 and 100")
    }
    return nil
}

func validateMembershipType(t model.MembershipType) error { ... }
func validatePetGender(gender string) error { ... }
func validatePetStatus(status string) error { ... }
```

### repository/

**責務**: DB アクセスの永続化処理のみ。

- Repository インターフェースをこの層で定義
- GORM 操作に集中
- DB エラーをセンチネルエラーに変換（`gorm.ErrRecordNotFound` → `apperrors.WrapNotFound`）
- 業務判断を書かない

```go
// repository/owner_repository.go
type OwnerRepository interface {
    FindAll(ctx context.Context, clinicID uint64, page, limit int, search string) ([]model.Owner, int64, error)
    FindByID(ctx context.Context, clinicID, id uint64) (*model.Owner, error)
    CreateWithPets(ctx context.Context, owner *model.Owner, pets []model.Pet) error
    Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
    Delete(ctx context.Context, clinicID, id uint64) error
}
```

### model/

**責務**: GORM モデル（DB スキーマ対応）。

- 現時点では GORM モデル兼業務データ構造として使用
- `json:"-"` で内部フィールド（`DeletedAt`, `PasswordHash` 等）を非公開
- Value Object / entity 分離は複雑化が必要になった時点で対応

### errors/

**責務**: アプリケーション全体で使うセンチネルエラーと `AppError` 型。

- `model/` に混ぜない（モデルの都合とエラーの都合は別）
- `ErrNotFound`, `ErrInvalidInput`, `ErrAlreadyExists`, `ErrUnauthorized`, `ErrForbidden`

### middleware/

**責務**: JWT 認証・CORS・リクエストログ等の横断的処理。

- `*gin.Context` を扱う
- handler 層と同様、Gin に依存してよい

---

## フロントエンド・アーキテクチャ

Feature-Based 構造 + Dependency Inversion。詳細は以下を参照：

- **[.claude/CLAUDE.md](../.claude/CLAUDE.md)**: ディレクトリ構成、React 19 パターン、Feature Indexing、Performance ルール
- **[DESIGN_SYSTEM.md](./DESIGN_SYSTEM.md)**: カラーパレット、デザイントークン
- **[frontend/CODING_RULES.md](../frontend/CODING_RULES.md)**: 型安全、memo 最適化、useActionState

---

## 依存方向

```
main.go
  │
  ├── handler  ←─── gin.Context のみここ
  │     │
  │     └── service（インターフェース経由）
  │           │
  │           └── repository（インターフェース経由）
  │                 │
  │                 └── model（GORM モデル）
  │
  └── middleware
        └── errors
```

**依存は常に下方向のみ。外側の都合（HTTP）を内側（service/repository）に漏らさない。**

---

## バリデーション責務の分離

| 種別 | 場所 | 例 |
|---|---|---|
| HTTP 入力チェック | `handler/*_request.go` | `binding:"required"`, length, format |
| 業務ルールチェック | `service/validators.go` | 割引率 0-100, enum 値, 複数フィールド整合性 |

`internal/validation/` パッケージは作らない。validation の責務が2種類に分かれるため、専用パッケージは「共通ゴミ箱」になりやすい。

---

## DI 配線（main.go）

```go
// cmd/api/main.go
db       := db.NewPostgres(cfg)
ownerRepo := repository.NewOwnerRepository(db)
ownerSvc  := service.NewOwnerService(ownerRepo)
h        := handler.New(cfg, &service.Services{Owner: ownerSvc, ...})
h.RegisterRoutes(r)
```

`di/container.go` は作らない。この規模では `main.go` への直書きで十分。

---

## 設計ルール（禁止事項）

| 禁止 | 理由 |
|---|---|
| `service` に `binding:` タグを書く | service が HTTP 入力形式を知ることになる |
| `service` に `*gin.Context` を渡す | service が Gin に依存する |
| `handler` から `repository` を直接呼ぶ | service 層をバイパスして業務ロジックが handler に漏れる |
| `model` をそのまま API レスポンスに使う | DB スキーマと API 契約が結合する |
| `validation/` パッケージを作る | HTTP/業務の両バリデーションが混ざり責務が曖昧になる |
| `di/container.go` を `internal/` に作る | DI 配線は `main.go` に集約するのが原則 |

---

## ファイル命名規則

リソースごとに横並びで揃える。

```
handler/owner_handler.go
handler/owner_request.go
handler/owner_response.go
service/owner_service.go
repository/owner_repository.go
model/owner.go
```

35テーブル規模では、この規則で全テーブルを横展開するのが最も追いやすい。

---

## 将来の拡張方針

現時点では軽量に保つ。複雑化したときだけ進化させる。

| 状況 | 対応 |
|---|---|
| 会計など複雑な業務処理が増えた | `service/billing/` にサブパッケージを切る |
| 集計・レポート系クエリが増えた | `service/query_service.go` を追加 |
| 外部サービス連携が増えた | `internal/infra/` を追加してインターフェース化 |
| マイクロサービス分割が必要になった | その時点で Clean Architecture に移行 |

**今は重くしない。必要な場所だけ必要な時に重くする。**
