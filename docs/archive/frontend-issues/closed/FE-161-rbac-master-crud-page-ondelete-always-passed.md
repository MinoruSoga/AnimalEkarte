# FE-161: MasterCRUDPage — onDeleteRequest が canDelete に関係なく常に渡され、SidePanel に削除ボタンが常時表示される

## 概要

`MasterCRUDPage.tsx` の `renderSidePanel` 呼び出しで `onDeleteRequest: crud.setPendingDelete` が `canDelete` のチェックなしに常に渡されている。その結果、`canDelete=false` のユーザーでも全マスタ設定の SidePeek パネルに削除ボタン（ゴミ箱アイコン）が表示される。

## 根本原因

```tsx
// MasterCRUDPage.tsx 行 122-128 — onDeleteRequest が常に渡される ❌
renderSidePanel({
  item: crud.panelItem,
  onClose: crud.handleClose,
  onSave: handleSave,
  onDeleteRequest: crud.setPendingDelete,  // ← canDelete チェックなし
  readOnly: !canEdit,
})
```

```tsx
// SidePeekToolbar.tsx 行 20-28 — onDelete が truthy なら必ずゴミ箱ボタンを表示
{onDelete ? (
  <button type="button" onClick={onDelete}>
    <Trash2 />
  </button>
) : null}
```

データフロー:
1. `MasterCRUDPage` → `onDeleteRequest: crud.setPendingDelete`（常に truthy）
2. `OccupationSidePanel` → `onDelete={item !== null ? () => onDeleteRequest(item) : undefined}`
3. `MasterSidePanel` → `SidePeekToolbar` → ゴミ箱ボタン **常に表示** ❌

## 影響範囲

`MasterCRUDPage` を使用する全マスタ設定ページ（SidePanel が開いている状態）:
- 職種設定、主訴設定、入院種別設定、保険設定、スタッフ設定
- インタビューテンプレート設定、診断名設定、トリミングメニュー設定 等

## 期待する挙動

`canDelete=false` の場合、SidePeek パネルのヘッダーにゴミ箱ボタンが表示されない。

## 修正方針

### MasterCRUDPage.tsx の onDeleteRequest を canDelete で条件付け

```tsx
// MasterCRUDPage.tsx — 修正後
renderSidePanel({
  item: crud.panelItem,
  onClose: crud.handleClose,
  onSave: handleSave,
  onDeleteRequest: canDelete ? crud.setPendingDelete : undefined,  // ← 追加
  readOnly: !canEdit,
})
```

合わせて `SidePanelRenderProps<T>` の型定義で `onDeleteRequest` をオプショナルにする:

```tsx
// MasterCRUDPage.tsx の型定義
export interface SidePanelRenderProps<T> {
  item: T | null;
  onClose: () => void;
  onSave: (data: unknown) => void;
  onDeleteRequest?: (item: T) => void;  // ← ? を追加
  readOnly?: boolean;
}
```

各 `XxxSidePanel` でも `onDeleteRequest` をオプショナルとして扱う:

```tsx
// OccupationSettings.tsx の OccupationSidePanel
const OccupationSidePanel = memo(function OccupationSidePanel({
  item, onClose, onSave, onDeleteRequest, readOnly,
}: {
  item: Occupation | null;
  onClose: () => void;
  onSave: (d: OccupationFormData) => void;
  onDeleteRequest?: (i: Occupation) => void;  // ← ? を追加
  readOnly?: boolean;
}) {
  return (
    <MasterSidePanel
      onDelete={item !== null && onDeleteRequest ? () => onDeleteRequest(item) : undefined}
      readOnly={readOnly}
      ...
    />
  );
```

## 優先度

**HIGH** — 全マスタ設定ページで `canDelete=false` のユーザーに削除ボタンが表示され、クリック時に API DELETE → 403 が発生する。1 ファイル（`MasterCRUDPage.tsx`）の修正で全マスタページを一括修正できる。

## 関連ファイル

- `frontend/src/features/master/components/MasterCRUDPage.tsx` (行 126) ← 主要修正箇所
- `frontend/src/components/shared/SidePeek/SidePeekToolbar.tsx` (行 20-28)
- `frontend/src/features/master/routes/OccupationSettings.tsx` (行 37-90)
- 他全マスタ XxxSidePanel コンポーネント
- 発見日: 2026-04-07（RBAC Phase 3 テスト中）
