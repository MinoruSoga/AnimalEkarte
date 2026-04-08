# BUG-221: delete/update mutation が useTransition でラップされていない（7箇所）

## 概要

`useMutation().mutate()` の呼び出しが `useTransition` でラップされていない。
これにより、API 書き込み中の UI が応答不能になり、ユーザーがレンダリング中の React ツリーを操作できない。
参照実装 (`owners/`) では全ての mutation を `startDeleteTransition(() => { mutate(...) })` で包んでいる。

## 現状コード（7箇所 — 実コード確認済み）

### 1. `features/reservations/hooks/use-reservation-management.ts:177,235,262,283`

```typescript
// 4箇所の mutate() が useTransition なし
updateMutation.mutate(updatePayload, { ... });  // :177
updateMutation.mutate(updatePayload, { ... });  // :235
updateMutation.mutate(updatePayload, { ... });  // :262
deleteMutation.mutate(deleteTarget.id, { ... }); // :283
```

### 2. `features/hospitalization/routes/HospitalizationForm.tsx:105`

```typescript
deleteMutation.mutate(hospitalizationId, {
  onSuccess: () => { ... },
});
```

### 3. `features/vaccinations/routes/VaccinationList.tsx:144`

```typescript
deleteMutation.mutate(pendingDeleteId, {
  onSuccess: () => { ... },
  onError: (error) => handleApiError(error, "削除"),
});
```

### 4. `features/estimates/routes/EstimateDetail.tsx:27`

```typescript
const handleDelete = () => {       // ← useCallback すら使っていない
  if (!id) return;
  deleteEstimate(id, {
    onSuccess: () => navigate('/estimates'),
  });
};
```

### 5. `features/estimates/routes/EstimateList.tsx:174-178`

```typescript
const handleDeleteConfirm = useCallback(() => {
  if (deleteModal.item == null) return;
  deleteEstimate(deleteModal.item);  // ← useTransition なし
  deleteModal.close();
}, [deleteModal, deleteEstimate]);
```

### 6. `features/master/routes/CompanySettings.tsx:128`

```typescript
updateMutation.mutate(req, { ... });
```

### 7. `features/hospital-settings/routes/ClinicMasterSettings.tsx:273`

```typescript
deleteMutation.mutate(pendingDeleteId, { ... });
```

## 比較: 正しい実装（プロジェクト内参照実装）

```typescript
// features/owners/routes/OwnersList.tsx
const [isPending, startDeleteTransition] = useTransition();

const handleDeleteConfirm = useCallback(() => {
  if (pendingDeleteOwnerId === null) return;
  startDeleteTransition(() => {
    deleteOwnerFn(String(pendingDeleteOwnerId), {
      onSuccess: () => { ... },
      onError: (error) => handleApiError(error, "削除"),
    });
  });
}, [pendingDeleteOwnerId, deleteOwnerFn]);
```

## 修正方針

各ファイルに以下のパターンを適用する:

```typescript
// 1. useTransition を追加
const [isDeletePending, startDeleteTransition] = useTransition();

// 2. mutate を startDeleteTransition でラップ
const handleDeleteConfirm = useCallback(() => {
  if (!target) return;
  startDeleteTransition(() => {
    deleteMutation.mutate(target.id, {
      onSuccess: () => { ... },
      onError: (error) => handleApiError(error, "削除"),
    });
  });
}, [target, deleteMutation]); // isPending は deps から外す（startDeleteTransition は安定）

// 3. ローディング状態は isPending を使用
<Button disabled={isDeletePending}>削除</Button>
```

## 影響範囲

| ファイル | 行 | mutation の種類 |
|---------|-----|----------------|
| `features/reservations/hooks/use-reservation-management.ts` | 177,235,262,283 | update×3, delete×1 |
| `features/hospitalization/routes/HospitalizationForm.tsx` | 105 | delete |
| `features/vaccinations/routes/VaccinationList.tsx` | 144 | delete |
| `features/estimates/routes/EstimateDetail.tsx` | 27 | delete（useCallback なし） |
| `features/estimates/routes/EstimateList.tsx` | 174 | delete |
| `features/master/routes/CompanySettings.tsx` | 128 | update |
| `features/hospital-settings/routes/ClinicMasterSettings.tsx` | 273 | delete |

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/code-style.md` — `rerender-transitions`
> API 書き込みに `useTransition`

### `.claude/rules/typescript-react.md` §8
> API ミューテーションは `useTransition`

### プロジェクト内参照実装
`features/owners/routes/OwnersList.tsx` — `startDeleteTransition(() => { deleteOwnerFn(...) })`

## 優先度

**Medium** — パフォーマンス改善。削除/更新中の UI ブロッキングが解消される。実害（データ破損等）はなし。

## 関連チケット

なし
