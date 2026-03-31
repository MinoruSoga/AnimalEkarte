# FE-049: auth ドメイン — models.ts 型移行

**親タスク**: [TASK-012](../../docs/tasks/open/TASK-012-models-ts-type-migration-by-domain.md)
**Status**: Open
**Priority**: High
**Date Created**: 2026-03-18

## Summary

auth feature の型定義が models.ts と完全に非接続。UserType, StaffRole, Permission, AuthUser, ClinicMembership 等8個の手書き型がバックエンドスキーマと同期されていない。

## 現状

```typescript
// features/auth/ 内
// models.ts からの import: なし
// 手書き型: UserType, StaffRole, Permission, ClinicMembership, AuthClinic, AuthUser, AuthContextValue, ProtectedRouteProps
// Zod スキーマで部分バリデーションあり
```

## models.ts 対応型

| 手書き型 | models.ts 対応型 |
|---------|----------------|
| `UserType` | `UserType`（models.ts に存在） |
| `StaffRole` | `StaffRole`（models.ts に存在） |
| `Permission` | `PermissionType`（models.ts に存在） |
| `ClinicMembership` | `UserClinicMembership`（models.ts に存在） |
| `AuthUser` | `UserAccount`（models.ts に存在） |
| `AuthClinic` | `Clinic`（models.ts に存在） |
| `AuthContextValue` | UI固有型 — 手書き許容（features/auth/types/ に配置） |
| `ProtectedRouteProps` | UI固有型 — 手書き許容 |

## 必要な変更

1. `UserType`, `StaffRole`, `PermissionType` を models.ts から re-export
2. `AuthUser` を `UserAccount` ベースで導出
3. `ClinicMembership` を `UserClinicMembership` ベースで導出
4. UI固有型（AuthContextValue, ProtectedRouteProps）は features/auth/types/ に残す

## 完了条件

- [ ] models.ts に対応がある型は全て models.ts から import
- [ ] UI固有型は features/auth/types/ に配置
- [ ] `npm run build` 成功・型エラーなし
