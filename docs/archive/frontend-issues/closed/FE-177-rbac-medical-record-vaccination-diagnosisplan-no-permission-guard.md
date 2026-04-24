# FE-177: カルテ内ワクチン記録・治療プラン — usePermission 完全欠落（MedicalRecordVaccination・MedicalRecordDiagnosisPlan）

## 概要

カルテ（MedicalRecordForm）のサブコンポーネントである `MedicalRecordVaccination` と `MedicalRecordDiagnosisPlan` に `usePermission` が実装されていない。ワクチン記録の保存（POST）と治療プランの追加・更新・削除（POST/PATCH/DELETE）が権限チェックなしで実行できる。

## 影響範囲

| ファイル | 問題操作 | API 呼び出し | 深刻度 |
|---------|---------|------------|--------|
| `MedicalRecordVaccination.tsx` | ワクチン記録保存（行 119: `onSave={handleSave}`） | POST `/v1/vaccinations` | HIGH |
| `MedicalRecordDiagnosisPlan.tsx` | 治療プラン追加（行 104: `handleAddRow`）・更新（行 99: `handleUpdateItem`）・削除（行 84: `handleRemoveItem`） | POST/PATCH/DELETE `/v1/treatments` | HIGH |
| `MedicalRecordEstimate.tsx` | 見積作成（行 121: `createEstimate.mutateAsync`）・更新（行 119: `updateEstimate.mutateAsync`） | POST/PATCH `/v1/estimates` | HIGH |
| `ClinicalPlanSection.tsx` | 臨床プラン更新（行 51: `updateMutation.mutateAsync`） | PATCH `/v1/medical-records/*/clinical-plan` | MEDIUM |

## 根本原因

```tsx
// MedicalRecordVaccination.tsx — usePermission なし ❌
export const MedicalRecordVaccination = memo(function MedicalRecordVaccination({ petId }) {
  // usePermission 呼び出しなし

  const handleSave = useCallback(async () => {
    // canCreate チェックなし → POST /v1/vaccinations ❌
    await createVaccinationMutation.mutateAsync({ pet_id, vaccine_id, date, ... });
  }, [...]);

  return (
    <VaccinationForm
      onSave={handleSave}  // canCreate チェックなしで常に保存可能 ❌
    />
  );
});
```

```tsx
// MedicalRecordDiagnosisPlan.tsx — usePermission なし ❌
const deleteMutation = useDeleteTreatment(medicalRecordId ?? "");

const handleRemoveItem = useCallback((id) => {
  // canDelete チェックなし → DELETE /v1/treatments/:id ❌
  deleteMutation.mutate(String(id));
}, [deleteMutation]);

const handleAddRow = useCallback(() => {
  // canCreate チェックなし → POST /v1/treatments ❌
  createMutation.mutate({ ... });
}, [...]);

<TreatmentTable
  onRemove={handleRemoveItem}  // canDelete なし ❌
  onAddRow={handleAddRow}      // canCreate なし ❌
  onUpdate={handleUpdateItem}  // canEdit なし ❌
/>
```

親コンポーネント `MedicalRecordForm.tsx` は `canEdit`/`canDelete` を取得しているが（行 156）、これらの子コンポーネントに権限情報を渡していない（行 364-367: `<MedicalRecordTreatment medicalRecordId ...>` のみ）。

## 修正方針

### MedicalRecordVaccination.tsx

```tsx
const { canCreate } = usePermission("vaccinations");  // または "medical-records"

const handleSave = useCallback(async () => {
  if (!canCreate) return;  // canCreate チェック追加
  // ... 保存処理
}, [canCreate, ...]);

<VaccinationForm
  onSave={canCreate ? handleSave : undefined}  // canCreate でガード
/>
```

### MedicalRecordDiagnosisPlan.tsx

```tsx
const { canCreate, canEdit, canDelete } = usePermission("medical-records");

const handleRemoveItem = useCallback((id) => {
  if (!canDelete) return;  // canDelete チェック
  deleteMutation.mutate(String(id));
}, [canDelete, deleteMutation]);

<TreatmentTable
  onRemove={canDelete ? handleRemoveItem : undefined}
  onAddRow={canCreate ? handleAddRow : undefined}
  onUpdate={canEdit ? handleUpdateItem : undefined}
/>
```

## 優先度

**HIGH** — カルテ内でのワクチン接種記録・治療プランは診療・会計に直結する重要データ。`canCreate=false` / `canDelete=false` ユーザーがデータを追加・削除しようとすると 403 エラーが発生する。

## 関連ファイル

- `frontend/src/features/medical-records/components/MedicalRecordVaccination.tsx` (行 44-91: handleSave, 行 119: onSave)
- `frontend/src/features/medical-records/components/MedicalRecordDiagnosisPlan.tsx` (行 63-65: mutations, 行 83-100: handlers, 行 178-185: TreatmentTable)
- `frontend/src/features/medical-records/components/MedicalRecordEstimate.tsx` (行 37-41: mutations, 行 119-121: 保存処理)
- `frontend/src/features/medical-records/routes/MedicalRecordForm.tsx` (行 156: usePermission, 権限情報を子に未渡し)
- 発見日: 2026-04-08（RBAC Phase 2/3 テスト中）
- 関連: FE-160（TreatmentTable/TreatmentsTab 削除ボタン未ガード）
