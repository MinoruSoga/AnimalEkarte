# FE-168: マスタ設定（独自実装ページ）— canEdit/canDelete ガード未実装・保存ボタン常時表示

## 概要

`MasterCRUDPage` を使わず独自実装している以下のマスタ設定ページで、RBAC ガードが不完全または完全に欠如している。

## 影響範囲

| ファイル | 問題 | 深刻度 |
|---------|------|--------|
| `CompanySettings.tsx` | `usePermission` 未使用。「保存」ボタンに権限チェックなし | HIGH |
| `DiagnosisSettings.tsx` | `canDelete` 未取得。`MasterSidePanel.onDelete` が canDelete チェックなし。readOnly 未伝播 | HIGH |
| `TreatmentPlanMaster.tsx` | `canDelete` 未取得。`MasterSidePanel.onDelete` が canDelete チェックなし。readOnly 未伝播 | HIGH |
| `TrimmingSettings.tsx` | `canCreate` のみ取得。`canEdit`・`canDelete` 未取得。readOnly 未伝播 | HIGH |

---

## 1. CompanySettings.tsx（会社・法人設定）

**根本原因**: `usePermission` が一切呼ばれていない。

```tsx
// CompanySettings.tsx — usePermission なし ❌
export function CompanySettings() {
  // usePermission(ResourceHospitalSettings) が呼ばれていない

  return (
    <button type="button" onClick={handleSave}>
      保存  {/* canEdit チェックなし ❌ */}
    </button>
  );
}
```

**修正方針**:
```tsx
const { canEdit } = usePermission(ResourceHospitalSettings);
// 保存ボタンを canEdit でガード
{canEdit ? (
  <button type="button" onClick={handleSave}>保存</button>
) : null}
```

---

## 2. DiagnosisSettings.tsx（診断名設定）

**根本原因**: `canCreate, canEdit` は取得済みだが `canDelete` 未取得。`MasterSidePanel` に `readOnly` が渡されず、`onDelete` が canDelete チェックなしで常に渡される。

```tsx
// DiagnosisSettings.tsx 行 476 — canDelete なし ❌
const { canCreate, canEdit } = usePermission(ResourceMasterMedical);

// 行 149-172 — readOnly なし、onDelete 無条件 ❌
<MasterSidePanel
  action={handleAction}
  onDelete={item !== null ? () => onDeleteRequest(item) : undefined}
  // readOnly={!canEdit}  ← なし
  // onDelete も canDelete チェックなし
>
```

**修正方針**:
```tsx
const { canCreate, canEdit, canDelete } = usePermission(ResourceMasterMedical);

<MasterSidePanel
  action={canEdit ? handleAction : undefined}
  onDelete={item !== null && canDelete ? () => onDeleteRequest(item) : undefined}
  readOnly={!canEdit}  // ← 追加
>
```

---

## 3. TreatmentPlanMaster.tsx（処置プランマスタ）

**根本原因**: `canCreate, canEdit` は取得済みだが `canDelete` 未取得。`MasterSidePanel` に `readOnly` が渡されず、`onDelete` が canDelete チェックなしで常に渡される。

```tsx
// TreatmentPlanMaster.tsx 行 461 — canDelete なし ❌
const { canCreate, canEdit } = usePermission(ResourceMasterMedical);

// 行 170-208 — readOnly なし、onDelete 無条件 ❌
<MasterSidePanel
  action={handleAction}
  onDelete={item !== null ? onDeleteRequest : undefined}
  // readOnly={!canEdit}  ← なし
>
```

**修正方針**: DiagnosisSettings と同様。

---

## 4. TrimmingSettings.tsx（トリミングメニュー設定）

**根本原因**: `canCreate` のみ取得。`canEdit` も `canDelete` も取得していない。2 つの `MasterSidePanel`（コース設定・メニュー設定）に `readOnly` が渡されない。

```tsx
// TrimmingSettings.tsx 行 501 — canCreate のみ ❌
const { canCreate } = usePermission(ResourceMasterTrimming);

// 行 144 と 344 — MasterSidePanel に readOnly なし ❌
<MasterSidePanel ...>  {/* readOnly={!canEdit} なし */}
```

**修正方針**:
```tsx
const { canCreate, canEdit, canDelete } = usePermission(ResourceMasterTrimming);

<MasterSidePanel
  ...
  readOnly={!canEdit}
  onDelete={canDelete ? handleDeleteRequest : undefined}
>
```

---

## 優先度

**HIGH** — 全て実際のデータ変更操作（保存・削除）をガードしていない。特に `CompanySettings` は `usePermission` が一切存在しないため最優先。

## 関連ファイル

- `frontend/src/features/master/routes/CompanySettings.tsx`
- `frontend/src/features/master/routes/DiagnosisSettings.tsx` (行 476, 149-172, 253-289)
- `frontend/src/features/master/routes/TreatmentPlanMaster.tsx` (行 461, 170-208)
- `frontend/src/features/master/routes/TrimmingSettings.tsx` (行 501, 144, 344)
- 発見日: 2026-04-08（RBAC Phase 2/3 テスト中）
- 関連: FE-157（MasterCRUDPage 系の行クリック問題）、FE-161（onDeleteRequest ガード漏れ）、FE-167（MedicineSettings 同様の問題）
