# FE-102: useModalState カスタムフック共有化

**Status**: Closed
**Priority**: Medium
**Affects**: owners, vaccinations, inventory, trimming, examinations, medical-records, estimates
**Date Created**: 2026-03-24
**Related**: TASK-024

## Summary

削除確認ダイアログやモーダルの開閉を管理する `useState<T | null>` パターンが 7〜10 箇所で重複実装されており、状態変数名（`pendingDeleteOwner` / `deleteTarget` / `deleteTargetId`）と構造（オブジェクト vs 単一ID）も不統一。`useModalState<T>()` カスタムフックで統一する。

## 現状のコード

```typescript
// frontend/src/features/owners/routes/OwnersList.tsx:130-136, 510-534
const [pendingDeleteOwner, setPendingDeleteOwner] = useState<{
  id: string;
  name: string;
} | null>(null);

const handleDeleteRequest = useCallback((ownerId: string, ownerName: string) => {
  setPendingDeleteOwner({ id: ownerId, name: ownerName });
}, []);
// ...
<ConfirmDialog
  open={!!pendingDeleteOwner}
  onClose={() => !isDeleting && setPendingDeleteOwner(null)}
  ...
/>
```

```typescript
// frontend/src/features/trimming/routes/TrimmingList.tsx:197, 203-208
const [deleteTarget, setDeleteTarget] = useState<{ id: string; label: string } | null>(null);

const handleDeleteClick = useCallback((record: TrimmingRecord) => {
  setDeleteTarget({ id: record.id, label: `${record.ownerName} - ${record.petName}` });
}, []);
```

```typescript
// frontend/src/features/medical-records/routes/MedicalRecords.tsx:52
const [deleteTarget, setDeleteTarget] = useState<{ id: string; label: string } | null>(null);
```

```typescript
// frontend/src/features/estimates/routes/EstimateList.tsx:71
// 構造が異なる（単一IDのみ）
const [deleteTargetId, setDeleteTargetId] = useState<string | null>(null);
```

## 必要な変更

### 1. useModalState フック作成

```typescript
// frontend/src/hooks/useModalState.ts（新規作成）

import { useState, useCallback } from "react";

export function useModalState<T>(initialValue: T | null = null) {
  const [item, setItem] = useState<T | null>(initialValue);

  const open = useCallback((value: T) => setItem(value), []);
  const close = useCallback(() => setItem(null), []);

  return {
    item,
    isOpen: item !== null,
    open,
    close,
  };
}
```

### 2. 各ページの置き換え（例: TrimmingList.tsx）

```typescript
// Before:
const [deleteTarget, setDeleteTarget] = useState<{ id: string; label: string } | null>(null);
const handleDeleteClick = useCallback((record) => {
  setDeleteTarget({ id: record.id, label: `${record.ownerName} - ${record.petName}` });
}, []);

// After:
import { useModalState } from "@/hooks/useModalState";

const deleteModal = useModalState<{ id: string; label: string }>();

const handleDeleteClick = useCallback((record: TrimmingRecord) => {
  deleteModal.open({ id: record.id, label: `${record.ownerName} - ${record.petName}` });
}, [deleteModal.open]);

// レンダー部分
<ConfirmDialog
  open={deleteModal.isOpen}
  onClose={deleteModal.close}
  title={`「${deleteModal.item?.label}」を削除しますか？`}
  ...
/>
```

同様の置き換えを以下で実施:
- `OwnersList.tsx`（`pendingDeleteOwner` + `selectedPet`）
- `VaccinationList.tsx`
- `InventoryList.tsx`
- `MedicalRecords.tsx`
- `EstimateList.tsx`（`deleteTargetId` → `deleteModal.item.id` に統一）

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] barrel index 経由 import なし（`hooks/useModalState` を直接 import）
- [ ] `useCallback` で `open` / `close` を安定化（`memo()` の前提条件）
- [ ] `FC` / `forwardRef` なし

## 依存関係

- Backend 変更なし。FE-101 とも独立。単独で実施可能。

## 完了条件

- [ ] `frontend/src/hooks/useModalState.ts` が作成されている
- [ ] 対象ページの削除確認ダイアログ開閉が `useModalState` を使用している
- [ ] 削除確認 → 削除実行 → ダイアログ閉じる の一連フローが変化なし
- [ ] `pnpm lint` エラーなし
- [ ] `pnpm build` エラーなし

## クローズ情報

- **Closed At**: 2026-03-24
- **変更ファイル**:
  - `frontend/src/hooks/use-modal-state.ts` — 新規作成
  - `frontend/src/features/trimming/routes/TrimmingList.tsx` — `deleteTarget` → `deleteModal`
  - `frontend/src/features/medical-records/routes/MedicalRecords.tsx` — `deleteTarget` → `deleteModal`
  - `frontend/src/features/estimates/routes/EstimateList.tsx` — `deleteTargetId` → `deleteModal`
  - `frontend/src/features/owners/routes/OwnersList.tsx` — `pendingDeleteOwner` → `deleteModal`、`selectedPet` → `petModal`
