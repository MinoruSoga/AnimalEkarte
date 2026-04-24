# FE-052: shifts ドメイン — models.ts 型移行

**親タスク**: [TASK-012](../../docs/tasks/open/TASK-012-models-ts-type-migration-by-domain.md)
**Status**: Open
**Priority**: High
**Date Created**: 2026-03-18

## Summary

shifts feature の型定義が models.ts と完全に非接続。BackendShift, ShiftType, Shift, CreateShiftInput, UpdateShiftInput, ShiftStaff が全て手書き。

## models.ts 対応型

| 手書き型 | models.ts 対応型 |
|---------|----------------|
| `ShiftType` | `ShiftType`（models.ts に存在） |
| `BackendShift` / `Shift` | `ShiftEntry`（models.ts に存在） |
| `CreateShiftInput` | `Omit<ShiftEntry, 'id' \| ...>` で導出 |
| `UpdateShiftInput` | `Partial<CreateShiftInput>` で導出 |
| `ShiftStaff` | `Staff`（models.ts に存在）から導出 |

## 必要な変更

1. `ShiftType` を models.ts から import
2. `Shift` を models.ts の `ShiftEntry` に置き換え（transform関数を追加）
3. Request 型を models.ts から導出
4. コンポーネント内の型参照を更新

## 完了条件

- [ ] models.ts の `ShiftEntry`/`ShiftType` を使用
- [ ] 手書き interface が残っていない（UI固有型は features/shifts/types/ に配置OK）
- [ ] `pnpm build` 成功・型エラーなし
