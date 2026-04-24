# BE-112: `discount` リソース追加 + 全 service の権限チェック

**Status**: Closed (2026-04-14, commit 8fcd1382)
**Priority**: High
**Affects**: `permission.go`, owner/treatment/hospitalization/estimate/accounting service, マイグレーション
**Date Created**: 2026-04-14
**Related**: BUG-372, FE-250

## Summary

新規 `Resource = "discount"` を定義し、5 つの service の更新メソッドで discount フィールドが含まれる場合に `discount:edit` / `discount:create` 権限をチェック。既存環境への自動マイグレーションも含む。

## 現状のコード

### 既存リソース定義
```go
// backend/internal/model/permission.go:6-32
type Resource string
const (
    ResourceReception Resource = "reception"
    ResourceOwners    Resource = "owners"
    // ... 24 リソース定義済み
)
```

### Discount フィールドの存在箇所
```go
// backend/internal/handler/owner_request.go:40,62
DiscountRate float64  `json:"discount_rate"`     // CreateOwnerRequest
DiscountRate *float64 `json:"discount_rate"`     // UpdateOwnerRequest

// backend/internal/handler/treatment_request.go:17,36
DiscountRate float64  `json:"discount_rate"`     // Create
DiscountRate *float64 `json:"discount_rate"`     // Update

// backend/internal/handler/estimate_request.go:15,30
DiscountAmount int64  `json:"discount_amount"`   // Create
DiscountAmount *int64 `json:"discount_amount"`   // Update

// backend/internal/handler/accounting_request.go:42
DiscountAmount *int64 `json:"discount_amount"`

// backend/internal/model/hospitalization.go:116-117
DiscountRate   float64 `gorm:"type:numeric(5,2);default:0"`
DiscountAmount int64   `gorm:"default:0"`

// backend/internal/model/estimate.go:32,60-61
DiscountAmount int64
DiscountRate   float64
DiscountAmount int64
```

### 既存権限チェック実装パターン
```go
// backend/internal/handler/handler.go:178-181
accountings.PATCH("/:id", h.RequirePermission(string(model.ResourceAccounting), "edit"), h.UpdateAccounting)
```

## 必要な変更

### 1. リソース定義追加

```go
// backend/internal/model/permission.go

const (
    // ... 既存リソース ...
    ResourceMasterMerchandise Resource = "master-merchandise"

    // BUG-372: 割引フィールド専用権限
    ResourceDiscount Resource = "discount"
)

// AllResources に追加
var AllResources = []Resource{
    // ... 既存 24 リソース ...
    ResourceDiscount,
}
```

### 2. 権限チェックヘルパー（共通化）

```go
// backend/internal/service/discount_permission.go (新規)

package service

import (
    "context"

    apperrors "github.com/animal-ekarte/backend/internal/errors"
    "github.com/animal-ekarte/backend/internal/model"
)

// PermissionChecker は actor に対するリソース×アクション権限の判定インターフェース。
// auth サービスが実装する。
type PermissionChecker interface {
    HasPermission(ctx context.Context, resource model.Resource, action string) (bool, error)
}

// requireDiscountPermission は discount リソースの指定アクション権限を要求する。
// 権限なしなら apperrors.ErrForbidden を返す。
func requireDiscountPermission(ctx context.Context, checker PermissionChecker, action string) error {
    ok, err := checker.HasPermission(ctx, model.ResourceDiscount, action)
    if err != nil {
        return apperrors.Wrap(err, "failed to check discount permission")
    }
    if !ok {
        return apperrors.WrapForbidden("割引フィールドの編集権限がありません")
    }
    return nil
}
```

### 3. Service 変更（5 ファイル）

#### 3.1 OwnerService

```go
// backend/internal/service/owner_service.go

func (s *ownerService) Update(ctx context.Context, input *UpdateOwnerInput) (*model.Owner, error) {
    // BUG-372: discount_rate が変更される場合のみ権限チェック
    if input.DiscountRate != nil {
        // 既存値と同じなら権限不要
        existing, err := s.repo.FindByID(ctx, input.ClinicID, input.ID)
        if err != nil {
            return nil, apperrors.Wrap(err, "failed to find owner")
        }
        if !floatEquals(*input.DiscountRate, existing.DiscountRate) {
            if err := requireDiscountPermission(ctx, s.permChecker, "edit"); err != nil {
                return nil, err
            }
        }
    }
    // ... 既存処理 ...
}

func (s *ownerService) Create(ctx context.Context, input *CreateOwnerInput) (*model.Owner, error) {
    // BUG-372: 0 以外の discount_rate を指定する場合のみ権限チェック
    if input.DiscountRate != 0 {
        if err := requireDiscountPermission(ctx, s.permChecker, "create"); err != nil {
            return nil, err
        }
    }
    // ... 既存処理 ...
}

// 浮動小数比較ヘルパー
func floatEquals(a, b float64) bool {
    const epsilon = 0.0001
    diff := a - b
    return diff < epsilon && diff > -epsilon
}
```

#### 3.2 TreatmentService / HospitalizationService / EstimateService / AccountingService

同様のパターンを各 service に適用:
- `*float64` / `*int64` の nil チェック
- 既存値との比較
- `requireDiscountPermission(ctx, checker, "edit"|"create")`

具体的な対象:
- `TreatmentService.Create/Update` — `discount_rate`
- `HospitalizationService.Create/Update` — `discount_rate`, `discount_amount`
- `EstimateService.Create/Update` および EstimateItem 系 — `discount_amount`, `discount_rate`
- `AccountingService.Update` — `discount_amount`（payment フィールド経由、`accounting_service.go:53`）
- BillingItem 系 — `discount_rate`, `discount_amount`

### 4. 各 Service への PermissionChecker 注入

```go
// 各 service の構造体・コンストラクタに permChecker を追加

type ownerService struct {
    repo        repository.OwnerRepository
    permChecker PermissionChecker  // 追加
}

func NewOwnerService(repo repository.OwnerRepository, permChecker PermissionChecker) OwnerService {
    return &ownerService{repo: repo, permChecker: permChecker}
}
```

`cmd/api/main.go` の DI 配線で auth サービスを各 service に渡す。

### 5. context から actor を取得する仕組み

`PermissionChecker` 実装は `ctx` から actor (staff_id) を取得し、その staff の権限グループ UNION から `discount` リソースの該当アクション権限を判定する。

既存実装の確認:
- `backend/internal/middleware/auth.go` で context に staff_id を埋め込んでいるはず → `extractStaffID(ctx)` ヘルパーが存在するか確認の上、PermissionChecker 実装で利用

### 6. 既存環境マイグレーション

```sql
-- backend/migrations/004_add_discount_permission.sql (新規)

-- 既存全権限グループに discount リソース行を追加
-- is_system_admin スタッフが所属するグループは ON、それ以外は OFF
INSERT INTO permission_group_rules (group_id, resource, can_view, can_create, can_edit, can_delete, created_at, updated_at)
SELECT
    pg.id AS group_id,
    'discount' AS resource,
    EXISTS (
        SELECT 1
        FROM staff_permission_groups spg
        JOIN staffs s ON s.id = spg.staff_id
        WHERE spg.group_id = pg.id AND s.is_system_admin = true
    ) AS can_view,
    EXISTS (
        SELECT 1 FROM staff_permission_groups spg
        JOIN staffs s ON s.id = spg.staff_id
        WHERE spg.group_id = pg.id AND s.is_system_admin = true
    ) AS can_create,
    EXISTS (
        SELECT 1 FROM staff_permission_groups spg
        JOIN staffs s ON s.id = spg.staff_id
        WHERE spg.group_id = pg.id AND s.is_system_admin = true
    ) AS can_edit,
    false AS can_delete, -- 削除アクションは割引には不要
    now(), now()
FROM permission_groups pg
WHERE NOT EXISTS (
    SELECT 1 FROM permission_group_rules pgr
    WHERE pgr.group_id = pg.id AND pgr.resource = 'discount'
);
```

### 7. tygo codegen

```bash
make codegen
# → frontend/src/types/generated/models.ts に ResourceDiscount = "discount" が自動生成される
```

## API レスポンス形式

権限なしで discount フィールドを変更しようとした場合:

```json
// PATCH /api/clinics/1/owners/100
// Body: { "discount_rate": 30 }
// 既存 discount_rate が 0 で、actor に discount:edit 権限がない場合
// → 403 Forbidden
{
  "code": "forbidden",
  "message": "割引フィールドの編集権限がありません",
  "timestamp": "..."
}
```

## フロントエンド影響

- `make codegen` で `models.ts` に `ResourceDiscount = "discount"` が自動追加される
- FE-250 で `usePermission("discount")` を使った UI 制御が必要

## 完了条件

- [ ] `ResourceDiscount` 追加 + `AllResources` 登録
- [ ] `requireDiscountPermission` ヘルパー実装
- [ ] PermissionChecker インターフェース定義 + auth service 実装
- [ ] 5 service に権限チェック追加（owner / treatment / hospitalization / estimate / accounting）
- [ ] DI 配線変更（cmd/api/main.go）
- [ ] マイグレーション 004 追加 + 既存環境への適用確認
- [ ] テストケース追加（テーブル駆動）
  - 正常: 権限あり → 200
  - 異常: 権限なし + 値変更あり → 403
  - 正常: 権限なし + 既存値と同じ → 200（ゼロ値再送）
  - 正常: 権限なし + discount_rate を指定しない → 200
- [ ] `go test ./... -race` パス
- [ ] `golangci-lint run` パス
- [ ] `make codegen` で models.ts に反映確認
