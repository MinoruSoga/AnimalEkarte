# TASK-013: DailyVitalsSection に memo() を追加

## 概要

`DailyVitalsSection.tsx`（257行）が `memo()` でラップされていない。隣接する `DailyCareLogsSection`（`memo()` 適用済み）と非対称であり、親コンポーネント `DailyRecordsTab`（memo済み）内での state 更新（`selectedDate` 等）のたびに毎回再レンダリングされる。

## 優先度

HIGH

## 影響ファイル

| ファイル | 行 |
|---------|-----|
| `frontend/src/features/hospitalization/components/DailyRecordsTab/DailyVitalsSection.tsx` | L48 |

## 規約違反

`.claude/CLAUDE.md`:
> `DataTable`, `NotionFilter`, `Pagination`, `SidePeekPanel` は `memo()` 適用済み。新規共有コンポーネントも同様に適用すること。

`.claude/rules/code-style.md`:
> 独立した大きいセクションは `memo()` で囲む。必ず props ハンドラを `useCallback` で安定化すること。

## 修正方針

```typescript
// Before
export function DailyVitalsSection({ ... }: Props) {
  return ...;
}

// After
import { memo } from "react";

export const DailyVitalsSection = memo(function DailyVitalsSection({ ... }: Props) {
  return ...;
});
```

合わせて、親コンポーネント（`DailyRecordsTab`）から `DailyVitalsSection` に渡しているハンドラが `useCallback` で安定化されているか確認すること。

## テスト

- `DailyRecordsTab` の `selectedDate` が変化したとき、`DailyVitalsSection` が再レンダリングされないことを React DevTools Profiler で確認
