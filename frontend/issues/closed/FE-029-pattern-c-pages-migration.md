# FE-029: パターンC マスタ設定ページ（Diagnosis, Trimming）を共通 hook + レイアウトに移行

**Status**: Open
**Priority**: Medium
**Affects**: master feature — DiagnosisSettings, TrimmingSettings
**Date Created**: 2026-03-17
**Related**: TASK-007, FE-026, FE-027

## Summary

タブ2面構成の DiagnosisSettings（645行）と TrimmingSettings（704行）を useMasterCRUD + MasterListPage に移行する。タブごとに独立した CRUD state を持つため、useMasterCRUD を2回呼び出すパターンになる。

## 対象ページ

| ページ | 行数 | タブ構成 | 移行後目安 |
|--------|------|---------|----------|
| `DiagnosisSettings.tsx` | 645 | カテゴリ / 病名 | ~350 |
| `TrimmingSettings.tsx` | 704 | コース / オプション | ~350 |

## 移行パターン

```typescript
// DiagnosisSettings.tsx（移行後）
const categoryCrud = useMasterCRUD({
  data: categories,
  createMutation: useCreateDiagnosisCategory(),
  updateMutation: useUpdateDiagnosisCategory(),
  deleteMutation: useDeleteDiagnosisCategory(),
  entityLabel: "診断カテゴリ",
});

const nameCrud = useMasterCRUD({
  data: names,
  createMutation: useCreateDiagnosisName(),
  updateMutation: useUpdateDiagnosisName(),
  deleteMutation: useDeleteDiagnosisName(),
  entityLabel: "診断名",
});

// 現在のタブに応じて crud を切替
const activeCrud = activeTab === "category" ? categoryCrud : nameCrud;
```

## 依存関係

- **FE-026** と **FE-027** が先に完了している必要がある
- **FE-028** の後に着手（パターンA/B で hook + レイアウトの実績を積んでから）

## 完了条件

- [ ] 2ページで useMasterCRUD を使用（各ページ2回呼び出し）
- [ ] タブ切替が正常動作
- [ ] 各タブの CRUD が独立動作
- [ ] DnD 機能が維持されている（Diagnosis のカテゴリタブ）
- [ ] 型エラーなし（`docker compose exec frontend npm run build` パス）
