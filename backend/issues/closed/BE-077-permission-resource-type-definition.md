# BE-077: Resource 型定数の定義・DB CHECK 制約・codegen 出力確認

**Status**: Open
**Priority**: High
**Affects**: model/permission.go（新規）, migrations/001_init.sql, handler/permission_group_handler.go, migrations/002_seed_master.sql
**Date Created**: 2026-03-29
**Related**: TASK-048, FE-132, BUG-056

---

## Summary

権限リソースキー（`"owners"`, `"medical-records"` 等）が文字列リテラルとして
バラバラに使用されており型安全性がない。Go 側に `Resource` 型定数を定義し、
tygo 経由でフロントエンドに自動生成する。また DB の `permission_group_rules.resource`
に CHECK 制約を追加して不正なキーの混入を防ぐ。

---

## 実装手順

### 1. `backend/internal/model/permission.go` を新規作成

`PermissionGroupRule` モデルとリソース定数を同ファイルに定義する。

```go
package model

// Resource は権限グループルールのリソースキーを表す型。
// tygo が models.ts に const として出力するため、フロントエンドが参照できる。
type Resource string

const (
    ResourceDashboard       Resource = "dashboard"
    ResourceOwners          Resource = "owners"
    ResourceReservations    Resource = "reservations"
    ResourceMedicalRecords  Resource = "medical-records"
    ResourceHospitalization Resource = "hospitalization"
    ResourceTrimming        Resource = "trimming"
    ResourceExaminations    Resource = "examinations"
    ResourceAccounting      Resource = "accounting"
    ResourceVaccinations    Resource = "vaccinations"
    ResourceCheckups        Resource = "checkups"
    ResourceInventory       Resource = "inventory"
    ResourceEstimates       Resource = "estimates"
    ResourceShifts          Resource = "shifts"
    ResourceMaster          Resource = "master"
    ResourceHospitalSettings Resource = "hospital-settings"
)

// AllResources はすべての有効なリソースキーの一覧。
// DB CHECK 制約の生成や permission_group_rules のシードに使用する。
var AllResources = []Resource{
    ResourceDashboard,
    ResourceOwners,
    ResourceReservations,
    ResourceMedicalRecords,
    ResourceHospitalization,
    ResourceTrimming,
    ResourceExaminations,
    ResourceAccounting,
    ResourceVaccinations,
    ResourceCheckups,
    ResourceInventory,
    ResourceEstimates,
    ResourceShifts,
    ResourceMaster,
    ResourceHospitalSettings,
}

// PermissionGroupRule は permission_group_rules テーブルのモデル。
// 既存モデル（clinic.go）から移動して型付けする。
type PermissionGroupRule struct {
    ID        uint64   `gorm:"primaryKey"           json:"id"`
    GroupID   uint64   `gorm:"not null"             json:"group_id"`
    Resource  Resource `gorm:"not null"             json:"resource"`
    CanView   bool     `gorm:"not null;default:false" json:"can_view"`
    CanCreate bool     `gorm:"not null;default:false" json:"can_create"`
    CanEdit   bool     `gorm:"not null;default:false" json:"can_edit"`
    CanDelete bool     `gorm:"not null;default:false" json:"can_delete"`
}
```

**注意点**:
- 既存の `clinic.go` に `PermissionGroupRule` が定義されている場合はそちらを削除して本ファイルに集約する
- `Resource` 型は `tygo.yaml` の設定によって `models.ts` に出力されることを確認する

### 2. tygo 設定確認（`backend/tygo.yaml`）

`Resource` 型と定数が `models.ts` に出力されるよう、`tygo.yaml` の対象パッケージに
`model` が含まれていることを確認する。出力例:

```typescript
// frontend/src/types/generated/models.ts に追加される想定
export type Resource = string
export const ResourceDashboard       = "dashboard"
export const ResourceOwners          = "owners"
export const ResourceReservations    = "reservations"
export const ResourceMedicalRecords  = "medical-records"
export const ResourceHospitalization = "hospitalization"
export const ResourceTrimming        = "trimming"
export const ResourceExaminations    = "examinations"
export const ResourceAccounting      = "accounting"
export const ResourceVaccinations    = "vaccinations"
export const ResourceCheckups        = "checkups"
export const ResourceInventory       = "inventory"
export const ResourceEstimates       = "estimates"
export const ResourceShifts          = "shifts"
export const ResourceMaster          = "master"
export const ResourceHospitalSettings = "hospital-settings"
```

`make codegen` を実行して上記が `models.ts` に追加されることを確認する。

### 3. DB CHECK 制約の追加（`backend/migrations/001_init.sql`）

`permission_group_rules` テーブルの `resource` カラムに CHECK 制約を追加する。

```sql
-- 変更前
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

-- 変更後
CREATE TABLE permission_group_rules (
    id         BIGSERIAL   PRIMARY KEY,
    group_id   bigint      NOT NULL REFERENCES permission_groups(id) ON DELETE CASCADE,
    resource   varchar(50) NOT NULL,
    can_view   boolean     NOT NULL DEFAULT false,
    can_create boolean     NOT NULL DEFAULT false,
    can_edit   boolean     NOT NULL DEFAULT false,
    can_delete boolean     NOT NULL DEFAULT false,
    CONSTRAINT uk_permission_group_rules UNIQUE (group_id, resource),
    CONSTRAINT chk_permission_group_rules_resource CHECK (resource IN (
        'dashboard', 'owners', 'reservations', 'medical-records',
        'hospitalization', 'trimming', 'examinations', 'accounting',
        'vaccinations', 'checkups', 'inventory', 'estimates',
        'shifts', 'master', 'hospital-settings'
    ))
);
```

**注意**: リリース前は `001_init.sql` 直接編集でよい（incremental migration 不要）。

### 4. ハンドラーでの文字列リテラル置換

`backend/internal/handler/permission_group_handler.go` で `resource` を文字列として
受け取っている箇所に `model.Resource` 型へのキャスト・バリデーションを追加する。

```go
// permission_group_handler.go の ruleRequest にバリデーションを追加
type ruleRequest struct {
    Resource  model.Resource `json:"resource"   binding:"required"`
    CanView   bool           `json:"can_view"`
    CanCreate bool           `json:"can_create"`
    CanEdit   bool           `json:"can_edit"`
    CanDelete bool           `json:"can_delete"`
}

// service 層でリソースキーの有効性を検証
func isValidResource(r model.Resource) bool {
    for _, valid := range model.AllResources {
        if r == valid {
            return true
        }
    }
    return false
}
```

### 5. BUG-056（認可ミドルウェア）実装時の準備

`RequirePermission` ミドルウェアは本タスク完了後に `model.Resource` を使って実装する。
文字列リテラルを引数に取る API は作らない。

```go
// ✅ 本タスク完了後の形
protected.GET("/medical-records",
    RequirePermission(model.ResourceMedicalRecords, "view", h.permRepo),
    h.ListMedicalRecords,
)
```

---

## 確認コマンド

```bash
# Go ビルド確認
docker compose exec backend go build ./...

# codegen 実行（models.ts に Resource 定数が追加されることを確認）
make codegen

# DB リセット＆CHECK 制約確認
make reset
# psql で不正なリソースキーを INSERT してエラーになることを確認
# INSERT INTO permission_group_rules (group_id, resource) VALUES (1, 'invalid-key');
# → ERROR: new row violates check constraint "chk_permission_group_rules_resource"
```

---

## 受入条件

- [ ] `model.Resource` 型が定義されており、15定数が存在する
- [ ] `make codegen` 後に `models.ts` に `ResourceXxx` 定数が出力される
- [ ] `permission_group_rules.resource` に CHECK 制約が追加されている
- [ ] 不正なリソースキーの INSERT が DB レベルで弾かれる
- [ ] `docker compose exec backend go build ./...` 成功
- [ ] `docker compose exec backend golangci-lint run ./...` エラー 0 件
