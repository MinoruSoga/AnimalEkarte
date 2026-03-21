# FE-042: dashboard ドメイン — models.ts 型移行

**親タスク**: [TASK-012](../../docs/tasks/open/TASK-012-models-ts-type-migration-by-domain.md)
**Status**: Open
**Priority**: Medium
**Date Created**: 2026-03-18

## Summary

transforms/get-dashboard で models.ts の `ReservationAppointment` を import 済み。api/types.ts に DashboardAppointment/DashboardColumn/UpdateAppointmentStatusRequest 3個の手書き型が残存。

## 現状

### api/types.ts（3個手書き）
```typescript
import type { ReservationStatus } from "@/types";  // ← src/types/index.ts の手書き型を使用

export interface DashboardAppointment { ... }  // 手書き
export interface DashboardColumn { ... }       // 手書き（UI固有）
export interface UpdateAppointmentStatusRequest { ... }  // 手書き
```

### transforms.ts, get-dashboard.ts（models.ts import 済み ✅）
```typescript
import type { ReservationAppointment as BackendDashboardReservation } from "@/types/generated/models";
```

### src/types/index.ts から使用している手書き型
- `ReservationStatus` — models.ts に `ReservationStatus` 存在
- `Appointment` — 手書き
- `ReservationAppointment` — models.ts に同名型存在

## 必要な変更

1. `api/types.ts`: `DashboardAppointment` を `ReservationAppointment & { /* UI拡張 */ }` で導出
2. `api/types.ts`: `UpdateAppointmentStatusRequest` を models.ts から導出
3. `api/types.ts`: `ReservationStatus` の import 元を `@/types` → `@/types/generated/models` に変更
4. `DashboardColumn` は UI固有型のため手書き許容

## 完了条件

- [ ] DashboardAppointment が models.ts の `ReservationAppointment` をベース型として使用
- [ ] ReservationStatus が models.ts から直接 import
- [ ] `npm run build` 成功・型エラーなし
