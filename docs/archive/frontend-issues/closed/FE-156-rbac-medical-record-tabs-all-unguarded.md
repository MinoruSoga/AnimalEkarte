# FE-156: MedicalRecordForm — 全サブタブコンポーネントに canEdit ガードなし（システム的欠如）

## 概要

`/medical-records/:id` の編集フォームで、`usePermission("medical-records")` は `MedicalRecordForm.tsx` 本体では呼ばれているが、**各タブのサブコンポーネントに canEdit が一切渡されておらず**、閲覧のみユーザー（canEdit=false）が全タブで CRUD 操作を実行できる。

## 影響範囲（全 5 コンポーネント）

| ファイル | アクション | API 呼び出し | 深刻度 |
|---------|-----------|------------|--------|
| `MedicalRecordDiagnosisPlan.tsx` | 「行を追加（検索）」ボタン、TreatmentTable の行編集・削除 | POST/PATCH/DELETE `/v1/treatments` | HIGH |
| `MedicalRecordVaccination.tsx` | 「保存」ボタン | POST `/v1/vaccinations` | HIGH |
| `MedicalRecordImage.tsx` | ファイルアップロード、画像削除 | POST/DELETE `/v1/images` | HIGH |
| `MedicalRecordEstimate.tsx` | 「行追加」「行削除」、見積保存 | POST/PATCH `/v1/estimates` | HIGH |
| `MedicalRecordBillCheck.tsx` | 「医師確認」「確認取り消し」、治療明細 CRUD | POST/PATCH `/v1/billing-review`, POST/PATCH `/v1/treatments` | HIGH |

注: `MedicalRecordExamination.tsx` は GET のみ（表示専用）のため問題なし ✅

## 現状の挙動（バグ）

```tsx
// MedicalRecordForm.tsx — canEdit は取得しているが...
const { canEdit, canDelete } = usePermission("medical-records");

// 各タブに canEdit が渡されていない ❌
<MedicalRecordDiagnosisPlan
  plan={plan}
  setPlan={handleSetPlan}
  // canEdit={canEdit}  ← 渡されていない
/>

<MedicalRecordBillCheck
  medicalRecordId={recordId}
  // canEdit={canEdit}  ← 渡されていない
/>
```

閲覧のみユーザーが実行できてしまう操作:
1. 「行を追加（検索）」→ 治療プラン行を追加（API → 403）
2. 「保存」（ワクチン）→ ワクチン記録を保存（API → 403）
3. 画像アップロード → 403
4. 画像削除 → 403
5. 「医師確認」ボタンをクリック → 403
6. 見積の行追加・削除 → 403

## 根本原因

タブコンポーネント設計時に RBAC を考慮していなかった。`MedicalRecordForm.tsx` 本体で `canEdit` を取得しているが、サブコンポーネントの props に定義・注入がない。

## 修正方針

### Step 1: 各サブコンポーネントの Props に `canEdit` を追加

```tsx
// MedicalRecordDiagnosisPlan.tsx
export interface DiagnosisPlanProps {
  // ...既存 props
  canEdit?: boolean;  // 追加
}

// TreatmentTable に disabled={!canEdit} を渡す
<TreatmentTable
  ...
  disabled={!canEdit}  // 追加
/>
```

```tsx
// MedicalRecordVaccination.tsx
// 「保存」ボタンを canEdit でガード
{canEdit ? <Button onClick={handleSave}>保存</Button> : null}
```

```tsx
// MedicalRecordImage.tsx
// アップロードエリアと削除ボタンを canEdit でガード
{canEdit ? <FileUploadArea onFilesSelected={handleFilesSelected} /> : null}
<Button onClick={handleDeleteImage} disabled={!canEdit}>削除</Button>
```

```tsx
// MedicalRecordBillCheck.tsx
// 「医師確認」「確認取り消し」ボタンを canEdit でガード
{canEdit ? <Button onClick={handleConfirm}>医師確認</Button> : null}
// TreatmentTable に disabled={!canEdit || isConfirmed} を渡す
```

### Step 2: MedicalRecordForm.tsx から canEdit を注入

```tsx
<MedicalRecordDiagnosisPlan canEdit={canEdit} ... />
<MedicalRecordVaccination canEdit={canEdit} ... />
<MedicalRecordImage canEdit={canEdit} ... />
<MedicalRecordEstimate canEdit={canEdit} ... />
<MedicalRecordBillCheck canEdit={canEdit} ... />
```

## 優先度

**HIGH** — 閲覧のみユーザーが複数の API エンドポイント（ワクチン・画像・見積・医師確認）に対して変更操作を試みることができる。いずれも API 403 で実際のデータ変更は防がれるが、UI が誤った操作可能感を与え、403 エラーがコンソールに大量に出力される。

## 関連ファイル

- `frontend/src/features/medical-records/routes/MedicalRecordForm.tsx`
- `frontend/src/features/medical-records/components/MedicalRecordDiagnosisPlan.tsx`
- `frontend/src/features/medical-records/components/MedicalRecordVaccination.tsx`
- `frontend/src/features/medical-records/components/MedicalRecordImage.tsx`
- `frontend/src/features/medical-records/components/MedicalRecordEstimate.tsx`
- `frontend/src/features/medical-records/components/MedicalRecordBillCheck.tsx`
- 発見日: 2026-04-07（RBAC テスト中）
