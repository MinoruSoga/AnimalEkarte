# FE-091: ペット選択画面の前回来院日がISO形式で表示される

**Status**: Open
**Priority**: Low
**Affects**: ペット選択画面（診察・入院・トリミング等の患者選択）
**Date Created**: 2026-03-21
**Related**: BUG-007

## Summary

`components/shared/PetSelection/PetSelectionResultsTable.tsx:79` で前回来院日を `pet.lastVisit` としてそのまま表示しているため、`2015-08-28T00:00:00Z` のようなISO 8601形式が表示される。`formatDate()` を適用して `2015/08/28` 形式に変換する。

## 現状のコード

```typescript
// frontend/src/components/shared/PetSelection/PetSelectionResultsTable.tsx:79
<TableCell className="font-mono text-sm text-[#37352F] whitespace-nowrap py-2">
  {pet.lastVisit || "-"}  // ← formatDate() を呼んでいない
</TableCell>
```

```typescript
// frontend/src/lib/transforms/pet.ts:81
lastVisit: p.last_visit ?? undefined,
// last_visit は "2015-08-28T00:00:00Z" などのISO文字列
```

既存の `formatDate()` ユーティリティ:
```typescript
// frontend/src/utils/format/date.ts:11
// ISO形式を "YYYY/MM/DD" に変換する関数が既に存在する
export function formatDate(dateString: string | undefined | null): string { ... }
```

## 必要な変更

```typescript
// frontend/src/components/shared/PetSelection/PetSelectionResultsTable.tsx

// Before
import { ... } from "...";

// After: formatDate をインポート
import { formatDate } from "@/utils/format/date";

// Before（line:79）
{pet.lastVisit || "-"}

// After
{formatDate(pet.lastVisit)}
// formatDate は undefined/null で "-" を返すため || "-" は不要
```

## 影響範囲

`PetSelectionResultsTable.tsx` は `PetSelection.tsx` から呼ばれており、ペット選択を使う全ページに影響:
- `/examinations/select-pet`
- `/hospitalization/select-pet`
- `/trimming/select-pet`
- その他患者選択ページ

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] barrel index 経由 import なし（直接ファイルから import）

## 依存関係

なし

## 完了条件

- [ ] ペット選択画面の前回来院日が `2015/08/28` 形式で表示される
- [ ] データなし（undefined）の場合は `-` が表示される
- [ ] `npm run build` が通る
