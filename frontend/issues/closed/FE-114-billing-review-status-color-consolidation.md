# FE-114: BillingReviewSection STATUS_BADGE_CLASS を status-helpers.ts に統合

**Status**: Closed
**Priority**: Low
**Affects**: BillingReviewSection, status-helpers.ts
**Date Created**: 2026-03-25
**Related**: TASK-026

## Summary

`BillingReviewSection.tsx` にインラインで定義されている `STATUS_BADGE_CLASS` カラーマップを
`status-helpers.ts` の `getBillingReviewStatusColor()` 関数として移行する。
他のステータスカラー関数（`getAccountingStatusColor` 等）と同じパターンに統一する。

## 現状のコード

```typescript
// frontend/src/features/medical-records/components/BillingReviewSection/BillingReviewSection.tsx:28-40
import { BADGE, C } from "@/lib/design-tokens";

const STATUS_BADGE_CLASS: Record<BillingReviewStatus, string> = {
  pending: BADGE.yellow,
  confirmed: BADGE.green,
  returned: BADGE.red,
};

// 使用箇所（BillingReviewSection.tsx:87）
const badgeClass = STATUS_BADGE_CLASS[status];
// ...
<span className={`inline-flex items-center gap-1.5 px-2 py-0.5 text-xs font-medium rounded border ${badgeClass}`}>
```

```typescript
// frontend/src/utils/status-helpers.ts（末尾）
// 現在 getBillingReviewStatusColor は存在しない
// 既存のパターン（参照）:
export const getAccountingStatusColor = (status: string) => {
  switch (status) {
    case "会計待ち": return badge(N.orangeBg, N.orangeText, N.orangeBorder);
    case "会計済":   return badge(N.greenBg, N.greenText, N.greenBorder);
    case "キャンセル": return badge(N.grayBg, N.grayText, N.grayBorder);
    default: return "";
  }
};
```

## 必要な変更

### 1. status-helpers.ts に関数追加

```typescript
// frontend/src/utils/status-helpers.ts
// ファイル末尾に追記

import type { BillingReviewStatus } from "@/features/medical-records/types";

export const getBillingReviewStatusColor = (status: BillingReviewStatus) => {
  switch (status) {
    case "confirmed": return badge(N.greenBg, N.greenText, N.greenBorder);
    case "returned":  return badge(N.redBg, N.redText, N.redBorder);
    case "pending":
    default:          return badge(N.yellowBg, N.yellowText, N.yellowBorder);
  }
};
```

**注意**: `BillingReviewStatus` 型の import パスを確認すること。
`features/medical-records/types` または `features/medical-records/types/index.ts` に存在するはず。

### 2. BillingReviewSection.tsx の修正

```typescript
// frontend/src/features/medical-records/components/BillingReviewSection/BillingReviewSection.tsx

// Before - import
import { BADGE, C } from "@/lib/design-tokens";

// After - import（BADGE は不要になる、C は他で使用している場合のみ残す）
import { C } from "@/lib/design-tokens";
import { getBillingReviewStatusColor } from "@/utils/status-helpers";

// Before - 定数
const STATUS_BADGE_CLASS: Record<BillingReviewStatus, string> = {
  pending:   BADGE.yellow,
  confirmed: BADGE.green,
  returned:  BADGE.red,
};

// After - 定数削除（getBillingReviewStatusColor を直接使用）

// Before - 使用箇所（BillingReviewSection.tsx:87）
const badgeClass = STATUS_BADGE_CLASS[status];

// After
const badgeClass = getBillingReviewStatusColor(status);
```

## BillingReviewStatus 型の確認

実装前に以下を確認する:
```bash
grep -rn "BillingReviewStatus" frontend/src/features/medical-records/
```

型定義箇所を特定し、`status-helpers.ts` での import パスに使用する。

## UI 操作フロー

変更なし（BADGE.yellow/green/red と N.yellowBg/greenBg/redBg は同一 Notion カラーパレット）。

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] feature 間 import に注意（status-helpers.ts が medical-records の型を import）

## 注意: feature 間 import について

`status-helpers.ts` は `src/utils/` にある共有ユーティリティ。
`BillingReviewStatus` 型が `features/medical-records/types/` にのみ存在する場合、
`src/utils/` が `features/` を import することになり **アーキテクチャ違反** となる。

その場合は以下の代替案を採用する:
- **代替案A**: `BillingReviewStatus` を `src/types/` に移動してから import
- **代替案B**: `status-helpers.ts` への関数移行をやめ、`BillingReviewSection.tsx` 内の `STATUS_BADGE_CLASS` を `getBillingReviewStatusColor` ローカル関数にリネームするのみ（インライン定数→ローカル関数化）

実装前に型の配置を確認し、アーキテクチャ違反を避ける代替案を選択すること。

## 依存関係

なし（独立して着手可能）。

## 完了条件

- [ ] `BillingReviewSection.tsx` から `STATUS_BADGE_CLASS` インライン定数が消えている
- [ ] ステータスカラーが関数経由で取得されている
- [ ] feature 間 import の問題がない（代替案を採用した場合もアーキテクチャ準拠）
- [ ] `npm run lint` エラーなし
- [ ] `npm run build` エラーなし
