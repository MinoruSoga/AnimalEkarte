---
status: closed
closed_at: 2026-03-16
---

# [master] MedicineSettings: UI コンポーネントが他マスタページと不統一

## 優先度
中

## 種別
仕様違反 / UI一貫性

## 対象ファイル
`frontend/src/features/master/routes/MedicineSettings.tsx`

## 問題

`MedicineSettings.tsx` のみ、他のマスタページと異なる UI コンポーネントを使用している。

### 1. `<select>` ネイティブ要素を使用（サイドパネル内）

サイドパネルの「親カテゴリ」フィールドで shadcn/ui の `Select` ではなく、ネイティブ `<select>` を使用している。
他のすべてのサイドパネル（DiagnosisNameSidePanel など）は shadcn/ui `Select` を使用しており、UI が不統一。

```tsx
// 現状
<select onChange={...}>
  <option value="">カテゴリなし</option>
  {categories.map(c => <option key={c.id} value={c.id}>{c.name}</option>)}
</select>

// 期待
<Select onValueChange={...}>
  <SelectTrigger /><SelectContent>
    <SelectItem value="">カテゴリなし</SelectItem>
    ...
  </SelectContent>
</Select>
```

### 2. 独自 `StatusDot` コンポーネントを定義（ステータス表示）

`NotionStatusPill` を使わず、ファイル内にローカルの `StatusDot` コンポーネントを定義している。
全マスタページで `NotionStatusPill` に統一すること。

```tsx
// 現状（削除すべき）
const StatusDot = ({ isActive }: { isActive: boolean }) => (
  <span className={cn("inline-block h-2 w-2 rounded-full", isActive ? "bg-green-500" : "bg-gray-300")} />
);
```

## 修正方針

1. ネイティブ `<select>` を shadcn/ui `<Select>` に置き換える
2. ローカル `StatusDot` を削除し、`NotionStatusPill` に変更する
3. `StatusDot` のインポートも削除する
