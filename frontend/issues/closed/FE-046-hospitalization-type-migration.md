# FE-046: hospitalization ドメイン — models.ts 型移行（Request型 + types/index.ts 導出化）

**親タスク**: [TASK-012](../../docs/tasks/open/TASK-012-models-ts-type-migration-by-domain.md)
**Status**: Open
**Priority**: Medium
**Date Created**: 2026-03-18

## Summary

api/types.ts は models.ts（Hospitalization, CarePlanItem, VitalRecord 等）を使用済み。Request 型が手書き + types/index.ts に10個の手書き型が残存。

## 現状

### api/types.ts（models.ts import 済み ✅、Request型手書き ❌）
```typescript
import type { Hospitalization, CarePlanItem, VitalRecord, CareLogRecord, StaffNoteRecord, DailyRecord } from "@/types/generated/models";
// ↑ Response 型 OK
// CreateHospitalizationRequest — 手書き interface
// UpdateHospitalizationRequest — 手書き interface
```

### features/hospitalization/types/index.ts（10個手書き）
| 手書き型 | models.ts 対応型 |
|---------|----------------|
| `CreateCarePlanDTO` | `Omit<CarePlanItem, ...>` で導出 — models.ts 由来 ✅ |
| `UpdateCarePlanDTO` | `Partial<CarePlanItem>` — models.ts 由来 ✅ |
| `CreateVitalDTO` | models.ts から導出すべき |
| `CreateCareLogDTO` | models.ts から導出すべき |
| `CreateHospitalizationDTO` | `Omit<Hospitalization, ...>` — models.ts 由来 ✅ |
| `UpdateHospitalizationDTO` | `Partial<Hospitalization>` — models.ts 由来 ✅ |
| `Task` | UI固有型 — 手書き許容 |
| `TimelineItem` | UI固有型 — 手書き許容 |
| `HospitalizationFormData` | UI固有型 — 手書き許容 |

## 必要な変更

1. `api/types.ts`: Request 型を models.ts から Omit/Partial で導出
2. `types/index.ts`: DTO 型が models.ts 由来の型を正しく参照しているか確認・修正
3. UI固有型（Task, TimelineItem, HospitalizationFormData）はそのまま残す

## 完了条件

- [ ] api/types.ts の Request 型が models.ts から導出されている
- [ ] types/index.ts の DTO 型が models.ts の型を正しく参照
- [ ] `npm run build` 成功・型エラーなし
