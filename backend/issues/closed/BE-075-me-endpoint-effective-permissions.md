# BE-075: GET /me レスポンスに実効権限マップを追加

**Status**: Closed
**Priority**: High
**Affects**: auth_handler.go, user_account_repository.go, user_account_service.go
**Date Created**: 2026-03-26
**Related**: TASK-032, BE-073（先に完了必要）, FE-128

## Summary

`GET /me` のレスポンス `Permissions` フィールドを `map[string][]string`（旧: PermissionType配列）から
`map[string]map[string]ResourcePermission`（新: clinicID → resource → CRUD）に変更する。
実効権限はサーバーサイドでグループのUNIONとして計算して返す。

## 現状のコード

**`backend/internal/handler/auth_handler.go:40-53`:**
```go
// MeResponse は GET /me のレスポンス（フロントエンド AuthUser と対応）
type MeResponse struct {
    ID           string               `json:"id"`
    Email        string               `json:"email"`
    DisplayName  string               `json:"display_name"`
    UserType     string               `json:"user_type"`
    StaffRole    *string              `json:"staff_role"`
    JobTitle     *string              `json:"job_title"`
    AvatarURL    *string              `json:"avatar_url"`
    MainClinicID string               `json:"main_clinic_id"`
    Clinic       *MeClinicInfo        `json:"clinic"`
    Clinics      []MeClinicMembership `json:"clinics"`
    Permissions  map[string][]string  `json:"permissions"`  // ← 変更対象
}
```

**`backend/internal/handler/auth_handler.go:83-87`:**
```go
permMap := make(map[string][]string)
for _, p := range data.Permissions {
    clIDStr := strconv.FormatUint(p.ClinicID, 10)
    permMap[clIDStr] = append(permMap[clIDStr], string(p.Permission))
}
```

**`backend/internal/repository/user_account_repository.go`（UserAccountWithMemberships）:**
```go
type UserAccountWithMemberships struct {
    model.UserAccount
    Memberships []model.UserClinicMembership
    Permissions []model.UserPermission  // ← BE-073で廃止される
}
```

## 必要な変更

### 1. 新型定義を auth_handler.go に追加

```go
// ResourcePermission は1リソースのCRUD権限
type ResourcePermission struct {
    View   bool `json:"view"`
    Create bool `json:"create"`
    Edit   bool `json:"edit"`
    Delete bool `json:"delete"`
}

// ClinicEffectivePermissions は clinic_id → resource → CRUD のマップ
// 例: {"1": {"accounting": {View: true, Create: true, Edit: false, Delete: false}}}
type ClinicEffectivePermissions = map[string]map[string]ResourcePermission

// MeResponse の Permissions フィールドを変更
type MeResponse struct {
    // ... 他フィールドは変更なし
    Permissions ClinicEffectivePermissions `json:"permissions"`
}
```

### 2. Repository に実効権限取得メソッドを追加

**`backend/internal/repository/user_account_repository.go`:**

```go
// EffectivePermissionRow は実効権限計算用のクエリ結果行
type EffectivePermissionRow struct {
    ClinicID  uint64
    Resource  string
    CanView   bool
    CanCreate bool
    CanEdit   bool
    CanDelete bool
}

// FindEffectivePermissions はユーザーの全クリニックの実効権限を取得する（グループUNION計算）
func (r *userAccountRepository) FindEffectivePermissions(ctx context.Context, userID uint64) ([]EffectivePermissionRow, error) {
    var rows []EffectivePermissionRow
    err := r.db.WithContext(ctx).Raw(`
        SELECT
            pg.clinic_id,
            pgr.resource,
            bool_or(pgr.can_view)   AS can_view,
            bool_or(pgr.can_create) AS can_create,
            bool_or(pgr.can_edit)   AS can_edit,
            bool_or(pgr.can_delete) AS can_delete
        FROM user_permission_groups upg
        JOIN permission_groups pg ON pg.id = upg.group_id AND pg.deleted_at IS NULL
        JOIN permission_group_rules pgr ON pgr.group_id = pg.id
        WHERE upg.user_id = ?
        GROUP BY pg.clinic_id, pgr.resource
    `, userID).Scan(&rows).Error
    return rows, err
}
```

### 3. UserAccountWithMemberships の Permissions フィールドを変更

```go
// Before
type UserAccountWithMemberships struct {
    model.UserAccount
    Memberships []model.UserClinicMembership
    Permissions []model.UserPermission  // 旧
}

// After
type UserAccountWithMemberships struct {
    model.UserAccount
    Memberships        []model.UserClinicMembership
    EffectivePermRows  []EffectivePermissionRow  // 新
}
```

### 4. buildMeResponse の権限マップ構築ロジックを変更

**`backend/internal/handler/auth_handler.go`（buildMeResponse 内）:**

```go
// Before
permMap := make(map[string][]string)
for _, p := range data.Permissions {
    clIDStr := strconv.FormatUint(p.ClinicID, 10)
    permMap[clIDStr] = append(permMap[clIDStr], string(p.Permission))
}

// After
permMap := make(ClinicEffectivePermissions)
for _, row := range data.EffectivePermRows {
    clIDStr := strconv.FormatUint(row.ClinicID, 10)
    if permMap[clIDStr] == nil {
        permMap[clIDStr] = make(map[string]ResourcePermission)
    }
    permMap[clIDStr][row.Resource] = ResourcePermission{
        View:   row.CanView,
        Create: row.CanCreate,
        Edit:   row.CanEdit,
        Delete: row.CanDelete,
    }
}
```

### 5. system_admin / clinic_admin の全権限バイパス

system_admin・clinic_admin は DB 問い合わせなしで全権限 true を返す。

```go
// buildMeResponse 内または GetMe ハンドラ内で
if account.UserType == model.UserTypeSystemAdmin || account.UserType == model.UserTypeClinicAdmin {
    // 全リソース全権限 true のマップを構築して返す
    permMap = buildAllPermissionsMap(allClinicIDs)
}
```

`allClinicIDs` はメンバーシップから取得した clinicID 一覧。`buildAllPermissionsMap` は全リソースに対して `{View:true, Create:true, Edit:true, Delete:true}` を返すヘルパー。

## API レスポンス形式

```json
{
  "id": "123",
  "email": "staff@example.com",
  "user_type": "staff",
  "permissions": {
    "1": {
      "accounting":      { "view": true,  "create": true,  "edit": false, "delete": false },
      "reservations":    { "view": true,  "create": true,  "edit": true,  "delete": false },
      "medical-records": { "view": false, "create": false, "edit": false, "delete": false }
    }
  }
}
```

## 完了条件

- [ ] `GET /me` レスポンスの `permissions` が `{clinicId: {resource: {view, create, edit, delete}}}` 形式になっている
- [ ] staff ユーザーのグループUNIONが正しく計算されている（複数グループ所属時）
- [ ] グループに含まれないリソースは含まれない（フロント側でデフォルト false 扱い）
- [ ] system_admin / clinic_admin は全リソース全CRUD true が返る
- [ ] `docker compose exec backend go test ./... -v` が通る

## クローズ情報

- **Closed At**: 2026-03-26
- **変更ファイル**:
  - `backend/internal/repository/user_account_repository.go` — EffectivePermissionRow/UserAccountWithMemberships 追加、findEffectivePermissions SQL実装
  - `backend/internal/handler/auth_handler.go` — ResourcePermission/ClinicEffectivePermissions 型追加、MeResponse.Permissions 型変更、buildMeResponse 権限マップ構築ロジック更新、system_admin/clinic_admin バイパス実装
