# FE-185: フォームページの新規作成モードで canCreate でなく canEdit をチェックしている（系統的欠陥）

## 概要

新規作成と編集の両方を扱うフォームページ 5 件で、`usePermission` から `canCreate` を取得していない。新規作成モード（isEdit=false）でも `canEdit` でガードしているため、`canCreate=false, canEdit=true` のユーザーが新規レコードを作成できてしまう。

## 影響範囲

| ファイル | 新規作成ボタンラベル | 既存 SubmitButton ガード | 正しいガード | 深刻度 |
|---------|---------------------|------------------------|------------|--------|
| `VaccinationForm.tsx` | 「保存」 | `{canEdit ? ...}` | `{isEdit ? canEdit : canCreate ? ...}` | HIGH |
| `ExaminationForm.tsx` | 「保存」 | `{canEdit ? ...}` | `{isEdit ? canEdit : canCreate ? ...}` | HIGH |
| `TrimmingForm.tsx` | 「保存」 | `{canEdit ? ...}` | `{mode === "edit" ? canEdit : canCreate ? ...}` | HIGH |
| `InventoryForm.tsx` | 「登録」 / 「更新」 | `{canEdit ? ...}` | `{isEdit ? canEdit : canCreate ? ...}` | HIGH |
| `HospitalizationForm.tsx` | 「登録」 / 「更新」 | `{canEdit ? ...}` | `{isEdit ? canEdit : canCreate ? ...}` | HIGH |
| `MedicalRecordForm.tsx` | 「保存」（新規カルテ）| `{canEdit ? ...}` | `{isNewRecord ? canCreate : canEdit ? ...}` | HIGH |
| `OwnerForm.tsx` | 「登録」 / 「更新」 | `{canEdit ? ...}` | `{isEdit ? canEdit : canCreate ? ...}` | HIGH |
| `AccountingDetail.tsx` | 「会計を確定する」（新規） | `{canEdit ? ...}` | `{!id ? canCreate : canEdit ? ...}` | HIGH |

## 根本原因

```tsx
// VaccinationForm.tsx 行 37 — canCreate 未取得 ❌
const { canEdit, canDelete } = usePermission("vaccinations");
//      ↑ canCreate がない

// 行 164-170: 新規作成モードも canEdit でガード ❌
const { isEdit } = useVaccinationForm(id);  // isEdit フラグはある

{canEdit ? (  // ← 新規のとき canCreate であるべきだが canEdit を使用 ❌
  <SubmitButton>保存</SubmitButton>
) : null}
```

```tsx
// InventoryForm.tsx 行 271 — canCreate 未取得 ❌
const { canEdit } = usePermission("inventory");
//      ↑ canCreate も canDelete もない

// 行 337-342: 「登録」ボタンも canEdit でガード ❌
headerAction={
  canEdit ? (
    <SubmitButton>
      {isEdit ? "更新" : "登録"}  {/* 「登録」も canEdit でガード ❌ */}
    </SubmitButton>
  ) : null
}
```

```tsx
// HospitalizationForm.tsx 行 41 — canCreate 未取得 ❌
const { canEdit, canDelete } = usePermission("hospitalization");

// 行 170-176 ❌
{canEdit ? (
  <SubmitButton>
    {hospitalizationId ? "更新" : "登録"}  {/* 「登録」も canEdit でガード ❌ */}
  </SubmitButton>
) : null}
```

## 修正方針

```tsx
// 各フォームで canCreate を追加取得
const { canCreate, canEdit, canDelete } = usePermission("vaccinations");

// モードに応じてガードを切り替え
const canSubmit = isEdit ? canEdit : canCreate;

{canSubmit ? (
  <SubmitButton>保存</SubmitButton>
) : null}
```

```tsx
// InventoryForm.tsx 例
const { canCreate, canEdit } = usePermission("inventory");
const canSubmit = isEdit ? canEdit : canCreate;

headerAction={
  canSubmit ? (
    <SubmitButton>
      {isEdit ? "更新" : "登録"}
    </SubmitButton>
  ) : null
}
```

## 優先度

**HIGH** — RBAC でスタッフに `canCreate=false` を設定しても、`canEdit=true` であればすべての新規作成フォームから新規レコードを作成できてしまう。診療記録・入院・トリミング・在庫・検査の新規作成が制御できない。

## 関連ファイル

- `frontend/src/features/vaccinations/routes/VaccinationForm.tsx` (行 37: usePermission, 行 164-170: SubmitButton)
- `frontend/src/features/examinations/routes/ExaminationForm.tsx` (行 197: usePermission, 行 177-183: SubmitButton)
- `frontend/src/features/trimming/routes/TrimmingForm.tsx` (行 418: usePermission, 行 558-562: SubmitButton)
- `frontend/src/features/inventory/routes/InventoryForm.tsx` (行 271: usePermission, 行 337-342/376-381: SubmitButton)
- `frontend/src/features/hospitalization/routes/HospitalizationForm.tsx` (行 41: usePermission, 行 170-176: SubmitButton)
- `frontend/src/features/medical-records/routes/MedicalRecordForm.tsx` (行 156: usePermission, 行 454: SubmitButton, `isNewRecord` 変数あり)
- `frontend/src/features/owners/routes/OwnerForm.tsx` (行 439: usePermission, 行 573-577: SubmitButton, `isEdit` 変数あり)
- `frontend/src/features/accounting/routes/AccountingDetail.tsx` (行 949: usePermission, 行 572-586: SubmitButton, `!id` で新規モード判定)
- 発見日: 2026-04-08（RBAC Phase 2/3 テスト中）
- 関連: FE-166（フォームフィールドの disabled 欠落）
