# BUG-234: O(n) `.some()` / `.includes()` を Set/Map で O(1) に最適化（3箇所）

## 概要
`PetSelection.tsx`、`use-pet-selection.ts`、`use-reception-kanban.ts` で、選択済みIDの存在チェックに `.some()` または `.includes()` を使用している。選択済みIDを `Set` で管理することで O(1) のルックアップに改善できる。

## 現状コード

### `components/shared/PetSelection/PetSelection.tsx:49`
```typescript
// ❌ レンダーループ内で O(n) .some()
filteredPets.map(pet => {
  const isSelected = selectedPets.some(p => p.id === pet.id); // O(n)
  return <div key={pet.id} className={isSelected ? "..." : "..."}>...</div>;
})
```

### `hooks/use-pet-selection.ts:33`
```typescript
// ❌ O(n) .some() で存在チェック
const isPetSelected = (pet: Pet) => selectedPets.some(p => p.id === pet.id);
```

### `features/reception/hooks/use-reception-kanban.ts:95`
```typescript
// ❌ O(n) .includes() で visit type フィルタ
if (!selectedVisitTypes.includes(app.visitType)) return false;
```

## 修正方針

### `use-pet-selection.ts` — selectedPetIds を Set で管理
```typescript
// ✅ Set で O(1) ルックアップ
const [selectedPets, setSelectedPets] = useState<Pet[]>([]);
const selectedPetIdSet = useMemo(
  () => new Set(selectedPets.map(p => p.id)),
  [selectedPets]
);

const isPetSelected = useCallback(
  (pet: Pet) => selectedPetIdSet.has(pet.id), // O(1)
  [selectedPetIdSet]
);
```

### `PetSelection.tsx:49`
```typescript
// ✅ Set を受け取るか、isPetSelected(pet) を使う
const isSelected = selectedPetIdSet.has(pet.id); // O(1)
```

### `use-reception-kanban.ts:95`
```typescript
// ✅ Set で管理
const selectedVisitTypeSet = useMemo(
  () => new Set(selectedVisitTypes),
  [selectedVisitTypes]
);

if (!selectedVisitTypeSet.has(app.visitType)) return false; // O(1)
```

## 影響範囲

| ファイル | 行 | 現状 | 修正後 |
|---------|-----|------|-------|
| `PetSelection.tsx` | 49 | O(n) .some() | O(1) Set.has() |
| `use-pet-selection.ts` | 33 | O(n) .some() | O(1) Set.has() |
| `use-reception-kanban.ts` | 95 | O(n) .includes() | O(1) Set.has() |

## 準拠すべきプロジェクト規約

### `.claude/rules/performance-rules.md` — js-set-map-lookups
> Set/Map で O(1) ルックアップ — 繰り返し存在チェックには `.includes()` / `.some()` でなく `Set` を使う

## 優先度
**Low** — 選択済みペット数・訪問タイプ数が多くなければ実害は軽微。コードの一貫性改善と将来的なスケール対応が目的。

## 関連ファイル
- `frontend/src/components/shared/PetSelection/PetSelection.tsx:49`
- `frontend/src/hooks/use-pet-selection.ts:33`
- `frontend/src/features/reception/hooks/use-reception-kanban.ts:95`
