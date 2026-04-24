# FE-088: 飼主一覧の飼主Noソートを数値ソートに修正

**Status**: Open
**Priority**: Medium
**Affects**: 飼主一覧 (`/owners`) のソート機能
**Date Created**: 2026-03-21
**Related**: BUG-003

## Summary

`OwnersList.tsx` のソート処理がすべてのキーで `localeCompare` を使用しているため、「飼主No」カラムが文字列（辞書順）ソートになっている。`ownerNumber` キーに対しては数値比較に変更する。

## 現状のコード

```typescript
// frontend/src/features/owners/routes/OwnersList.tsx:229-243
const sortedData = useMemo(() => {
  if (activeSorts.length === 0) return [...filteredPets];
  const sorted = [...filteredPets];
  sorted.sort((a, b) => {
    for (const sort of activeSorts) {
      const key = sort.key as SortKey;
      const aVal = String(a[key as keyof typeof a] ?? "");
      const bVal = String(b[key as keyof typeof b] ?? "");
      const cmp = aVal.localeCompare(bVal, "ja");  // ← 全フィールドで文字列ソート
      if (cmp !== 0) return sort.direction === "asc" ? cmp : -cmp;
    }
    return 0;
  });
  return sorted;
}, [filteredPets, activeSorts]);
```

現象: `ownerNumber` が `"1", "10", "11", "2"` の順で並ぶ（文字列辞書順）。

## 必要な変更

```typescript
// frontend/src/features/owners/routes/OwnersList.tsx
const sortedData = useMemo(() => {
  if (activeSorts.length === 0) return [...filteredPets];
  const sorted = [...filteredPets];
  sorted.sort((a, b) => {
    for (const sort of activeSorts) {
      const key = sort.key as SortKey;
      let cmp: number;

      // ownerNumber は数値ソート
      if (key === "ownerNumber") {
        const aNum = Number(a[key as keyof typeof a] ?? 0);
        const bNum = Number(b[key as keyof typeof b] ?? 0);
        cmp = aNum - bNum;
      } else {
        const aVal = String(a[key as keyof typeof a] ?? "");
        const bVal = String(b[key as keyof typeof b] ?? "");
        cmp = aVal.localeCompare(bVal, "ja");
      }

      if (cmp !== 0) return sort.direction === "asc" ? cmp : -cmp;
    }
    return 0;
  });
  return sorted;
}, [filteredPets, activeSorts]);
```

**注意**: FE-086 (owners ベース表示への変更) と同じファイルを修正するため、FE-086 と同時 or 順次対応すること。

## プロジェクトルール遵守チェック

- [ ] `any` 型なし

## 依存関係

- FE-086 と変更対象ファイルが重なるため、調整が必要（どちらか先に対応し、もう一方はリベース/マージ）

## 完了条件

- [ ] 「飼主No」列クリックで 1, 2, 3, ..., 10, 11, 12 の順に数値ソートされる
- [ ] 他のカラム（飼主名、ペット名等）は従来通り日本語文字列ソート
- [ ] `pnpm build` が通る
