# FE-043: medical-records ドメイン — models.ts 型移行（Request型導出化）

**親タスク**: [TASK-012](../../docs/tasks/open/TASK-012-models-ts-type-migration-by-domain.md)
**Status**: Open
**Priority**: Medium
**Date Created**: 2026-03-18

## Summary

Response型は models.ts を使用済み。CreateMedicalRecordRequest 等の Request 型が手書きのため、models.ts の `MedicalRecord`/`ClinicalPlan` から Omit/Partial で導出する。

## 現状

### api/types.ts
```typescript
// Response: ✅ models.ts import 済み（ApiMedicalRecord）
// Request: ❌ CreateMedicalRecordRequest 手書き（visit_date, ClinicalPlan フラット構造）
```

### features/medical-records/types/index.ts（10個手書き）
- `InterviewHistoryItem` — UI固有型
- `TreatmentItemType` — models.ts に `TreatmentItemType` 存在
- `Treatment` — models.ts に `Treatment` 存在
- `CreateTreatmentInput` / `UpdateTreatmentInput` — models.ts から導出すべき
- `BulkReorderTreatmentsInput` — API固有型
- `BillingReviewStatus` — models.ts に `BillingReviewStatus` 存在
- `BillingReview` — models.ts に `BillingReview` 存在
- `ReturnBillingReviewInput` — API固有型
- `Vital` — models.ts に `Vital` 存在
- `CreateVitalInput` / `UpdateVitalInput` — models.ts から導出すべき

## 必要な変更

1. `api/types.ts`: `CreateMedicalRecordRequest` を models.ts ベースで導出
2. `types/index.ts`: models.ts に対応がある型（Treatment, BillingReview, Vital, BillingReviewStatus, TreatmentItemType）を import に置換
3. Input/Request 型は models.ts から Omit/Partial で導出
4. UI固有型（InterviewHistoryItem）はそのまま残す

## 完了条件

- [ ] api/types.ts の Request 型が models.ts から導出されている
- [ ] types/index.ts の手書き型のうち models.ts 対応分が import に置換
- [ ] `pnpm build` 成功・型エラーなし
