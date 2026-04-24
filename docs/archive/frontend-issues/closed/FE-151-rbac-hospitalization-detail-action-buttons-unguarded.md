# FE-151: HospitalizationDetail — CarePlan・DailyRecords の操作ボタン全て canEdit/canCreate ガードなし

## 概要

`HospitalizationDetail` 配下のサブコンポーネントに `usePermission` が実装されておらず、閲覧のみユーザーでも入院管理の詳細画面で以下の操作ができてしまう：

- ケアプラン追加・編集・削除
- 日次記録（バイタル・ケアログ・スタッフノート）の追加

## 影響ファイル

| ファイル | 問題箇所 | 詳細 |
|---------|---------|------|
| `CarePlanItemRow.tsx` | 行 82-85 | Edit・DeleteIconButton — canEdit/canDelete なし |
| `CarePlanSection.tsx` | 行 59-62 | 「プラン追加」ボタン — canCreate なし |
| `DailyRecordsTab.tsx` | 行 140 | 「この日の記録を作成」ボタン — canCreate なし |
| `DailyRecordsTab.tsx` | `handleAddVital`, `handleAddCareLog`, `handleAddStaffNote` | canCreate なし |
| `DailyVitalsSection.tsx` | 行 97 | 「追加」ボタン → 保存ダイアログ — canCreate なし |

注: `HospitalizationDetailActions` の「退院処理」「入院情報の編集」はすでに `usePermission("hospitalization")` でガードされている ✅

## 現状の挙動（バグ）

閲覧のみユーザーが入院詳細画面を開くと：
1. ケアプランの各行に編集（鉛筆）・削除アイコンが表示され、クリック可能
2. 「プラン追加」ボタンが表示される
3. 「この日の記録を作成」ボタンが表示される
4. バイタル・ケアログ・スタッフノートの「追加」ボタンが表示される

## 修正方針

各サブコンポーネントで `usePermission("hospitalization")` を呼ぶか、`HospitalizationDetail` から `canEdit`, `canCreate`, `canDelete` を props として渡す。

```tsx
// CarePlanSection.tsx
const { canCreate } = usePermission("hospitalization");

{canCreate ? (
  <Button variant="primary" onClick={handleOpenCreate}>
    <Plus /> プラン追加
  </Button>
) : null}

// CarePlanItemRow.tsx
const { canEdit, canDelete } = usePermission("hospitalization");

{canEdit ? <Button variant="ghost" size="sm" onClick={() => onEdit(plan)}><Edit2 /></Button> : null}
{canDelete ? <DeleteIconButton onClick={() => onDelete(plan.id)} /> : null}
```

## 優先度

HIGH — 閲覧のみユーザーが操作ボタンを目視でき、クリックするとAPIエラー（403）になる

## 関連

- `frontend/src/features/hospitalization/components/CarePlan/CarePlanItemRow.tsx`
- `frontend/src/features/hospitalization/components/CarePlan/CarePlanSection.tsx`
- `frontend/src/features/hospitalization/components/DailyRecordsTab/DailyRecordsTab.tsx`
- `frontend/src/features/hospitalization/components/DailyRecordsTab/DailyVitalsSection.tsx`
- BUG-RBAC テスト 2026-04-07
