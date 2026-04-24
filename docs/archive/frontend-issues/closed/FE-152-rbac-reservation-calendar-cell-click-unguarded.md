# FE-152: 予約管理 — カレンダーセルクリックで canCreate ガードなく予約作成フォームが開く

## 概要

`ReservationManagement.tsx` で `canCreate` を取得して「新規予約登録」ボタンはガードされているが、カレンダービュー（WeekView・MonthView）のセルクリックによる予約作成フォームオープン（`handleTimeSlotClick`）に `canCreate` チェックがない。

## 影響範囲

- `frontend/src/features/reservations/routes/ReservationManagement.tsx:265`
- `frontend/src/features/reservations/hooks/use-reservation-management.ts:122-135`
- 権限: `can_create = false` のユーザー

## 現状の挙動（バグ）

```tsx
// use-reservation-management.ts — canCreate チェックなし
const handleTimeSlotClick = useCallback(
  (date: Date) => {
    const stub: ReservationFormData = { ... };
    handleOpenForm(stub);  // ← canCreate に関わらず実行される
  },
  [handleOpenForm]
);
```

```tsx
// ReservationManagement.tsx
<WeekView
  onTimeSlotClick={handleTimeSlotClick}  // ← 閲覧のみユーザーでも渡されている
/>
```

閲覧のみユーザーがカレンダーの時間スロットをクリックすると予約作成フォームが開く。フォーム内の「保存」ボタンは `SubmitButton` で pending 制御されているが、RBAC ガード自体はない（API 403 が返るのみ）。

## 期待する挙動

`canCreate` が false の場合、カレンダーセルクリックでフォームが開かないこと。

## 修正方針

```tsx
// ReservationManagement.tsx
<WeekView
  onTimeSlotClick={canCreate ? handleTimeSlotClick : undefined}
/>
```

WeekView・MonthView 側で `onTimeSlotClick` が undefined の場合はクリックハンドラを登録しない実装になっているか確認が必要。なっていない場合は条件分岐を追加する。

## 優先度

MEDIUM — フォームを開いて保存しようとすると API 403 になるため実害はないが、ユーザー体験が悪い

## 関連

- `frontend/src/features/reservations/routes/ReservationManagement.tsx`
- `frontend/src/features/reservations/hooks/use-reservation-management.ts`
- `frontend/src/features/reservations/components/WeekView.tsx`
- BUG-RBAC テスト 2026-04-07
