# BUG-230: レンダーループ内で O(n) .find() を繰り返し（HospitalizationBoard, MedicineSettings）

## 概要
`HospitalizationBoard` および `MedicineSettings` で、レンダーループ内で `.find()` を呼び出しているため O(N×M) の計算量になっている。リストの先頭に `Map` を構築することで O(N+M) に改善できる。

## 現状コード

### `features/hospitalization/routes/HospitalizationBoard.tsx`
```typescript
// ❌ getOccupant() が各セルのレンダーで O(n) .find() を実行
// ボード全体で O(ケージ数 × 入院件数) の計算量
{cages.map(cage => (
  <BoardCell
    key={cage.id}
    occupant={hospitalizations.find(h => h.cageId === cage.id)} // ループ内 find
  />
))}
```

### `features/master/routes/MedicineSettings.tsx`
```typescript
// ❌ レンダー時に category を .find() で都度解決
{medicines.map(medicine => (
  <MedicineRow
    key={medicine.id}
    category={categories.find(c => c.id === medicine.categoryId)} // ループ内 find
  />
))}
```

## 修正方針

`useMemo` でレンダー前に `Map` を構築し、O(1) ルックアップに変換する。

### `features/hospitalization/routes/HospitalizationBoard.tsx`
```typescript
// ✅ Map を先に構築
const occupantByCageId = useMemo(
  () => new Map(hospitalizations.map(h => [h.cageId, h])),
  [hospitalizations]
);

{cages.map(cage => (
  <BoardCell
    key={cage.id}
    occupant={occupantByCageId.get(cage.id)} // O(1)
  />
))}
```

### `features/master/routes/MedicineSettings.tsx`
```typescript
// ✅ Map を先に構築
const categoryById = useMemo(
  () => new Map(categories.map(c => [c.id, c])),
  [categories]
);

{medicines.map(medicine => (
  <MedicineRow
    key={medicine.id}
    category={categoryById.get(medicine.categoryId)} // O(1)
  />
))}
```

## 準拠すべきプロジェクト規約

### `.claude/rules/performance-rules.md` — js-index-maps
> レンダーループ内で `.find()` を繰り返す場合は `Map` でインデックスを構築して O(1) ルックアップに変換する

### プロジェクト内参照実装
`frontend/CODING_RULES.md` Section 12 — `js-index-maps` の `Map` 構築パターン

## 優先度
**Medium** — HospitalizationBoard はケージ数 × 入院件数が増えると顕著に遅くなる。早めに対処すべき。

## 関連ファイル
- `frontend/src/features/hospitalization/routes/HospitalizationBoard.tsx`
- `frontend/src/features/master/routes/MedicineSettings.tsx`
