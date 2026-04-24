# FE-111: DashboardDetailModal の STATUS_COLOR を status-colors.ts に統合

**Status**: Closed
**Priority**: Low
**Affects**: features/dashboard/components/DashboardDetailModal.tsx
**Date Created**: 2026-03-25
**Related**: TASK-025

## Summary

`DashboardDetailModal.tsx` にインライン定義された `STATUS_COLOR` を、プロジェクト共通の `utils/constants/status-colors.ts` に移動して一元管理する。機能変更なし。

## 現状のコード

```typescript
// frontend/src/features/dashboard/components/DashboardDetailModal.tsx:46-53
const STATUS_COLOR: Record<string, string> = {
  "受付予約": `${C.bgAccentLight} ${C.textAccentDark} ${C.borderAccentLight}`,
  "受付済":   `${C.bgStatusGreen} ${C.textStatusGreen} ${C.borderStatusGreen}`,
  "診療中":   `${C.bgStatusPurple} ${C.textStatusPurple} ${C.borderStatusPurple}`,
  "会計待ち": `${C.bgWarning50} ${C.textWarningIcon} ${C.borderWarning20}`,
  "会計済":   `${C.bgActive} ${C.text} ${C.borderLight}`,
};
```

```typescript
// 使用箇所（DashboardDetailModal.tsx:377）
<Badge variant="outline"
  className={`${STATUS_COLOR[currentStatus] ?? "bg-gray-100 text-gray-600 border-gray-200"} px-3 py-1 text-sm font-medium border shrink-0`}
>
```

現在の `status-colors.ts` は予約ステータス（confirmed/pending/checked_in 等の英語キー）を管理しているが、このコンポーネントは日本語ラベル（受付予約/受付済 等）をキーとする別定義。

## 必要な変更

### 1. `frontend/src/utils/constants/status-colors.ts` に追加

```typescript
// ──────────────────────────────────────────────
// ダッシュボード表示用 日本語ステータスカラーマップ
// ──────────────────────────────────────────────

import { C } from "@/lib/design-tokens";

/** DashboardDetailModal のステータスバッジカラー（日本語キー） */
export const DASHBOARD_STATUS_COLORS: Record<string, string> = {
  "受付予約": `${C.bgAccentLight} ${C.textAccentDark} ${C.borderAccentLight}`,
  "受付済":   `${C.bgStatusGreen} ${C.textStatusGreen} ${C.borderStatusGreen}`,
  "診療中":   `${C.bgStatusPurple} ${C.textStatusPurple} ${C.borderStatusPurple}`,
  "会計待ち": `${C.bgWarning50} ${C.textWarningIcon} ${C.borderWarning20}`,
  "会計済":   `${C.bgActive} ${C.text} ${C.borderLight}`,
} as const;

/** 未知ステータスのフォールバッククラス */
export const DASHBOARD_STATUS_COLOR_FALLBACK = "bg-gray-100 text-gray-600 border-gray-200";
```

### 2. `DashboardDetailModal.tsx` の変更

```typescript
// Before: インライン定義削除
// const STATUS_COLOR: Record<string, string> = { ... };  ← 削除

// import 追加
import {
  DASHBOARD_STATUS_COLORS,
  DASHBOARD_STATUS_COLOR_FALLBACK,
} from "@/utils/constants/status-colors";

// 使用箇所変更
<Badge variant="outline"
  className={`${DASHBOARD_STATUS_COLORS[currentStatus] ?? DASHBOARD_STATUS_COLOR_FALLBACK} px-3 py-1 text-sm font-medium border shrink-0`}
>
```

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] barrel index 経由 import なし（`@/utils/constants/status-colors` を直接 import）
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）

## 依存関係

- BE 変更なし
- 他 FE イシューとの依存なし（独立して実装可）

## 完了条件

- [ ] `DashboardDetailModal.tsx` に `STATUS_COLOR` のインライン定義が存在しない
- [ ] `DASHBOARD_STATUS_COLORS` が `status-colors.ts` にエクスポートされている
- [ ] ダッシュボード詳細モーダルのステータスバッジ表示が変化なし
- [ ] `pnpm build` パス
- [ ] `pnpm lint` パス（エラー 0）

## クローズ情報

- **Closed At**: 2026-03-25
- **変更ファイル**:
  - `utils/constants/status-colors.ts` — DASHBOARD_STATUS_COLORS / DASHBOARD_STATUS_COLOR_FALLBACK 追加
  - `features/dashboard/components/DashboardDetailModal.tsx` — インライン STATUS_COLOR 削除・import に置き換え
