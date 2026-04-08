# BUG-226: `.filter().map()` による二重イテレーション（4箇所）

## 概要

配列を `.filter()` した後に `.map()` をチェーンしている箇所が 4件ある。
これは配列を 2回走査するため、`reduce` や `flatMap` を使った 1パスに統一すべき。
特に `trimming/api/get-trimmings.ts` は API レスポンス全体を対象とするため影響が大きい。

## 現状コード（4箇所 — 実コード確認済み）

### 1. `features/trimming/api/get-trimmings.ts:18`

```typescript
// ❌ API レスポンス全件を filter → map で 2回走査
return data.data.filter((d) => d.pet?.id != null).map(transformTrimming);
```

### 2. `features/reception/routes/Reception.tsx:64`

```typescript
// ❌ staffs を filter → map で 2回走査（useMemo 内）
...staffs.filter((s) => s.is_active).map((s) => ({ id: s.name, name: s.name })),
```

### 3. `features/medical-records/components/VitalsTab/VitalsGraph.tsx:192`

```typescript
// ❌ METRICS を filter → map で 2回走査（JSX 内）
{METRICS.filter((m) => activeMetrics.has(m.key)).map((m) => (
  <Line key={m.key} ... />
))}
```

### 4. `features/medical-records/components/MedicalRecordVaccination.tsx:34`

```typescript
// ❌ vaccinesMaster を filter → map で 2回走査（useMemo 内）
() => vaccinesMaster.filter((v) => v.isActive).map((v) => ({ value: v.id, label: v.name })),
```

## 修正方針

### パターン A: `reduce` による 1パス

```typescript
// Before (get-trimmings.ts:18)
return data.data.filter((d) => d.pet?.id != null).map(transformTrimming);

// After
return data.data.reduce<Trimming[]>((acc, d) => {
  if (d.pet?.id != null) acc.push(transformTrimming(d));
  return acc;
}, []);
```

### パターン B: `flatMap` による 1パス（短く書ける場合）

```typescript
// Before (Reception.tsx:64)
...staffs.filter((s) => s.is_active).map((s) => ({ id: s.name, name: s.name }))

// After
...staffs.flatMap((s) =>
  s.is_active ? [{ id: s.name, name: s.name }] : []
)
```

```typescript
// Before (MedicalRecordVaccination.tsx:34)
vaccinesMaster.filter((v) => v.isActive).map((v) => ({ value: v.id, label: v.name }))

// After
vaccinesMaster.flatMap((v) =>
  v.isActive ? [{ value: v.id, label: v.name }] : []
)
```

### VitalsGraph (JSX 内)

```typescript
// Before (VitalsGraph.tsx:192)
{METRICS.filter((m) => activeMetrics.has(m.key)).map((m) => (
  <Line key={m.key} ... />
))}

// After
{METRICS.flatMap((m) =>
  activeMetrics.has(m.key) ? [<Line key={m.key} ... />] : []
)}
```

## 影響範囲

| ファイル | 行 | 対象配列 | 規模 |
|---------|-----|---------|------|
| `features/trimming/api/get-trimmings.ts` | 18 | API レスポンス全件 | 影響大 |
| `features/reception/routes/Reception.tsx` | 64 | staffs（スタッフ一覧） | 影響中 |
| `features/medical-records/components/VitalsTab/VitalsGraph.tsx` | 192 | METRICS（固定 5件程度） | 影響小 |
| `features/medical-records/components/MedicalRecordVaccination.tsx` | 34 | vaccinesMaster | 影響中 |

## 準拠すべきプロジェクト規約・ベストプラクティス

### Vercel React Best Practices — `js-combine-iterations`
> Combine multiple filter/map into one loop.
> Chained .filter().map() iterates the array twice unnecessarily.

## 優先度

**Low** — `get-trimmings.ts` は API レスポンス処理のため修正優先度が相対的に高い。
残り 3件は小さな配列（5〜数十件）のため実害は軽微。コードの一貫性のために対応する。
