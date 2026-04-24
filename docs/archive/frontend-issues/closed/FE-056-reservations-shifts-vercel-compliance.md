# FE-056: 予約・シフト — memo/barrel index/useTransition 準拠

**Status**: Open
**Priority**: Medium
**Affects**: `features/reservations/`, `features/shifts/`
**Date Created**: 2026-03-18
**Related**: TASK-013

## Summary

予約・シフト feature の Vercel React Best Practices 違反を修正する。
主な修正: barrel index 除去、WeekView 内サブコンポーネントの memo() 化、MonthView の memo() 化、ShiftFormDialog の useTransition 化。

## 現状のコード

### 1. barrel index（reservations）

```typescript
// frontend/src/features/reservations/components/index.ts:1-3
export { MonthView } from "./MonthView";
export { ReservationDetailModal } from "./ReservationDetailModal";
export { WeekView } from "./WeekView";

// frontend/src/features/reservations/api/index.ts:1-28
export { getReservations, useGetReservations } from "./get-reservations";
export { getReservation, useGetReservation, ... } from "./get-reservation";
// ... 28行の barrel export
```

### 2. barrel index（shifts）

```typescript
// frontend/src/features/shifts/components/ShiftCalendar/index.ts:1-2
export { ShiftCalendar } from "./ShiftCalendar";
export type { StaffItem } from "./ShiftCalendar";
```

### 3. WeekView 内サブコンポーネントに memo() なし

```typescript
// frontend/src/features/reservations/components/WeekView.tsx
// :123-138 — TimeSidebar: memo() なし
function TimeSidebar() { ... }

// :141-341 — AppointmentCard: 200行、memo() なし
function AppointmentCard({ appointment, layoutStyle, onClick, onUpdate, dynamicColorMap }) { ... }

// :344-464 — DayColumn: 120行、memo() なし
function DayColumn({ date, appointments, onAppointmentClick, onTimeSlotClick, ... }) { ... }
```

### 4. MonthView に memo() なし

```typescript
// frontend/src/features/reservations/components/MonthView.tsx:45
export function MonthView({ currentDate, appointments, onAppointmentClick, dynamicColorMap }: MonthViewProps) {
  // useMemo for rows 内部は正しい
}
```

### 5. ShiftFormDialog — useTransition なし

```typescript
// frontend/src/features/shifts/components/ShiftFormDialog/ShiftFormDialog.tsx:79-84
const isPending =
  createShift.isPending || updateShift.isPending || deleteShift.isPending;
// ❌ mutation.isPending の OR 結合。useTransition が標準。
```

### 6. ShiftCalendarPage — lazy init パターン

```typescript
// frontend/src/features/shifts/routes/ShiftCalendarPage.tsx:15
const [yearMonth, setYearMonth] = useState<string>(getInitialYearMonth);
// 関数参照を渡しているので実質 lazy init だが、インライン化すべき
```

## 必要な変更

### 1. barrel index 除去

**reservations/components/index.ts** — 削除

**reservations/api/index.ts** — 削除

**shifts/components/ShiftCalendar/index.ts** — 削除

**全 import パスを直接ファイル import に変更**:
```typescript
// ❌ Before
import { MonthView } from "../components";
// ✅ After
import { MonthView } from "../components/MonthView";

// ❌ Before
import { ShiftCalendar } from "@/features/shifts/components/ShiftCalendar";
// ✅ After
import { ShiftCalendar } from "@/features/shifts/components/ShiftCalendar/ShiftCalendar";
```

### 2. WeekView サブコンポーネント memo() 化

```typescript
// WeekView.tsx

// TimeSidebar — 静的コンテンツなので memo() で再レンダー完全防止
const TimeSidebar = memo(function TimeSidebar() {
  return (
    <div className={...}>
      {HOURS.map((hour) => (...))}
    </div>
  );
});

// AppointmentCard — 200行、props 多数。memo() 必須
const AppointmentCard = memo(function AppointmentCard({
  appointment, layoutStyle, onClick, onUpdate, dynamicColorMap
}: AppointmentCardProps) {
  // ...
});

// DayColumn — 120行、props 多数。memo() 必須
const DayColumn = memo(function DayColumn({
  date, appointments, onAppointmentClick, onTimeSlotClick, onAppointmentUpdate, dynamicColorMap
}: DayColumnProps) {
  // ...
});
```

### 3. MonthView memo() 化

```typescript
// MonthView.tsx
export const MonthView = memo(function MonthView({
  currentDate, appointments, onAppointmentClick, dynamicColorMap
}: MonthViewProps) {
  // ...existing implementation...
});
```

### 4. ShiftFormDialog useTransition 化

```typescript
// ShiftFormDialog.tsx
const [isSavePending, startSaveTransition] = useTransition();

const handleSubmit = useCallback(async (e: React.FormEvent) => {
  e.preventDefault();
  startSaveTransition(async () => {
    const input = { ... };
    if (isEdit && editShift) {
      await updateShift.mutateAsync({ id: editShift.id, input });
    } else {
      await createShift.mutateAsync(input);
    }
    onClose();
  });
}, [form, staffId, date, isEdit, editShift, updateShift, createShift, onClose]);
```

### 5. ShiftCalendarPage lazy init インライン化

```typescript
// ShiftCalendarPage.tsx:15
// Before
const [yearMonth, setYearMonth] = useState<string>(getInitialYearMonth);

// After（関数をインライン化）
const [yearMonth, setYearMonth] = useState<string>(() => {
  const now = new Date();
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, "0")}`;
});
```

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし（index.ts 削除済み）
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `useTransition` で pending 管理
- [ ] 型は `models.ts` から導出

## 依存関係

- 依存なし（独立して着手可能）

## 完了条件

- [ ] `reservations/components/index.ts` 削除済み
- [ ] `reservations/api/index.ts` 削除済み
- [ ] `shifts/components/ShiftCalendar/index.ts` 削除済み
- [ ] 全 import が直接ファイル import に変更済み
- [ ] WeekView: TimeSidebar, AppointmentCard, DayColumn が memo() で囲まれている
- [ ] MonthView が memo() で囲まれている
- [ ] ShiftFormDialog が useTransition で pending 管理している
- [ ] `pnpm build` パス
- [ ] `pnpm lint` パス
- [ ] 予約カレンダー（月/週表示）の操作が正常動作
- [ ] シフトカレンダーの登録・編集・削除が正常動作
