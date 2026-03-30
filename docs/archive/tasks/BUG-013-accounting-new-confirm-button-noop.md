# BUG-013: 新規会計登録で「会計を確定する」ボタンが機能しない

## 種類
バグ

## 発見日
2026-03-24

## 再現手順
1. `/accounting` に遷移
2. 「新規会計登録」をクリック
3. ペット選択画面でペットを選択 → `/accounting/new?petId=XX` に遷移
4. 「物販・その他追加」から品目を追加
5. 「丁度」ボタンでお預かり金額を設定（ボタンが有効化される）
6. 「会計を確定する」をクリック

## 期待動作
- POST `/api/v1/accountings` が呼び出され、新規会計レコードが作成される
- 一覧ページに遷移し、作成した会計が表示される

## 実際の動作
- ネットワークリクエストが一切発生しない（silent failure）
- ページ遷移もトーストも発生しない
- ボタンをクリックしても何も起きない

## 原因調査
`frontend/src/features/accounting/routes/AccountingDetail.tsx:628` の `handleComplete` 関数：

```typescript
const handleComplete = useCallback(() => {
    if (!accounting || !calculation || !id) return;  // ← id が undefined のため早期リターン
    // ...
    await updateAccounting(id, { ... });  // ← 既存レコードの更新のみ実装
}, [...]);
```

新規作成時（`/accounting/new`）は `useParams()` の `id` が `undefined` のため、`!id` 条件で即リターンされる。
さらに `createAccounting` の呼び出しが実装されていない。

### 関連問題
同じページで患者情報（飼い主名・ペット名）がハードコード値（"新規 飼い主様" / "新規 ペットちゃん"）になっており、実際に選択したペットの情報が表示されない：

```typescript
return {
    id: "acc_new",
    ownerName: "新規 飼い主様",   // ← petId から取得していない
    petName: "新規 ペットちゃん", // ← petId から取得していない
    ...
};
```

## 影響
- 新規会計の登録が完全に不可能（クリティカル）
- 既存レコードの編集（`/accounting/:id`）は正常動作

## 調査対象ファイル
- `frontend/src/features/accounting/routes/AccountingDetail.tsx`
  - `handleComplete` 関数（L628）: `!id` チェックの除去 + `createAccounting` 呼び出し追加
  - `baseAccounting` useMemo（L498）: petId から pet/owner 情報を取得するよう修正
- `frontend/src/features/accounting/api/create-accounting.ts`: 新規作成API関数
