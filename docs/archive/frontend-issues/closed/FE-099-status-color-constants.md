# FE-099: ステータスカラー定数集約

**Status**: Closed
**Priority**: Medium
**Affects**: reservations/components/, dashboard/components/
**Date Created**: 2026-03-24
**Related**: TASK-024

## Summary

予約ステータス（confirmed/pending/checked_in 等 7種）のカラー定義が `ReservationDetailModal.tsx` に直接ハードコードされており、`AppointmentCard.tsx` では visitType（初診/再診）のカラーが三項演算子でインライン記述されている。これらを共有定数ファイルに集約する。

## 現状のコード

```typescript
// frontend/src/features/reservations/components/ReservationDetailModal.tsx:46-54
// STATUS_OPTIONS が ReservationDetailModal 内にハードコード
const STATUS_OPTIONS: StatusOption[] = [
  { value: "confirmed",       label: "予約確定", dotColor: "bg-emerald-500", bgColor: "bg-emerald-50",  textColor: "text-emerald-700" },
  { value: "pending",         label: "仮予約",   dotColor: "bg-sky-500",     bgColor: "bg-sky-50",      textColor: "text-sky-700" },
  { value: "checked_in",      label: "受付済",   dotColor: "bg-blue-500",    bgColor: "bg-blue-50",     textColor: "text-blue-700" },
  { value: "in_consultation", label: "診療中",   dotColor: "bg-violet-500",  bgColor: "bg-violet-50",   textColor: "text-violet-700" },
  { value: "accounting",      label: "会計待ち", dotColor: "bg-amber-500",   bgColor: "bg-amber-50",    textColor: "text-amber-700" },
  { value: "completed",       label: "完了",     dotColor: "bg-gray-400",    bgColor: "bg-gray-50",     textColor: "text-gray-600" },
  { value: "cancelled",       label: "キャンセル", dotColor: "bg-red-500",   bgColor: "bg-red-50",      textColor: "text-red-700" },
];
```

```typescript
// frontend/src/features/reservations/components/ReservationDetailModal.tsx:72-77
// visitType アクセントカラーもインラインに直書き
function getVisitTypeAccent(visitType: string) {
  return visitType === "初診"
    ? { border: "border-red-200",  bg: "bg-red-50",  text: "text-red-700",  dot: "bg-red-500" }
    : { border: "border-blue-200", bg: "bg-blue-50", text: "text-blue-700", dot: "bg-blue-500" };
}
```

```typescript
// frontend/src/features/dashboard/components/AppointmentCard.tsx:141-146
// 初診/再診バッジのカラーもインライン三項演算子
className={`text-sm px-[7.5px] h-[22px] ...
  ${appointment.visitType === "初診"
    ? `bg-[#D3E5EF]/60 text-[#183B56]/90 border-[#B8D4E3]/50`
    : `bg-[#F7F6F3]/60 text-[#37352F]/90 border-[rgba(55,53,47,0.09)]/50`}`}
```

## 必要な変更

### 1. 共有定数ファイル作成

```typescript
// frontend/src/utils/constants/status-colors.ts（新規作成）

// 予約ステータスカラーマップ
export const RESERVATION_STATUS_COLORS = {
  confirmed:       { label: "予約確定", dot: "bg-emerald-500", bg: "bg-emerald-50",  text: "text-emerald-700" },
  pending:         { label: "仮予約",   dot: "bg-sky-500",     bg: "bg-sky-50",      text: "text-sky-700" },
  checked_in:      { label: "受付済",   dot: "bg-blue-500",    bg: "bg-blue-50",     text: "text-blue-700" },
  in_consultation: { label: "診療中",   dot: "bg-violet-500",  bg: "bg-violet-50",   text: "text-violet-700" },
  accounting:      { label: "会計待ち", dot: "bg-amber-500",   bg: "bg-amber-50",    text: "text-amber-700" },
  completed:       { label: "完了",     dot: "bg-gray-400",    bg: "bg-gray-50",     text: "text-gray-600" },
  cancelled:       { label: "キャンセル", dot: "bg-red-500",   bg: "bg-red-50",      text: "text-red-700" },
} as const;

export type ReservationStatus = keyof typeof RESERVATION_STATUS_COLORS;

// visitType（初診/再診）カラーマップ
export const VISIT_TYPE_COLORS = {
  初診: {
    border: "border-red-200",  bg: "bg-red-50",  text: "text-red-700",  dot: "bg-red-500",
    badgeBg: "bg-[#D3E5EF]/60", badgeText: "text-[#183B56]/90", badgeBorder: "border-[#B8D4E3]/50",
  },
  再診: {
    border: "border-blue-200", bg: "bg-blue-50", text: "text-blue-700", dot: "bg-blue-500",
    badgeBg: "bg-[#F7F6F3]/60", badgeText: "text-[#37352F]/90", badgeBorder: "border-[rgba(55,53,47,0.09)]/50",
  },
} as const;

export type VisitType = keyof typeof VISIT_TYPE_COLORS;

// ユーティリティ関数
export function getReservationStatusColor(status: string) {
  return RESERVATION_STATUS_COLORS[status as ReservationStatus]
    ?? RESERVATION_STATUS_COLORS.pending; // fallback
}

export function getVisitTypeColor(visitType: string) {
  return VISIT_TYPE_COLORS[visitType as VisitType]
    ?? VISIT_TYPE_COLORS["再診"]; // fallback
}
```

### 2. ReservationDetailModal.tsx の変更

```typescript
// frontend/src/features/reservations/components/ReservationDetailModal.tsx
// Before: ファイル内 STATUS_OPTIONS 定数 + getVisitTypeAccent 関数を削除
// After: 共有定数から import

import {
  RESERVATION_STATUS_COLORS,
  getReservationStatusColor,
  getVisitTypeColor,
} from "@/utils/constants/status-colors";

// STATUS_OPTIONS → RESERVATION_STATUS_COLORS に置き換え
// getVisitTypeAccent → getVisitTypeColor に置き換え
```

### 3. AppointmentCard.tsx の変更

```typescript
// frontend/src/features/dashboard/components/AppointmentCard.tsx:141-146
// Before: インライン三項演算子
// After: 共有定数使用

import { getVisitTypeColor } from "@/utils/constants/status-colors";

// ...
const visitColor = getVisitTypeColor(appointment.visitType);
// className に visitColor.badgeBg, visitColor.badgeText, visitColor.badgeBorder を使用
```

## プロジェクトルール遵守チェック

- [ ] `any` 型なし（`as const` + keyof で型安全）
- [ ] barrel index 経由 import なし（`utils/constants/status-colors` を直接 import）
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `FC` / `forwardRef` なし

## 依存関係

- Backend 変更なし。FE-098 とも独立。単独で実施可能。

## 完了条件

- [ ] `frontend/src/utils/constants/status-colors.ts` が作成されている
- [ ] `ReservationDetailModal.tsx` のインライン `STATUS_OPTIONS` 定数が削除されている
- [ ] `ReservationDetailModal.tsx` のインライン `getVisitTypeAccent` 関数が削除されている
- [ ] `AppointmentCard.tsx` のインライン三項演算子カラー指定が `getVisitTypeColor` に置き換えられている
- [ ] 予約カレンダー・ダッシュボードの各ステータス・バッジ色が変更前と同一
- [ ] `pnpm lint` エラーなし
- [ ] `pnpm build` エラーなし

## クローズ情報

- **Closed At**: 2026-03-24
- **変更ファイル**:
  - `frontend/src/utils/constants/status-colors.ts` — 新規作成（RESERVATION_STATUS_COLORS / VISIT_TYPE_COLORS / ユーティリティ関数）
  - `frontend/src/features/reservations/components/ReservationDetailModal.tsx` — STATUS_OPTIONS・getVisitTypeAccent を削除し共有定数に移行
  - `frontend/src/features/dashboard/components/AppointmentCard.tsx` — visitType インライン三項演算子を getVisitTypeColor に移行
