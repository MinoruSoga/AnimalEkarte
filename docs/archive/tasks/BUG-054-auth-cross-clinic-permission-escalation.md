# BUG-054: 認証 — クロスクリニック権限昇格バグ

**ステータス**: Closed（実装確認済み 2026-03-29）
> `use-auth.tsx:109-121` で `user.permissions[currentClinicId]` によるスコープが実装済みであることを確認。本バグは存在しない。

## 概要

`hasPermission` が `currentClinicId` でフィルタせず全グループを走査するため、
複数クリニックに所属する `staff` ユーザーが、現在選択していないクリニックの
権限グループを通じて意図しないアクセス権を取得できる。

## 重要度

**CRITICAL** — データ分離の根幹に関わるセキュリティバグ。

## 再現条件

1. ユーザー A を clinic-1（権限グループ: 管理者）と clinic-2（権限グループ: 一般）に所属させる
2. clinic-2 にログイン（`currentClinicId = clinic-2`）
3. `hasPermission("medical-records", "delete")` を呼ぶ
4. 期待値: `false`（一般グループは medical-records.delete 不可）
5. 実際: `true`（clinic-1 の管理者グループのルールが適用される）

## 原因

`frontend/src/features/auth/hooks/use-auth.tsx` の `hasPermission` 実装が
`user.permissionGroups` を `currentClinicId` でフィルタせず全件走査している。

```typescript
// ❌ 現在の実装（バグあり）
return user.permissionGroups.some((group) =>
  group.rules.some((rule) => { ... })
);
// ↑ clinic-1 のグループも clinic-2 のグループも両方チェックされる
```

## 修正方針

### フロントエンド (`use-auth.tsx`)

`AuthUser.permissions` を `{ clinicId → { resource → CRUD } }` 構造に変更し、
`currentClinicId` でスコープしてから判定する。

```typescript
// ✅ 修正後
function hasPermission(resource: string, action: string): boolean {
  if (!user || !currentClinicId) return false;
  if (user.userType === "system_admin" || user.userType === "clinic_admin") return true;
  const clinicPerms = user.permissions[currentClinicId];
  if (!clinicPerms) return false;
  const resourcePerms = clinicPerms[resource];
  if (!resourcePerms) return false;
  switch (action) {
    case "view":   return resourcePerms.canView;
    case "create": return resourcePerms.canCreate;
    case "edit":   return resourcePerms.canEdit;
    case "delete": return resourcePerms.canDelete;
    default:       return false;
  }
}
```

### バックエンド (`/me` レスポンス)

`permissions` フィールドを `{ clinicId → { resource → CRUD } }` 形式で返す。
フロントがグループ構造を再計算しなくて済むよう、実効権限はバックエンドで計算済みにする。

```go
// MeResponse.Permissions の構造
type ResourcePermission struct {
    View   bool `json:"canView"`
    Create bool `json:"canCreate"`
    Edit   bool `json:"canEdit"`
    Delete bool `json:"canDelete"`
}
// map[clinicID]map[resource]ResourcePermission
```

バックエンドはすでに `findEffectivePermissions()` で clinic 単位の実効権限を計算しているため、
レスポンス構造をこの形式に揃えるだけでよい。

## 影響範囲

- `frontend/src/features/auth/hooks/use-auth.tsx` — `hasPermission` 実装
- `frontend/src/features/auth/types/index.ts` — `AuthUser` 型（`permissionGroups` → `permissions`）
- `backend/internal/handler/auth_handler.go` — `/me` レスポンス構造
- 全 `usePermission(resource)` 呼び出し箇所（型変更の影響確認）

## 関連

- `docs/AUTH.md` §8 権限チェックパターン
- BUG-033: スタッフからユーザー一覧が取得可能（同系統の認可不備）
