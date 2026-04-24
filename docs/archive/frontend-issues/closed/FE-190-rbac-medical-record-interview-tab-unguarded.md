# FE-190: カルテ問診タブ — MedicalRecordInterview に canEdit ガードなし

## 概要

`/medical-records/:id` の **問診タブ** (`MedicalRecordInterview`) および子コンポーネント
`InterviewChiefComplaint` に `canEdit` が一切渡されておらず、
`canEdit=false` ユーザーが以下の操作を実行できる状態になっている。

FE-156（カルテタブ群の RBAC 修正）は DiagnosisPlan・Vaccination・Image・Estimate・BillCheck を対象としており、**問診タブは対象外**だった。

## 再現手順

1. `canEdit=false` ユーザー（例: 一般 花子）でログイン
2. `/medical-records/22` などカルテ詳細ページに移動
3. 「問診」タブを確認

## 確認された問題

### 1. 「マスタ編集」ボタン × 2 が表示・操作可能 ❌

```tsx
// InterviewChiefComplaint.tsx — canEdit ガードなし
<button
  onClick={() => navigate(paths.settings.interview.chiefComplaint.getHref())}
>
  マスタ編集  {/* canEdit=false でも表示 ❌ */}
</button>

<button
  onClick={() => navigate(paths.settings.interview.interviewTemplate.getHref())}
>
  マスタ編集  {/* canEdit=false でも表示 ❌ */}
</button>
```

### 2. 主訴区分 Select が操作可能 ❌

```tsx
<Select
  value={chiefComplaintCategoryId ? String(chiefComplaintCategoryId) : ""}
  onValueChange={(value) => setChiefComplaintCategoryId(value ? Number(value) : null)}
  disabled={isLoading}  // ← isLoading のみ。canEdit ガードなし ❌
>
```

### 3. 定型文挿入ボタン群が操作可能 ❌

```tsx
{templates.map((tmpl) => (
  <Button
    onClick={() => onInsertTemplate(tmpl.text)}
    // disabled={!canEdit} ← なし ❌
  >
    {tmpl.label}
  </Button>
))}
```

### 4. 主訴詳細・治療方針テキストエリアが編集可能 ❌

`InterviewTreatmentPolicy` も同様に `canEdit` 未受け取り。

## 影響コンポーネント

| ファイル | 問題 |
|---------|------|
| `MedicalRecordInterview.tsx` | `canEdit` prop なし、子コンポーネントに渡せない |
| `InterviewChiefComplaint.tsx` | 「マスタ編集」ボタン × 2、Select、テンプレートボタン群、textarea |
| `InterviewTreatmentPolicy.tsx` | 治療方針 textarea |

## 根本原因

```tsx
// MedicalRecordForm.tsx — canEdit を MedicalRecordInterview に渡していない
<MedicalRecordInterview
  chiefComplaint={chiefComplaint}
  setChiefComplaint={handleSetChiefComplaint}
  chiefComplaintCategoryId={chiefComplaintCategoryId}
  setChiefComplaintCategoryId={handleSetChiefComplaintCategoryId}
  treatmentPolicy={treatmentPolicy}
  setTreatmentPolicy={handleSetTreatmentPolicy}
  historyItems={historyItems}
  // canEdit={canEdit}  ← 渡されていない ❌
/>
```

## 修正方針

### Option A: usePermission を各コンポーネント内で呼ぶ（MedicalRecordDiagnosisPlan と同パターン）

```tsx
// InterviewChiefComplaint.tsx
const { canEdit } = usePermission("medical-records");

// マスタ編集ボタンを隠す
{canEdit ? (
  <button onClick={() => navigate(...)}>マスタ編集</button>
) : null}

// Select を disabled に
<Select disabled={isLoading || !canEdit}>

// テンプレートボタンを disabled に
<Button disabled={!canEdit} onClick={() => onInsertTemplate(tmpl.text)}>
```

### Option B: canEdit を props で渡す

```tsx
// MedicalRecordInterview.tsx
interface MedicalRecordInterviewProps {
  // ...
  canEdit?: boolean;  // 追加
}

// MedicalRecordForm.tsx
<MedicalRecordInterview ... canEdit={canEdit} />
```

Option A（usePermission 内部呼び出し）が既存パターンと一致するため推奨。

## 優先度

**HIGH** — 閲覧のみユーザーが主訴・治療方針を変更し、マスタ設定ページへの遷移ボタンが操作可能。

## 関連ファイル

- `frontend/src/features/medical-records/components/InterviewChiefComplaint.tsx`
- `frontend/src/features/medical-records/components/InterviewTreatmentPolicy.tsx`
- `frontend/src/features/medical-records/components/MedicalRecordInterview.tsx`
- `frontend/src/features/medical-records/routes/MedicalRecordForm.tsx`
- 発見日: 2026-04-08（RBAC Phase 3 テスト中）
- 関連: FE-156（同パターン・タブ群対応済み）
