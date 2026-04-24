# FE-207: use-hospitalization-form.ts で Date.now() を一時 ID として使用

## 概要

`use-hospitalization-form.ts` の `addTreatmentPlan()` で新規治療プランアイテムの
一時 ID として `Date.now().toString()` を使用している。
同一ミリ秒内に複数のアイテムが追加されると重複 ID が発生する。
また React StrictMode では問題が顕在化する可能性がある。

## 現状コード

### `frontend/src/features/hospitalization/hooks/use-hospitalization-form.ts:198-207`
```ts
const addTreatmentPlan = () => {
  const newPlan: TreatmentPlan = {
    id: Date.now().toString(),  // ← 同一ミリ秒で重複の可能性
    treatmentContent: "",
    memo: "",
    insurance: false,
    unitPrice: 0,
    quantity: 1,
    discount: 0,
    discountAmount: 0,
    subtotal: 0,
  };
```

## 問題点

- 1ミリ秒以内に複数追加すると `id` が重複 → React の `key` に使用した場合リストが壊れる
- 同一プロジェクトの他実装では `crypto.randomUUID()` を使用している

## 修正方針

```ts
// Before
id: Date.now().toString(),

// After
id: crypto.randomUUID(),
```

`crypto.randomUUID()` はモダンブラウザ・Node.js 14.17+ でサポート済み。インポート不要。

## 影響範囲

| 対象 | 行番号 | 状態 |
|------|--------|------|
| `frontend/src/features/hospitalization/hooks/use-hospitalization-form.ts` | 200 | 要修正 |

## 優先度
**Low** — 実用上は稀なケース。ただし idiomatic でない実装のため修正推奨。

## 関連ファイル
- `frontend/src/features/hospitalization/hooks/use-hospitalization-form.ts:200`
