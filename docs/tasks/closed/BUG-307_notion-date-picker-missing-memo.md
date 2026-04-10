# BUG-307: NotionDatePicker — memo() ラッピング欠落

## 概要

`frontend/src/components/shared/NotionDatePicker/NotionDatePicker.tsx` は615行の共有コンポーネントだが、`memo()` でラッピングされていない。`DataTable`, `NotionFilter`, `Pagination` など他の共有コンポーネントはすべて `memo()` 適用済み。

## 影響ファイル

| ファイル | 違反箇所 |
|---------|---------|
| `frontend/src/components/shared/NotionDatePicker/NotionDatePicker.tsx` | line 166 |

## 違反箇所と修正

```tsx
// Before (line 1)
import { useState, useCallback, useMemo } from "react";

// After
import { useState, useCallback, useMemo, memo } from "react";

// Before (line 166)
export function NotionDatePicker(props: NotionDatePickerProps) {
  ...
}

// After
export const NotionDatePicker = memo(function NotionDatePicker(props: NotionDatePickerProps) {
  ...
});
```

## 適用ルール

- `shared/` コンポーネントは `memo()` 適用必須（`.claude/CLAUDE.md` — Shared Component memo()）

## ステータス

✅ 修正済み
