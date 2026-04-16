# BUG-268: 7 リストページで handlePageChange useCallback の deps に pagination オブジェクトを指定

## 概要
`handlePageChange` を包む `useCallback` の deps 配列に `pagination`（`usePagination` が返すオブジェクト）をそのまま指定している。`usePagination` は毎レンダーで新しいオブジェクトリテラルを返すため（`return { currentPage, goToPage, ... }`）、`pagination` は毎レンダーで参照が変わる。結果として `handlePageChange` も毎レンダーで再生成される。`pagination.goToPage`（`useCallback` で安定化された関数）を直接 deps に指定することで不要な再生成を防げる。

## 現状コード

### 例: `frontend/src/features/checkups/routes/CheckupsList.tsx:106-117`
```typescript
const handlePageChange = useCallback((page: number) => {
  pagination.goToPage(page);           // ← goToPage のみ使用
  setSearchParams((prev) => {
    const next = new URLSearchParams(prev);
    if (page === 1) {
      next.delete("page");
    } else {
      next.set("page", String(page));
    }
    return next;
  }, { replace: true });
}, [pagination, setSearchParams]);
//   ^^^^^^^^^^ オブジェクト全体 — 毎レンダーで参照変化
```

`usePagination` の実装（`frontend/src/hooks/use-pagination.ts:82-94`）は毎レンダーで新しいオブジェクトを返す：
```typescript
return {
  currentPage: safePage,
  // ...
  goToPage,   // ← useCallback([totalPages]) で安定化済み
  // ...
};
```

**問題**: `pagination` 参照が毎レンダーで変わるため `handlePageChange` も毎レンダーで再生成。`Pagination` コンポーネントが `memo()` されている場合、`onPageChange` prop が毎回変わり再レンダーを引き起こす。

## 影響範囲

| ファイル | 行番号 |
|---------|-------|
| `frontend/src/features/accounting/routes/AccountingList.tsx` | 229 |
| `frontend/src/features/checkups/routes/CheckupsList.tsx` | 117 |
| `frontend/src/features/inventory/routes/InventoryList.tsx` | 149 |
| `frontend/src/features/examinations/routes/ExaminationsList.tsx` | 147 |
| `frontend/src/features/hospitalization/routes/HospitalizationList.tsx` | 227 |
| `frontend/src/features/owners/routes/OwnersList.tsx` | 239 |
| `frontend/src/features/vaccinations/routes/VaccinationList.tsx` | 131 |

## 修正方針

全7ファイルで統一修正。`pagination` → `pagination.goToPage` に変更：

```typescript
// Before
const handlePageChange = useCallback((page: number) => {
  pagination.goToPage(page);
  setSearchParams(...);
}, [pagination, setSearchParams]);

// After
// rerender-dependencies: pagination オブジェクト → stable な goToPage 関数のみ deps に指定
const handlePageChange = useCallback((page: number) => {
  pagination.goToPage(page);
  setSearchParams(...);
}, [pagination.goToPage, setSearchParams]);
```

`pagination.goToPage` は `usePagination` 内で `useCallback([totalPages])` によって安定化されているため、deps として安全に指定できる。

## 準拠すべきプロジェクト規約

### `.claude/rules/code-style.md` — rerender-dependencies
> `useCallback` deps にはオブジェクトを入れない — primitive を抽出して使う

### プロジェクト内参照実装
- `features/owners/routes/OwnersList.tsx` — `pendingDeleteOwnerId`（primitive id）を deps に（BUG-222 で修正済み参照）

## 優先度
**Medium** — `Pagination` コンポーネントは `memo()` 適用済みのため、`handlePageChange` の再生成が毎レンダーで `Pagination` 再レンダーを引き起こしている。7ファイルで同一パターン。

## 関連チケット
- BUG-222: useCallback deps にオブジェクト/配列（同一パターン、別ドメイン）

## 関連ファイル
- `frontend/src/hooks/use-pagination.ts:63-68` — `goToPage` が `useCallback([totalPages])` で安定化
