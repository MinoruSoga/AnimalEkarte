# FE-001: カルテ保存の treatment-plans.ts URL・フィールド名を clinical-plan に修正

**Status**: Closed
**Closed At**: 2026-03-23

## クローズ情報

- **変更ファイル**:
  - `frontend/src/features/medical-records/api/treatment-plans.ts` — URL `/treatment-plans` → `/clinical-plan`、型名 `UpdateTreatmentPlanRequest` → `UpdateClinicalPlanRequest`、フィールド名修正
  - `frontend/src/features/medical-records/hooks/use-medical-record-form.ts` — `treatmentPlanReq` のフィールド名を BE 契約に合わせて修正

---

**Status**: Open
**Priority**: High
**Affects**: カルテ詳細画面の「保存」ボタン（診察プランタブ）
**Date Created**: 2026-03-23
**Related**: BUG-008, BE-053

## Summary

`treatment-plans.ts` が `PATCH /v1/medical-records/:id/treatment-plans` を呼んでいるが、
正しくは `PATCH /v1/medical-records/:id/clinical-plan` であり、かつリクエストフィールド名も不一致。
修正により `updateTreatmentPlanMutation` が 404 エラーを返す問題が解消される。

## 現状のコード

```typescript
// frontend/src/features/medical-records/api/treatment-plans.ts:1-22
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";

export interface UpdateTreatmentPlanRequest {
  plan?: string;                        // BE は "treatment_policy"
  assessment?: string;                  // BE は "diagnosis_details"
  diagnosis_1_category_id?: number | null;  // BE は "diagnosis_category_id"
  diagnosis_1_name_id?: number | null;      // BE は "diagnosis_name_id"
  diagnosis_2_category_id?: number | null;  // BE と一致（BE-053 で対応）
  diagnosis_2_name_id?: number | null;      // BE と一致（BE-053 で対応）
}

export const useUpdateTreatmentPlan = (recordId: string) => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: UpdateTreatmentPlanRequest) =>
      axios.patch(`/v1/medical-records/${recordId}/treatment-plans`, input),  // ← URL が間違い
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["medical-record", recordId] });
    },
  });
};
```

**バックエンド側の正しいエンドポイント**:
```
PATCH /v1/medical-records/:id/clinical-plan  （clinical_plan_handler.go:75）
```

**バックエンド側の正しいフィールド名**（`clinical_plan_request.go`）:
```go
type updateClinicalPlanRequest struct {
    PhysicalExam         *string `json:"physical_exam"`
    DiagnosisCategoryID  *uint64 `json:"diagnosis_category_id"`     // FE: diagnosis_1_category_id
    DiagnosisNameID      *uint64 `json:"diagnosis_name_id"`          // FE: diagnosis_1_name_id
    DiagnosisDetails     *string `json:"diagnosis_details"`          // FE: assessment
    TreatmentPolicy      *string `json:"treatment_policy"`           // FE: plan
    // Diagnosis2CategoryID / Diagnosis2NameID は BE-053 で追加
}
```

## 必要な変更

### 1. `frontend/src/features/medical-records/api/treatment-plans.ts` を修正

```typescript
// After:
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";

export interface UpdateClinicalPlanRequest {
  treatment_policy?: string;
  diagnosis_details?: string;
  diagnosis_category_id?: number | null;
  diagnosis_name_id?: number | null;
  diagnosis_2_category_id?: number | null;
  diagnosis_2_name_id?: number | null;
}

export const useUpdateTreatmentPlan = (recordId: string) => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: UpdateClinicalPlanRequest) =>
      axios.patch(`/v1/medical-records/${recordId}/clinical-plan`, input),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["medical-record", recordId] });
    },
  });
};
```

### 2. `frontend/src/features/medical-records/hooks/use-medical-record-form.ts` のリクエスト組み立てを修正

```typescript
// Before（use-medical-record-form.ts:153-160）:
const treatmentPlanReq = {
  plan: plan !== DEFAULT_PLAN ? plan : undefined,
  assessment: assessment !== DEFAULT_ASSESSMENT ? assessment : undefined,
  diagnosis_1_category_id: diagnosis1CategoryId,
  diagnosis_1_name_id: diagnosis1NameId,
  diagnosis_2_category_id: diagnosis2CategoryId,
  diagnosis_2_name_id: diagnosis2NameId,
};

// After:
const treatmentPlanReq: UpdateClinicalPlanRequest = {
  treatment_policy: plan !== DEFAULT_PLAN ? plan : undefined,
  diagnosis_details: assessment !== DEFAULT_ASSESSMENT ? assessment : undefined,
  diagnosis_category_id: diagnosis1CategoryId ?? undefined,
  diagnosis_name_id: diagnosis1NameId ?? undefined,
  diagnosis_2_category_id: diagnosis2CategoryId ?? undefined,
  diagnosis_2_name_id: diagnosis2NameId ?? undefined,
};
```

## UI 操作フロー（修正後）

1. カルテ詳細 `/medical-records/:id` を開く
2. 診察プランタブで「診断名」「治療方針」「評価・所見」を入力
3. 「保存」ボタンをクリック
4. `PATCH /v1/medical-records/:id/clinical-plan` が 200 を返す
5. 「カルテを更新しました」トーストが表示される（404 エラーが出なくなる）

## 依存関係

- BE-053 が完了している必要がある（`diagnosis_2_*` フィールド対応）
- BE-053 なしでも `treatment_policy`, `diagnosis_details`, `diagnosis_category_id`, `diagnosis_name_id` の修正は可能

## プロジェクトルール遵守チェック

- [x] `any` 型なし
- [x] `FC` / `forwardRef` なし（API ファイル）
- [x] barrel index 経由 import なし
- [x] `useTransition` で pending 管理済み（`useUpdateTreatmentPlan` は `useTransition` 内で呼ばれている）
- [x] 型は既存パターンと同様に手書き interface（API リクエスト型は models.ts に対応モデルなし）

## 完了条件

- [ ] `PATCH /v1/medical-records/:id/clinical-plan` が 200 を返す（404 が解消）
- [ ] カルテ詳細画面で「保存」後に「カルテを更新しました」トーストが表示される
- [ ] `docker compose exec frontend pnpm build` が通る（型エラーなし）
- [ ] `docker compose exec frontend pnpm lint` が通る（ESLint エラーなし）
