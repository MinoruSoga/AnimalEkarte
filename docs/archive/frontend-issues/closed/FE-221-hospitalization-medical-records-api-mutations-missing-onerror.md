# FE-221: hospitalization daily-records・medical-records API の useMutation に onError がない

## 概要

入院の日次記録・医療記録関連の `useMutation` フックで `onError` が未設定。
記録追加・更新・削除の失敗がサイレントに握り潰される。

## 影響ファイル

### `frontend/src/features/hospitalization/api/`

#### `create-hospitalization.ts`
| フック | 行 | 問題 |
|--------|---|------|
| `useCreateHospitalization` | 25-34 | onError なし |

#### `delete-hospitalization.ts`
| フック | 行 | 問題 |
|--------|---|------|
| `useDeleteHospitalization` | 8-17 | onError なし |

#### `daily-records.ts`
| フック | 行 | 問題 |
|--------|---|------|
| `useCreateDailyRecord` | 108-123 | onError なし |
| `useCreateDailyVital` | 125-140 | onError なし |
| `useCreateCareLog` | 142-157 | onError なし |
| `useCreateStaffNote` | 159-174 | onError なし |

### `frontend/src/features/medical-records/api/`

| ファイル | フック | 行 | 問題 |
|---------|--------|---|------|
| `create-medical-record.ts` | `useCreateMedicalRecord` | 17-26 | onError なし |
| `delete-medical-record.ts` | `useDeleteMedicalRecord` | 8-17 | onError なし |
| `treatment-plans.ts` | `useUpdateTreatmentPlan` | 13-22 | onError なし |
| `estimates.ts` | `useCreateEstimate` | 9-21 | onError なし |
| `save-estimate.ts` | `useCreateEstimateRecord` | 33-42 | onError なし |
| `save-estimate.ts` | `useUpdateEstimateRecord` | 44-53 | onError なし |
| `inquiries.ts` | `useUpdateInquiry` | 10-19 | onError なし |
| `billing-review.ts` | `useConfirmBillingReview` | 31-51 | onError なし |
| `billing-review.ts` | `useReturnBillingReview` | 54-78 | onError なし |

## 修正方針

```ts
// Before
export const useCreateDailyRecord = () => {
  return useMutation({
    mutationFn: createDailyRecord,
    onSuccess: () => { ... },
    // onError なし
  });
};

// After
export const useCreateDailyRecord = () => {
  return useMutation({
    mutationFn: createDailyRecord,
    onSuccess: () => { ... },
    onError: (error) => handleApiError(error, "日次記録の追加"),
  });
};
```

## 注意

同じ hospitalization feature でも以下は **正しく実装済み**（変更不要）:
- `care-plan-items.ts` — `onError: (error) => handleApiError(...)` 実装済み

medical-records の一部（clinical-plan.ts など）は既に正しい実装済みのため除外。

## 準拠すべきプロジェクト規約

### `.claude/rules/error-handling.md`
> すべての `catch` ブロックおよび `onError` コールバックで `handleApiError(error, "コンテキスト")` を呼び出す。

## 優先度
**High** — カルテ作成・入院記録追加・診察プラン更新の失敗がユーザーに通知されない。
`useConfirmBillingReview` / `useReturnBillingReview` は会計確認フローへの影響があるため特に重要。

## 関連ファイル
- `frontend/src/features/hospitalization/api/create-hospitalization.ts`
- `frontend/src/features/hospitalization/api/delete-hospitalization.ts`
- `frontend/src/features/hospitalization/api/daily-records.ts`
- `frontend/src/features/medical-records/api/create-medical-record.ts`
- `frontend/src/features/medical-records/api/delete-medical-record.ts`
- `frontend/src/features/medical-records/api/treatment-plans.ts`
- `frontend/src/features/medical-records/api/estimates.ts`
- `frontend/src/features/medical-records/api/save-estimate.ts`
- `frontend/src/features/medical-records/api/inquiries.ts`
- `frontend/src/features/medical-records/api/billing-review.ts`
