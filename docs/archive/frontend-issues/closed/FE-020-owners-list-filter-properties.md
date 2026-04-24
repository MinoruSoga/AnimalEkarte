# FE-020: 飼主一覧 — フィルタプロパティ追加

**Status**: Open
**Priority**: Medium
**Affects**: owners feature — 飼主一覧
**Date Created**: 2026-03-17
**Related**: TASK-005

## Summary

飼主一覧は NotionFilter を使用しているが、FilterProperty が空配列のため「+ フィルタを追加」ボタンが表示されない。種（犬/猫）等のフィルタプロパティを追加して、他の一覧ページと同じ操作体験を提供する。

## 現状のコード

```typescript
// frontend/src/features/owners/routes/OwnersList.tsx:256
<NotionFilter
  properties={[]}  // ← 空配列のため FilterAddPopover が return null
  activeFilters={[]}
  onFilterChange={() => {}}
  searchTerm={searchTerm}
  onSearchChange={setSearchTerm}
  searchPlaceholder="飼主名、ペット名、飼主No、種別..."
  count={...}
/>
```

## 必要な変更

```typescript
// frontend/src/features/owners/routes/OwnersList.tsx

const OWNER_FILTER_PROPERTIES: FilterProperty[] = [
  {
    key: "species",
    label: "種",
    type: "select",
    icon: PawPrint,
    options: [
      { value: "犬", label: "犬" },
      { value: "猫", label: "猫" },
      { value: "鳥", label: "鳥" },
      { value: "うさぎ", label: "うさぎ" },
      { value: "ハムスター", label: "ハムスター" },
      { value: "その他", label: "その他" },
    ],
  },
  {
    key: "status",
    label: "生死",
    type: "select",
    icon: Heart,
    options: [
      { value: "alive", label: "生存" },
      { value: "deceased", label: "死亡" },
    ],
  },
];
```

フィルタ適用ロジック（クライアント側フィルタ）:
```typescript
const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);

const filteredPets = useMemo(() => {
  let result = pets;

  // ActiveFilter からフィルタ適用
  const speciesFilter = activeFilters.find((f) => f.key === "species");
  if (speciesFilter) {
    result = result.filter((p) => p.species === speciesFilter.value);
  }
  const statusFilter = activeFilters.find((f) => f.key === "status");
  if (statusFilter) {
    // alive/deceased でフィルタ
  }

  // テキスト検索
  if (deferredSearchTerm) { ... }

  return result;
}, [pets, activeFilters, deferredSearchTerm]);
```

## 完了条件

- [ ] 飼主一覧に「+ フィルタを追加」ボタンが表示される
- [ ] 種（犬/猫/鳥...）でフィルタ可能
- [ ] 生死でフィルタ可能
- [ ] フィルタがピルで表示・削除可能
- [ ] テキスト検索との併用が動作
- [ ] 型エラーなし（`docker compose exec frontend pnpm build` パス）
