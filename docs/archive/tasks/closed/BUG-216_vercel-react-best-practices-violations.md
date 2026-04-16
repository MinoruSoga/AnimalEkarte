# BUG-216: Vercel React Best Practices 違反（grep 検証済み）

| 項目 | 内容 |
|------|------|
| 優先度 | **Medium** |
| カテゴリ | パフォーマンス / Vercel React Best Practices |
| 検証方法 | grep による全文検索で確認 |

## 1. `rendering-conditional-render` — && 条件レンダー

**features/ 内: 0件。** 全てのドメインで `? (...) : null` パターンを使用しており準拠済み。

components/ui/ 内（shadcn/ui 由来）に 4件あるが、修正対象外：
- `chart.tsx:200,233`, `sidebar.tsx:615`, `input-otp.tsx:58`, `resizable.tsx:45`

## 2. `rerender-transitions` — useDeferredValue 未使用（1箇所）

`features/hospitalization/routes/HospitalizationList.tsx:3,139-176`

`searchTerm` を直接 `useMemo` の deps に渡してフィルタリング。`useDeferredValue` 未使用。
他の全リスト画面（OwnersList, VaccinationList, AccountingList 等）は正しく使用済み。

```typescript
// 現状
const filtered = useMemo(() => {
  if (searchTerm) { ... }  // searchTerm を直接使用
}, [allHospitalizations, statusFilter, searchTerm, activeFilters]);

// 修正
const deferredSearch = useDeferredValue(searchTerm);
const filtered = useMemo(() => {
  if (deferredSearch) { ... }
}, [allHospitalizations, statusFilter, deferredSearch, activeFilters]);
```

## 3. `bundle-dynamic-imports` — lazy() + Suspense 使用状況

**広く使用済み。** grep で 23 箇所確認。主要モーダル（PetEditModal, OwnerSearchModal, ReservationFormModal, ShiftFormDialog, TreatmentSearchDialog 等）は全て lazy() でロード。

## 4. `async-parallel` — Promise.all 使用状況

**使用済み。** owners/loaders.ts, hospitalization/hooks 等で Promise.all / Promise.allSettled を使用。

## 5. デザイントークン違反（grep count 確認済み）

features/ 全体で Tailwind ハードコード色が **19箇所/9ファイル**:

| ドメイン | ファイル数 | 箇所数 |
|---------|----------|--------|
| hospitalization | 7 | 15 |
| trimming | 1 | 2 |
| medical-records | 1 | 2 |

**hospitalization に集中** — 他ドメインはほぼ準拠済み。

## 6. `rerender-memo` — shared コンポーネントの memo() 適用状況

**memo() 適用済み（grep 確認）:**
DataTable, SidePeekPanel, Pagination, NotionFilter, SortPill, SortPopover, FilterRuleRow, FilterAddPopover, OwnerSearchModal, HistoryFilterPanel, CharCountTextarea, PermissionBadges + 内部コンポーネント多数

**memo() 未適用の export function（大型コンポーネント）:**

これらが memo 対象かはユースケースによる。親が頻繁に再レンダーされる場合のみ影響：
- `PageLayout`, `FormHeader`, `ReservationFormModal`, `TreatmentSearchDialog`, `PatientInfoCard`
- `ConfirmDialog`, `MasterSelectModal`, `NavigationBlocker`

CLAUDE.md 規約「新規共有コンポーネントも memo() を適用すること」に従えば対象だが、
ダイアログ系は開閉時のみ描画されるため実質的なパフォーマンス影響は小さい。

## 総評

Vercel React Best Practices への準拠度は**高い**。

| ルール | 準拠状況 |
|--------|---------|
| `rendering-conditional-render` (&&禁止) | **100%** — features/ 内で0件 |
| `rerender-transitions` (useDeferredValue) | **97%** — 1件のみ未使用（HospitalizationList） |
| `bundle-dynamic-imports` (lazy) | **良好** — 23箇所で使用 |
| `async-parallel` (Promise.all) | **良好** — loaders 等で使用 |
| デザイントークン | **90%** — hospitalization に集中（15箇所） |
| `rerender-memo` (shared memo) | **良好** — 主要コンポーネントは適用済み |
