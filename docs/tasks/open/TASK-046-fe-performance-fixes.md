# TASK-046: FE パフォーマンス修正（renderRow / COLUMNS / useHospitalizations）

**作成日**: 2026-03-27
**ステータス**: Open
**優先度**: Medium
**領域**: Frontend

---

## 概要

複数のリストコンポーネントで `useCallback`/`useMemo` の適用漏れが検出された。
`useHospitalizations` が React Query を使わず `useState + useEffect` パターンを使用しており、二重データ取得の可能性がある。

---

## 対応項目

### 1. `MedicalRecords.tsx:144-170` — `COLUMNS` を `useMemo` に変更

```tsx
// Before（レンダーごとに再生成）
const COLUMNS = [...]  // コンポーネント内部で定義

// After
const COLUMNS = useMemo(() => [...], [directionFor, toggleSort]);
// または依存がないカラムはモジュール定数に巻き上げ
```

### 2. `EstimateList.tsx:172-211` — `renderRow` を `useCallback` でラップ

```tsx
// Before
const renderRow = (estimate: Estimate) => (...)

// After
const renderRow = useCallback((estimate: Estimate) => (...), [navigate, ...deps]);
```

また同ファイルの URL ハードコード (`/estimates/${estimate.id}`) を `config/paths.ts` 経由に変更する。

### 3. `ExaminationsList.tsx`, `InventoryList.tsx`, `VaccinationList.tsx` — `renderRow` を `useCallback` または `memo` コンポーネントに変更

`DataTable` に渡すインライン `renderRow` アロー関数をレンダー外に出す。
`TrimmingList.tsx` の `TrimmingTableRow` (`memo`) パターンを参照すること。

### 4. `hospitalization/hooks/use-hospitalizations.ts` — `useGetHospitalizations` に一本化

```ts
// Before: useState + useEffect
const [data, setData] = useState([]);
const [isLoading, setIsLoading] = useState(true);
useEffect(() => { loadData(); }, [...]);

// After: React Query に一本化
// useGetHospitalizations（既存）を使用し、use-hospitalizations.ts を廃止
```

`HospitalizationList.tsx` で `useHospitalizationList` を使っている箇所を `useGetHospitalizations` に置き換える。

---

## 受入条件

- [ ] `MedicalRecords.tsx` の `COLUMNS` が `useMemo` またはモジュール定数になっている
- [ ] `EstimateList.tsx` の `renderRow` が `useCallback` でラップされている
- [ ] URL が `paths.ts` 経由になっている（`EstimateList.tsx`）
- [ ] `ExaminationsList`, `InventoryList`, `VaccinationList` の `renderRow` が安定化されている
- [ ] `use-hospitalizations.ts` が廃止または `useGetHospitalizations` に委譲している
- [ ] `docker compose exec frontend npm run lint` エラー 0 件
- [ ] `docker compose exec frontend npx tsc --noEmit` エラー 0 件
