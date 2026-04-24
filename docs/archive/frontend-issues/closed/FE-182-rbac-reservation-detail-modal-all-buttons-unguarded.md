# FE-182: 予約詳細モーダル — 編集・削除・ステータス変更・カルテ作成ボタン権限ガード完全欠落（ReservationDetailModal）

## 概要

`ReservationDetailModal`（予約詳細モーダル）に `usePermission` が一切実装されていない。モーダル内のすべてのアクションボタン（編集・削除・ステータス変更・カルテ作成等）が権限チェックなしで表示・実行できる。

`ReservationManagement.tsx`（親）は `canCreate` のみ取得しているが、`canEdit`/`canDelete` を子モーダルに渡していない。

## 影響範囲

| ファイル | 問題 UI | API 呼び出し | 深刻度 |
|---------|---------|------------|--------|
| `ReservationDetailModal.tsx` | 編集ボタン（行 148）・削除/取消ボタン（行 142）・ステータス変更（行 155/180/194）・カルテ作成（行 169）・会計へ進む（行 225）| PUT/DELETE `/v1/reservations`, POST `/v1/medical-records` 等 | HIGH |

## 根本原因

```tsx
// ReservationDetailModal.tsx — usePermission なし ❌
// ActionButtons function — 全ボタンにガードなし

// 行 142-145: 取消ボタン — canDelete チェックなし ❌
<Button onClick={onCancelReservation}>取消</Button>

// 行 148-151: 編集ボタン — canEdit チェックなし ❌
<Button onClick={onEdit}>編集</Button>

// 行 155-157: 受付済にするボタン — canEdit チェックなし ❌
<Button onClick={onMarkAsReceived}>受付済にする</Button>

// 行 169-175: カルテ作成ボタン — canCreate チェックなし ❌
<Button onClick={onCreateRecord}>カルテ作成</Button>

// 行 180-182: 診察を開始するボタン — canEdit チェックなし ❌
// 行 194-196: 診察を終了するボタン — canEdit チェックなし ❌
// 行 225-229: 会計へ進むボタン — canEdit チェックなし ❌
// 行 240-242: 完了/リストから削除ボタン — canDelete チェックなし ❌
```

```tsx
// ReservationManagement.tsx — canCreate のみ取得 ❌
const { canCreate } = usePermission("reservations");
// canEdit, canDelete を取得せず、モーダルに渡していない ❌
```

## 修正方針

```tsx
// ReservationManagement.tsx
const { canCreate, canEdit, canDelete } = usePermission("reservations");

<ReservationDetailModal
  canEdit={canEdit}
  canDelete={canDelete}
/>
```

```tsx
// ReservationDetailModal.tsx
interface ReservationDetailModalProps {
  canEdit?: boolean;
  canDelete?: boolean;
}

// 取消ボタン
{canDelete ? <Button onClick={onCancelReservation}>取消</Button> : null}

// 編集ボタン
{canEdit ? <Button onClick={onEdit}>編集</Button> : null}

// ステータス変更ボタン群
{canEdit ? <Button onClick={onMarkAsReceived}>受付済にする</Button> : null}
```

## 優先度

**HIGH** — 予約管理は診療フローの入口。`canEdit=false` のスタッフがステータスを誤変更したり、`canDelete=false` のスタッフが予約を取消できてしまう。

## 関連ファイル

- `frontend/src/features/reservations/components/ReservationDetailModal.tsx` (行 138-276: ActionButtons)
- `frontend/src/features/reservations/routes/ReservationManagement.tsx` (行 52: `const { canCreate } = usePermission("reservations")`, canEdit/canDelete 未取得)
- 発見日: 2026-04-08（RBAC Phase 2/3 テスト中）
- 関連: FE-159（reservations feature の usePermission 未使用として報告済み — 本チケットが詳細）
