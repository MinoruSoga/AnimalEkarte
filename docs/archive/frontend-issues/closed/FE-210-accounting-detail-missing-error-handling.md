# FE-210: AccountingDetail に isError 未チェックと handleUpdateItemTax のエラー処理欠落

## 概要

`AccountingDetail.tsx` で2つのエラーハンドリング問題が存在する:
1. `useGetAccountingDetail` の `isError` をチェックしておらず、エラー時に空白ページが表示される
2. `handleUpdateItemTax` に `try/catch` がなく、税率更新失敗時にユーザーに通知されない

## 問題1: isError 未チェック

### `frontend/src/features/accounting/routes/AccountingDetail.tsx:768付近`
```tsx
// Before: isLoading のみチェック、isError なし
const { data: accounting, isLoading } = useGetAccountingDetail(id);

if (id && isLoading) return <div>読み込み中...</div>;
if (!accounting || !calculation) return <div>データが見つかりません</div>;
// → API エラー時も「データが見つかりません」と表示。エラー原因が不明

// After
const { data: accounting, isLoading, isError } = useGetAccountingDetail(id);

if (id && isLoading) return <LoadingFallback />;
if (isError) return <ErrorFallback />;
if (!accounting) return <div className={`...`}>データが見つかりません</div>;
```

## 問題2: handleUpdateItemTax に try/catch なし

### `frontend/src/features/accounting/routes/AccountingDetail.tsx:1031-1040付近`
```tsx
// Before: try/catch なし、handleApiError なし
const handleUpdateItemTax = useCallback(
  (itemId: string, taxType: TaxType, taxRate: number) => {
    if (!id) return;
    startTaxUpdateTransition(async () => {
      await updateBillingItem(itemId, { tax_type: taxType, tax_rate: taxRate });
      queryClient.invalidateQueries({ queryKey: queryKeys.accountings.detail(id) });
    });
  },
  [id, queryClient]
);

// After: try/catch + handleApiError を追加
const handleUpdateItemTax = useCallback(
  (itemId: string, taxType: TaxType, taxRate: number) => {
    if (!id) return;
    startTaxUpdateTransition(async () => {
      try {
        await updateBillingItem(itemId, { tax_type: taxType, tax_rate: taxRate });
        queryClient.invalidateQueries({ queryKey: queryKeys.accountings.detail(id) });
      } catch (error) {
        handleApiError(error, "税率更新");
      }
    });
  },
  [id, queryClient]
);
```

## 影響範囲

| 対象 | 行番号 | 状態 |
|------|--------|------|
| `frontend/src/features/accounting/routes/AccountingDetail.tsx` | isError 未チェック箇所 | 要修正 |
| `frontend/src/features/accounting/routes/AccountingDetail.tsx:1031-1040` | handleUpdateItemTax | 要修正 |

## 準拠すべきプロジェクト規約

### `.claude/rules/error-handling.md`
> すべての `catch` ブロックで `handleApiError` を呼び出す。

### プロジェクト内参照実装
- `frontend/src/features/accounting/routes/AccountingList.tsx:350-351` — isLoading + isError + ErrorFallback/LoadingFallback で正しく実装済み

## 優先度
**High** — 税率更新失敗時のエラーが握り潰される。API エラー時のデータ整合性が損なわれる。

## 関連ファイル
- `frontend/src/features/accounting/routes/AccountingDetail.tsx`
- `frontend/src/lib/handle-api-error.ts`
- `frontend/src/components/shared/DataStates/` — LoadingFallback/ErrorFallback
