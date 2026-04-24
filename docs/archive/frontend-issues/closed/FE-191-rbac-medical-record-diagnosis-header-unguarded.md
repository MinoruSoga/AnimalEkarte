# FE-191: カルテ診察/治療プランタブ — DiagnosisHeader 系コンポーネントに RBAC なし

## 概要

`/medical-records/:id` の **診察/治療プランタブ** で表示される `DiagnosisHeader`
およびその子コンポーネント（`DiagnosisHeaderChiefComplaint`、`DiagnosisHeaderDiagnosis`、
`DiagnosisHeaderPhysicalExam`）に `canEdit` ガードが一切なく、
`canEdit=false` ユーザーが診断情報を自由に編集できる。

FE-156 の修正対象は `TreatmentTable`（行追加/削除）のみで、
`DiagnosisHeader` は対象外だった。

## 再現手順

1. `canEdit=false` ユーザーでログイン
2. `/medical-records/22` 等カルテ詳細へ移動
3. 「診察/治療プラン」タブを開く

## 確認された問題

### 操作可能なフィールド（全て disabled なし）

| フィールド | コンポーネント |
|-----------|--------------|
| 身体検査所見 textarea | `DiagnosisHeaderPhysicalExam` |
| 診断カテゴリ select（診断1・診断2） | `DiagnosisHeaderDiagnosis` |
| 診断病名 select（診断1・診断2） | `DiagnosisHeaderDiagnosis` |
| 診断詳細 textarea | `DiagnosisHeaderDiagnosis` |
| 治療方針 textarea | `DiagnosisHeaderChiefComplaint` |
| 引用ボタン × 2 | `DiagnosisHeaderChiefComplaint` |

## 根本原因

```tsx
// MedicalRecordDiagnosisPlan.tsx — canEdit を DiagnosisHeader に渡していない
<DiagnosisHeader
  chiefComplaint={chiefComplaint}
  policy={plan}
  setPolicy={setPlan}
  diagnosisDetails={assessment}
  setDiagnosisDetails={setAssessment}
  diagnosis1CategoryId={diagnosis1CategoryId}
  setDiagnosis1CategoryId={setDiagnosis1CategoryId}
  // ...
  // canEdit={canEdit}  ← 渡されていない ❌
/>
```

```tsx
// DiagnosisHeader.tsx — canEdit/usePermission なし
// disabled, readOnly の実装ゼロ
```

## 修正方針

`MedicalRecordDiagnosisPlan` 内で既に `const { canCreate, canEdit, canDelete } = usePermission(...)` を
呼んでいるため、同パターンで各子コンポーネントに canEdit を渡す。

または `DiagnosisHeader` 内で `usePermission` を内部呼び出しする（MedicalRecordBillCheck 方式）。

```tsx
// DiagnosisHeader.tsx
const { canEdit } = usePermission("medical-records");

// 各 textarea / select に disabled を追加
<Textarea disabled={!canEdit} ... />
<Select disabled={!canEdit} ... />

// 引用ボタンを条件付きレンダー
{canEdit ? <Button onClick={handleQuote}>引用</Button> : null}
```

## 優先度

**HIGH** — 閲覧のみユーザーが診断内容（診断病名・身体検査所見・治療方針）を変更できる。
保存ボタンは canEdit ガードで非表示だが、Enter キー送信のリスクあり。

## 関連ファイル

- `frontend/src/features/medical-records/components/DiagnosisHeader.tsx`
- `frontend/src/features/medical-records/components/DiagnosisHeaderChiefComplaint.tsx`
- `frontend/src/features/medical-records/components/DiagnosisHeaderDiagnosis.tsx`
- `frontend/src/features/medical-records/components/DiagnosisHeaderPhysicalExam.tsx`
- `frontend/src/features/medical-records/components/MedicalRecordDiagnosisPlan.tsx`
- 発見日: 2026-04-08
- 関連: FE-156（カルテタブ RBAC 修正 — TreatmentTable のみ対応、DiagnosisHeader は未対応）
