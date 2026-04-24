# FE-179: カルテ内 MedicalRecordBillCheck・BillingReviewSection — usePermission 完全欠落（会計確認・返戻操作無防備）

## 概要

カルテ（MedicalRecordForm）内の会計確認セクション `MedicalRecordBillCheck` および `BillingReviewSection` に `usePermission` が一切実装されていない。会計確認・返戻・処置明細の更新・削除が権限チェックなしで実行できる。

## 影響範囲

| ファイル | 問題操作 | API 呼び出し | 深刻度 |
|---------|---------|------------|--------|
| `MedicalRecordBillCheck.tsx` | 会計確認（行 47: `confirmMutation.mutateAsync`）・返戻（行 59: `returnMutation.mutate`）・処置更新（行 95: `updateTreatmentMutation.mutate`）・処置削除（行 101: `deleteMutation.mutate`） | POST `/v1/billing-reviews/confirm`, POST `/v1/billing-reviews/return`, PATCH/DELETE `/v1/treatments` | HIGH |
| `BillingReviewSection.tsx` | 会計レビュー mutations（行 54/62） | POST/PATCH `/v1/billing-reviews/*` | HIGH |

## 根本原因

```tsx
// MedicalRecordBillCheck.tsx — usePermission なし ❌
const confirmMutation = useConfirmBillingReview(...);
const returnMutation = useReturnBillingReview(...);
const updateTreatmentMutation = useUpdateTreatment(...);
const deleteMutation = useDeleteTreatment(...);

// 行 47: canEdit チェックなし → 会計確認 ❌
confirmMutation.mutateAsync();

// 行 59: canEdit チェックなし → 返戻 ❌
returnMutation.mutate();

// 行 95: canEdit チェックなし → 処置更新 ❌
updateTreatmentMutation.mutate({ ... });

// 行 101: canDelete チェックなし → 処置削除 ❌
deleteMutation.mutate(id);

// 行 186-207: 確認/返戻ボタンに canEdit ガードなし ❌
<Button onClick={handleConfirm}>会計確認</Button>
<Button onClick={handleReturn}>返戻</Button>
```

```tsx
// BillingReviewSection.tsx — usePermission なし ❌
// 行 54/62: mutations に canEdit チェックなし ❌
```

## 修正方針

```tsx
// MedicalRecordBillCheck.tsx
const { canEdit, canDelete } = usePermission("accounting");  // または "medical-records"

const handleConfirm = useCallback(async () => {
  if (!canEdit) return;
  await confirmMutation.mutateAsync();
}, [canEdit, confirmMutation]);

// 確認・返戻ボタンを canEdit でガード
{canEdit ? (
  <Button onClick={handleConfirm}>会計確認</Button>
) : null}
{canEdit ? (
  <Button onClick={handleReturn}>返戻</Button>
) : null}

// 処置削除を canDelete でガード
<DeleteIconButton
  onClick={canDelete ? () => deleteMutation.mutate(id) : undefined}
/>
```

## 優先度

**HIGH** — 会計確認・返戻は診療・会計フローに直結する重要操作。`canEdit=false` ユーザーが誤って確認ボタンを押すと 403 または不正な会計確定が発生しうる。

## 関連ファイル

- `frontend/src/features/medical-records/components/MedicalRecordBillCheck.tsx` (行 36-39: mutations, 行 44-101: handlers, 行 186-207: 確認/返戻ボタン)
- `frontend/src/features/medical-records/components/BillingReviewSection/BillingReviewSection.tsx` (行 54/62: mutations)
- `frontend/src/features/medical-records/routes/MedicalRecordForm.tsx` (行 156: usePermission, 子コンポーネントに未渡し)
- 発見日: 2026-04-08（RBAC Phase 2/3 テスト中）
- 関連: FE-177（カルテ内 Vaccination/DiagnosisPlan/Estimate）、FE-178（VitalsTab/CheckupsTab/TreatmentsTab）
