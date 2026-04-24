# FE-167: 医薬品マスタ（MedicineSettings）— MedicineSidePanel に readOnly が未実装で canEdit=false でも保存・削除ボタン表示

## 概要

`/settings/medicine` の `MedicineSettings.tsx` では、`MedicineSidePanel` が `MasterCRUDPage` を経由しない独自実装になっているため、FE-157/FE-161 の MasterCRUDPage 側の修正が適用されない。`MedicineSidePanelProps` に `readOnly` が定義されておらず、`canEdit=false` / `canDelete=false` でも SidePanel の保存ボタン・削除ボタンが常時表示される。

## 根本原因

```tsx
// MedicineSettings.tsx 行 216-227 — readOnly が Props にない ❌
interface MedicineSidePanelProps {
  isEditing: boolean;
  selectedMedicine: Medicine | null;
  // ...
  handleSave: () => void;
  handleDeleteRequest: () => void;
  // readOnly?: boolean;  ← なし
}

// 行 410 — canCreate と canEdit のみ取得、canDelete なし ❌
const { canCreate, canEdit } = usePermission(ResourceMasterMedical);

// 行 261-271 — MasterSidePanel に readOnly が渡されない ❌
<MasterSidePanel
  isNew={!selectedMedicine}
  title={formData.name}
  onClose={handleCloseEdit}
  action={handleAction}
  onDelete={selectedMedicine ? handleDeleteRequest : undefined}  // canDelete チェックなし
  // readOnly={!canEdit}  ← なし
>
```

## 影響

`canEdit=false` / `canDelete=false` のユーザーが医薬品をクリックすると：
1. SidePanel が開く（FE-157 の行クリック問題と連動）
2. 「保存」ボタンが表示される → `handleAction()` 実行 → API PATCH → 403
3. 「削除」ボタン（ゴミ箱）が表示される → `handleDeleteRequest()` → API DELETE → 403

## 修正方針

```tsx
// 1. usePermission から canDelete も取得
const { canCreate, canEdit, canDelete } = usePermission(ResourceMasterMedical);

// 2. MedicineSidePanelProps に readOnly を追加
interface MedicineSidePanelProps {
  // ...既存プロパティ...
  readOnly?: boolean;  // ← 追加
}

// 3. MedicineSidePanel の引数に readOnly を追加
const MedicineSidePanel = memo(function MedicineSidePanel({
  // ...既存引数...
  readOnly,  // ← 追加
}: MedicineSidePanelProps) {
  return (
    <MasterSidePanel
      // ...
      onDelete={!readOnly && selectedMedicine ? handleDeleteRequest : undefined}  // canDelete ガード
      readOnly={readOnly}  // ← 追加（SidePeekFooter に保存ボタン非表示を伝播）
    >
```

// 4. 呼び出し側（MedicineSettings main、行 843-863 付近）で readOnly を渡す
```tsx
<MedicineSidePanel
  // ...
  readOnly={!canEdit}  // ← 追加
/>
```

## 優先度

**HIGH** — 医薬品マスタは在庫・処方に直結する重要データ。`canEdit=false` ユーザーが医薬品情報（用法・用量・単価等）を変更しようとして API エラーが発生する。

## 関連ファイル

- `frontend/src/features/master/routes/MedicineSettings.tsx` (行 216-227, 261-271, 410)
- `frontend/src/components/shared/SidePeek/MasterSidePanel.tsx`
- 発見日: 2026-04-08（RBAC Phase 2/3 テスト中）
- 関連: FE-157（全マスタ設定 readOnly 伝播漏れ）、FE-161（MasterCRUDPage onDeleteRequest ガード漏れ）
