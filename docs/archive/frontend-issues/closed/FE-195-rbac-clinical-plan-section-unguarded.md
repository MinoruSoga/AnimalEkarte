# FE-195: カルテ診察/治療プランタブ — ClinicalPlanSection に RBAC なし

## 概要

`/medical-records/:id` の **診察/治療プランタブ** 下部セクション
（`ClinicalPlanSection`：「診察所見・診断・治療方針」）に
`canEdit` ガードが一切なく、`canEdit=false` ユーザーが以下を操作できる。

FE-191 で `DiagnosisHeader` 系コンポーネントは修正済みだが、
同タブ内の `ClinicalPlanSection` が対象外だった。

## 確認された問題

### 操作可能なフィールド

| フィールド | コンポーネント | 状態 |
|-----------|--------------|------|
| 身体検査所見 textarea | `ClinicalPlanSection` | `disabled:false` ❌ |
| 診断カテゴリ `<input type="text">` | `ClinicalPlanSection` | `disabled:false` ❌ |
| 診断病名 `<input type="text">` | `ClinicalPlanSection` | `disabled:false` ❌ |
| 診断詳細 textarea | `ClinicalPlanSection` | `disabled:false` ❌ |
| 治療方針 textarea | `ClinicalPlanSection` | `disabled:false` ❌ |

### Enter キー送信リスク

`<input type="text">` フィールド（診断カテゴリ・病名）は
main form `<form action={formAction}>` の内部に存在するため、
**Enter キー押下 → formAction 実行 → `updateTreatmentPlanMutation.mutateAsync` → API 403** が発生する。

```
// use-medical-record-form.ts — canEdit チェックなし
case "診察/治療プラン": {
  await updateTreatmentPlanMutation.mutateAsync(treatmentPlanPayload);  // canEdit チェックなし ❌
  break;
}
```

また `formState.success = true` になった場合（他タブから操作した場合など）、
`clinicalPlanSaveRef.current()` → `ClinicalPlanSection.handleSave()` →
`updateMutation.mutateAsync` も権限チェックなしで実行される。

## 根本原因

```tsx
// ClinicalPlanSection.tsx — usePermission/canEdit が一切存在しない
export function ClinicalPlanSection({ medicalRecordId, onRegisterSave }: ClinicalPlanSectionProps) {
  const updateMutation = useUpdateClinicalPlan(medicalRecordId);

  const handleSave = useCallback(async () => {
    // canEdit チェックなし ❌
    await updateMutation.mutateAsync(input);
  }, [...]);
```

```tsx
// use-medical-record-form.ts — formAction に canEdit チェックなし
case "診察/治療プラン": {
  await updateTreatmentPlanMutation.mutateAsync(payload);  // ❌
}
```

## セキュリティ影響

- UI から直接データ変更試行が可能（API は 403 を返す）
- Enter キーによる formAction 実行 → API 403
- 視覚的に canEdit=false ユーザーが編集者と同じ UI を見る（UX問題）
- 保存ボタンは `{canSubmit ? <SubmitButton> : null}` で非表示のため、
  通常クリック保存はできない

## 修正方針

### Option A: ClinicalPlanSection に canEdit prop を追加

```tsx
// ClinicalPlanSection.tsx
interface ClinicalPlanSectionProps {
  medicalRecordId: string;
  onRegisterSave?: (fn: () => Promise<void>) => void;
  canEdit?: boolean;   // 追加
}

// フィールドに disabled={!canEdit} を追加
<CharCountTextarea
  value={physicalExam}
  onChange={setPhysicalExam}
  disabled={!canEdit}    // ← 追加
  ...
/>

// handleSave に canEdit ガードを追加
const handleSave = useCallback(async () => {
  if (!canEdit) return;  // ← 追加
  await updateMutation.mutateAsync(input);
}, [canEdit, ...]);
```

### Option B: formAction に canEdit チェックを追加

```tsx
// use-medical-record-form.ts
case "診察/治療プラン": {
  if (!canEdit) break;   // ← 追加（Enter キー対策）
  await updateTreatmentPlanMutation.mutateAsync(payload);
}
```

**推奨: Option A + Option B の組み合わせ**

## 優先度

**HIGH** — 閲覧のみユーザーが診察所見・診断情報を操作でき、
Enter キーで API 呼び出し（→ 403）が発生する。
FE-191 の修正漏れ（DiagnosisHeader は修正済み、ClinicalPlanSection は未対応）。

## 関連ファイル

- `frontend/src/features/medical-records/components/ClinicalPlanSection/ClinicalPlanSection.tsx`
- `frontend/src/features/medical-records/hooks/use-medical-record-form.ts`
- `frontend/src/features/medical-records/routes/MedicalRecordForm.tsx`
- 発見日: 2026-04-08（RBAC Phase 3 テスト中）
- 関連: FE-191（DiagnosisHeader 系修正済み、本コンポーネントは対象外）
