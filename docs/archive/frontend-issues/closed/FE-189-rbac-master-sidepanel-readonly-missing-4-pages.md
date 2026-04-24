# FE-189: マスタ設定 4 ページで SidePanel の readOnly 未実装（保存ボタン常時表示）

## 概要

以下のマスタ設定ページでは `canEdit=false` のユーザーが行クリックで SidePanel を開いても、
保存ボタンが表示されフォームフィールドが操作可能なままになっている。
`MasterCRUDPage` は `readOnly: !canEdit` を `renderSidePanel` callback に渡しているが、
各 SidePanel コンポーネントの props 型に `readOnly` が未定義で処理されていない。

## 影響ページ

| ファイル | ルート | 状態 |
|---------|-------|------|
| `ServiceTypeSettings.tsx` | `/settings/service-type` | readOnly 未定義 ❌ |
| `AnimalSpeciesSettings.tsx` | `/settings/animal-species` | readOnly 未定義 ❌ |
| `CageSettings.tsx` | `/settings/cage` | readOnly 未定義 ❌ |
| `MerchandiseItemSettings.tsx` | `/settings/merchandise-items` | readOnly 未定義 ❌ |

注: `StaffSettings.tsx` の問題は FE-188 として別途起票済み（props 型には readOnly があるが renderSidePanel callback で受け取り漏れ）。

## 根本原因

```tsx
// ServiceTypeSettings.tsx — SidePanel props 型に readOnly がない ❌
const ServiceTypeSidePanel = memo(function ServiceTypeSidePanel({
  item, onClose, onSave, onDeleteRequest,
}: {
  item: ServiceType | null;
  onClose: () => void;
  onSave: (d: ServiceTypeFormData) => void;
  onDeleteRequest?: (i: ServiceType) => void;
  // readOnly?: boolean; ← 未定義
}) {
  return (
    <MasterSidePanel
      action={handleAction}        // ← 常に保存アクション（readOnly チェックなし）
      onDelete={...}               // ← 常に削除アクション（readOnly チェックなし）
      // readOnly={readOnly}       ← MasterSidePanel に渡していない
    />
  );
});
```

`renderSidePanel={(props) => <ServiceTypeSidePanel key={props.item?.id ?? "new"} {...props} />}` で
`{...props}` スプレッドしても TypeScript の型に `readOnly` がないため動作しない。

## 修正方法

各 SidePanel に `readOnly?: boolean` を追加し `MasterSidePanel` に渡す。

```tsx
// ServiceTypeSettings.tsx
interface ServiceTypeSidePanelProps {
  item: ServiceType | null;
  onClose: () => void;
  onSave: (d: ServiceTypeFormData) => void;
  onDeleteRequest?: (i: ServiceType) => void;
  readOnly?: boolean;   // ← 追加
}

const ServiceTypeSidePanel = memo(function ServiceTypeSidePanel({
  item, onClose, onSave, onDeleteRequest, readOnly,
}: ServiceTypeSidePanelProps) {
  ...
  return (
    <MasterSidePanel
      ...
      action={readOnly ? undefined : handleAction}   // ← 修正
      onDelete={readOnly ? undefined : (item !== null && onDeleteRequest ? () => onDeleteRequest(item) : undefined)}
      readOnly={readOnly}   // ← 追加
    />
  );
});
```

同様の修正を `AnimalSpeciesSettings.tsx`、`CageSettings.tsx`、`MerchandiseItemSettings.tsx` に適用。

## 優先度

**HIGH** — 閲覧のみユーザーが各マスタの設定を変更・保存できる状態。

## 関連ファイル

- `frontend/src/features/master/routes/ServiceTypeSettings.tsx`
- `frontend/src/features/master/routes/AnimalSpeciesSettings.tsx`
- `frontend/src/features/master/routes/CageSettings.tsx`
- `frontend/src/features/master/routes/MerchandiseItemSettings.tsx`
- 発見日: 2026-04-08（RBAC Phase 3 テスト中）
- 関連: FE-188（StaffSettings の同問題）、FE-161〜FE-176（マスタ設定 readOnly 修正）の修正漏れ
