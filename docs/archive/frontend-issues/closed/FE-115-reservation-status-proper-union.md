# FE-115: status-colors.ts の重複 ReservationStatus export を削除

**Status**: Closed
**Priority**: Low
**Affects**: frontend/src/utils/constants/status-colors.ts
**Date Created**: 2026-03-25
**Related**: TASK-027

## Summary

`status-colors.ts:21` に `export type ReservationStatus = keyof typeof RESERVATION_STATUS_COLORS` が定義されているが、
プロジェクト全体で誰もインポートしていない（grep 確認済み）。
`src/types/index.ts:179` の `type ReservationStatus = ReservationAppointment["status"]` が canonical 定義であり、
`ReservationAppointment.status` は手書きの proper union（`"confirmed" | "pending" | ...`）のため型安全性は既に担保されている。
未使用の重複 export を削除し、`status as ReservationStatus` キャストを `@/types` import に切り替える。

## 現状のコード

```typescript
// frontend/src/types/index.ts:163-179 — canonical 定義（変更不要）
export interface ReservationAppointment {
  // ...
  status: "confirmed" | "pending" | "checked_in" | "in_consultation" | "accounting" | "completed" | "cancelled";
}
export type ReservationStatus = ReservationAppointment["status"];
// ↑ すでに proper union。型安全。

// frontend/src/utils/constants/status-colors.ts:21 — 重複・未使用 export
export type ReservationStatus = keyof typeof RESERVATION_STATUS_COLORS;
// ↑ 誰もインポートしていない（grep 確認済み）

// frontend/src/utils/constants/status-colors.ts:27-31 — キャストに使用中
export function getReservationStatusColor(status: string) {
  return (
    RESERVATION_STATUS_COLORS[status as ReservationStatus] ??  // ← ローカル型を参照
    RESERVATION_STATUS_COLORS.pending
  );
}
```

## 必要な変更

### 1. `src/utils/constants/status-colors.ts:21` — 重複 export を削除

```typescript
// Before（line 21）
export type ReservationStatus = keyof typeof RESERVATION_STATUS_COLORS;

// After: この行を削除
```

### 2. `src/utils/constants/status-colors.ts` — import 追加 + キャスト維持

```typescript
// Before（importなし。line 5のみ）
import { C } from "@/lib/design-tokens";

// After
import { C } from "@/lib/design-tokens";
import type { ReservationStatus } from "@/types";

// getReservationStatusColor のキャストはそのまま維持（変更不要）
export function getReservationStatusColor(status: string) {
  return (
    RESERVATION_STATUS_COLORS[status as ReservationStatus] ??
    RESERVATION_STATUS_COLORS.pending
  );
}
```

**循環依存なし**: `@/types/index.ts` は `status-colors.ts` を import していないことを確認済み。

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] barrel index 経由 import なし
- [ ] 型は `models.ts` から導出 — `ReservationStatus` は `@/types` の手書き union から導出（適切）

## 依存関係

なし（他イシューと独立）

## 完了条件

- [ ] `src/utils/constants/status-colors.ts` に `export type ReservationStatus` が存在しない
- [ ] `status-colors.ts` に `import type { ReservationStatus } from "@/types"` が追加されている
- [ ] `pnpm build` 型エラーゼロ
- [ ] `pnpm lint` エラーゼロ
