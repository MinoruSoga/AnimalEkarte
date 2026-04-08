# BUG-228: 軽量計算に useMemo を使っている（2箇所）

## 概要

`useMemo` のオーバーヘッド（比較コスト + クロージャ生成）が、
メモ化する計算コスト自体を上回っている箇所が 2件ある。
`useMemo` は複雑な計算や配列変換のためにあり、単純な関数呼び出し・恒等変換には不要。

## 現状コード（2箇所 — 実コード確認済み）

### 1. `components/shared/OwnerSearchModal/OwnerSearchModal.tsx:110`

```typescript
// ❌ useMemo が owners をそのまま返すだけ（恒等変換）
const filteredOwners = useMemo(() => owners, [owners]);
```

`owners` がそのまま返されているだけで、フィルタリングも変換も行っていない。
`useMemo` のメモ化オーバーヘッドがゼロのメリットを上回る。

### 2. `features/master/routes/MedicineSettings.tsx:435`

```typescript
// ❌ 単純な null チェック + 2プロパティ比較を useMemo でラップ
const isCategory = useMemo(() => isCategoryMedicine(selectedMedicine), [selectedMedicine]);

// isCategoryMedicine の実装 (line 409-411):
function isCategoryMedicine(m: Medicine | null): boolean {
  return m != null && !m.parentId && m.price === 0;
}
```

`isCategoryMedicine` は null チェックと 2つのプロパティ比較のみ。
これほど軽量な計算は `useMemo` なしで毎レンダーに実行しても全く問題ない。

## 修正方針

### 1. OwnerSearchModal.tsx

```typescript
// Before:
const filteredOwners = useMemo(() => owners, [owners]);

// After: useMemo を削除して直接使用
const filteredOwners = owners;
```

### 2. MedicineSettings.tsx

```typescript
// Before:
const isCategory = useMemo(() => isCategoryMedicine(selectedMedicine), [selectedMedicine]);

// After: 直接計算
const isCategory = isCategoryMedicine(selectedMedicine);
```

## なぜ StaffSettings の `useMemo(() => data ?? [], [data])` は例外か

`StaffSettings.tsx:382-391` の `useMemo(() => allOccupationsData ?? [], [allOccupationsData])` は
一見同じパターンに見えるが、下流の `useMemo([allOccupations])` の参照安定化が目的であり除外する:

```typescript
const allOccupations = useMemo(() => allOccupationsData ?? [], [allOccupationsData]);
//                                                                ↑
const occupationSelectItems = useMemo(
  () => allOccupations.filter(...).map(...),
  [allOccupations],  // ← allOccupations 参照が安定しないと毎回再計算される
);
```

`allOccupationsData` が `undefined` の間、`?? []` は毎レンダーで新しい `[]` を生成する。
`useMemo` でラップすることで同じ `[]` 参照が返り、下流の `useMemo` の無駄な再計算を防いでいる。
これは `rerender-memo-with-default-value` を useMemo で解決している正当なパターン。

## 影響範囲

| ファイル | 行 | 問題 |
|---------|-----|------|
| `components/shared/OwnerSearchModal/OwnerSearchModal.tsx` | 110 | 恒等変換の useMemo |
| `features/master/routes/MedicineSettings.tsx` | 435 | 2プロパティ比較の useMemo |

## 準拠すべきプロジェクト規約・ベストプラクティス

### Vercel React Best Practices — `rerender-simple-expression-in-memo`
> Avoid memo for simple primitives.
> Don't use useMemo if the memoization overhead exceeds the computation cost.

## 優先度

**Low** — 機能的な影響はなし。コードの明確化と僅かなメモリ節約。
