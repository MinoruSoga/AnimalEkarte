# BUG-165: 複数一覧ページで isLoading/isError 状態を UI にフィードバックしていない

## 概要

`VaccinationList`、`InventoryList`、`HospitalizationList` の 3 ページで、
データ取得中・取得失敗時の UI フィードバックが実装されていない。
各ページのフィルタフックが `isLoading` / `isError` を expose していないため、
空テーブルが表示されたままになる。

## 再現手順

1. 各ページ（ワクチン一覧 / 在庫一覧 / 入院一覧）をネットワーク低速でリロード
2. または、バックエンドを一時停止した状態でページを開く
3. **結果**: ワクチン/在庫/入院一覧が空テーブルのまま表示され、エラーメッセージも出ない

## 期待する動作

- ローディング中は `<LoadingFallback />` を表示する
- データ取得失敗時は `<ErrorFallback />` を表示する

## 現状コード

### `frontend/src/features/vaccinations/routes/VaccinationList.tsx:84付近`
```tsx
// Before: isLoading / isError を返さないフック
const { data: filteredRecords, allVaccinations } = useFilterVaccinations(
  deferredSearchTerm, filters, activeFilters
);
// ↑ isLoading / isError がないため確認不可能
// Line 252: データ取得中も count={filteredRecords.length} を表示してしまう
```

### `frontend/src/features/inventory/routes/InventoryList.tsx:111付近`
```tsx
// Before: isLoading / isError を返さないフック
const { data: filteredItems, summary } = useInventory({ ... });
// ↑ isLoading / isError がないため確認不可能
// Line 294: データ取得中も count={filteredItems.length} を表示してしまう
```

### `frontend/src/features/hospitalization/routes/HospitalizationList.tsx:105付近`
```tsx
// Before: isLoading チェックなし
const { data: allHospitalizations = [] } = useGetHospitalizations(dateFilters);
// ↑ isLoading を分割代入していないため未確認
// 入院ボード / 一覧ビューが空のまま表示される
```

### 比較: 正しい実装（プロジェクト内参照実装）
```tsx
// frontend/src/features/medical-records/routes/MedicalRecords.tsx:196-197
if (isLoading) return <LoadingFallback />;
if (isError) return <ErrorFallback />;

// frontend/src/features/checkups/routes/CheckupsList.tsx:70,183-184
const { data: checkups = [], isLoading, error } = useGetCheckups(filters);
if (isLoading) return <LoadingFallback />;
if (error) return <ErrorFallback />;
```

## 影響範囲

| ファイル | 問題箇所 | 状態 |
|---------|---------|------|
| `features/vaccinations/routes/VaccinationList.tsx` | フックが isLoading/isError を未 expose | 未修正 |
| `features/vaccinations/hooks/use-filter-vaccinations.ts` | isLoading/isError を返していない | 未修正 |
| `features/inventory/routes/InventoryList.tsx` | フックが isLoading/isError を未 expose | 未修正 |
| `features/inventory/hooks/use-inventory.ts` | isLoading/isError を返していない | 未修正 |
| `features/hospitalization/routes/HospitalizationList.tsx` | isLoading 未チェック | 未修正 |

## 修正方針

### 1. フック側で isLoading/isError を expose

**`features/vaccinations/hooks/use-filter-vaccinations.ts`**
```ts
// After: isLoading / isError を返す
const { data, isLoading, isError } = useGetVaccinations(clinicId);

return {
  data: filteredRecords,
  allVaccinations,
  isLoading,
  isError,
};
```

**`features/inventory/hooks/use-inventory.ts`**（同様のパターン）
```ts
const { data, isLoading, isError } = useGetInventoryItems(clinicId);

return {
  data: filteredItems,
  summary,
  isLoading,
  isError,
};
```

### 2. 各ページコンポーネントに早期リターンを追加

```tsx
import { LoadingFallback } from "@/components/shared/DataStates/LoadingFallback";
import { ErrorFallback } from "@/components/shared/DataStates/ErrorFallback";

// VaccinationList.tsx / InventoryList.tsx
const { data: filteredRecords, isLoading, isError } = useFilterXxx(...);

if (isLoading) return <LoadingFallback />;
if (isError) return <ErrorFallback />;
```

**`features/hospitalization/routes/HospitalizationList.tsx:105`**
```tsx
const { data: allHospitalizations = [], isLoading, isError } = useGetHospitalizations(dateFilters);

if (isLoading) return <LoadingFallback />;
if (isError) return <ErrorFallback />;
```

## 準拠すべきプロジェクト規約

### `.claude/CLAUDE.md` — Frontend ベストプラクティス
`features/checkups/routes/CheckupsList.tsx` が参照実装として `isLoading`/`error` を適切に処理している。

## 優先度
**Medium** — データなし状態と取得中状態の区別がつかず、UX 上の混乱を招く。

## 関連チケット
- FE-247: 受付カンバンの初期ローディングスケルトン欠如（同種）
- BUG-163: MedicalRecordForm の null リターン（同種）
- BUG-164: ShiftCalendarPage のローディング状態欠如（同種）

## 関連ファイル
- `frontend/src/features/vaccinations/routes/VaccinationList.tsx:84,252`
- `frontend/src/features/vaccinations/hooks/use-filter-vaccinations.ts`
- `frontend/src/features/inventory/routes/InventoryList.tsx:111,294`
- `frontend/src/features/inventory/hooks/use-inventory.ts`
- `frontend/src/features/hospitalization/routes/HospitalizationList.tsx:105`
- `frontend/src/components/shared/DataStates/LoadingFallback.tsx`
- `frontend/src/components/shared/DataStates/ErrorFallback.tsx`
