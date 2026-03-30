# BUG-056: バックエンド — API レベルの認可チェック未実装

## 概要

バックエンド API に `user_type` / 権限グループベースの認可チェックが実装されていない。
`clinic_id` によるマルチテナント分離は実装済みだが、リソース×アクション単位の
アクセス制御はフロントエンドの UI 制御のみに依存しており、**直接 API コールで完全に回避できる**。

## 重要度

**CRITICAL** — フロントエンドを迂回した直接 API 呼び出しで全権限制御が無効化される。

## 現状の問題

### 認可ミドルウェアが存在しない

`backend/internal/middleware/` に認証ミドルウェア (`auth.go`) はあるが、
認可（Authorization）ミドルウェアは存在しない。

### `extractUserType()` が定義されているが未使用

`backend/internal/handler/response.go` に `extractUserType()` が定義済みだが、
ほぼすべてのハンドラーで呼ばれていない。

### `ErrForbidden` が定義されているが未使用

`backend/internal/errors/` に `ErrForbidden` が定義済みだが使用箇所がほぼゼロ。

## 再現手順

```bash
# staff ユーザーの JWT で権限グループを削除できる
curl -X DELETE http://localhost:8080/api/v1/permission-groups/1 \
  -H "Cookie: auth_token=<staff_jwt>"
# → 200 OK（本来は 403 Forbidden であるべき）

# staff ユーザーで別ユーザーを作成できる
curl -X POST http://localhost:8080/api/v1/users \
  -H "Cookie: auth_token=<staff_jwt>" \
  -d '{"email":"hacked@example.com","user_type":"clinic_admin"}'
# → 201 Created（本来は 403 Forbidden であるべき）
```

## 修正方針

### 1. `RequirePermission` ミドルウェアの追加

`backend/internal/middleware/require_permission.go` を新規作成。

```go
// RequirePermission はリソース×アクション単位の認可チェックを行うミドルウェア
// system_admin / clinic_admin は全権限バイパス
// staff は permission_groups テーブルから実効権限を確認
func RequirePermission(resource, action string, permRepo PermissionRepository) gin.HandlerFunc {
    return func(c *gin.Context) {
        userType := c.GetString("user_type")
        // admin 系はバイパス
        if userType == string(model.UserTypeSystemAdmin) ||
           userType == string(model.UserTypeClinicAdmin) {
            c.Next()
            return
        }
        userID   := c.GetUint64("user_id")
        clinicID := c.GetUint64("clinic_id")
        allowed, err := permRepo.HasPermission(c.Request.Context(), userID, clinicID, resource, action)
        if err != nil || !allowed {
            RespondError(c, apperrors.WrapForbidden(resource+"/"+action))
            c.Abort()
            return
        }
        c.Next()
    }
}
```

### 2. ルート定義への適用

`backend/cmd/api/handler.go` のルート登録に追加。

```go
// 例: 医療記録
records := protected.Group("/medical-records")
records.GET("",    RequirePermission("medical-records", "view",   h.permRepo), h.ListMedicalRecords)
records.POST("",   RequirePermission("medical-records", "create", h.permRepo), h.CreateMedicalRecord)
records.PUT("/:id",RequirePermission("medical-records", "edit",   h.permRepo), h.UpdateMedicalRecord)
records.DELETE("/:id", RequirePermission("medical-records", "delete", h.permRepo), h.DeleteMedicalRecord)
```

### 3. clinic_admin 専用エンドポイントの保護

権限グループ管理・ユーザー管理は `RequireClinicAdmin` ミドルウェアで保護。

```go
func RequireClinicAdmin() gin.HandlerFunc {
    return func(c *gin.Context) {
        userType := c.GetString("user_type")
        if userType != string(model.UserTypeSystemAdmin) &&
           userType != string(model.UserTypeClinicAdmin) {
            RespondError(c, apperrors.WrapForbidden("clinic_admin required"))
            c.Abort()
            return
        }
        c.Next()
    }
}

// 適用例
users := protected.Group("/users")
users.GET("",    RequireClinicAdmin(), h.ListUsers)
users.POST("",   RequireClinicAdmin(), h.CreateUser)
users.DELETE("/:id", RequireClinicAdmin(), h.DeleteUser)

permGroups := protected.Group("/permission-groups")
permGroups.GET("",    h.ListPermissionGroups)  // 全員閲覧可
permGroups.POST("",   RequireClinicAdmin(), h.CreatePermissionGroup)
permGroups.PUT("/:id",RequireClinicAdmin(), h.UpdatePermissionGroup)
permGroups.DELETE("/:id", RequireClinicAdmin(), h.DeletePermissionGroup)
```

### 4. `HasPermission` リポジトリメソッドの追加

実効権限をDBから直接チェックするメソッドを追加（または既存の実装を再利用）。

```go
// PermissionRepository.HasPermission
func (r *permissionRepository) HasPermission(
    ctx context.Context,
    userID, clinicID uint64,
    resource, action string,
) (bool, error) {
    // user_permission_groups → permission_groups → permission_group_rules を JOIN
    // bool_or() で UNION した実効権限を返す
    var result bool
    query := `
        SELECT COALESCE(bool_or(CASE $4
            WHEN 'view'   THEN pgr.can_view
            WHEN 'create' THEN pgr.can_create
            WHEN 'edit'   THEN pgr.can_edit
            WHEN 'delete' THEN pgr.can_delete
            ELSE false END), false)
        FROM user_permission_groups upg
        JOIN permission_groups pg ON pg.id = upg.group_id
            AND pg.clinic_id = $2 AND pg.deleted_at IS NULL
        JOIN permission_group_rules pgr ON pgr.group_id = pg.id
            AND pgr.resource = $3
        WHERE upg.user_id = $1
    `
    err := r.db.WithContext(ctx).Raw(query, userID, clinicID, resource, action).Scan(&result).Error
    return result, err
}
```

## 優先対応エンドポイント（リスク順）

| エンドポイント | 問題 | 必要な保護 |
|---|---|---|
| `DELETE /users/:id` | staff がユーザーを削除可能 | `RequireClinicAdmin` |
| `POST /users` | staff がユーザーを作成可能 | `RequireClinicAdmin` |
| `DELETE /permission-groups/:id` | staff が権限グループを削除可能 | `RequireClinicAdmin` |
| `POST /permission-groups` | staff が権限グループを作成可能 | `RequireClinicAdmin` |
| `DELETE /medical-records/:id` | 閲覧権限しかない staff でも削除可能 | `RequirePermission("medical-records","delete")` |
| `DELETE /accounting/:id` | 閲覧権限しかない staff でも削除可能 | `RequirePermission("accounting","delete")` |
| `POST/PUT /master/*` | 一般グループが master を変更可能 | `RequirePermission("master","create/edit")` |

## 影響範囲

- `backend/internal/middleware/require_permission.go` — 新規作成
- `backend/internal/middleware/require_clinic_admin.go` — 新規作成
- `backend/internal/repository/permission_repository.go` — `HasPermission` メソッド追加
- `backend/cmd/api/handler.go` — 全ルート定義に認可ミドルウェアを追加
- `backend/internal/handler/` — `extractUserType()` の活用

## 派生イシュー

| イシュー | 領域 | 内容 |
|---------|------|------|
| BE-080 | Backend | `RequirePermission` / `RequireClinicAdmin` ミドルウェア実装・全ルートへの適用 |
| FE-133 | Frontend | サイドバーメニューを権限でフィルタリング（`canView = false` のリソースを非表示） |
| FE-134 | Frontend | `usePermission` hook を各 feature コンポーネントに統合（ボタン・フォームの表示制御） |

## 依存タスク

| チケット | 関係 |
|---------|------|
| **TASK-048 / BE-077（先に完了必要）** | `RequirePermission` ミドルウェアは `model.Resource` 型を引数に取る設計（文字列リテラルでなく型定数）。BE-077 が完了して `model.Resource` 型と全リソース定数が定義された後に実装すること。 |

## 関連

- `docs/AUTH.md` §7 アプリケーション層認可設計
- TASK-048: 権限リソース定義の単一情報源化（本チケットの前提条件）
- BUG-054: クロスクリニック権限昇格（フロントエンド側の問題、実装確認済みで Closed）
- BUG-033: staff がユーザー一覧を取得可能
