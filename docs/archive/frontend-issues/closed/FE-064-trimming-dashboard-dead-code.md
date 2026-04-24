# FE-064: トリミング・当日の受付 — 未使用フィルタ API + import 除去

**Status**: Open
**Priority**: Medium
**Affects**: `features/trimming/api/`, `features/trimming/routes/`
**Date Created**: 2026-03-18
**Related**: TASK-014

## Summary

trimming から未使用フィルタ API 関数6件と TrimmingForm 内の未使用 import 1件を削除する。
dashboard はデッドコードなし。

## 現状のコード

### 1. trimming — 6関数が未使用

```typescript
// frontend/src/features/trimming/api/get-trimming.ts
// 以下6関数が一度もインポートされていない:
// :21-27  getTrimmingsByPetId()
// :30-35  useGetTrimmingsByPetId()
// :39-45  getTrimmingsByOwnerId()
// :48-53  useGetTrimmingsByOwnerId()
// :57-63  getTrimmingsByStatus()
// :66-71  useGetTrimmingsByStatus()

// frontend/src/features/trimming/api/index.ts:8-14
// 上記6関数の barrel 再エクスポート（未使用）
```

### 2. trimming — 未使用 import

```typescript
// frontend/src/features/trimming/routes/TrimmingForm.tsx:36
import { useGetTrimmingsByPetId } from "../api/get-trimming";
// ❌ インポートされているがコンポーネント内で一度も使用されていない
```

### 3. dashboard — デッドコードなし ✅

全 API 関数・コンポーネントが使用されていることを確認済み。

## 必要な変更

### 1. get-trimming.ts

未使用6関数を削除。`getTrimming()` と `useGetTrimming()` は使用されているため残す。

### 2. api/index.ts

対応する再エクスポート行（8-14行目）を削除。

### 3. TrimmingForm.tsx

```typescript
// Before: line 36
import { useGetTrimmingsByPetId } from "../api/get-trimming";

// After: この行を削除
```

## 完了条件

- [ ] trimming: 6関数が `get-trimming.ts` から削除されている
- [ ] trimming: `index.ts` の対応再エクスポートが削除されている
- [ ] trimming: `TrimmingForm.tsx` の未使用 import が削除されている
- [ ] `pnpm build` パス
- [ ] `pnpm lint` パス
