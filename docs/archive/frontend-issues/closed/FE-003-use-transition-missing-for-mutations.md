# FE-003: useTransition 未適用 — 書き込み操作のpending状態管理

## 問題
30件以上の mutation 呼び出しが `useMutation` の `isPending` を
直接使用しており、`useTransition` でラップされていない。
`useTransition` を使うことでUIのレスポンシブ性と
Suspenseとの統合が改善される。

## 適用済み（参照実装）
- `features/owners/hooks/useOwnerForm.ts` — `startSaveTransition`
- `features/owners/routes/OwnersList.tsx` — `startDeleteTransition`
- `features/trimming/hooks/useTrimmingForm.ts` — save/delete両方
- `features/master/routes/*.tsx`（CageSettings, StaffSettings 等）

## 未適用の主要ファイル
- `features/accounting/routes/AccountingDetail.tsx`
- `features/estimates/routes/EstimateForm.tsx`
- `features/hospitalization/routes/HospitalizationForm.tsx`
- `features/medical-records/` 配下の複数 mutation

## 修正方針
```ts
const [isSaving, startSaveTransition] = useTransition();

const handleSave = useCallback(() => {
  startSaveTransition(async () => {
    await mutateFn(data);
  });
}, [mutateFn, data]);
```

ボタンの `disabled={isSaving}` に `isPending` から切り替える。

## 優先度
LOW（パフォーマンス改善・段階的適用で可）
