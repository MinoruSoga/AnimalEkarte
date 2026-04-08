# BUG-230: ループ内で O(n) の .find() を繰り返している（2箇所）

## 概要

配列を繰り返し `.find()` で線形探索している箇所が 2件ある。
Map を事前に構築することで O(1) で検索できる。
特に `HospitalizationBoard` はケージ一覧のレンダーループ内で毎回 `.find()` を呼んでおり、
ケージ数 × 在院患者数 = O(N×M) の計算コストがある。

## 現状コード（2箇所 — 実コード確認済み）

### 1. `features/hospitalization/components/HospitalizationBoard.tsx:157-158,171`

```typescript
// ❌ ヘルパー関数がレンダーループ内で毎回 O(n) 探索
const getOccupant = (cageId: string) => {
  return hospitalizations.find(h => h.cageId === cageId && h.status === "入院中");
};

// JSX レンダーループ内で呼ばれる
{cagesByArea.flatMap(area => area.cages).map((cage) => {
  const occupant = getOccupant(String(cage.id));  // ← 毎回 O(n) scan
  return <CageCard key={cage.id} occupant={occupant} ... />;
})}
```

ケージが 20個あり在院患者が 10名いれば、毎レンダーで 200回の比較が発生する。

### 2. `features/master/routes/MedicineSettings.tsx:555-556`

```typescript
// ❌ DnD の handleDragEnd イベントハンドラで 2回 O(n) 探索
const handleDragEnd = useCallback((event: DragEndEvent) => {
  const { active, over } = event;
  const activeMedicine = orderedMedicines.find((m) => m.id === active.id);   // O(n)
  const overMedicine   = orderedMedicines.find((m) => m.id === over?.id);     // O(n)
  ...
}, [orderedMedicines, ...]);
```

DnD イベントは頻繁には発生しないが、`orderedMedicines` が deps にある結果、
`orderedMedicines` が変わるたびにコールバック全体が再生成される問題もある。

## 修正方針

### 1. HospitalizationBoard.tsx — useMemo で Map を事前構築

```typescript
// useMemo で cageId → Hospitalization の Map を一度だけ構築（O(n)）
const occupantByCageId = useMemo(() =>
  new Map(
    hospitalizations
      .filter(h => h.status === "入院中")
      .map(h => [h.cageId, h])
  ),
  [hospitalizations]
);

// ヘルパー関数を削除し、Map で O(1) 検索
{cagesByArea.flatMap(area => area.cages).map((cage) => {
  const occupant = occupantByCageId.get(String(cage.id));  // ✅ O(1)
  return <CageCard key={cage.id} occupant={occupant} ... />;
})}
```

### 2. MedicineSettings.tsx — Map を deps に追加（または BUG-222 と合わせて修正）

```typescript
// useMemo で id → Medicine の Map を構築
const orderedMedicinesById = useMemo(
  () => new Map(orderedMedicines.map(m => [m.id, m])),
  [orderedMedicines]
);

const handleDragEnd = useCallback((event: DragEndEvent) => {
  const activeMedicine = orderedMedicinesById.get(String(active.id));   // ✅ O(1)
  const overMedicine   = orderedMedicinesById.get(String(over?.id));     // ✅ O(1)
}, [orderedMedicinesById, ...]);  // Map が stable なら deps が安定
```

## 影響範囲

| ファイル | 行 | 計算量 | 状況 |
|---------|-----|--------|------|
| `features/hospitalization/components/HospitalizationBoard.tsx` | 157-158,171 | O(N×M)、レンダーごと | 影響大（常時発生） |
| `features/master/routes/MedicineSettings.tsx` | 555-556 | O(n)×2、DnD 時 | 影響小（操作時のみ） |

## 準拠すべきプロジェクト規約・ベストプラクティス

### Vercel React Best Practices — `js-index-maps`
> Build Map for repeated lookups.
> Using .find() inside render or frequently-called handlers is O(n) per call.
> Build a Map once (O(n)) and look up in O(1).

### プロジェクト内参照実装
`features/master/routes/MedicineSettings.tsx:474-477` — `medicinesById` Map が既に構築済み:
```typescript
const medicinesById = useMemo(
  () => new Map(medicines?.map(m => [m.id, m])),
  [medicines]
);
```
同ファイルの他の箇所では正しく Map を使っているが、`orderedMedicines` 用の Map が不足している。

## 優先度

**Medium** — `HospitalizationBoard` はレンダーごとに実行されるため、入院患者が多い病院では影響が出る可能性がある。`MedicineSettings` は操作時のみのため影響は小さい。
