# FE-219: useMutation の onError 欠落（examinations / vaccinations / inventory / trimming / pets / hospital-settings）

## 概要

複数 feature の `useMutation` フックに `onError` コールバックが設定されていない。
API エラー発生時にユーザーへの通知が行われず、データ保存・削除の失敗がサイレントに握り潰される。

## 影響範囲

### `frontend/src/features/examinations/api/`

| ファイル | フック | 問題 |
|---------|--------|------|
| `create-examination.ts:17-26` | `useCreateExamination` | onError なし |
| `update-examination.ts:18-33` | `useUpdateExamination` | onError なし |

### `frontend/src/features/vaccinations/api/`

| ファイル | フック | 問題 |
|---------|--------|------|
| `create-vaccination.ts:17-26` | `useCreateVaccination` | onError なし |
| `update-vaccination.ts:18-33` | `useUpdateVaccination` | onError なし |
| `delete-vaccination.ts:8-17` | `useDeleteVaccination` | onError なし |

### `frontend/src/features/inventory/api/inventory.ts`

| フック | 行 | 問題 |
|--------|---|------|
| `useCreateInventoryItem` | 47-55 | onError なし |
| `useUpdateInventoryItem` | 62-70 | onError なし |
| `useDeleteInventoryItem` | 76-84 | onError なし |

### `frontend/src/features/trimming/api/`

| ファイル | フック | 問題 |
|---------|--------|------|
| `delete-trimming.ts:8-17` | `useDeleteTrimming` | onError なし |

### `frontend/src/features/pets/api/`

| ファイル | フック | 問題 |
|---------|--------|------|
| `create-pet.ts:13-22` | `useCreatePet` | onError なし |
| `update-pet.ts:16-27` | `useUpdatePet` | onError なし |
| `delete-pet.ts:8-17` | `useDeletePet` | onError なし |

### `frontend/src/features/hospital-settings/api/clinics.ts`

| フック | 行 | 問題 |
|--------|---|------|
| `useCreateClinic` | 88-96 | onError なし |
| `useUpdateClinic` | 98-107 | onError なし |
| `useDeleteClinic` | 109-117 | onError なし |

## 追加: ClinicMasterSettings の isLoading/isError 未処理

`frontend/src/features/hospital-settings/routes/ClinicMasterSettings.tsx:152`
- `useGetClinics()` を呼んでいるが `isLoading`/`isError` を UI にフィードバックしていない
- API エラー時に空の画面が表示される

## 修正方針

```ts
// Before
export const useCreateExamination = () => {
  return useMutation({
    mutationFn: createExamination,
    onSuccess: () => { ... },
    // onError なし
  });
};

// After
export const useCreateExamination = () => {
  return useMutation({
    mutationFn: createExamination,
    onSuccess: () => { ... },
    onError: (error) => handleApiError(error, "診察記録の作成"),
  });
};
```

各フックに同様のパターンで `onError` を追加する。

## 参照実装

- `frontend/src/features/accounting/api/create-accounting.ts` — `onError` 正しく実装済み
- `frontend/src/features/vaccinations/api/create-vaccination.ts` の `useCreateVaccination` が参照として機能すれば理想だが、上記の通り現在も欠落している

## 準拠すべきプロジェクト規約

### `.claude/rules/error-handling.md`
> すべての `catch` ブロックおよび `onError` コールバックで `handleApiError(error, "コンテキスト")` を呼び出す。

## 優先度
**High** — ワクチン接種・診察記録・在庫・ペット情報の操作失敗がユーザーに通知されない。

## 関連ファイル
- `frontend/src/features/examinations/api/create-examination.ts`
- `frontend/src/features/examinations/api/update-examination.ts`
- `frontend/src/features/vaccinations/api/create-vaccination.ts`
- `frontend/src/features/vaccinations/api/update-vaccination.ts`
- `frontend/src/features/vaccinations/api/delete-vaccination.ts`
- `frontend/src/features/inventory/api/inventory.ts`
- `frontend/src/features/trimming/api/delete-trimming.ts`
- `frontend/src/features/pets/api/create-pet.ts`
- `frontend/src/features/pets/api/update-pet.ts`
- `frontend/src/features/pets/api/delete-pet.ts`
- `frontend/src/features/hospital-settings/api/clinics.ts`
- `frontend/src/features/hospital-settings/routes/ClinicMasterSettings.tsx`
