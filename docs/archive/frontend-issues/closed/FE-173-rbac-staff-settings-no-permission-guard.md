# FE-173: スタッフ管理（StaffSettings）— usePermission 完全欠落・スタッフ追加・編集・削除が無条件に実行可能

## 概要

`/settings/staff` の `StaffSettings.tsx` は `usePermission` を一切呼び出していない。スタッフの追加・編集・削除が権限チェックなしで実行できる状態になっている。スタッフ管理は RBAC の実行主体（誰がどの権限を持つか）に直結するため、特に高リスク。

## 根本原因

```tsx
// StaffSettings.tsx — usePermission なし ❌
export function StaffSettings() {
  // usePermission(ResourceHospitalSettings) が呼ばれていない

  // MasterCRUDPage に resource を渡している場合でも、
  // StaffSettings 自体での canEdit/canDelete チェックなし

  // スタッフ追加ボタン — canCreate チェックなし ❌
  <AddButton onClick={handleAdd}>スタッフを追加</AddButton>

  // renderRow — canEdit チェックなし ❌
  (item, onEdit, canEdit) => <RowActionButton onClick={() => onEdit(item)} />
  // ↑ canEdit が undefined/null になる可能性（resource 未指定の場合）
}
```

`MasterCRUDPage` 経由で `canEdit` が渡される設計になっていても、`StaffSettings` が `usePermission` を呼ばずに適切な resource を指定しなければ、`canEdit` は常に `true` または `undefined` になる。

## 影響

1. `canCreate=false` のユーザーが「スタッフを追加」ボタンを押して新規スタッフを作成できる
2. `canEdit=false` のユーザーが既存スタッフ情報（名前・役職・メールアドレス等）を変更できる
3. `canDelete=false` のユーザーがスタッフを削除できる

**スタッフ管理は RBAC の実行主体**: スタッフを追加・変更できれば間接的に権限グループを操作でき、権限昇格攻撃（Privilege Escalation）の入口になる。

## 修正方針

```tsx
// StaffSettings.tsx
const { canCreate, canEdit, canDelete } = usePermission(ResourceHospitalSettings);

// MasterCRUDPage に渡す resource を明示（canEdit/canDelete が正確に渡るように）
<MasterCRUDPage
  resource={ResourceHospitalSettings}
  canCreate={canCreate}
  // ...
/>
```

または `MasterCRUDPage` 内部の `usePermission(resource)` で自動的に取得する（FE-161 の修正と連動）。

## 優先度

**CRITICAL** — スタッフ管理は権限制御の実行主体。権限なしユーザーによるスタッフ追加・編集・削除は Privilege Escalation のリスクを含む。FE-164（ClinicMasterSettings）・FE-168（CompanySettings 等）と同カテゴリだが、データの機密性が最高。

## 関連ファイル

- `frontend/src/features/master/routes/StaffSettings.tsx`
- 発見日: 2026-04-08（RBAC Phase 2/3 テスト中）
- 関連: FE-161（MasterCRUDPage onDeleteRequest 常時渡し）、FE-168（独自実装マスタ設定 RBAC 漏れ）、FE-172（PermissionGroupSettings canDelete 漏れ）
