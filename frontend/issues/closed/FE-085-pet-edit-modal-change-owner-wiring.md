# FE-085: PetEditModal onChangeOwner 配線

**Status**: Open
**Priority**: High
**Affects**: owners feature（PetEditModal）、app/pages（OwnerFormPage）
**Date Created**: 2026-03-19
**Related**: TASK-023, FE-078

## Summary

FE-078 で `PetEditModal` に `onChangeOwner` prop を追加したが、呼び出し元（`OwnerForm.tsx`, `OwnersList.tsx`）から prop が渡されていない。飼主変更ボタンが表示されず、機能が動作しない。

## 現状のコード

### PetEditModal（onChangeOwner prop は定義済み）

```typescript
// frontend/src/features/owners/components/PetEditModal.tsx:57-65
interface PetEditModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  ownerName?: string;
  petData?: PetFormData;
  onSave: (data: PetFormData) => void;
  onChangeOwner?: (newOwner: { id: string; name: string }) => void;  // ← 追加済み
}
```

### OwnerForm.tsx（onChangeOwner 未配線）

```typescript
// frontend/src/features/owners/routes/OwnerForm.tsx:589-596
<PetEditModal
  key={editingPet?.id ?? "new"}
  open={petModalOpen}
  onOpenChange={setPetModalOpen}
  ownerName={ownerData.ownerName || "新規飼主"}
  petData={editingPet ?? undefined}
  onSave={handleSavePet}
  // ← onChangeOwner が渡されていない
/>
```

### OwnerFormPage.tsx（petMutations に updatePetMutate あり）

```typescript
// frontend/src/app/pages/OwnerFormPage.tsx:14-30
export function OwnerFormPage() {
  const { mutate: updatePetMutate } = useUpdatePet();
  const petMutations: PetMutations = {
    updatePetMutate: (args, { onSuccess, onError }) =>
      updatePetMutate(args, { onSuccess, onError }),
    // ...
  };
  return <OwnerForm petMutations={petMutations} />;
}
```

## 必要な変更

### 1. OwnerForm.tsx に onChangeOwner ハンドラ追加

```typescript
// frontend/src/features/owners/routes/OwnerForm.tsx

// petMutations.updatePetMutate を使って owner_id を変更
const handlePetChangeOwner = useCallback(
  (newOwner: { id: string; name: string }) => {
    if (!editingPet?.id) return;
    petMutations.updatePetMutate(
      { id: editingPet.id, req: { owner_id: Number(newOwner.id) } },
      {
        onSuccess: () => {
          toast.success(`飼主を ${newOwner.name} に変更しました`);
          setPetModalOpen(false);
          // ペットリストを再取得（React Query invalidation で自動）
        },
        onError: () => toast.error("飼主変更に失敗しました"),
      },
    );
  },
  [editingPet, petMutations, setPetModalOpen],
);
```

### 2. PetEditModal に onChangeOwner を渡す

```typescript
// frontend/src/features/owners/routes/OwnerForm.tsx
<PetEditModal
  key={editingPet?.id ?? "new"}
  open={petModalOpen}
  onOpenChange={setPetModalOpen}
  ownerName={ownerData.ownerName || "新規飼主"}
  petData={editingPet ?? undefined}
  onSave={handleSavePet}
  onChangeOwner={handlePetChangeOwner}  // ← 追加
/>
```

### 3. OwnersList.tsx の PetEditModal にも同様に配線（必要な場合）

`OwnersList.tsx:519` の PetEditModal にも `onChangeOwner` を渡す。OwnersList はペット一覧からの編集なので、飼主変更機能が必要。

## UI 操作フロー

1. 飼主フォームで既存ペットの「編集」をクリック → PetEditModal が開く
2. モーダルヘッダーの「飼主変更」ボタンが表示される（isEdit && onChangeOwner）
3. クリック → OwnerSearchModal が開く
4. 飼主を検索・選択
5. 確認ダイアログ → 確定
6. PATCH `/api/v1/pets/{id}` with `{ owner_id: newOwnerId }`
7. toast「飼主を {新飼主名} に変更しました」

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] cross-feature import なし（petMutations は props 注入）
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `useCallback` でハンドラ安定化

## 依存関係

- FE-078 が完了済み（PetEditModal に onChangeOwner prop 定義済み）
- Backend のペット更新 API は owner_id 変更サポート済み

## 完了条件

- [ ] ペット編集モーダル（既存ペット）に「飼主変更」ボタンが表示される
- [ ] 飼主変更ボタン → 検索 → 選択 → 確認 → PATCH API → 成功 toast
- [ ] 新規ペット登録時は「飼主変更」ボタンが表示されない
- [ ] `npm run build` パス
