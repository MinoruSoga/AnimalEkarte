# FE-045: vaccinations ドメイン — models.ts 型移行（Request型導出化）

**親タスク**: [TASK-012](../../docs/tasks/open/TASK-012-models-ts-type-migration-by-domain.md)
**Status**: Open
**Priority**: Medium
**Date Created**: 2026-03-18

## Summary

Response型は models.ts を使用済み。CreateVaccinationRequest/UpdateVaccinationRequest が手書きのため、models.ts の `Vaccination` から導出する。

## 現状

### api/types.ts（models.ts import 済み ✅、Request型手書き ❌）
```typescript
import type { Vaccination } from "@/types/generated/models";
// CreateVaccinationRequest — 手書き interface
// UpdateVaccinationRequest — 手書き interface
```

### features/vaccinations/types/index.ts
- 手書き型なし（空 or re-export のみ）

### hooks/useVaccinationForm.ts
- `VaccinationFormState` — UI固有 interface（手書き許容）

### src/types/index.ts から使用している手書き型
- `VaccinationRecord` — models.ts の `Vaccination` と対応

## 必要な変更

1. `api/types.ts`: Request 型を models.ts の `Vaccination` から Omit/Partial で導出
2. `VaccinationRecord`（src/types/index.ts 手書き）への依存があれば models.ts に置換

## 完了条件

- [ ] api/types.ts の Request 型が models.ts から導出されている
- [ ] src/types/index.ts の手書き VaccinationRecord への依存が解消
- [ ] `npm run build` 成功・型エラーなし
