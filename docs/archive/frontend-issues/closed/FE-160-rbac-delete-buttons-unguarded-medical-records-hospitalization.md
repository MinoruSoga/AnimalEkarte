# FE-160: MedicalRecord・Hospitalization — 削除ボタンが canDelete=false でも表示される（システム的欠如）

## 概要

`canDelete=false` のユーザーに対して、以下のコンポーネントで削除アイコンボタンが非表示にならない。これらのコンポーネントは `usePermission` を呼び出しておらず、親コンポーネントから `canDelete` も受け取っていない。

## 影響範囲

| ファイル | 問題 UI | API 呼び出し | 深刻度 |
|---------|---------|------------|--------|
| `MedicalRecord: CheckupsTab.tsx` | `<DeleteIconButton>` (行 317) + 編集鉛筆ボタン (行 310) | DELETE/PATCH `/v1/checkups/:id` | HIGH |
| `MedicalRecord: VitalsTab.tsx` | `<DeleteIconButton>` (行 495) + 編集鉛筆ボタン (行 488) | DELETE/PATCH `/v1/vitals/:id` | HIGH |
| `MedicalRecord: TreatmentRow.tsx` | `<DeleteIconButton>` (行 382) | DELETE `/v1/treatments/:id` | HIGH |
| `MedicalRecord: TreatmentsTab.tsx` | `onDelete` 連携 + 「明細を追加」ボタン (行 312) canCreate チェックなし | DELETE/POST `/v1/treatments` | HIGH |
| `MedicalRecord: TreatmentTable.tsx` | `onDelete` 連携 | DELETE `/v1/treatments/:id` | HIGH |
| `Hospitalization: CarePlanTab.tsx` | `<DeleteIconButton>` (行 207) | DELETE `/v1/care-plans/:id` | HIGH |
| `Hospitalization: HospitalizationTreatmentTable.tsx` | `<DeleteIconButton>` (usePermission なし) | DELETE `/v1/treatments/:id`（入院版） | HIGH |

注: `MedicalRecordImage.tsx` の画像削除は FE-156 で報告済み（canEdit ガードと共通）

## 現状の挙動（バグ）

```tsx
// CheckupsTab.tsx — canDelete チェックなし ❌
<DeleteIconButton
  onClick={() => handleDeleteCheckup(checkup.id)}
  // canDelete のガードなし
/>

// VitalsTab.tsx — canDelete チェックなし ❌
<DeleteIconButton
  onClick={() => handleDeleteVital(vital.id)}
/>

// CarePlanTab.tsx — ItemRow コンポーネント内、canDelete チェックなし ❌
function ItemRow({ item, onEdit, onDelete, isDeleting }: ItemRowProps) {
  return (
    <DeleteIconButton
      onClick={() => onDelete(item.id)}
      // canDelete がない
    />
  );
}
```

`canDelete=false` のユーザーがカルテ・入院詳細を開くと：
1. 健診記録の行に削除ボタンが表示され、クリックすると DELETE → 403
2. バイタル記録の行に削除ボタンが表示され、クリックすると DELETE → 403
3. 治療計画の行に削除ボタンが表示され、クリックすると DELETE → 403
4. ケアプランの行に削除ボタンが表示され（3 件確認）、クリックすると DELETE → 403

## 根本原因

これらのコンポーネントは `MedicalRecordForm` または `HospitalizationDetail` から権限情報を受け取っていない。コンポーネント設計時に RBAC を考慮していなかった。

- `CheckupsTab`, `VitalsTab`, `CarePlanTab` は自コンポーネント内で API フック（useDeleteXxx）を呼び出す独立型コンポーネント
- `usePermission("medical-records")` / `usePermission("hospitalization")` を自分で呼び出していない
- 親コンポーネントから `canDelete` prop を受け取る設計になっていない

**特記**: `HospitalizationDetail.tsx` 自体も `usePermission` を呼び出しておらず、`DailyRecordsTab`・`CarePlanTab` に権限情報を props で渡していない（行 54-72）。子コンポーネントが props 受け取り方式を採用する場合、親の修正も必須。

## 修正方針

### 方針 A: 各コンポーネントで usePermission を直接呼び出す（推奨）

```tsx
// CheckupsTab.tsx
export function CheckupsTab({ medicalRecordId }: CheckupsTabProps) {
  const { canDelete } = usePermission("medical-records");  // ← 追加
  // ...
  return (
    <DeleteIconButton
      onClick={() => handleDeleteCheckup(checkup.id)}
      style={{ display: canDelete ? undefined : 'none' }}
    />
    // または
    // {canDelete ? <DeleteIconButton onClick={...} /> : null}
  );
}
```

同様に `VitalsTab`, `TreatmentRow`, `CarePlanTab` の `ItemRow` も修正。

### 方針 B: 親コンポーネントから canDelete を props で注入

```tsx
// MedicalRecordForm.tsx
const { canEdit, canDelete } = usePermission("medical-records");
<CheckupsTab medicalRecordId={recordId} canDelete={canDelete} />
<VitalsTab medicalRecordId={recordId} canDelete={canDelete} />
```

方針 A が推奨（コンポーネントが自律的に権限を管理できるため）。

## 優先度

**HIGH** — `canDelete=false` のユーザーがカルテ・入院詳細を開くたびに、削除ボタンが表示され、クリック時に DELETE → 403 エラーが発生する。実際のデータ削除は 403 で防がれるが、UI が誤った操作可能感を与える。

## 関連ファイル

- `frontend/src/features/medical-records/components/CheckupsTab/CheckupsTab.tsx` (行 317)
- `frontend/src/features/medical-records/components/VitalsTab/VitalsTab.tsx` (行 495)
- `frontend/src/features/medical-records/components/TreatmentsTab/TreatmentRow.tsx`
- `frontend/src/features/medical-records/components/TreatmentsTab/TreatmentsTab.tsx`
- `frontend/src/features/medical-records/components/TreatmentTable.tsx`
- `frontend/src/features/hospitalization/components/CarePlanTab/CarePlanTab.tsx` (行 207)
- `frontend/src/features/hospitalization/components/HospitalizationTreatmentTable.tsx`
- 発見日: 2026-04-07（RBAC Phase 3 テスト中）
