# FE-170: シフト管理（ShiftFormDialog）— usePermission 未使用・保存・削除ボタン常時表示

## 概要

`/shifts` の `ShiftCalendarPage.tsx` は `usePermission("shifts")` から `canCreate`, `canEdit` を取得しているが、その権限情報を `ShiftFormDialog` へ渡していない。`ShiftFormDialog.tsx` は自身では `usePermission` を呼び出さず、保存ボタンと削除確認ダイアログが権限に関係なく常時表示される。

## 根本原因

```tsx
// ShiftCalendarPage.tsx 行 18 — canCreate/canEdit は取得しているが ❌
const { canCreate, canEdit } = usePermission("shifts");

// ShiftFormDialog に権限情報を渡していない ❌
<ShiftFormDialog
  // canEdit={canEdit}  ← なし
  // canDelete={canDelete}  ← canDelete 自体も未取得
  {...props}
/>
```

```tsx
// ShiftFormDialog.tsx — usePermission なし ❌
export function ShiftFormDialog({ ... }: ShiftFormDialogProps) {
  // usePermission 呼び出しなし

  // 保存ボタン — 権限ガードなし ❌
  <SubmitButton>保存</SubmitButton>

  // 削除ダイアログ — 権限ガードなし ❌（行 147-162）
  <DeleteConfirmDialog onConfirm={handleDelete} />
}
```

加えて `ShiftCalendarPage.tsx` では `canDelete` を `usePermission` から取得していない。

## 影響

`canEdit=false` / `canDelete=false` のユーザーがシフト枠をクリックすると：
1. `ShiftFormDialog` が開く（カレンダー行クリック → 編集ダイアログ open）
2. 全フォームフィールドが入力可能
3. 「保存」ボタンが表示される → PATCH `/v1/shifts/:id` → 403
4. 削除確認ダイアログが呼び出せる → DELETE `/v1/shifts/:id` → 403

## 修正方針

```tsx
// ShiftCalendarPage.tsx — canDelete も追加取得
const { canCreate, canEdit, canDelete } = usePermission("shifts");

// ShiftFormDialog に権限を渡す
<ShiftFormDialog
  canEdit={canEdit}
  canDelete={canDelete}
  {...props}
/>
```

```tsx
// ShiftFormDialog.tsx — Props に追加
interface ShiftFormDialogProps {
  canEdit?: boolean;
  canDelete?: boolean;
  // ...既存 Props...
}

export function ShiftFormDialog({ canEdit = false, canDelete = false, ...}: ShiftFormDialogProps) {
  return (
    <>
      <fieldset disabled={!canEdit}>
        {/* フォームフィールド */}
      </fieldset>

      {canEdit ? (
        <SubmitButton>保存</SubmitButton>
      ) : null}

      {canDelete ? (
        <DeleteConfirmDialog onConfirm={handleDelete} />
      ) : null}
    </>
  );
}
```

## 優先度

**HIGH** — シフト管理は病院スタッフの勤務管理に直結する。`canEdit=false` ユーザーがシフトを変更・削除を試みると API エラーが発生する。

## 関連ファイル

- `frontend/src/features/shifts/routes/ShiftCalendarPage.tsx` (行 18: usePermission, canDelete 未取得)
- `frontend/src/features/shifts/components/ShiftFormDialog/ShiftFormDialog.tsx` (保存ボタン、削除ダイアログ)
- 発見日: 2026-04-08（RBAC Phase 2/3 テスト中）
- 関連: FE-166（フォームフィールド disabled 漏れ）
