# FE-040: estimates ドメイン — models.ts 型移行

**親タスク**: [TASK-012](../../docs/tasks/open/TASK-012-models-ts-type-migration-by-domain.md)
**Status**: Open
**Priority**: High
**Date Created**: 2026-03-18

## Summary

estimates feature の型定義が models.ts と完全に非接続。BackendEstimate/BackendEstimateItem を手書きしている。models.ts の `Estimate`/`EstimateItem` を使用するように移行する。

## 現状のコード

```typescript
// frontend/src/features/estimates/api/types.ts
// models.ts からの import なし
// BackendEstimate, BackendEstimateItem を手書きで定義
```

## 必要な変更

1. `features/estimates/api/types.ts` — 手書き型を削除し、models.ts の `Estimate`/`EstimateItem` を import
2. Request 型を models.ts から `Omit`/`Partial`/`Pick` で導出
3. transform 関数があれば入出力型を models.ts 由来に統一
4. コンポーネント・hooks 内の型参照を更新

## 型マッピング

### api/types.ts
| 手書き型 | models.ts 対応型 |
|---------|----------------|
| `BackendEstimate` | `Estimate` |
| `BackendEstimateItem` | `EstimateItem` |
| `BackendOwnerSummary` | `Owner` から Pick で導出 |
| `EstimateListResponse` | API固有型（手書き許容） |
| `CreateEstimateRequest` | `Omit<Estimate, 'id' \| 'created_at' \| ...>` で導出 |
| `UpdateEstimateRequest` | `Partial<CreateEstimateRequest>` で導出 |

### features/estimates/types/index.ts
| 手書き型 | models.ts 対応型 |
|---------|----------------|
| `EstimateStatus` | `EstimateStatus`（models.ts に存在） |
| `Estimate` | UI変換後の型 — transform + ReturnType パターンに移行 |
| `EstimateLineItem` | `EstimateItem`（models.ts に存在） |

## 完了条件

- [ ] `features/estimates/` 内で models.ts の型を import して使用
- [ ] 手書き interface が残っていない（Request型は Omit/Partial 導出OK）
- [ ] `npm run build` 成功・型エラーなし
