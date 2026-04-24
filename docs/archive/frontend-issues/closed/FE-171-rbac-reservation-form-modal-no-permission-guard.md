# FE-171: ReservationFormModal（shared）— usePermission 完全欠落・「予約を確定」ボタン常時表示

## 概要

`frontend/src/components/shared/ReservationFormModal/` の `ReservationFormModal.tsx` は `usePermission` を一切呼び出していない。このコンポーネントは複数の機能（reception, reservations 等）から呼ばれる shared コンポーネントであるため、`canEdit=false` のユーザーでも予約の新規作成・編集ができる状態になっている。

## 根本原因

```tsx
// ReservationFormModal.tsx — usePermission なし ❌
export function ReservationFormModal({
  isOpen,
  onClose,
  onSave,
  // ...
}: ReservationFormModalProps) {
  // usePermission 呼び出しなし

  // 行 311-316 — 「予約を確定」「更新する」ボタンに権限ガードなし ❌
  <Button onClick={handleSave}>
    {isEditMode ? "更新する" : "予約を確定"}
  </Button>
}

// ReservationFormFields.tsx — usePermission なし ❌
// 行 125-294 — 全フォームフィールド（時間・スタッフ・サービス種別等）に disabled なし ❌
```

呼び出し元（Reception.tsx, ReservationManagement.tsx 等）でモーダルを開く際に権限チェックをしていても、モーダル自体には権限ガードがないため、opened state が真になれば誰でも保存操作ができる。

## 影響

`canEdit=false` / `canCreate=false` のユーザーが予約フォームモーダルを開くと：
1. 全フォームフィールド（日時・スタッフ・サービス種別・メモ等）が入力可能
2. 「予約を確定」「更新する」ボタンが表示され、クリックすると POST/PATCH `/v1/reservations` → 403
3. 患者選択（PatientSelectionTable）も権限チェックなしで操作可能

## 修正方針

### 方針 A: Props 経由で親から権限を注入（推奨）

`ReservationFormModal` は shared コンポーネントのため、`usePermission` を直接呼ぶより呼び出し元から `canEdit`/`canCreate` を受け取るパターンが適切。

```tsx
// ReservationFormModalProps に追加
interface ReservationFormModalProps {
  canEdit?: boolean;   // 既存予約の編集権限
  canCreate?: boolean; // 新規予約の作成権限
  // ...既存 Props...
}

export function ReservationFormModal({
  canEdit = false,
  canCreate = false,
  isEditMode,
  ...
}: ReservationFormModalProps) {
  const canSave = isEditMode ? canEdit : canCreate;

  return (
    <>
      <fieldset disabled={!canSave}>
        <ReservationFormFields ... />
      </fieldset>

      {canSave ? (
        <Button onClick={handleSave}>
          {isEditMode ? "更新する" : "予約を確定"}
        </Button>
      ) : null}
    </>
  );
}
```

呼び出し元では：
```tsx
// Reception.tsx
const { canCreate: canCreateReservation, canEdit: canEditReservation } = usePermission(ResourceReservations);

<ReservationFormModal
  canCreate={canCreateReservation}
  canEdit={canEditReservation}
  ...
/>
```

### 方針 B: ReservationFormModal 内で usePermission を直接呼び出す

```tsx
const { canCreate, canEdit } = usePermission(ResourceReservations);
const canSave = isEditMode ? canEdit : canCreate;
```

shared コンポーネントでリソース名をハードコードするため、方針 A を推奨。

## 優先度

**HIGH** — `ReservationFormModal` は reception・reservations 等複数ページから呼ばれる shared コンポーネント。権限チェックが一切ないため、呼び出し元でのガードに完全依存している状態。特に FE-165 で報告された Reception.tsx の onEdit 無条件渡しと組み合わさると、閲覧のみユーザーが予約を自由に編集・作成できる。

## 関連ファイル

- `frontend/src/components/shared/ReservationFormModal/ReservationFormModal.tsx` (行 311-316: 保存ボタン)
- `frontend/src/components/shared/ReservationFormModal/ReservationFormFields.tsx` (行 125-294: フォームフィールド)
- `frontend/src/components/shared/ReservationFormModal/PatientSelectionTable.tsx`
- 発見日: 2026-04-08（RBAC Phase 2/3 テスト中）
- 関連: FE-165（Reception onEdit/onCancel 権限ガード漏れ）、FE-162（ReservationManagement canDelete 漏れ）
