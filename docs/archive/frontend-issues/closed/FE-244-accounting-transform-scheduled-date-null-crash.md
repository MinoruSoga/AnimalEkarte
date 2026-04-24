# FE-244: accounting transforms.ts — scheduled_date の null チェックなしでクラッシュする可能性

## 概要

`frontend/src/features/accounting/api/transforms.ts` の `scheduled_date` フィールドで
null/undefined チェックなしに `.slice()` を呼び出している。
`scheduled_date` が null の場合、`TypeError: Cannot read properties of null (reading 'slice')` でクラッシュする。

## 問題コード

### `frontend/src/features/accounting/api/transforms.ts:66付近`

```ts
// Before: null チェックなし → scheduled_date が null の場合クランタイムクラッシュ
scheduledDate: data.scheduled_date.slice(0, 10),

// After: null チェックを追加
scheduledDate: data.scheduled_date ? data.scheduled_date.slice(0, 10) : null,
// または
scheduledDate: data.scheduled_date?.slice(0, 10) ?? null,
```

## 影響

会計データの取得時に `scheduled_date` が null の会計レコードが存在すると、
transforms 関数がクラッシュし、会計一覧・詳細ページ全体がエラーになる。

## 優先度
**Critical** — 本番データに null の scheduled_date が存在する場合、会計ページが完全に壊れる。

## 関連ファイル
- `frontend/src/features/accounting/api/transforms.ts`
