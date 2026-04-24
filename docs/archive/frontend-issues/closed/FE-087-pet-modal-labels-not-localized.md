# FE-087: ペット登録モーダルのラベル「CFBE」「BREED」を日本語化

**Status**: Won't Fix
**Priority**: Medium
**Affects**: 飼主編集画面 (`/owners/:id`) のペット追加モーダル
**Date Created**: 2026-03-21
**Related**: BUG-002

## Summary

`features/owners/components/PetEditModal.tsx` でペット種別ラベルが「CFBE」、品種ラベルが「BREED」と英語のまま表示されている。日本語ラベル（「動物種」「品種」）に修正する。

## 現状のコード

```typescript
// frontend/src/features/owners/components/PetEditModal.tsx:160
if (!formData.animalSpeciesId) errors.animalSpeciesId = "CFBEを選択してください";

// frontend/src/features/owners/components/PetEditModal.tsx:275
CFBE <span className={C.textRequired}>*</span>

// frontend/src/features/owners/components/PetEditModal.tsx:333-334
<Label htmlFor="breed" className={LABEL_CLS}>
  BREED
</Label>
```

## 必要な変更

```typescript
// PetEditModal.tsx:160 バリデーションメッセージ
// Before
if (!formData.animalSpeciesId) errors.animalSpeciesId = "CFBEを選択してください";
// After
if (!formData.animalSpeciesId) errors.animalSpeciesId = "動物種を選択してください";

// PetEditModal.tsx:275 ラベルテキスト
// Before
CFBE <span className={C.textRequired}>*</span>
// After
動物種 <span className={C.textRequired}>*</span>

// PetEditModal.tsx:333-334 ラベルテキスト
// Before
<Label htmlFor="breed" className={LABEL_CLS}>
  BREED
</Label>
// After
<Label htmlFor="breed" className={LABEL_CLS}>
  品種
</Label>
```

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）

## 依存関係

なし（テキスト変更のみ）

## 完了条件

- [ ] ペット追加モーダルで「CFBE」→「動物種」が表示される
- [ ] ペット追加モーダルで「BREED」→「品種」が表示される
- [ ] バリデーションエラーメッセージも「動物種を選択してください」になる
- [ ] `pnpm build` が通る
- [ ] `pnpm lint` がエラーなし
