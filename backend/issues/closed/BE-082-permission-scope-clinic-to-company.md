# BE-082: 権限スコープを clinic_id から company_id へ変更

**Status**: Open
**Priority**: High
**Affects**: migrations/001_init.sql, migrations/002_seed_master.sql, internal/handler/permission_group_handler.go, internal/service/permission_group_service.go, internal/repository/permission_group_repository.go, internal/repository/user_account_repository.go, internal/handler/auth_handler.go
**Date Created**: 2026-03-29
**Related**: TASK-049, FE-139

---

## Summary

`permission_groups.clinic_id` を `company_id` に変更し、
権限グループを全院共通で定義できるようにする。

---

## 実装手順

### 1. DB スキーマ変更（`migrations/001_init.sql`）

```sql
-- 変更前
CREATE TABLE permission_groups (
    id          BIGSERIAL   PRIMARY KEY,
    clinic_id   bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    name        text        NOT NULL,
    description text        NOT NULL DEFAULT '',
    color       varchar(7)  NOT NULL DEFAULT '#6366F1',
    deleted_at  timestamptz NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_permission_groups_clinic ON permission_groups(clinic_id) WHERE deleted_at IS NULL;

-- 変更後
CREATE TABLE permission_groups (
    id          BIGSERIAL   PRIMARY KEY,
    company_id  bigint      NOT NULL REFERENCES company(id) ON DELETE RESTRICT,  -- clinic_id → company_id
    name        text        NOT NULL,
    description text        NOT NULL DEFAULT '',
    color       varchar(7)  NOT NULL DEFAULT '#6366F1',
    deleted_at  timestamptz NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_permission_groups_company ON permission_groups(company_id) WHERE deleted_at IS NULL;
```

`user_permission_groups` テーブルは変更不要（`group_id` 参照のみ）。

### 2. Seed データ更新（`migrations/002_seed_master.sql`）

```sql
-- 変更前: clinic_id = 3 で3グループ
INSERT INTO permission_groups (id, clinic_id, name, ...) VALUES
    (1, 3, '管理者', ...),
    (2, 3, '執行',   ...),
    (3, 3, '一般',   ...);

-- 変更後: company_id = 1（シングルトン）で3グループ
INSERT INTO permission_groups (id, company_id, name, ...) VALUES
    (1, 1, '管理者', '全機能フルアクセス・権限設定管理', '#EF4444'),
    (2, 1, '執行',   '業務全般閲覧・権限設定変更',       '#6366F1'),
    (3, 1, '一般',   '基本的な業務操作',                 '#10B981');
```

### 3. Go モデル変更（`internal/model/clinic.go` または `permission.go`）

```go
// 変更前
type PermissionGroup struct {
    ID        uint64  `gorm:"primaryKey"`
    ClinicID  uint64  `gorm:"not null"`
    Name      string  `gorm:"not null"`
    // ...
}

// 変更後
type PermissionGroup struct {
    ID        uint64  `gorm:"primaryKey"`
    CompanyID uint64  `gorm:"not null"`  // ClinicID → CompanyID
    Name      string  `gorm:"not null"`
    // ...
}
```

### 4. リポジトリ変更（`internal/repository/permission_group_repository.go`）

```go
// 変更前: clinic_id でフィルタ
func (r *permissionGroupRepository) List(ctx context.Context, clinicID uint64) ([]model.PermissionGroup, error) {
    var groups []model.PermissionGroup
    err := r.db.WithContext(ctx).
        Where("clinic_id = ? AND deleted_at IS NULL", clinicID).
        Find(&groups).Error
    return groups, err
}

// 変更後: company_id でフィルタ（company はシングルトンなので ID = 1 固定も可）
func (r *permissionGroupRepository) List(ctx context.Context, companyID uint64) ([]model.PermissionGroup, error) {
    var groups []model.PermissionGroup
    err := r.db.WithContext(ctx).
        Where("company_id = ? AND deleted_at IS NULL", companyID).
        Find(&groups).Error
    return groups, err
}
```

`FindByID`, `Create`, `Update`, `Delete` も同様に `clinicID` → `companyID` に置き換える。

### 5. サービス・ハンドラー変更

ハンドラーで `extractClinicID(c)` していた箇所を `extractCompanyID(c)` に変更する。

`company_id` は JWT に含まれていないため、以下のいずれかで取得する：

**Option A（推奨）**: JWT の `clinic_id` から `clinics.company_id` を JOIN して取得

```go
// middleware/auth.go で clinic_id → company_id を解決してコンテキストにセット
clinicID := claims.ClinicID
var clinic model.Clinic
db.Select("company_id").Where("id = ?", clinicID).First(&clinic)
c.Set("company_id", clinic.CompanyID)
```

**Option B**: company はシングルトンなので常に ID = 1 を使う（簡易）

```go
// 権限グループ操作では company_id = 1 固定
const defaultCompanyID uint64 = 1
```

Option B はハードコードだが、シングルトン設計上は問題ない。

### 6. `findEffectivePermissions()` の SQL 変更（`internal/repository/user_account_repository.go`）

```go
// 変更前: clinic_id で JOIN
func (r *userAccountRepository) findEffectivePermissions(ctx context.Context, userID uint64) ([]EffectivePermissionRow, error) {
    err := r.db.WithContext(ctx).Raw(`
        SELECT
            pg.clinic_id,   -- ← 削除
            pgr.resource,
            bool_or(pgr.can_view)   AS can_view,
            bool_or(pgr.can_create) AS can_create,
            bool_or(pgr.can_edit)   AS can_edit,
            bool_or(pgr.can_delete) AS can_delete
        FROM user_permission_groups upg
        JOIN permission_groups pg ON pg.id = upg.group_id AND pg.deleted_at IS NULL
        JOIN permission_group_rules pgr ON pgr.group_id = pg.id
        WHERE upg.user_id = ?
        GROUP BY pg.clinic_id, pgr.resource   -- ← clinic_id 削除
    `, userID).Scan(&rows).Error
    return rows, err
}

// 変更後: clinic_id フィルタなし（全院共通権限）
func (r *userAccountRepository) findEffectivePermissions(ctx context.Context, userID uint64) ([]EffectivePermissionRow, error) {
    err := r.db.WithContext(ctx).Raw(`
        SELECT
            pgr.resource,
            bool_or(pgr.can_view)   AS can_view,
            bool_or(pgr.can_create) AS can_create,
            bool_or(pgr.can_edit)   AS can_edit,
            bool_or(pgr.can_delete) AS can_delete
        FROM user_permission_groups upg
        JOIN permission_groups pg ON pg.id = upg.group_id AND pg.deleted_at IS NULL
        JOIN permission_group_rules pgr ON pgr.group_id = pg.id
        WHERE upg.user_id = ?
        GROUP BY pgr.resource
    `, userID).Scan(&rows).Error
    return rows, err
}
```

### 7. `/me` レスポンスのフラット化（`internal/handler/auth_handler.go`）

```go
// 変更前: clinic_id でネスト
type MeResponse struct {
    // ...
    Permissions map[string]map[string]ResourcePermission `json:"permissions"`
    // { "1": { "medical-records": { "view": true, "create": true, ... } } }
}

// 変更後: フラット
type MeResponse struct {
    // ...
    Permissions map[string]ResourcePermission `json:"permissions"`
    // { "medical-records": { "view": true, "create": true, ... } }
}

func buildMeResponse(user *model.UserAccount, perms []EffectivePermissionRow) MeResponse {
    permissions := make(map[string]ResourcePermission)
    for _, p := range perms {
        permissions[p.Resource] = ResourcePermission{
            View:   p.View,
            Create: p.Create,
            Edit:   p.Edit,
            Delete: p.Delete,
        }
    }
    // system_admin / clinic_admin は全リソース全権限
    if user.UserType == model.UserTypeSystemAdmin || user.UserType == model.UserTypeClinicAdmin {
        for _, resource := range model.AllResources {
            permissions[string(resource)] = ResourcePermission{View: true, Create: true, Edit: true, Delete: true}
        }
    }
    return MeResponse{
        // ...
        Permissions: permissions,
    }
}
```

---

## 確認コマンド

```bash
make reset
docker compose exec backend go build ./...
docker compose exec backend go test ./... -v
docker compose exec backend golangci-lint run ./...
```

---

## 受入条件

- [ ] `permission_groups` テーブルに `company_id` カラムがあり `clinic_id` が削除されている
- [ ] 権限グループの CRUD が正しく動作する
- [ ] `GET /v1/me` の `permissions` フィールドが `{ resource → CRUD }` のフラット構造になっている
- [ ] `system_admin` / `clinic_admin` は全リソース全権限が返る
- [ ] `staff` は割り当てられたグループの権限が返る
- [ ] `docker compose exec backend go build ./...` 成功
- [ ] `docker compose exec backend golangci-lint run ./...` エラー 0 件
- [ ] `docker compose exec backend go test ./... -v` 成功
