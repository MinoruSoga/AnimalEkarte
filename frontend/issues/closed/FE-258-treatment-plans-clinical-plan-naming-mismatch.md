# FE-258: treatment-plans.ts のファイル名・hook 名と実エンドポイントの乖離

**Status**: Open  
**Priority**: Medium  
**Type**: Refactor  
**Date Created**: 2026-04-19  
**Background**: FE-001 で URL を `/treatment-plans` → `/clinical-plan` に修正したが、ファイル名・hook 名が追随していない残留問題

## 現状の不統一

| 要素 | 現在の語 | あるべき語 |
|------|---------|----------|
| ファイル名 | `treatment-plans.ts` | `clinical-plan.ts` |
| hook 名 | `useUpdateTreatmentPlan` | `useUpdateClinicalPlan` |
| 型名 | `UpdateClinicalPlanRequest` | ✓（既に clinical-plan） |
| エンドポイント | `/v1/medical-records/{id}/clinical-plan` | ✓（既に clinical-plan） |

```typescript
// frontend/src/features/medical-records/api/treatment-plans.ts（現状）
export interface UpdateClinicalPlanRequest { ... }   // 型名は clinical-plan ✓

export const useUpdateTreatmentPlan = ...            // hook 名は treatment-plan ✗
  axios.patch(`.../clinical-plan`, input)            // エンドポイントは clinical-plan ✓
```

## 混在の経緯

FE-001 で URL・型名は `clinical-plan` に修正されたが、
ファイル名 `treatment-plans.ts` と hook 名 `useUpdateTreatmentPlan` が変更されなかった。

## 対応

```
frontend/src/features/medical-records/api/treatment-plans.ts
  → clinical-plan.ts にリネーム

useUpdateTreatmentPlan
  → useUpdateClinicalPlan にリネーム

参照箇所（hook を import している全ファイル）を更新:
  - hooks/use-medical-record-form.ts
```

## 完了条件

- [ ] ファイル名を `clinical-plan.ts` にリネーム
- [ ] `useUpdateTreatmentPlan` → `useUpdateClinicalPlan` にリネーム
- [ ] `use-medical-record-form.ts` の import・呼び出し箇所を更新
- [ ] lint / 型チェック / ビルドが通る
