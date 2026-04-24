# FE-128: AuthContext 型更新 + usePermission hook 実装

**Status**: Closed
**Priority**: High
**Affects**: features/auth/types/index.ts, features/auth/hooks/use-auth.tsx（新規: use-permission.ts）
**Date Created**: 2026-03-26
**Related**: TASK-032, BE-075（先に完了必要）, FE-129

## Summary

`Permission` 型（旧 PermissionType 文字列配列）を廃止し、`EffectivePermissions`（resource×CRUD マップ）に移行する。
`hasPermission(permission)` のシグネチャを変更し、`usePermission(resource)` hook を新規作成する。

## 現状のコード

**`frontend/src/features/auth/types/index.ts`:**
```typescript
// 廃止対象
export type Permission =
  | "account_admin"
  | "medical"
  | "medical_read"
  | "trimming"
  | "billing"
  | "reception"
  | "hospitalization"
  | "master_admin"
  | "shift_admin"
  | "inventory";

// 変更対象
export interface AuthUser {
  // ...
  permissions: ClinicPermissions; // Record<clinicId, Permission[]>
}
```

**`frontend/src/features/auth/hooks/use-auth.tsx:hasPermission`:**
```typescript
// 廃止対象のシグネチャ
const hasPermission = useCallback(
  (permission: Permission): boolean => {
    if (!user || !currentClinicId) return false;
    if (user.userType === "system_admin") return true;
    if (user.userType === "clinic_admin") {
      return user.clinics.some(c => c.clinicId === currentClinicId);
    }
    const clinicPerms = user.permissions[currentClinicId];
    if (!clinicPerms) return false;
    return clinicPerms.includes(permission);
  },
  [user, currentClinicId],
);
```

## 必要な変更

### 1. 型定義の更新（`frontend/src/features/auth/types/index.ts`）

```typescript
// Before: 削除
export type Permission = "account_admin" | "medical" | ...;
export type ClinicPermissions = Record<string, Permission[]>;

// After: 追加
export type ResourceAction = "view" | "create" | "edit" | "delete";

export interface ResourcePermission {
  view:   boolean;
  create: boolean;
  edit:   boolean;
  delete: boolean;
}

// resource → CRUD
export type ResourcePermissions = Record<string, ResourcePermission>;

// clinicId → resource → CRUD
export type ClinicEffectivePermissions = Record<string, ResourcePermissions>;

// AuthUser の permissions フィールド型変更
export interface AuthUser {
  id:           string;
  email:        string;
  displayName:  string;
  userType:     UserType;
  staffRole:    StaffRole | null;
  avatarUrl:    string | null;
  mainClinicId: string;
  clinic:       AuthClinic | null;
  clinics:      ClinicMembership[];
  permissions:  ClinicEffectivePermissions;  // ← 型変更
}
```

### 2. hasPermission のシグネチャ変更（`use-auth.tsx`）

```typescript
// Before: 削除
const hasPermission = useCallback((permission: Permission): boolean => { ... }, [...]);

// After: 変更
const hasPermission = useCallback(
  (resource: string, action: ResourceAction): boolean => {
    if (!user || !currentClinicId) return false;
    // system_admin / clinic_admin はバイパス（バックエンドが全権限 true で返す想定だが念のため）
    if (user.userType === "system_admin" || user.userType === "clinic_admin") return true;
    const clinicPerms = user.permissions[currentClinicId];
    if (!clinicPerms) return false;
    const resourcePerms = clinicPerms[resource];
    if (!resourcePerms) return false;
    return resourcePerms[action] === true;
  },
  [user, currentClinicId],
);
```

### 3. hasAnyPermission を削除（不要になる）

旧 `hasAnyPermission` は PermissionType の配列チェック用だったため削除。
用途は `usePermission` hook で代替する。

### 4. 新規: usePermission hook（`frontend/src/features/auth/hooks/use-permission.ts`）

```typescript
import { useAuth } from "@/features/auth/hooks/use-auth";
import type { ResourceAction } from "@/features/auth/types";

export interface UsePermissionResult {
  canView:   boolean;
  canCreate: boolean;
  canEdit:   boolean;
  canDelete: boolean;
}

/**
 * 現在のユーザーが指定リソースに対して持つ権限を返す。
 * system_admin / clinic_admin は常に true。
 * staff はグループUNIONから計算された実効権限（バックエンド計算済み）を使用。
 */
export function usePermission(resource: string): UsePermissionResult {
  const { hasPermission } = useAuth();
  return {
    canView:   hasPermission(resource, "view"),
    canCreate: hasPermission(resource, "create"),
    canEdit:   hasPermission(resource, "edit"),
    canDelete: hasPermission(resource, "delete"),
  };
}
```

### 5. AuthContext の値型を更新

```typescript
// use-auth.tsx の AuthContextValue
interface AuthContextValue {
  // ...変更なし
  hasPermission: (resource: string, action: ResourceAction) => boolean;
  // hasAnyPermission 削除
}
```

### 6. transformMeResponse の型変換（`features/auth/api/` 内の変換関数）

バックエンドから受け取る `permissions` は既に `{clinicId: {resource: {view, create, edit, delete}}}` 形式なので、
変換ロジックは単純なマッピングになる。既存の `Permission[]` → `EffectivePermissions` の変換は不要。

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `useTransition` で pending 管理（`useState(false)` 禁止）
- [ ] 型は `models.ts` から導出（手書き interface は ResourcePermission のみ許容）

## 依存関係

- BE-075 が完了していること（/me レスポンスの `permissions` 形式が変更済み）

## 完了条件

- [ ] `Permission` 型・`ClinicPermissions` 型が削除されている
- [ ] `ResourcePermission`, `ResourcePermissions`, `ClinicEffectivePermissions` 型が追加されている
- [ ] `AuthUser.permissions` の型が `ClinicEffectivePermissions` になっている
- [ ] `hasPermission(resource, action)` のシグネチャに変更されている
- [ ] `usePermission("accounting")` が `{canView, canCreate, canEdit, canDelete}` を返す
- [ ] `pnpm build` で型エラーなし（`hasPermission` の旧呼び出し箇所がすべて更新されている）
- [ ] `pnpm lint` パス

## クローズ情報

- **Closed At**: 2026-03-26
- **変更ファイル**:
  - `frontend/src/features/auth/types/index.ts` — Permission/ClinicPermissions削除、ResourcePermission/ResourcePermissions/ClinicEffectivePermissions/ResourceAction追加
  - `frontend/src/features/auth/hooks/use-auth.tsx` — hasPermission シグネチャ変更、hasAnyPermission削除
  - `frontend/src/features/auth/api/transforms.ts` — permissions schema を新形式に更新
  - `frontend/src/features/auth/hooks/use-permission.ts` — 新規作成
  - `frontend/src/features/auth/index.ts` — export更新
