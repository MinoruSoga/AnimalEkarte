# FE-139: 権限スコープを clinic 単位からフラット構造へ変更

**Status**: Open
**Priority**: High
**Affects**: features/auth/types/index.ts, features/auth/hooks/use-auth.tsx, features/auth/hooks/use-permission.ts
**Date Created**: 2026-03-29
**Related**: TASK-049, BE-082（先に完了必要）
**Blocked by**: BE-082（`/me` レスポンスのフラット化が前提）

---

## Summary

BE-082 完了後、`/me` レスポンスの `permissions` が
`{ clinicId → { resource → CRUD } }` から `{ resource → CRUD }` にフラット化される。

フロントエンドの `AuthUser` 型・`hasPermission()` 関数・`usePermission()` hook を
新しいフラット構造に対応させ、`currentClinicId` スコープを削除する。

---

## 変更内容

### 1. `AuthUser` 型変更（`features/auth/types/index.ts`）

```typescript
// 変更前（現在の実装: ResourcePermission のフィールド名は view/create/edit/delete）
interface ResourcePermission {
  view:   boolean;
  create: boolean;
  edit:   boolean;
  delete: boolean;
}

interface AuthUser {
  id: string;
  email: string;
  userType: UserType;
  // clinic_id → { resource → CRUD } のネスト構造
  permissions: Record<string, Record<string, ResourcePermissions>>;
  clinics: ClinicMembership[];
}

// 変更後
interface AuthUser {
  id: string;
  email: string;
  userType: UserType;
  // resource → CRUD のフラット構造
  permissions: Record<string, ResourcePermissions>;
  clinics: ClinicMembership[];
}
```

### 2. `hasPermission()` 変更（`features/auth/hooks/use-auth.tsx`）

```typescript
// 変更前: currentClinicId でスコープ
const hasPermission = useCallback(
  (resource: string, action: ResourceAction): boolean => {
    if (!user || !currentClinicId) return false;
    if (user.userType === "system_admin" || user.userType === "clinic_admin") return true;
    const clinicPerms = user.permissions[currentClinicId];  // ← 削除
    if (!clinicPerms) return false;
    const resourcePerms = clinicPerms[resource];
    if (!resourcePerms) return false;
    return resourcePerms[action] === true;
  },
  [user, currentClinicId],
);

// 変更後: clinic スコープ不要
const hasPermission = useCallback(
  (resource: string, action: ResourceAction): boolean => {
    if (!user) return false;
    if (user.userType === "system_admin" || user.userType === "clinic_admin") return true;
    const resourcePerms = user.permissions[resource];  // フラットに参照
    if (!resourcePerms) return false;
    return resourcePerms[action] === true;
  },
  [user],  // currentClinicId を deps から削除
);
```

### 3. `/me` レスポンスの変換（`features/auth/api/` の transforms）

```typescript
// BE-082 完了後の /me レスポンス形式（permissions がフラット化される）
interface MeResponse {
  id: string;
  email: string;
  userType: string;
  permissions: {
    [resource: string]: {
      view:   boolean;
      create: boolean;
      edit:   boolean;
      delete: boolean;
    }
  };
  clinics: ClinicInfo[];
}

// transform 関数
export function transformMeResponse(me: MeResponse): AuthUser {
  return {
    id: String(me.id),
    email: me.email,
    userType: me.userType as UserType,
    permissions: me.permissions,  // フラット構造をそのまま使用
    clinics: me.clinics.map(transformClinic),
  };
}
```

---

## `currentClinicId` は権限以外の目的で維持

`currentClinicId` は「どの院のデータを表示するか」のために引き続き必要。
権限チェックの `deps` から削除するだけで、`switchClinic()` 等は変更しない。

```typescript
// ✅ 変更不要: データ取得の clinic_id スコープ
const { data: medicalRecords } = useGetMedicalRecords({ clinicId: currentClinicId });

// ✅ 変更: 権限チェックから currentClinicId を削除
const hasPermission = useCallback(
  (resource, action) => {
    // currentClinicId を参照しない
  },
  [user],  // currentClinicId を deps から削除
);
```

---

## 変更ファイル一覧

| ファイル | 変更内容 |
|---------|---------|
| `features/auth/types/index.ts` | `AuthUser.permissions` 型をフラット化 |
| `features/auth/hooks/use-auth.tsx` | `hasPermission` から `currentClinicId` スコープを削除 |
| `features/auth/api/transforms.ts`（または相当） | `/me` レスポンスの変換ロジックを更新 |

---

## 受入条件

- [ ] `hasPermission("medical-records", "view")` が `currentClinicId` なしで正しく動作する
- [ ] `staff` ユーザーの `usePermission()` が正しい権限を返す
- [ ] `clinic_admin` / `system_admin` は全権限 true
- [ ] `switchClinic()` 後も権限が変わらない（全院同一権限）
- [ ] `docker compose exec frontend pnpm build` 成功
- [ ] `docker compose exec frontend pnpm lint` エラー 0 件
- [ ] `docker compose exec frontend pnpm test:run` 成功
