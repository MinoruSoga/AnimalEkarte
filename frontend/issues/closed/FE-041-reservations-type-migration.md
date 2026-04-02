# FE-041: reservations ドメイン — models.ts 型移行

**親タスク**: [TASK-012](../../docs/tasks/open/TASK-012-models-ts-type-migration-by-domain.md)
**Status**: Open
**Priority**: Medium
**Date Created**: 2026-03-18

## Summary

transforms/create/update で models.ts の `ReservationAppointment` を import 済み。api/types.ts の Request 型 + types/index.ts の手書き型を models.ts から導出する。

## 現状

### api/types.ts（Request型手書き）
```typescript
// CreateReservationRequest — 手書き interface
// UpdateReservationRequest — 手書き interface
```

### transforms, create/update（models.ts import 済み ✅）
```typescript
import type { ReservationAppointment as BackendReservation } from "@/types/generated/models";
```

### features/reservations/types/index.ts（3個手書き）
- `ReservationFormData` — UI固有型（手書き許容だが models.ts の型を参照すべき）
- `ReservationFormSaveHandler` — UI固有型
- 各種 re-export（`@/types` index.ts の手書き型を re-export）

### src/types/index.ts から使用している手書き型
- `ReservationAppointment`（手書き）— models.ts に同名型が存在
- `ReservationStatus`（手書き）— models.ts に `ReservationStatus` が存在

## 必要な変更

1. `api/types.ts`: Request 型を models.ts の `ReservationAppointment` から `Omit`/`Partial` で導出
2. `types/index.ts`: `@/types` の手書き ReservationAppointment を models.ts import に置換
3. ReservationFormData 内で models.ts の型を参照する形に修正

## 完了条件

- [ ] api/types.ts の Request 型が models.ts から導出されている
- [ ] types/index.ts が src/types/index.ts の手書き型ではなく models.ts を参照
- [ ] `npm run build` 成功・型エラーなし
