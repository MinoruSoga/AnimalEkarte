# FE-163: 見積管理 — EstimateForm・EstimateDetail で usePermission が未使用（全ボタン無条件表示）

## 概要

`/estimates/new`・`/estimates/:id/edit` の `EstimateForm.tsx` と `/estimates/:id` の `EstimateDetail.tsx` で `usePermission` が呼び出されておらず、`canEdit=false` / `canDelete=false` のユーザーに対して保存・編集・削除ボタンが無条件に表示される。

## 影響範囲

| ファイル | 問題箇所 | 操作 | API | 深刻度 |
|---------|---------|------|-----|--------|
| `EstimateForm.tsx` | `SubmitButton` (行 297-303) | 作成・更新 | POST/PATCH `/v1/estimates` | HIGH |
| `EstimateDetail.tsx` | 「編集」ボタン (行 63) | 編集画面へ遷移 | — (UI) | MEDIUM |
| `EstimateDetail.tsx` | 「削除」ボタン (行 73) | 削除確認ダイアログ表示 → DELETE | DELETE `/v1/estimates/:id` | HIGH |

注: `EstimateList.tsx` は `canCreate`・`canEdit`・`canDelete` を正しくガード済み ✅

## 現状の挙動（バグ）

```tsx
// EstimateForm.tsx — usePermission なし ❌
export function EstimateForm() {
  // usePermission("estimates") が呼ばれていない
  
  return (
    <form action={formAction}>
      ...
      <SubmitButton>         {/* canEdit チェックなし ❌ */}
        {isEdit ? '更新' : '作成'}
      </SubmitButton>
    </form>
  );
}

// EstimateDetail.tsx — usePermission なし ❌
export function EstimateDetail() {
  // usePermission("estimates") が呼ばれていない

  return (
    <>
      <Button onClick={handleEdit}>編集</Button>   {/* canEdit チェックなし ❌ */}
      <Button onClick={handleDelete}>削除</Button> {/* canDelete チェックなし ❌ */}
    </>
  );
}
```

## 修正方針

```tsx
// EstimateForm.tsx
const { canEdit } = usePermission("estimates");

// SubmitButton を canEdit でガード
{canEdit ? (
  <SubmitButton>{isEdit ? '更新' : '作成'}</SubmitButton>
) : null}

// EstimateDetail.tsx
const { canEdit, canDelete } = usePermission("estimates");

{canEdit ? <Button onClick={handleEdit}>編集</Button> : null}
{canDelete ? <Button onClick={handleDelete}>削除</Button> : null}
```

## 優先度

**HIGH** — `canEdit=false` でも見積の作成・更新が試みられ（フォームがアクセス可能）、`canDelete=false` でも削除ボタンが表示される。API は 403 で防ぐが、UI の誤った操作可能感が問題。

## 関連ファイル

- `frontend/src/features/estimates/routes/EstimateForm.tsx` (行 297-303)
- `frontend/src/features/estimates/routes/EstimateDetail.tsx` (行 63, 73)
- 発見日: 2026-04-07（RBAC Phase 2/3 テスト中）
