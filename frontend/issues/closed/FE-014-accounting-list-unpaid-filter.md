# FE-014: 会計一覧 — 未払いフィルタUI追加

**Status**: Open
**Priority**: Medium
**Affects**: accounting feature — 一覧画面
**Date Created**: 2026-03-17
**Related**: TASK-002

## Summary

会計一覧画面にステータスフィルタ（全件 / 会計待ち / 会計済 / キャンセル）を追加する。Backend API は既に `?status=waiting` をサポート済みのため、フロントエンドのみの変更。

## 現状のコード

### API hook — フィルタパラメータ未送信

```typescript
// frontend/src/features/accounting/api/get-accountings.ts:14-17
export const getAccountings = async (): Promise<Accounting[]> => {
  const { data } = await axios.get<AccountingsListResponse>("/v1/accountings");
  // ← status パラメータなし
  return data.data.map(transformToAccounting);
};
```

### 一覧画面 — テキスト検索のみ

```typescript
// frontend/src/features/accounting/routes/Accounting.tsx:69-82
const [searchTerm, setSearchTerm] = useState("");
const deferredSearch = useDeferredValue(searchTerm);

const filteredRecords = useMemo(() => {
  if (!deferredSearch) return accountings;
  return accountings.filter((r) =>
    r.ownerName.toLowerCase().includes(lowerTerm) ||
    r.petName.toLowerCase().includes(lowerTerm),
  );
}, [accountings, deferredSearch]);

// 行137-142: SearchFilterBar UI（テキスト検索のみ）
<SearchFilterBar
  searchTerm={searchTerm}
  onSearchChange={setSearchTerm}
  placeholder="飼主名、ペット名..."
  count={filteredRecords.length}
/>
```

### ステータスラベル定義 — 既に存在

```typescript
// frontend/src/features/accounting/routes/Accounting.tsx:31-36
const ACCOUNTING_STATUS_LABELS: Record<AccountingStatus, string> = {
  waiting: "会計待ち",
  pending: "会計待ち",
  completed: "会計済",
  cancelled: "キャンセル",
};
```

### 型定義

```typescript
// frontend/src/features/accounting/types/index.ts:1
export type AccountingStatus = "waiting" | "completed" | "cancelled" | "pending";
```

### Backend API — 既にステータスフィルタ対応済み

```go
// backend/internal/handler/accounting_handler.go:46-49
var status *string
if s := c.Query("status"); s != "" {
  status = &s
}
// backend/internal/repository/accounting_repository.go:41-43
if status != nil {
  q = q.Where("status = ?", *status)
}
```

## 必要な変更

### 1. API hook — status パラメータ追加

```typescript
// frontend/src/features/accounting/api/get-accountings.ts
// Before:
export const getAccountings = async (): Promise<Accounting[]> => {
  const { data } = await axios.get<AccountingsListResponse>("/v1/accountings");
  return data.data.map(transformToAccounting);
};

// After:
export const getAccountings = async (
  status?: AccountingStatus,
): Promise<Accounting[]> => {
  const params: Record<string, string> = {};
  if (status) params.status = status;
  const { data } = await axios.get<AccountingsListResponse>("/v1/accountings", { params });
  return data.data.map(transformToAccounting);
};

export const useGetAccountings = (status?: AccountingStatus) => {
  return useQuery({
    queryKey: ["accountings", { status }],
    queryFn: () => getAccountings(status),
  });
};
```

### 2. 一覧画面 — ステータスドロップダウン追加

```typescript
// frontend/src/features/accounting/routes/Accounting.tsx
// SearchFilterBar の横にステータス Select を追加

const [statusFilter, setStatusFilter] = useState<AccountingStatus | "">("");

// フィルタ選択肢:
// "" → 全件
// "waiting" → 会計待ち（未払い）
// "completed" → 会計済
// "cancelled" → キャンセル

// Select UI（shadcn/ui の Select コンポーネントを使用）
<Select value={statusFilter} onValueChange={setStatusFilter}>
  <SelectTrigger className="w-[140px]">
    <SelectValue placeholder="すべて" />
  </SelectTrigger>
  <SelectContent>
    <SelectItem value="">すべて</SelectItem>
    <SelectItem value="waiting">会計待ち</SelectItem>
    <SelectItem value="completed">会計済</SelectItem>
    <SelectItem value="cancelled">キャンセル</SelectItem>
  </SelectContent>
</Select>
```

### 3. API 呼び出しの変更

```typescript
// useGetAccountings に statusFilter を渡す
const { data: accountings = [], isLoading } = useGetAccountings(
  statusFilter || undefined,
);
```

## UI 操作フロー

1. ユーザーが会計一覧画面を開く
2. デフォルトは「すべて」（全件表示）
3. ドロップダウンから「会計待ち」を選択
4. 未払い（status=waiting）のレコードのみ表示される
5. テキスト検索との併用が可能

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `useDeferredValue` でフィルタ遅延（テキスト検索は既に対応済み）
- [ ] 型は `models.ts` から導出

## 依存関係

- Backend 変更不要（`?status=waiting` は既にサポート済み）

## 完了条件

- [ ] `get-accountings.ts` に status パラメータ追加
- [ ] `useGetAccountings` の queryKey に status を含める
- [ ] ステータス Select UI を SearchFilterBar の横に配置
- [ ] 「会計待ち」選択で未払いのみ表示
- [ ] 「すべて」選択で全件表示
- [ ] テキスト検索との併用が動作
- [ ] 型エラーなし（`docker compose exec frontend npm run build` パス）
- [ ] ESLint エラーなし（`docker compose exec frontend npm run lint` パス）
