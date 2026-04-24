# FE-243: MedicalRecordBillCheck の returnMutation にエラーハンドリングがない

## 概要

`frontend/src/features/medical-records/components/MedicalRecordBillCheck.tsx` の
`handleReturn` 関数が `returnMutation.mutate()` を呼び出すが、
エラーハンドリングが設定されていない。会計確認の「差し戻し」操作が失敗してもユーザーに通知されない。

## 問題コード

### `MedicalRecordBillCheck.tsx:62-69付近`

```tsx
// Before: returnMutation に onError なし + handleReturn に try/catch なし
const handleReturn = () => {
  returnMutation.mutate(medicalRecordId, {
    onSuccess: () => { ... },
    // onError なし
  });
};

// After: onError を追加
const handleReturn = () => {
  returnMutation.mutate(medicalRecordId, {
    onSuccess: () => { ... },
    onError: (error) => handleApiError(error, "会計確認の差し戻し"),
  });
};
```

## 追加: デザイントークン違反

同ファイル内の bg-white 指定はデザイントークンに変換が必要：

| 行 | 違反 | 修正 |
|----|------|------|
| 138 | `bg-white` | `C.bgPage` |
| 163 | `bg-white` | `C.bgPage` |

## 備考

FE-221 の `billing-review.ts` の `useReturnBillingReview` の `onError` 欠落と連動した問題。
ただし、本チケットはコンポーネント側の `mutate()` 呼び出し時の `onError` コールバック欠落を対象とする。

## 準拠すべきプロジェクト規約

### `.claude/rules/error-handling.md`
> すべての `catch` ブロックおよび `onError` コールバックで `handleApiError(error, "コンテキスト")` を呼び出す。

## 優先度
**High** — 会計確認の差し戻し失敗が通知されない。会計フロー上の重要な操作。

## 関連ファイル
- `frontend/src/features/medical-records/components/MedicalRecordBillCheck.tsx`
- `frontend/src/features/medical-records/api/billing-review.ts`（FE-221 関連）
- `frontend/src/lib/handle-api-error.ts`
