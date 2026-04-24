# FE-178: カルテ内 VitalsTab・CheckupsTab・TreatmentsTab — usePermission 完全欠落（CRUD 全操作無防備）

## 概要

カルテ（MedicalRecordForm）内のタブコンポーネント `VitalsTab`・`CheckupsTab`・`TreatmentsTab` に `usePermission` が一切実装されていない。バイタル・検査・処置の追加・更新・削除が権限チェックなしで実行できる。

## 影響範囲

| ファイル | 問題操作 | API 呼び出し | 深刻度 |
|---------|---------|------------|--------|
| `VitalsTab.tsx` | バイタル追加（行 344: `useCreateVital`）・更新（行 360: `useUpdateVital`）・削除（行 379: `useDeleteVital`） | POST/PATCH/DELETE `/v1/medical-records/*/vitals` | HIGH |
| `CheckupsTab.tsx` | 検査追加（行 213: `useCreateCheckup`）・更新（行 229: `useUpdateCheckup`）・削除（行 248: `useDeleteCheckup`） | POST/PATCH/DELETE `/v1/medical-records/*/checkups` | HIGH |
| `TreatmentsTab.tsx` | 処置追加（行 120: `createMutation`）・更新（行 127: `updateMutation`）・削除（行 142: `deleteMutation`）・並び替え（行 157/173） | POST/PATCH/DELETE `/v1/medical-records/*/treatments` | HIGH |

## 根本原因

```tsx
// VitalsTab.tsx — usePermission なし ❌
const createMutation = useCreateVital(medicalRecordId);
const updateMutation = useUpdateVital(medicalRecordId);
const deleteMutation = useDeleteVital(medicalRecordId);

// 行 344: canCreate チェックなし → POST ❌
createMutation.mutate({ ... });

// 行 379: canDelete チェックなし → DELETE ❌
deleteMutation.mutate(vitalId);
```

```tsx
// TreatmentsTab.tsx — usePermission なし ❌
const createMutation = useCreateTreatment(medicalRecordId);
const updateMutation = useUpdateTreatment(medicalRecordId);
const deleteMutation = useDeleteTreatment(medicalRecordId);

// 行 120/127/142/157/173: 全 mutation に canCreate/canEdit/canDelete チェックなし ❌
```

親コンポーネント `MedicalRecordForm.tsx`（行 156）は `canEdit`/`canDelete` を取得しているが、これらのタブコンポーネントに権限情報を渡していない。

注: FE-160 では VitalsTab の編集ペンシルアイコン・CheckupsTab の編集ペンシルアイコン・TreatmentsTab の「明細を追加」ボタン（行 312）が未ガードとして報告済みだが、本チケットはそれらに加えて **usePermission 自体が存在しない** という根本問題を記載する。

## 修正方針

```tsx
// 各タブで usePermission を追加
const { canCreate, canEdit, canDelete } = usePermission("medical-records");

// 各 mutation ハンドラに権限チェック追加
const handleDelete = useCallback((id: string) => {
  if (!canDelete) return;
  deleteMutation.mutate(id);
}, [canDelete, deleteMutation]);

// 追加ボタンを canCreate でガード
{canCreate ? <Button onClick={handleAdd}>追加</Button> : null}

// 削除ボタンを canDelete でガード
<DeleteIconButton
  onClick={canDelete ? () => handleDelete(item.id) : undefined}
/>
```

## 優先度

**HIGH** — カルテ内バイタル・検査・処置は診療記録の中核データ。`canDelete=false` ユーザーが削除ボタンを押すと 403 が発生する。

## 関連ファイル

- `frontend/src/features/medical-records/components/VitalsTab/VitalsTab.tsx` (行 282-284: mutations, 行 344/360/379: mutate)
- `frontend/src/features/medical-records/components/CheckupsTab/CheckupsTab.tsx` (行 173-175: mutations, 行 213/229/248: mutate)
- `frontend/src/features/medical-records/components/TreatmentsTab/TreatmentsTab.tsx` (行 68-70: mutations, 行 120-173: mutate)
- `frontend/src/features/medical-records/routes/MedicalRecordForm.tsx` (行 156: usePermission, 権限情報を子タブに未渡し)
- 発見日: 2026-04-08（RBAC Phase 2/3 テスト中）
- 関連: FE-160（同タブ内削除ボタン未ガード部分報告済み）、FE-177（同カルテ内 Vaccination/DiagnosisPlan/Estimate）
