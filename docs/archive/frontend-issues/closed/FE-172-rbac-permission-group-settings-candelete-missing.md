# FE-172: 権限グループ設定（PermissionGroupSettings）— canDelete 未取得・削除が canEdit に依存

## 概要

`/settings/permission-groups` の `PermissionGroupSettings.tsx` は `usePermission(ResourceMasterPermission)` から `canEdit` のみを取得している。`canDelete` を取得しておらず、SidePanel の `onDelete` は `canEdit` に連動している（`canEdit=false` かつ `canDelete=true` のケースでは削除できない、逆に `canEdit=true` かつ `canDelete=false` では削除できてしまう）。

## 現状の挙動（バグ）

```tsx
// PermissionGroupSettings.tsx 行 218 — canEdit のみ取得 ❌
const { canEdit } = usePermission(ResourceMasterPermission);

// 行 328 — RowActionButton は canEdit でガード済み ✅
{canEdit ? <RowActionButton onClick={() => crud.handleEdit(item)} /> : null}

// 行 171 — onDeleteRequest は item !== null のみで渡す ❌
// canDelete チェックなし
onDelete={item !== null ? () => onDeleteRequest(item) : undefined}
```

`canEdit=true` かつ `canDelete=false` のユーザーが権限グループを開いた場合、SidePanel に削除ボタンが表示される（`canDelete` が未確認のため）。また `canEdit=false` かつ `canDelete=true` のユーザーは削除できない（RowActionButton が非表示になるため SidePanel が開けない）。

## 深刻度の理由

**権限グループ設定は RBAC の根幹を制御するデータ。** 誤った権限グループの削除は全ユーザーのアクセス権を破壊する可能性がある。canEdit/canDelete の独立した制御が必須。

## 修正方針

```tsx
// PermissionGroupSettings.tsx 行 218 を修正
const { canEdit, canDelete } = usePermission(ResourceMasterPermission);

// onDeleteRequest を canDelete でガード
onDelete={item !== null && canDelete ? () => onDeleteRequest(item) : undefined}
```

MasterCRUDPage を使っている場合は FE-161 の修正で対応される。`PermissionGroupSettings` が独自実装の場合は上記を直接修正する。

## 優先度

**HIGH** — 権限グループは全 RBAC の基盤データ。誤操作による権限グループ削除は全スタッフのアクセス制御を破壊する。

## 関連ファイル

- `frontend/src/features/master/routes/PermissionGroupSettings.tsx` (行 218: usePermission, 行 171: onDelete)
- 発見日: 2026-04-08（RBAC Phase 2/3 テスト中）
- 関連: FE-161（MasterCRUDPage onDeleteRequest 常時渡し）、FE-168（他の独自実装マスタ設定）
