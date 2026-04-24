# FE-021: 見積一覧 — NotionFilter 移行

**Status**: Open
**Priority**: Low
**Affects**: estimates feature — 見積一覧
**Date Created**: 2026-03-17
**Related**: TASK-005

## Summary

見積一覧（EstimateList）が SearchFilterBar のまま未移行。NotionFilter に置き換える。

## 現状のコード

```typescript
// frontend/src/features/estimates/routes/EstimateList.tsx:6
import { SearchFilterBar } from '@/components/shared/SearchFilterBar/SearchFilterBar';

// 行151
<SearchFilterBar ... />
```

## 必要な変更

SearchFilterBar → NotionFilter に置き換え。会計一覧（Accounting.tsx）を参照実装として同パターンで移行する。

## 完了条件

- [ ] NotionFilter を使用
- [ ] SearchFilterBar の import を削除
- [ ] テキスト検索が動作
- [ ] 型エラーなし（`docker compose exec frontend pnpm build` パス）
