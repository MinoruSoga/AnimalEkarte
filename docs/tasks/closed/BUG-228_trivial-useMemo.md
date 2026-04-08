# BUG-228: 軽量計算に useMemo を使用（OwnerSearchModal, MedicineSettings）

## 概要
`OwnerSearchModal` および `MedicineSettings` で、ほぼコストゼロの計算に `useMemo` を使用している。`useMemo` 自体にもオーバーヘッドがあるため、軽量な計算では削除した方がパフォーマンスが向上する。

**注意**: `MedicineSettings` の `useMemo(() => data ?? [], [data])` は downstream の `useMemo([allOccupations])` のための参照安定化が目的であり、削除対象外。削除すると連鎖的な再レンダーが発生するため除外。

## 現状コード

### `features/owners/components/OwnerSearchModal.tsx`（推定）
```typescript
// ❌ 軽量な boolean 計算に useMemo
const hasResults = useMemo(
  () => owners.length > 0,
  [owners.length]
);
```

### `features/master/routes/MedicineSettings.tsx`（削除対象の箇所のみ）
```typescript
// ❌ 軽量な primitive 計算に useMemo
const totalCount = useMemo(
  () => medicines.length,
  [medicines]
);
```

## 修正方針

軽量な計算は `useMemo` なしでレンダー中に直接計算する。

```typescript
// ✅ 直接計算
const hasResults = owners.length > 0;
const totalCount = medicines.length;
```

## 準拠すべきプロジェクト規約

### `frontend/CODING_RULES.md` Section 12 — rerender-simple-expression-in-memo
> 軽量計算（boolean, 単純プロパティアクセス）には `useMemo` を使わない

## 優先度
**Low** — useMemo のオーバーヘッドを削減。修正は10分（2箇所）。

## 関連ファイル
- `frontend/src/features/owners/components/OwnerSearchModal.tsx`
- `frontend/src/features/master/routes/MedicineSettings.tsx`（削除対象箇所のみ）
