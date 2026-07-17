---
name: gin-api-design
description: （本プロジェクト固有・Gin実装特化。汎用API設計原則はグローバル api-designer スキルを参照）REST API 設計パターン。リソース命名・HTTPステータスコード・ページネーション・エラーレスポンス設計。新規 API エンドポイント設計時に使用。
origin: ECC (adapted for AnimalEkarte)
---

# API 設計パターン

このプロジェクト（Go 1.25 / Gin）の REST API 設計標準。

## コードの配置（BE8 規約・2026-07-17 — 正本 = /.claude/skills/be8-package-refactor/SKILL.md §3）

新規エンドポイントの実装コードは次に配置する:

| 層 | 配置 | 備考 |
|---|---|---|
| handler | `internal/handler/`（フラット直下） | 分割は BE8-7 まで保留。ルート登録は `handler.go` / `master_routes.go` |
| service | `internal/service/<domain>/` **サブパッケージ** | フラット直下への新規追加は禁止。ドメイン間参照は consumer 側の小文字ローカル interface |
| repository | `internal/repository/<domain>/` **サブパッケージ** | 先例: `paymentmethod/`。共有ヘルパは `repohelpers`。**サブパッケージ内にさらにディレクトリを掘らない**（走査 lint が 1 階層のため不可視化する） |
| model | `internal/model/`（単一パッケージ） | 分割しない — 決定事項 |

パッケージ名 = 単数形・全小文字・連結（`trimmingcoursetype` 形式）。型名で パッケージ名を繰り返さない（`reservation.NewRepository`）。詳細規約 = `.claude/refs/go-language.md` §8。

## When to Activate

- 新規 API エンドポイントの設計
- 既存 API の設計レビュー
- ページネーション・フィルタリングの実装
- エラーレスポンスの設計

## URL 設計

```
# ✅ 正しい URL 設計
GET    /api/v1/owners                    # 一覧取得
GET    /api/v1/owners/:id                # 単体取得
POST   /api/v1/owners                    # 作成
PATCH  /api/v1/owners/:id                # 部分更新
DELETE /api/v1/owners/:id                # 削除

# サブリソース
GET    /api/v1/owners/:id/pets           # オーナーのペット一覧
POST   /api/v1/owners/:id/pets           # ペット追加

# アクション（動詞は最小限）
POST   /api/v1/owners/:id/activate       # アクティベート
POST   /api/v1/invoices/:id/pay          # 支払い処理
```

```
# ❌ 禁止パターン
GET    /api/v1/getOwners                 # 動詞をURL に含める
GET    /api/v1/owner                     # 単数形
GET    /api/v1/owner_list                # snake_case
```

## HTTP メソッドと Status Code

| 操作 | メソッド | 成功ステータス | 説明 |
|------|---------|--------------|------|
| 一覧取得 | GET | 200 OK | ページネーション付き配列 |
| 単体取得 | GET | 200 OK | 単体オブジェクト |
| 作成 | POST | 201 Created | 作成されたオブジェクト |
| 部分更新 | PATCH | 200 OK | 更新後のオブジェクト |
| 削除 | DELETE | 204 No Content | ボディなし |
| 存在しない | GET/PATCH/DELETE | 404 Not Found | |
| バリデーションエラー | POST/PATCH | 400 Bad Request | |
| 未認証 | * | 401 Unauthorized | |
| 権限なし | * | 403 Forbidden | |
| 競合（FK使用中） | DELETE | 409 Conflict | |

## Gin ハンドラーパターン

```go
// GET /api/v1/owners
func (h *OwnerHandler) List(c *gin.Context) {
    clinicID, ok := extractClinicID(c)
    if !ok {
        return // extractClinicID がエラーレスポンス済み
    }
    // ※ 拠点横断の一覧は resolveListClinicIDs(c) を使う
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

    owners, total, err := h.service.ListOwners(c.Request.Context(), clinicID, page, limit)
    if err != nil {
        RespondError(c, err)
        return
    }

    c.JSON(http.StatusOK, ListOwnersResponse{
        Data:  owners,
        Total: total,
        Page:  page,
        Limit: limit,
    })
}

// POST /api/v1/owners
func (h *OwnerHandler) Create(c *gin.Context) {
    var req CreateOwnerRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
        return
    }

    clinicID, ok := extractClinicID(c)
    if !ok {
        return
    }
    owner, err := h.service.CreateOwner(c.Request.Context(), clinicID, req)
    if err != nil {
        RespondError(c, err)
        return
    }

    c.JSON(http.StatusCreated, owner)
}
```

## レスポンス設計

### 一覧レスポンス
```go
type ListOwnersResponse struct {
    Data  []OwnerResponse `json:"data"`
    Total int64           `json:"total"`
    Page  int             `json:"page"`
    Limit int             `json:"limit"`
}
```

### エラーレスポンス（統一フォーマット）
```json
{
    "error": "リソースが見つかりません"
}
```

> `RespondError`（backend/internal/handler/response.go）が status ごとに `gin.H{"error": msg}` を返す（実装準拠。code/timestamp フィールドは存在しない）。

## Request Validation

```go
// 入力バリデーションは binding タグで宣言的に
type CreateOwnerRequest struct {
    Name    string `json:"name"    binding:"required,min=1,max=100"`
    Email   string `json:"email"   binding:"omitempty,email"`
    Phone   string `json:"phone"   binding:"omitempty,min=10,max=20"`
    Address string `json:"address" binding:"omitempty,max=500"`
}

type UpdateOwnerRequest struct {
    Name    *string `json:"name"    binding:"omitempty,min=1,max=100"`
    Email   *string `json:"email"   binding:"omitempty,email"`
    Phone   *string `json:"phone"   binding:"omitempty,min=10,max=20"`
}
```

## ページネーション

```go
// クエリパラメータ: ?page=1&limit=20&sort=created_at&order=desc
type PaginationParams struct {
    Page  int    `form:"page"  binding:"omitempty,min=1"`
    Limit int    `form:"limit" binding:"omitempty,min=1,max=100"`
    Sort  string `form:"sort"  binding:"omitempty,oneof=created_at updated_at name"`
    Order string `form:"order" binding:"omitempty,oneof=asc desc"`
}

func (p *PaginationParams) SetDefaults() {
    if p.Page == 0 { p.Page = 1 }
    if p.Limit == 0 { p.Limit = 20 }
    if p.Sort == "" { p.Sort = "created_at" }
    if p.Order == "" { p.Order = "desc" }
}

func (p *PaginationParams) Offset() int {
    return (p.Page - 1) * p.Limit
}
```

## フィルタリング

```go
// クエリパラメータ: ?name=佐藤&status=active
type OwnerFilterParams struct {
    Name   string `form:"name"`
    Status string `form:"status" binding:"omitempty,oneof=active inactive"`
}

func (r *OwnerRepository) List(ctx context.Context, clinicID uint64, filter OwnerFilterParams, pagination PaginationParams) ([]model.Owner, int64, error) {
    query := r.db.WithContext(ctx).Where("clinic_id = ? AND deleted_at IS NULL", clinicID)

    if filter.Name != "" {
        // escapeLike で % / _ をエスケープしてから LIKE に渡す（パターン汚染防止）
        query = query.Where("name LIKE ?", "%"+escapeLike(filter.Name)+"%")
    }
    if filter.Status != "" {
        query = query.Where("status = ?", filter.Status)
    }

    var total int64
    query.Count(&total)

    var owners []model.Owner
    err := query.Order(fmt.Sprintf("%s %s", pagination.Sort, pagination.Order)).
        Offset(pagination.Offset()).
        Limit(pagination.Limit).
        Find(&owners).Error

    return owners, total, err
}
```

`escapeLike` の実装は golang-gin-clean-arch スキルの repository-pattern.md / SKILL.md を参照。

## API 設計チェックリスト

- [ ] URL は複数形の名詞（`/owners` ✅、`/getOwners` ❌）
- [ ] 適切な HTTP メソッド（CRUD → GET/POST/PATCH/DELETE）
- [ ] 正しい HTTP ステータスコード（201 for POST、204 for DELETE）
- [ ] 全入力に binding バリデーション
- [ ] エラーは必ず `RespondError(c, err)` で統一
- [ ] 一覧は `total`, `page`, `limit` を含む
- [ ] clinic_id は middleware から取得（クエリパラメータ禁止）
- [ ] PATCH は ポインタ型リクエスト
