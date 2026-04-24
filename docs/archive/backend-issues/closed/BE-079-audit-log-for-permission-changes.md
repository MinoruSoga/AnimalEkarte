# BE-079: 権限変更操作の監査ログ実装

**Status**: Open
**Priority**: Medium
**Affects**: migrations/001_init.sql, internal/model/audit_log.go（新規）, internal/handler/permission_group_handler.go, internal/handler/auth_handler.go, internal/handler/user_account_handler.go
**Date Created**: 2026-03-29
**Related**: BUG-056, TASK-048

---

## Summary

権限グループの作成・変更・削除、ユーザーへの権限割り当て変更、ログイン成功/失敗が
**一切記録されていない**。

- 誰がいつ権限を変更したか追跡できない
- 不正アクセス試行を事後検知できない
- コンプライアンス要件（医療情報システム）を満たさない

---

## 記録が必要な操作

| 操作 | テーブル/エンドポイント | 重要度 |
|------|----------------------|--------|
| 権限グループ作成 | `POST /v1/permission-groups` | 高 |
| 権限グループ更新（名前・説明） | `PUT /v1/permission-groups/:id` | 高 |
| 権限グループ削除 | `DELETE /v1/permission-groups/:id` | 高 |
| 権限ルール更新（view/create/edit/delete） | `POST /v1/permission-groups/:id/rules` | **最高** |
| ユーザーへのグループ割り当て変更 | `PUT /v1/users/:id/permission-groups` | **最高** |
| ログイン成功 | `POST /v1/auth/login` | 高 |
| ログイン失敗 | `POST /v1/auth/login`（失敗時） | 高 |
| ログアウト | `POST /v1/auth/logout` | 中 |

---

## 実装手順

### 1. `audit_logs` テーブル追加（`migrations/001_init.sql`）

```sql
CREATE TABLE audit_logs (
    id           BIGSERIAL    PRIMARY KEY,
    clinic_id    bigint       NULL,                           -- NULL: ログイン失敗など clinic 確定前
    actor_id     bigint       NULL,                           -- 操作者 user_id（NULL: 未認証）
    actor_type   varchar(30)  NOT NULL,                       -- system_admin / clinic_admin / staff / anonymous
    action       varchar(50)  NOT NULL,                       -- permission_group.create 等
    resource     varchar(50)  NOT NULL,                       -- permission_group / user_permission / auth
    resource_id  bigint       NULL,                           -- 対象レコードの ID
    old_value    jsonb        NULL,                           -- 変更前の値（JSON）
    new_value    jsonb        NULL,                           -- 変更後の値（JSON）
    ip_address   inet         NULL,
    user_agent   text         NULL,
    created_at   timestamp    DEFAULT CURRENT_TIMESTAMP
);

-- 検索インデックス
CREATE INDEX idx_audit_logs_clinic    ON audit_logs(clinic_id, created_at DESC);
CREATE INDEX idx_audit_logs_actor     ON audit_logs(actor_id, created_at DESC);
CREATE INDEX idx_audit_logs_resource  ON audit_logs(resource, resource_id, created_at DESC);
```

**注意**: `audit_logs` は削除禁止（`deleted_at` カラムを持たない）。
操作の証跡を改ざん・削除できないようにする。

### 2. `AuditLog` モデル（`backend/internal/model/audit_log.go`）

```go
package model

import "time"

// AuditLog は権限変更・認証操作の記録。削除禁止テーブル。
type AuditLog struct {
    ID         uint64     `gorm:"primaryKey"  json:"id"`
    ClinicID   *uint64    `gorm:"default:null" json:"clinic_id"`
    ActorID    *uint64    `gorm:"default:null" json:"actor_id"`
    ActorType  string     `gorm:"not null"    json:"actor_type"`
    Action     string     `gorm:"not null"    json:"action"`
    Resource   string     `gorm:"not null"    json:"resource"`
    ResourceID *uint64    `gorm:"default:null" json:"resource_id"`
    OldValue   []byte     `gorm:"type:jsonb"  json:"old_value"`
    NewValue   []byte     `gorm:"type:jsonb"  json:"new_value"`
    IPAddress  string     `gorm:"default:null" json:"ip_address"`
    UserAgent  string     `gorm:"default:null" json:"user_agent"`
    CreatedAt  time.Time  `json:"created_at"`
}

func (AuditLog) TableName() string { return "audit_logs" }

// Action 定数
const (
    AuditActionPermissionGroupCreate     = "permission_group.create"
    AuditActionPermissionGroupUpdate     = "permission_group.update"
    AuditActionPermissionGroupDelete     = "permission_group.delete"
    AuditActionPermissionRulesUpdate     = "permission_rules.update"
    AuditActionUserPermissionGroupSet    = "user_permission_group.set"
    AuditActionAuthLoginSuccess          = "auth.login.success"
    AuditActionAuthLoginFailure          = "auth.login.failure"
    AuditActionAuthLogout                = "auth.logout"
)
```

### 3. `AuditRepository`（`backend/internal/repository/audit_repository.go`）

```go
package repository

import (
    "context"
    "encoding/json"
    "gorm.io/gorm"
    "github.com/.../model"
)

type AuditRepository interface {
    Create(ctx context.Context, log *model.AuditLog) error
}

type auditRepository struct {
    db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) AuditRepository {
    return &auditRepository{db: db}
}

func (r *auditRepository) Create(ctx context.Context, log *model.AuditLog) error {
    return r.db.WithContext(ctx).Create(log).Error
}

// MarshalJSON は old_value / new_value を jsonb 用にシリアライズするヘルパー。
func MarshalJSON(v any) ([]byte, error) {
    if v == nil {
        return nil, nil
    }
    return json.Marshal(v)
}
```

### 4. ハンドラーへの組み込み例

#### 権限ルール更新時

```go
// permission_group_handler.go — SetPermissionGroupRules の onSuccess
func (h *Handler) SetPermissionGroupRules(c *gin.Context) {
    // ... 既存処理 ...

    // 変更前の値を取得
    oldRules, _ := h.svc.PermissionGroup.GetRules(ctx, groupID)
    oldJSON, _ := repository.MarshalJSON(oldRules)
    newJSON, _ := repository.MarshalJSON(req.Rules)

    actorID := extractUserID(c)
    clinicID := extractClinicID(c)

    _ = h.auditRepo.Create(ctx, &model.AuditLog{
        ClinicID:   &clinicID,
        ActorID:    &actorID,
        ActorType:  c.GetString("user_type"),
        Action:     model.AuditActionPermissionRulesUpdate,
        Resource:   "permission_group",
        ResourceID: &groupID,
        OldValue:   oldJSON,
        NewValue:   newJSON,
        IPAddress:  c.ClientIP(),
        UserAgent:  c.Request.UserAgent(),
    })
}
```

#### ログイン成功/失敗時

```go
// auth_handler.go — Login
// 成功時
_ = h.auditRepo.Create(ctx, &model.AuditLog{
    ClinicID:  &clinicID,
    ActorID:   &user.ID,
    ActorType: string(user.UserType),
    Action:    model.AuditActionAuthLoginSuccess,
    Resource:  "auth",
    IPAddress: c.ClientIP(),
    UserAgent: c.Request.UserAgent(),
})

// 失敗時（ユーザーが存在しない場合も記録）
_ = h.auditRepo.Create(ctx, &model.AuditLog{
    Action:    model.AuditActionAuthLoginFailure,
    Resource:  "auth",
    NewValue:  []byte(`{"email":"` + req.Email + `"}`),
    IPAddress: c.ClientIP(),
    UserAgent: c.Request.UserAgent(),
})
```

### 5. 監査ログ参照 API（clinic_admin 用・オプション）

```go
// GET /v1/audit-logs?resource=permission_group&from=2026-01-01&to=2026-03-31
// clinic_admin のみアクセス可能（RequireClinicAdmin ミドルウェアで保護）
```

---

## 注意事項

### 監査ログはベストエフォート

監査ログの書き込みに失敗しても、メインの操作は成功させる（ログ失敗でロールバックしない）。

```go
// ✅ エラーを無視してメイン処理を優先
_ = h.auditRepo.Create(ctx, &model.AuditLog{...})

// ❌ ログ失敗でトランザクションをロールバックしない
if err := h.auditRepo.Create(ctx, &log); err != nil {
    return err  // ← これはやらない
}
```

ただしログに失敗した場合は `slog.WarnContext` で記録する。

### 個人情報の扱い

`old_value` / `new_value` に含まれる情報が PII（個人識別情報）でないことを確認する。
グループ名・リソース名・CRUD フラグのみ記録。ユーザーの個人情報（氏名・メール）は含めない。

---

## 確認コマンド

```bash
# DB リセット（audit_logs テーブルが作成されることを確認）
make reset

# ビルド確認
docker compose exec backend go build ./...

# Lint
docker compose exec backend golangci-lint run ./...
```

---

## 受入条件

- [ ] `audit_logs` テーブルが DB に存在する
- [ ] 権限ルール更新時に `permission_rules.update` レコードが記録される
- [ ] ユーザーへのグループ割り当て変更時に `user_permission_group.set` レコードが記録される
- [ ] ログイン成功・失敗時に `auth.login.success` / `auth.login.failure` が記録される
- [ ] 監査ログの書き込み失敗がメインの操作をロールバックしない
- [ ] `audit_logs` テーブルへの DELETE が禁止されている（アプリから削除 API を提供しない）
- [ ] `docker compose exec backend go build ./...` 成功
- [ ] `docker compose exec backend golangci-lint run ./...` エラー 0 件
