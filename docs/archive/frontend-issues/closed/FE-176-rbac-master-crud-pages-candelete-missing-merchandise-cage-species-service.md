# FE-176: マスタ設定（MerchandiseItem/Cage/AnimalSpecies/ServiceType）— canDelete 未取得・SidePanel 削除ボタン常時表示

## 概要

以下の 4 つのマスタ設定ページで `usePermission` から `canDelete` を取得していない。各ページの `MasterSidePanel` に `onDeleteRequest` が権限チェックなしで渡されるため、`canDelete=false` のユーザーにも削除ボタンが表示される。

## 影響範囲

| ファイル | usePermission 取得権限 | onDeleteRequest 行番号 | 深刻度 |
|---------|----------------------|----------------------|--------|
| `MerchandiseItemSettings.tsx` | `canCreate, canEdit` のみ (行 211) | 行 134 — 無条件渡し | HIGH |
| `CageSettings.tsx` | `canCreate, canEdit` のみ (行 157) | 行 103 — 無条件渡し | HIGH |
| `AnimalSpeciesSettings.tsx` | `canEdit` のみ (行 79) | 行 68 — 無条件渡し | HIGH |
| `ServiceTypeSettings.tsx` | `canEdit` のみ (行 110) | 行 89 — 無条件渡し | HIGH |

## 根本原因

```tsx
// CageSettings.tsx（例）— canDelete 未取得 ❌
const { canCreate, canEdit } = usePermission(ResourceMasterHospitalization);

// 行 103 — onDeleteRequest に canDelete チェックなし ❌
<MasterSidePanel
  onDeleteRequest={item !== null ? () => onDeleteRequest(item) : undefined}
  // ← canDelete チェックなし
>
```

FE-161（MasterCRUDPage の onDeleteRequest 常時渡し）と同パターンだが、これらのページは MasterCRUDPage 経由でなく直接 MasterSidePanel を呼んでいるか、renderSidePanel コールバック内で同様の問題を持っている。

## 修正方針

```tsx
// 各ページで canDelete も取得
const { canCreate, canEdit, canDelete } = usePermission(ResourceMasterHospitalization);

// onDeleteRequest を canDelete でガード
<MasterSidePanel
  onDeleteRequest={item !== null && canDelete ? () => onDeleteRequest(item) : undefined}
>
```

## 優先度

**HIGH** — ケージ（病室）・動物種別・診療サービス種別・物販マスタは運用上重要なデータ。誤削除で診療・会計フローに影響が出る。

## 関連ファイル

- `frontend/src/features/master/routes/MerchandiseItemSettings.tsx` (行 211: usePermission, 行 134: onDeleteRequest)
- `frontend/src/features/master/routes/CageSettings.tsx` (行 157: usePermission, 行 103: onDeleteRequest)
- `frontend/src/features/master/routes/AnimalSpeciesSettings.tsx` (行 79: usePermission, 行 68: onDeleteRequest)
- `frontend/src/features/master/routes/ServiceTypeSettings.tsx` (行 110: usePermission, 行 89: onDeleteRequest)
- 発見日: 2026-04-08（RBAC Phase 2/3 テスト中）
- 関連: FE-161（MasterCRUDPage onDeleteRequest 常時渡し）、FE-168（CompanySettings 等の独自実装マスタ設定）
