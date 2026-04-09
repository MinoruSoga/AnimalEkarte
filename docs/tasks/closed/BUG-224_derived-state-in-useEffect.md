# BUG-224: useEffect で derived state を同期（owners/inventory 各1件）

## 概要
`useEffect` 内で derived state（他の state から計算できる値）を `setState` で同期しているパターンが2件ある。これは余分なレンダーサイクルを発生させ、React の推奨パターンに反する。derived state はレンダー中に直接計算すべきである。

## 現状コード

### `features/owners/routes/OwnersList.tsx`（または owners 内の hooks）
```typescript
// ❌ useEffect で derived state を同期
const [filteredOwners, setFilteredOwners] = useState<Owner[]>([]);

useEffect(() => {
  setFilteredOwners(owners.filter(o => o.name.includes(searchTerm)));
}, [owners, searchTerm]);
```

### `features/inventory/` 内の同様パターン
```typescript
// ❌ 同様の useEffect による同期
const [displayItems, setDisplayItems] = useState<InventoryItem[]>([]);

useEffect(() => {
  setDisplayItems(items.filter(/* ... */));
}, [items, someFilter]);
```

## 修正方針

`useState` + `useEffect` の組み合わせを廃止し、レンダー中に `useMemo` で直接計算する。

```typescript
// ✅ レンダー中に直接 derive
const filteredOwners = useMemo(
  () => owners.filter(o => o.name.includes(deferredSearchTerm)),
  [owners, deferredSearchTerm]
);
```

## 準拠すべきプロジェクト規約

### `.claude/rules/typescript-react.md` — rerender-derived-state-no-effect
> `useEffect` で derived state を同期しない。レンダー中に `useMemo` で直接計算する。

### プロジェクト内参照実装
`features/owners/routes/OwnersList.tsx` — `useDeferredValue` + `useMemo` でフィルタを直接計算するパターン

## 優先度
**Low** — 余分なレンダーサイクルを1回削減。機能的影響なし。

## 関連ファイル
- `frontend/src/features/owners/` — 該当 hooks または routes
- `frontend/src/features/inventory/` — 該当 hooks または routes
