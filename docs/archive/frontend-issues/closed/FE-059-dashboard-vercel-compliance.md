# FE-059: 当日の受付 — memo/useMemo 準拠

**Status**: Open
**Priority**: Medium
**Affects**: `features/dashboard/`
**Date Created**: 2026-03-18
**Related**: TASK-013

## Summary

当日の受付（dashboard）feature の Vercel React Best Practices 違反を修正する。
主な修正: 3コンポーネントの memo() 化、column list の useMemo 追加。

## 現状のコード

### 1. memo() 未適用（3コンポーネント）

```typescript
// components/AppointmentCard.tsx:47 — 198行、50件以上の予約カードを描画
export function AppointmentCard({
  appointment, columnTitle, onCardClick, isDragOverlay = false
}: AppointmentCardProps) {
  // useCallback handlers は正しく実装済み（:77-102）
  const handleKarteClick = useCallback(...);
  const handleAccountingClick = useCallback(...);
  const handleHospitalizationClick = useCallback(...);
  // ❌ しかしコンポーネント自体が memo() で囲まれていない
}

// components/KanbanColumn.tsx:22 — 75行、filteredColumns.map() で描画
export function KanbanColumn({
  data, onAddClick, onCardClick
}: KanbanColumnProps) {
  // ❌ memo() なし。親の Dashboard state 変更で全カラムが再レンダー
}

// components/DashboardDetailModal.tsx:302 — 474行、10+ callbacks
export function DashboardDetailModal({
  isOpen, onClose, appointment, onConfirm, onEdit, onCancel, currentStatus
}: DashboardDetailModalProps) {
  // useCallback + primitive 抽出は正しく実装済み（:314-350）
  const petId = appointment?.petId;
  const appointmentId = appointment?.id;
  const ownerId = appointment?.ownerId;
  // ❌ しかしコンポーネント自体が memo() で囲まれていない
}
```

### 2. column list の useMemo なし

```typescript
// routes/Dashboard.tsx:345-352
<div className="grid grid-cols-2 md:grid-cols-3 lg:flex gap-4 h-full w-full ...">
  {filteredColumns.map((column) => (
    <KanbanColumn
      key={column.title}
      data={column}
      onAddClick={addClickHandlers.get(column.title)}
      onCardClick={handleCardClick}
    />
  ))}
</div>
// ❌ filteredColumns.map() が useMemo で囲まれていない
// Dashboard の state 変更（modal open 等）で全カラム JSX が再生成される
```

**注**: 以下は既に正しく実装済み（変更不要）:
- `useMemo` for `doctors`（:54-58）✅
- `useMemo` for `addClickHandlers`（:175-184）✅
- `lazy()` + `Suspense` for DashboardDetailModal, ReservationFormModal ✅
- hoisted `NO_ADD_BUTTON_COLUMNS` ✅
- hoisted CSS constants in DashboardDetailModal ✅
- conditional rendering with ternary ✅

## 必要な変更

### 1. AppointmentCard memo() 化

```typescript
// components/AppointmentCard.tsx
export const AppointmentCard = memo(function AppointmentCard({
  appointment, columnTitle, onCardClick, isDragOverlay = false
}: AppointmentCardProps) {
  // ...existing useCallback handlers (already correct)...
  // ...existing JSX...
});
```

### 2. KanbanColumn memo() 化

```typescript
// components/KanbanColumn.tsx
export const KanbanColumn = memo(function KanbanColumn({
  data, onAddClick, onCardClick
}: KanbanColumnProps) {
  // ...existing implementation...
});
```

### 3. DashboardDetailModal memo() 化

```typescript
// components/DashboardDetailModal.tsx
export const DashboardDetailModal = memo(function DashboardDetailModal({
  isOpen, onClose, appointment, onConfirm, onEdit, onCancel, currentStatus
}: DashboardDetailModalProps) {
  // ...existing useCallback handlers with primitive extraction (already correct)...
  // ...existing JSX...
});
```

### 4. column list useMemo 追加

```typescript
// routes/Dashboard.tsx
const columnElements = useMemo(() =>
  filteredColumns.map((column) => (
    <KanbanColumn
      key={column.title}
      data={column}
      onAddClick={addClickHandlers.get(column.title)}
      onCardClick={handleCardClick}
    />
  )),
  [filteredColumns, addClickHandlers, handleCardClick]
);

// JSX 内
<div className="grid grid-cols-2 md:grid-cols-3 lg:flex gap-4 h-full w-full ...">
  {columnElements}
</div>
```

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `useTransition` で pending 管理
- [ ] 型は `models.ts` から導出

## 依存関係

- 依存なし（独立して着手可能）

## 完了条件

- [ ] AppointmentCard が memo() で囲まれている
- [ ] KanbanColumn が memo() で囲まれている
- [ ] DashboardDetailModal が memo() で囲まれている
- [ ] Dashboard.tsx の filteredColumns.map() が useMemo でキャッシュされている
- [ ] `pnpm build` パス
- [ ] `pnpm lint` パス
- [ ] 当日の受付画面のカンバン操作（ドラッグ&ドロップ、ステータス変更、詳細モーダル）が正常動作
