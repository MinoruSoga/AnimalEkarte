# FE-050: hospital-settings ドメイン — models.ts 型移行（Request型導出化）

**親タスク**: [TASK-012](../../docs/tasks/open/TASK-012-models-ts-type-migration-by-domain.md)
**Status**: Open
**Priority**: Medium
**Date Created**: 2026-03-18

## Summary

Response型は models.ts（Clinic, Staff）を使用済み。UpdateClinicRequest, CreateStaffRequest, UpdateStaffRequest が手書きのため、models.ts から導出する。

## features/hospital-settings/types/index.ts（1個手書き）
| 手書き型 | models.ts 対応型 |
|---------|----------------|
| `ClinicInfo` | `Clinic`（models.ts に存在）から Pick で導出可能 |

## 必要な変更

1. `api/types.ts`: `UpdateClinicRequest` を `Partial<Omit<Clinic, 'id' | ...>>` で導出
2. `api/types.ts`: `CreateStaffRequest` を `Omit<Staff, 'id' | 'created_at' | ...>` で導出
3. `api/types.ts`: `UpdateStaffRequest` を `Partial<CreateStaffRequest>` で導出
4. `types/index.ts`: `ClinicInfo` を models.ts の `Clinic` から Pick で導出

## 完了条件

- [ ] Request 型が models.ts から導出されている
- [ ] ClinicInfo が models.ts の Clinic から導出
- [ ] `pnpm build` 成功・型エラーなし
