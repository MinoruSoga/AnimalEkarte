# FE-116: status-colors.ts の未使用 VisitType export を削除

**Status**: Closed
**Priority**: Low
**Affects**: frontend/src/utils/constants/status-colors.ts
**Date Created**: 2026-03-25
**Related**: TASK-027

## Summary

`status-colors.ts` に `export type VisitType = keyof typeof VISIT_TYPE_COLORS`（= `"初診" | "再診"`）が定義されているが、プロジェクト全体で誰もインポートしていない（grep 確認済み）。
一方、`src/types/index.ts:180` に `export type VisitType = "first" | "revisit"`（API 値）が定義されており、これが canonical の `VisitType`。
不要な export を削除してコードを整理する。

## 現状のコード

```typescript
// frontend/src/utils/constants/status-colors.ts:62
export type VisitType = keyof typeof VISIT_TYPE_COLORS;
// ↑ "初診" | "再診" — 誰もインポートしていない未使用 export

// frontend/src/types/index.ts:180 — canonical VisitType
export type VisitType = "first" | "revisit";
// ↑ reservations feature で使用（API から来る英語値）

// 参考: 同ファイルの getVisitTypeColor（変更不要）
export function getVisitTypeColor(visitType: string) {
  if (visitType === "初診" || visitType === "first") {
    return VISIT_TYPE_COLORS["初診"];
  }
  return VISIT_TYPE_COLORS["再診"];
}
```

```typescript
// frontend/src/features/reservations/types/index.ts:1 — @/types から import
import type { Pet, VisitType, ReservationStatus } from "@/types";
// ↑ "first" | "revisit" の canonical 型を使用（正しい）

// status-colors.ts の VisitType を import しているファイル: なし（grep 確認済み）
```

## 必要な変更

### 1. `src/utils/constants/status-colors.ts:62` — 行を削除

```typescript
// Before
export type VisitType = keyof typeof VISIT_TYPE_COLORS;

// After: この行を削除
// （ファイル内での使用もないため、型変数自体が不要）
```

削除後、`VISIT_TYPE_COLORS` の内部型が必要な場合は `keyof typeof VISIT_TYPE_COLORS` をインライン使用する。
ただし現状の実装では `getVisitTypeColor(visitType: string)` が `string` を受け取るため、型変数は使用していない。

## UI 操作フロー

UIへの影響なし（型定義の削除のみ）

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] 型は `models.ts` から導出

## 依存関係

なし（他イシューと独立）

## 完了条件

- [ ] `status-colors.ts` に `export type VisitType` が存在しない
- [ ] `pnpm build` 型エラーゼロ
- [ ] `pnpm lint` エラーゼロ
