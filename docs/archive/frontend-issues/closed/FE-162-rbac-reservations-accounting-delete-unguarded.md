# FE-162: 予約管理・会計管理 — 削除ボタンが canDelete=false でも表示される

## 概要

`/reservations` と `/accounting/:id` で、`canDelete=false` のユーザーに対して削除ボタンが非表示にならない。

## 影響範囲

| ページ | コンポーネント | 問題箇所 | 深刻度 |
|-------|--------------|---------|--------|
| `/reservations` | `ReservationDetailModal.tsx` | `DeleteIconButton` 行 208 | HIGH |
| `/accounting/:id` | `AccountingDetail.tsx (ItemListCard)` | `DeleteIconButton` 行 195 | MEDIUM |

## 1. 予約管理（ReservationManagement / ReservationDetailModal）

### 根本原因

```tsx
// ReservationManagement.tsx 行 52 — canCreate しか取得していない ❌
const { canCreate } = usePermission("reservations");

// 行 289 — onDelete を canDelete チェックなしに常に渡している ❌
<ReservationDetailModal
  ...
  onDelete={handleDelete}  // canDelete 関係なく常に渡す
/>

// ReservationDetailModal.tsx 行 207-208 — onDelete が truthy なら削除ボタン表示
{onDelete ? (
  <DeleteIconButton onClick={() => onDelete(appointment)} />
) : null}
```

`canCreate` しか取得しておらず、`canEdit` も `canDelete` も参照していない。Phase 2 (FE-159) で指摘した「`usePermission` が reservations feature で参照されていない」問題の延長。

### 修正方針

```tsx
// ReservationManagement.tsx
const { canCreate, canEdit, canDelete } = usePermission("reservations");

// onDelete を canDelete で条件付け
<ReservationDetailModal
  ...
  onDelete={canDelete ? handleDelete : undefined}
/>
```

---

## 2. 会計管理詳細（AccountingDetail / ItemListCard）

### 根本原因

```tsx
// AccountingDetail.tsx 行 949 — canEdit しか取得していない ❌
const { canEdit } = usePermission("accounting");

// ItemListCard 行 194-196 — canDelete チェックなし ❌
{item.source === "manual" ? (
  <DeleteIconButton onClick={() => onDeleteItem(item.id)} />
) : null}
```

`canDelete` を取得しておらず、手動追加（`source === "manual"`）の明細行には削除ボタンが `canDelete` に関係なく常に表示される。

### 修正方針

```tsx
// AccountingDetail.tsx
const { canEdit, canDelete } = usePermission("accounting");  // canDelete 追加

// ItemListCard の DeleteIconButton を canDelete でガード
{item.source === "manual" && canDelete ? (  // canDelete 追加
  <DeleteIconButton onClick={() => onDeleteItem(item.id)} />
) : null}
```

---

## 期待する挙動

`canDelete=false` の場合：
1. 予約詳細モーダルの削除ボタンが非表示
2. 会計詳細の手動明細行の削除ボタンが非表示

## 優先度

- 予約管理: **HIGH** — 予約を削除できてしまう（API は 403 で防止）
- 会計管理: **MEDIUM** — 会計確定後の手動明細削除ボタンが表示（FE-153 の支払フォーム問題と同一ページ）

## 関連ファイル

- `frontend/src/features/reservations/routes/ReservationManagement.tsx` (行 52, 289)
- `frontend/src/features/reservations/components/ReservationDetailModal.tsx` (行 208)
- `frontend/src/features/accounting/routes/AccountingDetail.tsx` (行 195, 949)
- 発見日: 2026-04-07（RBAC Phase 3 テスト中）
- 関連: FE-159（予約管理で usePermission 未参照）、FE-153（会計詳細の支払フォーム canEdit ガード漏れ）、FE-182（ReservationDetailModal の全ボタン未ガードの詳細分析）
