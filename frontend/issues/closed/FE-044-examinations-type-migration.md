# FE-044: examinations ドメイン — models.ts 型移行（Request型 + src/types 依存解消）

**親タスク**: [TASK-012](../../docs/tasks/open/TASK-012-models-ts-type-migration-by-domain.md)
**Status**: Open
**Priority**: Medium
**Date Created**: 2026-03-18

## Summary

api/types.ts は models.ts（Examination, ExaminationItem）を使用済み。Request 型が手書き + transforms が src/types/index.ts の手書き型 `ExaminationRecord`/`ExaminationItem` を使用している。

## 現状

### api/types.ts（models.ts import 済み ✅、Request型手書き ❌）
```typescript
import type { Examination, ExaminationItem } from "@/types/generated/models";
// CreateExaminationRequest — 手書き interface
// UpdateExaminationRequest — 手書き interface
```

### transforms.ts（src/types/index.ts の手書き型を使用 ❌）
```typescript
import type { ExaminationRecord, ExaminationItem } from "@/types";
// ↑ src/types/index.ts の手書き ExaminationRecord を使用
// models.ts の Examination に統一すべき
```

### update-examination.ts
```typescript
import type { ExaminationRecord } from "@/types";
// ↑ 同様に手書き型を参照
```

## 必要な変更

1. `api/types.ts`: Request 型を models.ts の `Examination` から Omit/Partial で導出
2. `transforms.ts`: `ExaminationRecord` → models.ts の `Examination` を使用するように変更
3. `update-examination.ts`: 同様に models.ts 由来の型に変更
4. transform 関数の ReturnType で UI 向け型を推論（手書き ExaminationRecord を廃止）

## 依存関係

- FE-053（src/types/index.ts 整理）で `ExaminationRecord` が削除/移行される前提

## 完了条件

- [ ] api/types.ts の Request 型が models.ts から導出されている
- [ ] transforms が models.ts の Examination を直接使用
- [ ] src/types/index.ts の手書き ExaminationRecord への依存が解消
- [ ] `npm run build` 成功・型エラーなし
