# BUG-294: lib/handle-api-error.ts — console.error に DEV ガードなし

## 概要

`frontend/src/lib/handle-api-error.ts:39` の `console.error` が `import.meta.env.DEV` ガードなしで実装されており、本番環境でもコンソールにエラーが出力される。

## 問題

```typescript
// frontend/src/lib/handle-api-error.ts:38-40
// Non-Axios errors (unexpected)
console.error("Non-Axios Error:", err);  // ← DEVガードなし
toast.error(`${context}中に予期しないエラーが発生しました。`);
```

CLAUDE.mdの禁止事項: `console.log` 放置（本番コード汚染）。`console.error` も同様。

## 修正

```typescript
// DEVガードを追加
if (import.meta.env.DEV) {
  console.error("Non-Axios Error:", err);
}
toast.error(`${context}中に予期しないエラーが発生しました。`);
```

## ステータス

- [x] ドキュメント作成
- [x] 実装完了
