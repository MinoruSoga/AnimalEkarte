# 認証システム実装（ユーザー管理 CRUD）

## 概要
ログイン・ログアウト・自ユーザー情報取得は `handler/auth_handler.go` に実装済み。未実装なのはユーザーアカウント一覧・作成・更新・削除、および権限管理 API（`GET/PUT /v1/users/:id/permissions`）。これらを実装してユーザー管理画面を完成させる。

`POST /v1/auth/login`、`POST /v1/auth/logout`、`GET /v1/auth/me` は実装済みのため対象外。

## 優先度
high（ログイン機能はブロッキング済みのため残タスクを優先処理）

## 関連テーブル
- `user_accounts` (id, email, password_hash, display_name, user_type, staff_id, avatar_url, status, created_at, updated_at)
- `user_clinic_memberships` (id, user_id, clinic_id, is_main, created_at, updated_at)
- `user_permissions` (id, user_id, clinic_id, permission, created_at)

## 実装内容

### モデル
`model/user_account.go`（既存）の内容を確認し不足フィールドがあれば追加する。
`password_hash` フィールドは `json:"-"` タグを付与し、**絶対にレスポンスに含めない**。

### リポジトリ
`repository/user_account_repository.go` に以下を追加・確認:
- `List(ctx, clinicID uint64) ([]UserAccountWithMemberships, error)` — クリニックに所属するユーザー一覧
- `Create(ctx, account *model.UserAccount, memberships []model.UserClinicMembership) error` — トランザクションで作成
- `Update(ctx, id uint64, fields map[string]any) error`
- `Delete(ctx, id uint64) error` — soft delete または論理削除（status = inactive）
- `GetPermissions(ctx, userID, clinicID uint64) ([]model.UserPermission, error)`
- `SetPermissions(ctx, userID, clinicID uint64, permissions []model.PermissionType) error` — 既存を削除して再挿入（トランザクション）

### サービス
`service/user_account_service.go` に以下の Input DTO と処理を追加:
```go
type CreateUserAccountInput struct {
    Email       string
    Password    string  // bcrypt でハッシュ化して保存
    DisplayName string
    UserType    model.UserType
    StaffID     *uint64
    ClinicID    uint64
    IsMain      bool
}

type UpdateUserAccountInput struct {
    DisplayName *string
    UserType    *model.UserType
    StaffID     *uint64
    AvatarURL   *string
    Status      *string
}

type SetPermissionsInput struct {
    Permissions []string
}
```
- `CreateUserAccount`: パスワードを bcrypt でハッシュ化（コスト12以上）して DB 保存
- `UpdateUserAccount`: `buildUserAccountUpdateFields()` + `map[string]any` パターン
- `DeleteUserAccount`: status を `inactive` に変更（物理削除不可）
- `GetPermissions`: クリニック別権限一覧
- `SetPermissions`: 既存権限を削除して新規 permission レコードを一括挿入

`service/validators.go` に追加:
- `validatePassword(password string) error` — 最低8文字、追加要件はビジネス要件に準じる
- `validateUserType(userType string) error`
- `validatePermission(permission string) error`

### ハンドラ
新規ファイル `handler/user_account_handler.go`:
```go
func (h *Handler) ListUsers(c *gin.Context)
func (h *Handler) CreateUser(c *gin.Context)
func (h *Handler) UpdateUser(c *gin.Context)
func (h *Handler) DeleteUser(c *gin.Context)
func (h *Handler) GetUserPermissions(c *gin.Context)
func (h *Handler) SetUserPermissions(c *gin.Context)
func (h *Handler) RegisterUserRoutes(rg *gin.RouterGroup)
```

`handler/user_account_request.go`:
```go
type createUserRequest struct {
    Email       string  `json:"email"        binding:"required,email"`
    Password    string  `json:"password"     binding:"required,min=8"`
    DisplayName string  `json:"display_name" binding:"required"`
    UserType    string  `json:"user_type"    binding:"required"`
    StaffID     *uint64 `json:"staff_id"`
    IsMain      bool    `json:"is_main"`
}

type updateUserRequest struct {
    DisplayName *string `json:"display_name"`
    UserType    *string `json:"user_type"`
    StaffID     *uint64 `json:"staff_id"`
    AvatarURL   *string `json:"avatar_url"`
    Status      *string `json:"status"`
}

type setPermissionsRequest struct {
    Permissions []string `json:"permissions" binding:"required"`
}
```

`handler/user_account_response.go`:
```go
type userResponse struct {
    ID          string  `json:"id"`
    Email       string  `json:"email"`
    DisplayName string  `json:"display_name"`
    UserType    string  `json:"user_type"`
    StaffID     *uint64 `json:"staff_id,omitempty"`
    AvatarURL   string  `json:"avatar_url"`
    Status      string  `json:"status"`
    // password_hash は絶対に含めない
}
```

### ルート登録
`cmd/api/main.go` に以下を追加:
```go
// 認証不要
auth := v1.Group("/auth")
auth.POST("/login", h.Login)   // 実装済み
auth.POST("/logout", h.Logout) // 実装済み
auth.GET("/me", authMiddleware, h.GetMe) // 実装済み

// 認証必要
users := v1.Group("/users", authMiddleware)
users.GET("",      h.ListUsers)
users.POST("",     h.CreateUser)
users.PATCH("/:id", h.UpdateUser)
users.DELETE("/:id", h.DeleteUser)
users.GET("/:id/permissions", h.GetUserPermissions)
users.PUT("/:id/permissions", h.SetUserPermissions)
```

## 完了条件
- `GET /v1/users` がクリニックに所属するユーザー一覧を返す（クエリパラメータ `clinic_id` で絞り込み）
- `POST /v1/users` でユーザーを作成できる（パスワードは bcrypt ハッシュ化して保存）
- `PATCH /v1/users/:id` でユーザー情報を更新できる
- `DELETE /v1/users/:id` でユーザーを無効化できる（物理削除なし）
- `GET /v1/users/:id/permissions` で権限一覧を返す
- `PUT /v1/users/:id/permissions` で権限を一括更新できる
- **いかなるレスポンスにも `password_hash` が含まれない**
- 未認証リクエストは 401 を返す
- 不正な `user_type` / `permission` 値は 400 を返す
