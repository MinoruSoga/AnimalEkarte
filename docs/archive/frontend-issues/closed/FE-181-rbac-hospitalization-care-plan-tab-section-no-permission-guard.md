# FE-181: 入院 CarePlanTab・CarePlanSection・CarePlanItemRow — usePermission 完全欠落（ケアプラン CRUD 無防備）

## 概要

入院詳細の「ケアプラン」タブを構成する `CarePlanTab`・`CarePlanSection`・`CarePlanItemRow` に `usePermission` が一切実装されていない。ケアプランアイテムの追加・更新・削除が権限チェックなしで実行できる。

## 影響範囲

| ファイル | 問題操作 | API 呼び出し | 深刻度 |
|---------|---------|------------|--------|
| `CarePlanTab.tsx` | ケアプラン追加（行 291: 追加フォームボタン）・更新（行 201: 編集ボタン）・削除（行 206: DeleteIconButton） | POST/PATCH/DELETE `/v1/hospitalizations/*/care-plan-items` | HIGH |
| `CarePlanSection.tsx` | 追加（行 59-62: Plusボタン）・更新・削除（行 49: CarePlanItemRow に onDelete 渡し） | POST/PATCH/DELETE `/v1/hospitalizations/*/care-plan-items` | HIGH |
| `CarePlanItemRow.tsx` | 削除（行 85: DeleteIconButton `onClick={() => onDelete(plan.id)}`） | DELETE `/v1/hospitalizations/*/care-plan-items/:id` | HIGH |

## 根本原因

```tsx
// CarePlanTab.tsx — usePermission なし ❌
const createMutation = useCreateCarePlanItem();  // 行 307
const updateMutation = useUpdateCarePlanItem();  // 行 308
const deleteMutation = useDeleteCarePlanItem();  // 行 309

// 行 350: canCreate チェックなし → POST ❌
const handleAdd = useCallback(() => {
  createMutation.mutate({ ... });
}, [createMutation]);

// 行 339: canDelete チェックなし → DELETE ❌
const handleDelete = useCallback((id: string) => {
  deleteMutation.mutate(id);
}, [deleteMutation]);

// 行 201: 編集ボタンに canEdit ガードなし ❌
<EditIconButton onClick={() => handleEdit(item.id)} />

// 行 206: 削除ボタンに canDelete ガードなし ❌
<DeleteIconButton onClick={() => handleDelete(item.id)} />

// 行 291: 追加フォームボタンに canCreate ガードなし ❌
<Button type="submit" disabled={isSubmitting}>追加</Button>
```

```tsx
// CarePlanSection.tsx — usePermission なし ❌
// 行 59-62: 追加ボタンに canCreate ガードなし ❌
<button onClick={onAdd}>
  <Plus />
</button>

// 行 49: onDelete を CarePlanItemRow に canDelete チェックなく渡す ❌
<CarePlanItemRow onDelete={onDelete} />
```

```tsx
// CarePlanItemRow.tsx — usePermission なし ❌
// 行 85: 削除ボタンに canDelete ガードなし ❌
<DeleteIconButton onClick={() => onDelete(plan.id)} />
```

親コンポーネント（`HospitalizationDetail`）は `usePermission` を持たず、権限情報がこの階層まで伝播していない。

## 修正方針

### 方針 A: CarePlanTab で usePermission を取得（推奨）

```tsx
// CarePlanTab.tsx
const { canCreate, canEdit, canDelete } = usePermission("hospitalization");

// ミューテーション前に権限チェック
const handleDelete = useCallback((id: string) => {
  if (!canDelete) return;
  deleteMutation.mutate(id);
}, [canDelete, deleteMutation]);

// ボタンを canXxx でガード
{canCreate ? <Button type="submit">追加</Button> : null}
<EditIconButton onClick={canEdit ? () => handleEdit(item.id) : undefined} />
<DeleteIconButton onClick={canDelete ? () => handleDelete(item.id) : undefined} />
```

```tsx
// CarePlanSection.tsx
// onAdd / onDelete に canCreate / canDelete を条件付けて渡す
<button onClick={canCreate ? onAdd : undefined}>
  <Plus />
</button>
<CarePlanItemRow onDelete={canDelete ? onDelete : undefined} />
```

```tsx
// CarePlanItemRow.tsx
interface CarePlanItemRowProps {
  onDelete?: (id: string) => void;  // undefined なら非表示
}
{onDelete !== undefined ? (
  <DeleteIconButton onClick={() => onDelete(plan.id)} />
) : null}
```

## 優先度

**HIGH** — 入院患者のケアプラン（投薬スケジュール・処置計画等）は医療記録の重要データ。`canDelete=false` ユーザーが削除できてしまうと治療計画が失われる。

## 関連ファイル

- `frontend/src/features/hospitalization/components/CarePlanTab/CarePlanTab.tsx` (行 307-309: mutations, 行 291/201/206: ボタン)
- `frontend/src/features/hospitalization/components/CarePlan/CarePlanSection.tsx` (行 59-62: 追加ボタン, 行 49: onDelete 渡し)
- `frontend/src/features/hospitalization/components/CarePlan/CarePlanItemRow.tsx` (行 85: DeleteIconButton)
- 発見日: 2026-04-08（RBAC Phase 2/3 テスト中）
- 関連: FE-175（DailyRecordsTab サブセクション未ガード）、FE-160（入院管理削除ボタン未ガード）
