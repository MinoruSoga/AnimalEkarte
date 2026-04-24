# TASK-022: フロントエンド deep 監査 MEDIUM 問題 4件

## 概要

`app/` / `components/shared/` / `hooks/` レイヤーで発見された MEDIUM 優先度の問題。

## 優先度

MEDIUM

---

## 問題 1: use-pet-selection.ts — togglePetSelection に useCallback なし

### ファイル
`frontend/src/hooks/use-pet-selection.ts:18`

### 問題
`isPetSelected` は `useCallback` 適用済みだが、`togglePetSelection` は未適用。`memo()` コンポーネントに渡すと毎レンダーで参照が変わり不要な再レンダーが発生する。

### 修正案
```typescript
const togglePetSelection = useCallback((pet: Pet) => {
  setSelectedPets((prev) => {
    const isSelected = prev.some((p) => p.id === pet.id);
    return isSelected ? prev.filter((p) => p.id !== pet.id) : [...prev, pet];
  });
}, []); // setSelectedPets は安定参照なので deps 不要
```

---

## 問題 2: reservation-types — 非 null assertion (!) の使用

### ファイル
- `frontend/src/features/reservations/api/get-reservation-types.ts:54`
- `frontend/src/hooks/use-reservation-types.ts:64`

### 問題
`map.get(key)!.types.push(t)` が `!` 非 null assertion を使用している。`map.has(key)` の直後なので安全だが、規約では `!` 乱用を禁止している。

### 修正案
```typescript
// Before
map.get(key)!.types.push(t);

// After
const entry = map.get(key);
if (entry) entry.types.push(t);
```

---

## 問題 3: MasterSidePanel に memo() なし

### ファイル
`frontend/src/components/shared/SidePeek/MasterSidePanel.tsx`（119行）

### 問題
`DataTable`, `NotionFilter`, `Pagination`, `SidePeekPanel` には `memo()` が適用済みだが、`MasterSidePanel` のみ未適用で一貫性が欠ける。

### 修正案
```typescript
export const MasterSidePanel = memo(function MasterSidePanel({ ... }: Props) {
  return ...;
});
```

---

## 問題 4: router.tsx で auth/routes/Login へ deep import

### ファイル
`frontend/src/app/router.tsx:13`

### 問題
```typescript
// 現状: index.ts を経由しない deep import
const Login = lazy(() =>
  import("@/features/auth/routes/Login").then((m) => ({ default: m.Login })),
);
```

他のルートはすべて `@/features/xxx` index 経由なのに Login のみ例外。

### 修正案
`frontend/src/features/auth/index.ts` に `Login` が export されていれば:
```typescript
const Login = lazy(() =>
  import("@/features/auth").then((m) => ({ default: m.Login })),
);
```

**注意**: `lazy()` の named export と Vite のチャンク分割の相性を確認してから対応すること。現状の直接参照が意図的な場合はコメントを追加して許容することも可。
