# FE-155: MedicalRecordForm — 担当医ボタンクリックで PATCH /inquiries が 403（閲覧のみユーザー）

## 概要

`/medical-records/:id` の編集フォームで、閲覧のみユーザー（canEdit=false）が `PatientInfoCard` の「担当医」ボタンをクリックすると、`PATCH /api/v1/medical-records/:id/inquiries` が発火し、403 Forbidden が返される。

## 影響範囲

- `frontend/src/features/medical-records/routes/MedicalRecordForm.tsx`
- `frontend/src/components/shared/PatientInfoCard/PatientInfoCard.tsx`
- 権限: `can_edit = false` のユーザー（閲覧のみ）

## 現状の挙動（バグ）

担当医ボタン（PatientInfoCard 内）をクリックすると：

```
PATCH /api/v1/medical-records/22/inquiries → 403 Forbidden
```

XHR インターセプトで確認済み:
```js
[{"m":"PATCH","u":"/api/v1/medical-records/22/inquiries","s":403}]
```

## 根本原因

`MedicalRecordForm.tsx` は `<form action={formAction}>` で囲まれており、内部の `PatientInfoCard` の担当医ボタンが **`type="button"` を明示していないか、onClick ハンドラが `formAction` を間接的にトリガー** している。

```tsx
// PatientInfoCard.tsx — 担当医ボタン（type 未指定の場合 type="submit" として動作）
<Button
  id={staffButtonId}
  variant="ghost"
  size="sm"
  onClick={onStaffClick}
>
  {staffName}
  <ChevronDown />
</Button>
```

```tsx
// MedicalRecordForm.tsx — usePermission の利用
const { canEdit, canDelete } = usePermission("medical-records");

// SubmitButton は canEdit でガード済み ✅
{canEdit ? <SubmitButton>保存</SubmitButton> : null}

// しかし PatientInfoCard の担当医ボタンには canEdit ガードなし ❌
<PatientInfoCard
  ...
  onStaffClick={handleStaffClick}  // canEdit 関係なく渡されている
/>
```

## 期待する挙動

`canEdit` が false の場合：
1. 担当医ボタンがクリック不可（disabled）またはクリックしても PATCH が発火しない
2. スタッフ選択ダイアログが開かない

## 修正方針

### 方針 A: `onStaffClick` を canEdit で条件付き渡し（推奨）

```tsx
// MedicalRecordForm.tsx
<PatientInfoCard
  ...
  onStaffClick={canEdit ? handleStaffClick : undefined}
/>
```

`PatientInfoCard` 側で `onStaffClick` が undefined の場合はボタンを disabled にするか非表示にする。

### 方針 B: PatientInfoCard 側に `disabled` prop を追加

```tsx
// PatientInfoCard.tsx
<Button
  type="button"   // ← form submit を防止するため必須
  disabled={!canEdit}
  onClick={onStaffClick}
>
```

### 根本対策: Button に `type="button"` を必ず明示

`<form action={formAction}>` 内にある全ボタンで、submit ではないものには `type="button"` を明示すること（HTML デフォルトは `type="submit"`）。

## 優先度

**HIGH** — 閲覧のみユーザーがカルテを開くだけで API 403 エラーが発生し、コンソールに エラーが出る。実際のデータ変更は 403 によって防がれるが、「スタッフ選択ダイアログを開いて担当医を変更できそう」な誤認を生む。

## 関連

- `frontend/src/features/medical-records/routes/MedicalRecordForm.tsx`
- `frontend/src/components/shared/PatientInfoCard/PatientInfoCard.tsx`
- 発見日: 2026-04-07（RBAC テスト中）
- XHR 証拠: `[{"m":"PATCH","u":"/api/v1/medical-records/22/inquiries","s":403}]`
