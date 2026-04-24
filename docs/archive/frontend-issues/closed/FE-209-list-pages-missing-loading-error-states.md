# FE-209: 複数の一覧ページでローディング・エラー状態が未実装

## 概要

複数の一覧ページが API からのデータ取得時にローディング・エラー・空状態を適切に処理していない。
API エラー発生時にユーザーには何も表示されず（空のリストが表示される）、
UX とデバッグ性を著しく損なっている。

## 影響ファイル

| ファイルパス | 欠けている状態 | 重要度 |
|------------|--------------|--------|
| `hospitalization/routes/HospitalizationList.tsx` | Loading + Error + Empty | 高 |
| `vaccinations/routes/VaccinationList.tsx` | Loading + Error + Empty | 高 |
| `examinations/routes/ExaminationsList.tsx` | Error + Empty | 中 |
| `inventory/routes/InventoryList.tsx` | Loading + Error | 中 |

## 現状コード

### HospitalizationList.tsx
```tsx
// isLoading, isError が useGetHospitalizations から取得されていない
const { data: allHospitalizations = [] } = useGetHospitalizations(dateFilters);
// → API エラー時に空の入院リストが表示される
```

### VaccinationList.tsx
```tsx
// isLoading, isError が useFilterVaccinations から取得されていない
const { data: filteredRecords, allVaccinations } = useFilterVaccinations(
  deferredSearchTerm, filters, activeFilters
);
// → エラー時に空のリストが表示される
```

## 比較: 正しい実装（CheckupsList.tsx — 参照実装）

```tsx
const { data: checkups = [], isLoading, error } = useGetCheckups(filters);

if (isLoading) return <LoadingFallback />;
if (error) return <ErrorFallback />;
if (checkups.length === 0) {
  return <EmptyStateFallback message="定期健診記録がありません" />;
}
```

## 修正方針

各ファイルで以下のパターンを適用する。
共通コンポーネント `LoadingFallback`、`ErrorFallback`、`EmptyStateFallback` は
`@/components/shared/DataStates` に既に存在する。

### HospitalizationList.tsx
```tsx
// Before
const { data: allHospitalizations = [] } = useGetHospitalizations(dateFilters);

// After
const { data: allHospitalizations = [], isLoading, isError } = useGetHospitalizations(dateFilters);

if (isLoading) return <LoadingFallback />;
if (isError) return <ErrorFallback />;
```

### VaccinationList.tsx
```tsx
// After
const { data: filteredRecords, allVaccinations, isLoading, isError } = useFilterVaccinations(
  deferredSearchTerm, filters, activeFilters
);

if (isLoading) return <LoadingFallback />;
if (isError) return <ErrorFallback />;
```

### ExaminationsList.tsx
```tsx
// isError チェックを追加
if (isError) return <ErrorFallback />;
```

### InventoryList.tsx
```tsx
// isLoading, isError を追加
const { data: filteredItems, summary, isLoading, isError } = useInventory({...});

if (isLoading) return <LoadingFallback />;
if (isError) return <ErrorFallback />;
```

## 準拠すべきプロジェクト規約

### `.claude/rules/error-handling.md` — Frontend エラーハンドリング
> API エラー時は必ずユーザーに分かるフィードバックを表示すること。
> 空の状態と「データなし」を区別して表示すること。

### プロジェクト内参照実装
- `frontend/src/features/checkups/routes/CheckupsList.tsx` — Loading/Error/Empty すべて実装済みの参照実装

## 優先度
**High** — API エラー発生時にユーザーに何も表示されない。
ネットワーク障害・権限エラー時に空白画面になり、ユーザーが原因を特定できない。

## 関連ファイル
- `frontend/src/features/hospitalization/routes/HospitalizationList.tsx`
- `frontend/src/features/vaccinations/routes/VaccinationList.tsx`
- `frontend/src/features/examinations/routes/ExaminationsList.tsx`
- `frontend/src/features/inventory/routes/InventoryList.tsx`
- `frontend/src/components/shared/DataStates/` — 参照コンポーネント
