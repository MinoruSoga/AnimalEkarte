# FE-159: 複数フィーチャーページで行クリックが canEdit 無関係に編集 UI を開く（システム的欠如）

## 概要

RBAC Phase 2 テスト（canEdit=false）で、以下の 7 つのフィーチャーページにて、閲覧のみユーザーが行クリック・カードクリックで編集フォームを開ける、または編集用ボタンが無条件に表示されることが判明した。

## 影響範囲

| ページ | 問題の操作 | 深刻度 |
|-------|-----------|--------|
| `/owners` | 行クリックで `OwnerForm` に遷移（`OwnersList.tsx` 行 323: `onClick={() => handleRowClick(pet)}` canEdit チェックなし） | HIGH |
| `/examinations` | `ExaminationsList.tsx` 行 160: `<DataTableRow onClick={() => handleEdit(r.id)}>` — canEdit チェックなし | HIGH |
| `/hospitalization` | 行クリックで入院詳細が開く。DailyRecordsTab「この日の記録を作成」ボタン未ガード | HIGH |
| `/inventory` | 行クリックで在庫編集フォームが開く（全フィールド入力可） | HIGH |
| `/trimming` | 行クリックでトリミング編集フォームが開く（全フィールド入力可） | HIGH |
| `/vaccinations` | `DataTableRow onClick={() => handleEdit(r.id)}` — canEdit チェックなし | HIGH |
| `/accounting` | 行クリックで会計詳細が開く。支払方法ボタン・金額入力が操作可能（FE-153 参照） | HIGH |
| `/medical-records` | 行クリックでカルテ編集画面が開く。全サブタブで CRUD 操作可能（FE-155/156 参照） | HIGH |
| `/reservations` | イベントクリックで予約詳細モーダルに「編集」ボタン・ステータス変更 UI が表示される。`usePermission` が reservations feature で参照されていない | HIGH |

注: `/shifts` は `ShiftCell` 内で `canEdit` チェック済み ✅

## 各ページの根本原因

### `/hospitalization`（HospitalizationList / HospitalizationDetail）

```tsx
// HospitalizationList.tsx — 行クリックに canEdit チェックなし（推定）
<DataTableRow onClick={() => navigate(`/hospitalization/${item.id}`)}>
// → 閲覧のみユーザーでも詳細ページに遷移できる

// DailyRecordsTab.tsx — 「この日の記録を作成」ボタン
<Button onClick={handleCreateRecord}>この日の記録を作成</Button>
// → canEdit チェックなし
```

### `/inventory`（InventoryList）

```tsx
// InventoryList.tsx — 行クリックで編集フォームが開く
<DataTableRow onClick={() => onEdit(item)}>
// → canEdit チェックなし
```

### `/trimming`（TrimmingList）

```tsx
// TrimmingList.tsx — 行クリックで編集フォームが開く
<DataTableRow onClick={() => onEditTargetChange(item)}>
// → canEdit チェックなし（前セッションで usePermission のガードは追加済みだが
//    行クリック自体のガードが漏れている可能性あり）
```

### `/vaccinations`（VaccinationList）

```tsx
// VaccinationList.tsx — canEdit チェックなし
<DataTableRow onClick={() => handleEdit(r.id)}>
// → 全ユーザーが行クリックで編集フォームを開ける
```

### `/reservations`（ReservationCalendar / ReservationList）

```tsx
// reservations feature 全体で usePermission("reservations") が参照されていない
// イベントクリック時の詳細モーダルに「編集」ボタンが常時表示
// ステータス変更ドロップダウンも常時操作可能
```

## 共通パターン

これらのページはすべて以下のパターンに該当する：

1. `usePermission(resource)` は呼び出しているが `canEdit` を行クリックの条件に使っていない
2. または `usePermission` 自体を呼び出していない（reservations）

正しい実装例（`MedicineSettings.tsx:133`）:
```tsx
<SortableDataTableRow
  id={medicine.id}
  onClick={canEdit ? () => onEdit(medicine) : undefined}
>
```

## 期待する挙動

`canEdit=false` の場合：
1. 行クリックで編集フォームが開かない（またはクリック不可）
2. 「編集」ボタンが非表示
3. ステータス変更 UI が disabled または非表示
4. 「この日の記録を作成」等の CRUD ボタンが非表示

## 修正方針

### 各ページの DataTableRow / カードコンポーネント

```tsx
// ① usePermission を取得
const { canEdit } = usePermission("resource-name");

// ② onClick を canEdit で条件付け
<DataTableRow
  onClick={canEdit ? () => onEdit(item) : undefined}
>
```

### reservations feature

```tsx
// ReservationCalendar.tsx または ReservationModal.tsx に追加
const { canEdit } = usePermission("reservations");

// 「編集」ボタンをガード
{canEdit ? <Button onClick={handleEdit}>編集</Button> : null}

// ステータス変更 UI をガード
{canEdit ? <StatusChangeDropdown ... /> : <StatusBadge status={status} />}
```

### hospitalization DailyRecordsTab

```tsx
// DailyRecordsTab.tsx
const { canEdit } = usePermission("hospitalization");

{canEdit ? (
  <Button onClick={handleCreateRecord}>この日の記録を作成</Button>
) : null}
```

## 優先度

**HIGH** — 閲覧のみユーザーが 7 ページにわたって編集 UI を操作でき、403 エラーが大量発生する可能性がある。特に `/reservations` は `usePermission` 未使用のため、RBAC が完全に機能していない。

## 関連ファイル

- `frontend/src/features/owners/routes/OwnersList.tsx` (行 323: `onClick={() => handleRowClick(pet)}` canEdit チェックなし)
- `frontend/src/features/examinations/routes/ExaminationsList.tsx` (行 160: `DataTableRow onClick` canEdit チェックなし)
- `frontend/src/features/hospitalization/routes/HospitalizationList.tsx`
- `frontend/src/features/hospitalization/components/DailyRecordsTab/DailyRecordsTab.tsx`
- `frontend/src/features/inventory/routes/InventoryList.tsx`
- `frontend/src/features/trimming/routes/TrimmingList.tsx`
- `frontend/src/features/vaccinations/routes/VaccinationList.tsx`
- `frontend/src/features/accounting/routes/AccountingList.tsx`
- `frontend/src/features/medical-records/routes/MedicalRecordList.tsx`
- `frontend/src/features/reservations/routes/ReservationCalendar.tsx` (または類似)
- `frontend/src/features/estimates/routes/EstimateList.tsx` (行 185: `onClick={() => navigate(paths.estimates.detail...)}` canEdit/canDelete チェックなし)
- 発見日: 2026-04-07（RBAC Phase 2 テスト中）
- 関連: FE-153（会計支払フォーム）, FE-155（カルテ担当医ボタン）, FE-156（カルテサブタブ）, FE-163（EstimateDetail/Form の usePermission 欠落）
