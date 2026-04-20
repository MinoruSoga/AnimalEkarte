# CODE-QUALITY-227: GetEffectivePermissionsByStaffID のマルチテナント分離強化

## 概要

`backend/internal/repository/permission_group_repository.go` の
`GetEffectivePermissionsByStaffID` メソッドが Raw SQL で実装されており、
`staff_id` のみを条件にしているため、スタッフが所属する clinic の検証が行われていない。

## 該当コード

**ファイル:** `backend/internal/repository/permission_group_repository.go:149-176`

```go
func (r *permissionGroupRepository) GetEffectivePermissionsByStaffID(
    ctx context.Context, staffID uint64,
) ([]model.PermissionGroupRule, error) {
    var rules []model.PermissionGroupRule
    err := r.db.WithContext(ctx).Raw(`
        SELECT pgr.resource, ...
        FROM staff_permission_groups spg
        JOIN permission_groups pg
            ON pg.id = spg.group_id
            AND pg.deleted_at IS NULL
            AND pg.is_active = true
        JOIN permission_group_rules pgr
            ON pgr.group_id = pg.id
        WHERE spg.staff_id = ?   ← staff_id のみ。clinic_id のフィルタなし
        GROUP BY pgr.resource
    `, staffID).Scan(&rules).Error
    // ...
}
```

## 問題点

`staff_permission_groups` テーブルから `staff_id` で検索しているが、
そのスタッフが **どの clinic に所属しているか** の検証がない。

現在の認証フロー（middleware）で clinic 境界が別途検証されているため、
実害リスクは低いと考えられるが:

1. このメソッド単体でテストする際に clinic 境界が不明確
2. より将来的な呼び出しパターンの変化でクロステナント混入リスクが生まれる
3. コードを読む者が「clinic_id フィルタがないのは意図的か見落としか」判断できない

## 修正案

`staff_clinic_assignments` を JOIN に追加して clinic 境界を明示:

```go
// 修正後
err := r.db.WithContext(ctx).Raw(`
    SELECT pgr.resource,
           BOOL_OR(pgr.can_view)   AS can_view,
           BOOL_OR(pgr.can_create) AS can_create,
           BOOL_OR(pgr.can_edit)   AS can_edit,
           BOOL_OR(pgr.can_delete) AS can_delete
    FROM staff_permission_groups spg
    JOIN permission_groups pg
        ON pg.id = spg.group_id
        AND pg.deleted_at IS NULL
        AND pg.is_active = true
    JOIN permission_group_rules pgr
        ON pgr.group_id = pg.id
    JOIN staff_clinic_assignments sca   ← 追加: clinic 境界の明示
        ON sca.staff_id = spg.staff_id
        AND sca.deleted_at IS NULL
    WHERE spg.staff_id = ?
    GROUP BY pgr.resource
`, staffID).Scan(&rules).Error
```

または、インターフェースを変更して `clinicID` を引数に追加し、
`pg.clinic_id = ?` を WHERE に追加するアプローチも検討できる。

## インターフェース変更が必要な場合の影響範囲

`GetEffectivePermissionsByStaffID` の変更が必要な箇所:
- `permission_group_repository.go` — interface 定義 + 実装
- `permission_group_service.go` — `GetEffectivePermissions` の呼び出し
- `permission_group_service_test.go` — テストのモック定義
- 認証 middleware（呼び出し元） — 引数追加

## 優先度

MEDIUM — 現在の認証フローで clinic 境界は別途保証されているため緊急性は低い。
ただしコードの明示性（Principle of Explicit Security）として修正することを推奨。
Raw SQL を使っている場合は特に「意図的な省略か漏れか」のコメントを追加すること。
