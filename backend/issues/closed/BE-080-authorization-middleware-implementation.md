# BE-080: APIレベル認可チェック未実装 — RequirePermission ミドルウェア実装

**Status**: Open
**Priority**: Critical
**Affects**: internal/middleware/, backend/cmd/api/handler.go（ルート定義）
**Date Created**: 2026-03-29
**Related**: BUG-056, TASK-048（BE-077 先行必須）

---

## Summary

バックエンド API に認可（Authorization）ミドルウェアが存在しない。
`clinic_id` によるテナント分離は実装済みだが、リソース×アクション単位のアクセス制御は
フロントエンド UI 制御のみに依存しており、直接 API コールで完全に回避できる。

**依存関係**: BE-077（`model.Resource` 型定数定義）の完了後に実装すること。

---

## 実装手順

### 1. `RequireClinicAdmin` ミドルウェア（`internal/middleware/require_clinic_admin.go`）

```go
package middleware

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/animal-ekarte/backend/internal/model"
)

// RequireClinicAdmin は clinic_admin または system_admin のみ許可する
func RequireClinicAdmin() gin.HandlerFunc {
    return func(c *gin.Context) {
        userType, ok := c.Get("user_type")
        if !ok {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
            return
        }
        ut, ok := userType.(model.UserType)
        if !ok || (ut != model.UserTypeClinicAdmin && ut != model.UserTypeSystemAdmin) {
            c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
            return
        }
        c.Next()
    }
}
```

### 2. `RequirePermission` ミドルウェア（`internal/middleware/require_permission.go`）

```go
package middleware

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/animal-ekarte/backend/internal/model"
    "github.com/animal-ekarte/backend/internal/repository"
)

type PermissionAction string

const (
    ActionView   PermissionAction = "view"
    ActionCreate PermissionAction = "create"
    ActionEdit   PermissionAction = "edit"
    ActionDelete PermissionAction = "delete"
)

// RequirePermission は指定リソース×アクションの権限を確認する
func RequirePermission(repo repository.UserAccountRepository, resource model.Resource, action PermissionAction) gin.HandlerFunc {
    return func(c *gin.Context) {
        // system_admin / clinic_admin は全権限
        userType, _ := c.Get("user_type")
        if ut, ok := userType.(model.UserType); ok {
            if ut == model.UserTypeSystemAdmin || ut == model.UserTypeClinicAdmin {
                c.Next()
                return
            }
        }

        userID, ok := c.Get("user_id")
        if !ok {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
            return
        }

        uid, ok := userID.(uint64)
        if !ok {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
            return
        }

        has, err := repo.HasPermission(c.Request.Context(), uid, string(resource), string(action))
        if err != nil || !has {
            c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
            return
        }
        c.Next()
    }
}
```

### 3. `UserAccountRepository` に `HasPermission` メソッド追加

```go
// repository/user_account_repository.go
type UserAccountRepository interface {
    // 既存メソッド ...
    HasPermission(ctx context.Context, userID uint64, resource, action string) (bool, error)
}

func (r *userAccountRepository) HasPermission(ctx context.Context, userID uint64, resource, action string) (bool, error) {
    var count int64
    err := r.db.WithContext(ctx).Raw(`
        SELECT COUNT(*)
        FROM user_permission_groups upg
        JOIN permission_groups pg ON pg.id = upg.group_id AND pg.deleted_at IS NULL
        JOIN permission_group_rules pgr ON pgr.group_id = pg.id
        WHERE upg.user_id = ?
          AND pgr.resource = ?
          AND pgr.`+action+` = true
    `, userID, resource).Scan(&count).Error
    return count > 0, err
}
```

**注意**: `action` カラム名は動的クエリになるためホワイトリスト検証を必ず行うこと。

### 4. ルート定義にミドルウェアを追加（`cmd/api/handler.go`）

優先度の高いエンドポイントから段階的に適用：

```go
// clinic_admin 以上が必要なルート
users := v1.Group("/users", middleware.RequireClinicAdmin())
{
    users.POST("", h.CreateUser)
    users.DELETE("/:id", h.DeleteUser)
}

permGroups := v1.Group("/permission-groups", middleware.RequireClinicAdmin())
{
    permGroups.POST("", h.CreatePermissionGroup)
    permGroups.DELETE("/:id", h.DeletePermissionGroup)
}

// リソース別権限チェックが必要なルート
medicalRecords := v1.Group("/medical-records", authMiddleware)
{
    medicalRecords.DELETE("/:id", middleware.RequirePermission(repo, model.ResourceMedicalRecords, middleware.ActionDelete), h.DeleteMedicalRecord)
}
```

---

## 保護対象エンドポイント（優先順）

| 優先度 | エンドポイント | 必要な保護 |
|--------|--------------|-----------|
| 🔴 Critical | `POST /permission-groups` | `RequireClinicAdmin` |
| 🔴 Critical | `DELETE /permission-groups/:id` | `RequireClinicAdmin` |
| 🔴 Critical | `POST /users` | `RequireClinicAdmin` |
| 🔴 Critical | `DELETE /users/:id` | `RequireClinicAdmin` |
| 🟠 High | `DELETE /medical-records/:id` | `RequirePermission(medical-records, delete)` |
| 🟠 High | `DELETE /accounting/:id` | `RequirePermission(accounting, delete)` |
| 🟡 Medium | `POST/PUT /master/*` | `RequirePermission(master, create/edit)` |

---

## 確認コマンド

```bash
docker compose exec backend go build ./...
docker compose exec backend go test ./... -v
docker compose exec backend golangci-lint run ./...
# staff ユーザーで POST /permission-groups → 403 を確認
```

---

## 受入条件

- [ ] `middleware/require_clinic_admin.go` が存在し、staff ユーザーで保護ルートに → 403
- [ ] `middleware/require_permission.go` が存在し、権限なし staff → 403
- [ ] 全 Critical エンドポイントにミドルウェアが適用されている
- [ ] system_admin / clinic_admin は全エンドポイントに引き続きアクセス可能
- [ ] `docker compose exec backend go build ./...` 成功
- [ ] `docker compose exec backend go test ./... -v` 成功
