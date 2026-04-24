# FE-027: MasterListPage レイアウトコンポーネント作成

**Status**: Open
**Priority**: High
**Affects**: master feature — 全マスタ設定ページ共通
**Date Created**: 2026-03-17
**Related**: TASK-007, FE-026, FE-028

## Summary

マスタ設定ページ11ページで重複しているページレイアウト外枠（PageLayout + SidePanel + ConfirmDialog + NotionFilter）を1つの共通コンポーネント `MasterListPage` に抽出する。推定 ~220行の重複を削減。

## 重複パターン（現在11ページで繰り返し）

```tsx
<>
  <div className="flex h-full">
    <div className="flex-1 min-w-0">
      <PageLayout
        title="XXXマスタ"
        icon={<XxxIcon className="..." />}
        onBack={() => navigate(paths.settings.getHref())}
        maxWidth="max-w-full"
        headerAction={
          <PrimaryButton onClick={() => setEditTarget("new")}>
            新規登録
          </PrimaryButton>
        }
      >
        <NotionFilter
          properties={[]}
          activeFilters={[]}
          onFilterChange={() => {}}
          searchTerm={searchTerm}
          onSearchChange={setSearchTerm}
          searchPlaceholder="XXXで検索..."
          count={filteredItems.length}
        />
        <DataTable columns={COLUMNS}>
          {/* rows */}
        </DataTable>
      </PageLayout>
    </div>
    {editTarget !== null ? <XxxSidePanel ... /> : null}
  </div>
  <ConfirmDialog
    open={pendingDelete !== null}
    title="XXXの削除"
    description={`「${pendingDelete?.name}」を削除しますか？`}
    onConfirm={handleDeleteConfirm}
    onCancel={() => setPendingDelete(null)}
  />
</>
```

## 必要な変更

### 新規ファイル: `frontend/src/features/master/components/MasterListPage.tsx`

```typescript
interface MasterListPageProps {
  // ヘッダー
  title: string;
  icon: React.ReactNode;

  // CRUD（useMasterCRUD の return を渡す）
  searchTerm: string;
  onSearchChange: (term: string) => void;
  searchPlaceholder: string;
  count: number;
  onNew: () => void;

  // SidePanel
  sidePanel: React.ReactNode | null;

  // 削除確認
  deleteTarget: { name: string } | null;
  deleteTitle: string;
  onDeleteConfirm: () => void;
  onDeleteCancel: () => void;

  // テーブル（children）
  children: React.ReactNode;
}

export const MasterListPage = memo(function MasterListPage({
  title, icon,
  searchTerm, onSearchChange, searchPlaceholder, count, onNew,
  sidePanel,
  deleteTarget, deleteTitle, onDeleteConfirm, onDeleteCancel,
  children,
}: MasterListPageProps) {
  const navigate = useNavigate();
  return (
    <>
      <div className="flex h-full">
        <div className="flex-1 min-w-0">
          <PageLayout
            title={title}
            icon={icon}
            onBack={() => navigate(paths.settings.getHref())}
            maxWidth="max-w-full"
            headerAction={
              <PrimaryButton onClick={onNew}>新規登録</PrimaryButton>
            }
          >
            <NotionFilter
              properties={[]}
              activeFilters={[]}
              onFilterChange={() => {}}
              searchTerm={searchTerm}
              onSearchChange={onSearchChange}
              searchPlaceholder={searchPlaceholder}
              count={count}
            />
            {children}
          </PageLayout>
        </div>
        {sidePanel}
      </div>
      <ConfirmDialog
        open={deleteTarget !== null}
        title={deleteTitle}
        description={deleteTarget ? `「${deleteTarget.name}」を削除しますか？` : ""}
        onConfirm={onDeleteConfirm}
        onCancel={onDeleteCancel}
      />
    </>
  );
});
```

### Vercel Best Practices 準拠ポイント

- `memo()` でコンポーネント最適化
- children パターンで柔軟な合成
- 静的 JSX はモジュール定数に巻き上げ不要（props 駆動のため）

## 使用例（移行後の各ページ）

```tsx
// CageSettings.tsx（移行後）
<MasterListPage
  title="ケージマスタ"
  icon={<Building2 className="h-5 w-5" />}
  searchTerm={crud.searchTerm}
  onSearchChange={crud.setSearchTerm}
  searchPlaceholder="ケージ名で検索..."
  count={crud.filteredItems.length}
  onNew={crud.handleNew}
  sidePanel={crud.isEditing ? <CageSidePanel ... /> : null}
  deleteTarget={crud.pendingDelete}
  deleteTitle="ケージの削除"
  onDeleteConfirm={crud.handleDeleteConfirm}
  onDeleteCancel={crud.handleDeleteCancel}
>
  <DataTable columns={COLUMNS}>
    {crud.filteredItems.map((cage) => (
      <DataTableRow key={cage.id} ... />
    ))}
  </DataTable>
</MasterListPage>
```

## 完了条件

- [ ] `MasterListPage.tsx` 新規作成
- [ ] `memo()` で最適化
- [ ] CageSettings.tsx で動作確認（1ページ先行適用）
- [ ] 型エラーなし（`docker compose exec frontend pnpm build` パス）
- [ ] ESLint エラーなし（`docker compose exec frontend pnpm lint` パス）
