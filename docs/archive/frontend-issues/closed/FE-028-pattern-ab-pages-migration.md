# FE-028: パターンA/B マスタ設定ページ（7ページ）を共通 hook + レイアウトに移行

**Status**: Open
**Priority**: Medium
**Affects**: master feature — 7ページ
**Date Created**: 2026-03-17
**Related**: TASK-007, FE-026, FE-027

## Summary

FE-026（useMasterCRUD）と FE-027（MasterListPage）を使って、パターンA（単一リスト DnDなし 4ページ）とパターンB（単一リスト + DnD 3ページ）の計7ページを移行する。各ページ ~80行削減で合計 ~560行削減。

## 対象ページ

### パターンA（DnDなし）
| ページ | 行数 | 移行後目安 |
|--------|------|----------|
| `StaffSettings.tsx` | 352 | ~200 |
| `HospitalizationSettings.tsx` | 362 | ~200 |
| `InterviewTemplateSettings.tsx` | 284 | ~150 |
| `ChiefComplaintSettings.tsx` | 265 | ~140 |

### パターンB（DnD付き）
| ページ | 行数 | 移行後目安 |
|--------|------|----------|
| `CageSettings.tsx` | 483 | ~300 |
| `ServiceTypeSettings.tsx` | 359 | ~200 |
| `AnimalSpeciesSettings.tsx` | 278 | ~150 |

## 移行パターン（各ページ共通）

```typescript
// Before（~350行）:
// - useState x4、useCallback x5、useMemo x2
// - PageLayout + SidePeek + ConfirmDialog の外枠
// - DataTable + Row マッピング

// After（~200行）:
const crud = useMasterCRUD({ ... });

return (
  <MasterListPage
    title="XXXマスタ"
    {...crud.layoutProps}
  >
    <DataTable columns={COLUMNS}>
      {crud.filteredItems.map((item) => (
        <DataTableRow key={item.id} onClick={() => crud.handleEdit(item)}>
          {/* セル内容のみ — ページ固有の部分 */}
        </DataTableRow>
      ))}
    </DataTable>
  </MasterListPage>
);
```

DnD 付きページの追加対応:
- `useSortableList` hook はそのまま維持
- `DndContext` + `SortableContext` はページ内に残す
- `SortableDataTableRow` を使用

## 各ページの SidePanel 内容（ページ固有 — 共通化しない）

SidePanel の中身はページごとに異なるため、`MasterXxxSidePanel` として各ページに残す:
- StaffSettings: 名前 + 役割 + メールアドレス + 権限
- CageSettings: 名前 + 料金 + is_active
- 等

## Vercel Best Practices チェック

各ページで以下を確認・修正:
- [ ] `memo()` で SidePanel コンポーネントを最適化
- [ ] `useCallback` でハンドラ安定化（useMasterCRUD 内で対応済み）
- [ ] `useDeferredValue` で検索遅延（useMasterCRUD 内で対応済み）
- [ ] 静的 JSX はモジュール定数に巻き上げ
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] barrel index import 禁止

## 依存関係

- **FE-026** と **FE-027** が先に完了している必要がある

## 完了条件

- [ ] 7ページ全てで useMasterCRUD + MasterListPage を使用
- [ ] 各ページの state/ハンドラ定義が hook 呼び出しに置換
- [ ] 各ページのレイアウト外枠が MasterListPage に置換
- [ ] DnD 機能が維持されている
- [ ] SidePanel の CRUD が正常動作
- [ ] 型エラーなし（`docker compose exec frontend pnpm build` パス）
- [ ] ESLint エラーなし（`docker compose exec frontend pnpm lint` パス）
