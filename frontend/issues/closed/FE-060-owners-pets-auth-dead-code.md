# FE-060: 飼主・ペット・認証 — dead files + 再エクスポート除去

**Status**: Open
**Priority**: Medium
**Affects**: `features/auth/hooks/`, `hooks/use-pet.ts`
**Date Created**: 2026-03-18
**Related**: TASK-014

## Summary

認証 feature の重複ファイル2件を削除し、共通 hooks の未使用再エクスポート1件を除去する。

## 現状のコード

### 1. 重複 auth hook ファイル（2件）

```typescript
// frontend/src/features/auth/hooks/use-auth.ts（全11行）
// useAuth.tsx と重複する useAuth() 関数を export — 一度もインポートされていない
import { useContext } from "react";
import { AuthContext } from "./useAuth";
export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) throw new Error("useAuth must be used within AuthProvider");
  return context;
}

// frontend/src/features/auth/hooks/auth-context.ts（全4行）
// AuthContext の再エクスポート — 一度もインポートされていない
export { AuthContext } from "./useAuth";
```

**正規ファイル**: `features/auth/hooks/useAuth.tsx`（AuthProvider + useAuth hook + AuthContext を全て含む）
**auth/index.ts**: `useAuth.tsx` からのみエクスポートしており、上記2ファイルは参照されていない。

### 2. 未使用再エクスポート（use-pet.ts）

```typescript
// frontend/src/hooks/use-pet.ts:4
export { useGetPet };
// ❌ 一度もインポートされていない。コンポーネントは @/features/pets/api/get-pet から直接インポート
```

**同ファイル内の `usePetSearch()` と `usePetInfo()` は使用されているため、ファイル自体は残す。**

## 必要な変更

### 1. dead ファイル削除

```bash
rm frontend/src/features/auth/hooks/use-auth.ts
rm frontend/src/features/auth/hooks/auth-context.ts
```

### 2. use-pet.ts の未使用 export 除去

```typescript
// frontend/src/hooks/use-pet.ts
// Before: line 4
export { useGetPet };

// After: line 4 を削除
// usePetSearch() と usePetInfo() の export はそのまま残す
```

## 完了条件

- [ ] `auth/hooks/use-auth.ts` が削除されている
- [ ] `auth/hooks/auth-context.ts` が削除されている
- [ ] `hooks/use-pet.ts` から `useGetPet` の再エクスポートが除去されている
- [ ] `npm run build` パス
- [ ] `npm run lint` パス
