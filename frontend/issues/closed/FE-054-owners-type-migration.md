# FE-054: owners ドメイン — types/index.ts + src/types/owner.ts 手書き型の移行

**親タスク**: [TASK-012](../../docs/tasks/open/TASK-012-models-ts-type-migration-by-domain.md)
**Status**: Open
**Priority**: Medium
**Date Created**: 2026-03-18

## Summary

owners の api/types.ts と transforms は models.ts 移行済みだが、`features/owners/types/index.ts` に手書き型が5個、`src/types/owner.ts` が models.ts 非依存のまま残存している。

## 手書き型一覧

### features/owners/types/index.ts
| 手書き型 | models.ts 対応型 |
|---------|----------------|
| `PetGender` | `PetGender`（models.ts に存在） |
| `AcquisitionType` | `AcquisitionType`（models.ts に存在） |
| `DangerLevel` | `DangerLevel`（models.ts に存在） |
| `MembershipType` | `MembershipType`（models.ts に存在） |
| `PetFormData` | UI固有型 — 手書き許容 |
| `OwnerData` | UI固有型 — 手書き許容 |

### src/types/owner.ts
| 手書き型 | models.ts 対応型 |
|---------|----------------|
| `Owner` | `Owner`（models.ts に存在） |
| `CreateOwnerRequest` | `Omit<Owner, 'id' \| ...>` で導出すべき |
| `UpdateOwnerRequest` | `Partial<CreateOwnerRequest>` で導出すべき |

## 必要な変更

1. `features/owners/types/index.ts` — PetGender/AcquisitionType/DangerLevel/MembershipType を models.ts から re-export
2. `src/types/owner.ts` — Owner を models.ts から import、Request 型を Omit/Partial で導出
3. PetFormData/OwnerData は UI固有型として残す（ただし models.ts の型を参照する形に）

## 完了条件

- [ ] models.ts に対応がある enum 型（PetGender 等）は models.ts から import
- [ ] src/types/owner.ts が models.ts の Owner を使用
- [ ] `npm run build` 成功・型エラーなし
