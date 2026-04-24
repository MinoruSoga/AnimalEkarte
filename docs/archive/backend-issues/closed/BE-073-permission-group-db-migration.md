# BE-073: 権限グループ DB マイグレーション（3テーブル追加・user_permissions廃止）

**Status**: Closed
**Priority**: High
**Affects**: user_permissions, model/clinic.go, make codegen
**Date Created**: 2026-03-26
**Related**: TASK-032, BE-074, BE-075, BE-076

## Summary

`user_permissions` テーブルと `permission_type` ENUM を廃止し、グループベースの権限管理用に3テーブルを追加する。
モデル変更後に `make codegen` で `models.ts` を再生成する。

## 現状のコード

**`backend/migrations/001_init.sql`（廃止対象）:**
```sql
-- 廃止: permission_type ENUM
CREATE TYPE permission_type AS ENUM (
    'account_admin', 'medical', 'medical_read', 'trimming',
    'billing', 'reception', 'hospitalization', 'master_admin',
    'shift_admin', 'inventory'
);

-- 廃止: user_permissions テーブル
CREATE TABLE user_permissions (
    id         BIGSERIAL       PRIMARY KEY,
    user_id    bigint          NOT NULL REFERENCES user_accounts(id) ON DELETE CASCADE,
    clinic_id  bigint          NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    permission permission_type NOT NULL,
    granted_by bigint                   REFERENCES user_accounts(id) ON DELETE SET NULL,
    granted_at timestamptz              DEFAULT now(),
    CONSTRAINT uk_user_permissions UNIQUE (user_id, clinic_id, permission)
);
```

**`backend/internal/model/clinic.go`（廃止対象）:**
```go
// 廃止: PermissionType と UserPermission
type PermissionType string
const (
    PermissionAccountAdmin    PermissionType = "account_admin"
    // ... 10個
)

type UserPermission struct {
    ID         uint64         `gorm:"primaryKey;autoIncrement"`
    UserID     uint64         `gorm:"not null"`
    ClinicID   uint64         `gorm:"not null"`
    Permission PermissionType `gorm:"type:permission_type;not null"`
    GrantedBy  *uint64
    GrantedAt  time.Time      `gorm:"default:now()"`
}
```

## 必要な変更

### 1. 001_init.sql の変更

**削除する箇所:**
```sql
-- この ENUM 定義を削除
CREATE TYPE permission_type AS ENUM (...);

-- このテーブルを削除
CREATE TABLE user_permissions (...);
```

**追加する箇所（user_clinic_memberships の後に追加）:**
```sql
-- 権限グループ定義
CREATE TABLE permission_groups (
    id          BIGSERIAL    PRIMARY KEY,
    clinic_id   bigint       NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    name        varchar(100) NOT NULL,
    description text         NOT NULL DEFAULT '',
    color       varchar(7)   NOT NULL DEFAULT '#6B7280',
    created_at  timestamptz  NOT NULL DEFAULT now(),
    updated_at  timestamptz  NOT NULL DEFAULT now(),
    deleted_at  timestamptz
);

CREATE INDEX idx_permission_groups_clinic ON permission_groups(clinic_id) WHERE deleted_at IS NULL;

-- グループ×ページ×CRUD 権限ルール
CREATE TABLE permission_group_rules (
    id         BIGSERIAL   PRIMARY KEY,
    group_id   bigint      NOT NULL REFERENCES permission_groups(id) ON DELETE CASCADE,
    resource   varchar(50) NOT NULL,
    can_view   boolean     NOT NULL DEFAULT false,
    can_create boolean     NOT NULL DEFAULT false,
    can_edit   boolean     NOT NULL DEFAULT false,
    can_delete boolean     NOT NULL DEFAULT false,
    CONSTRAINT uk_permission_group_rules UNIQUE (group_id, resource)
);

-- ユーザー → グループ 紐付け（多対多）
CREATE TABLE user_permission_groups (
    user_id  bigint NOT NULL REFERENCES user_accounts(id) ON DELETE CASCADE,
    group_id bigint NOT NULL REFERENCES permission_groups(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, group_id)
);

CREATE INDEX idx_user_permission_groups_user ON user_permission_groups(user_id);
```

### 2. model/clinic.go の変更

**削除:**
```go
// 以下を全削除
type PermissionType string
const (
    PermissionAccountAdmin    PermissionType = "account_admin"
    PermissionMedical         PermissionType = "medical"
    PermissionMedicalRead     PermissionType = "medical_read"
    PermissionTrimming        PermissionType = "trimming"
    PermissionBilling         PermissionType = "billing"
    PermissionReception       PermissionType = "reception"
    PermissionHospitalization PermissionType = "hospitalization"
    PermissionMasterAdmin     PermissionType = "master_admin"
    PermissionShiftAdmin      PermissionType = "shift_admin"
    PermissionInventory       PermissionType = "inventory"
)

type UserPermission struct { ... }
```

**追加（同ファイル末尾）:**
```go
// PermissionGroup は権限グループ定義
type PermissionGroup struct {
    ID          uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
    ClinicID    uint64         `gorm:"not null"                json:"clinic_id"`
    Name        string         `gorm:"not null"                json:"name"`
    Description string         `gorm:"default:''"              json:"description"`
    Color       string         `gorm:"default:'#6B7280'"       json:"color"`
    CreatedAt   time.Time      `gorm:"autoCreateTime"          json:"created_at"`
    UpdatedAt   time.Time      `gorm:"autoUpdateTime"          json:"updated_at"`
    DeletedAt   gorm.DeletedAt `                               json:"deleted_at,omitempty"`

    // Relations
    Rules []PermissionGroupRule `gorm:"foreignKey:GroupID" json:"rules,omitempty"`
}

// PermissionGroupRule はグループ×ページ×CRUD 権限
type PermissionGroupRule struct {
    ID        uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
    GroupID   uint64 `gorm:"not null"                json:"group_id"`
    Resource  string `gorm:"not null"                json:"resource"`
    CanView   bool   `gorm:"not null;default:false"  json:"can_view"`
    CanCreate bool   `gorm:"not null;default:false"  json:"can_create"`
    CanEdit   bool   `gorm:"not null;default:false"  json:"can_edit"`
    CanDelete bool   `gorm:"not null;default:false"  json:"can_delete"`
}

// UserPermissionGroup はユーザー → グループ紐付け
type UserPermissionGroup struct {
    UserID  uint64 `gorm:"primaryKey" json:"user_id"`
    GroupID uint64 `gorm:"primaryKey" json:"group_id"`
}
```

### 3. make codegen の実行

```bash
make codegen
```

`frontend/src/types/generated/models.ts` に上記3モデルが追加される。

## 完了条件

- [ ] 001_init.sql から user_permissions テーブルと permission_type ENUM が削除されている
- [ ] 001_init.sql に permission_groups / permission_group_rules / user_permission_groups が追加されている
- [ ] model/clinic.go から PermissionType・UserPermission が削除されている
- [ ] model/clinic.go に PermissionGroup / PermissionGroupRule / UserPermissionGroup が追加されている
- [ ] `make codegen` 実行後に models.ts が更新されている（PermissionGroup 等が含まれる）
- [ ] `docker compose down && docker compose up` でDBが正常に起動する（マイグレーション成功）

## クローズ情報

- **Closed At**: 2026-03-26
- **変更ファイル**:
  - `backend/migrations/001_init.sql` — user_permissions/permission_type 削除、3新テーブル追加
  - `backend/migrations/002_seed_master.sql` — user_permissions シード削除、新グループデータ追加
  - `backend/migrations/003_seed_demo.sql` — insurance_ratio 型修正（70 → 0.70）
  - `backend/internal/model/clinic.go` — PermissionType/UserPermission 削除、3新モデル追加
  - `backend/internal/repository/user_account_repository.go` — 旧権限メソッド削除
  - `backend/internal/service/user_account_service.go` — 旧権限メソッド削除
  - `backend/internal/handler/user_account_handler.go` — 旧権限エンドポイント削除
  - `backend/internal/handler/auth_handler.go` — permissions 空マップに変更（BE-075 で実装）
  - `backend/internal/model/schema_drift_test.go` — 新モデル参照に更新
  - `frontend/src/types/generated/models.ts` — make codegen で自動更新
