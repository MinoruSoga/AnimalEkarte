# FE-246: shifts/api/transforms.ts — 日付フィールドの null チェックなしで .split() クラッシュ

## 概要

`frontend/src/features/shifts/api/transforms.ts` の日付フィールドで
null/undefined チェックなしに `.split("T")[0]` を呼び出している。
日付が null の場合 `TypeError: Cannot read properties of null (reading 'split')` でクラッシュする。

## 問題コード

### `frontend/src/features/shifts/api/transforms.ts:10付近`

```ts
// Before: null チェックなし → 日付が null の場合クラッシュ
date: data.date.split("T")[0],

// After
date: data.date ? data.date.split("T")[0] : "",
// または
date: data.date?.split("T")[0] ?? "",
```

## 関連

FE-244（accounting/transforms.ts の同種クラッシュバグ）と同パターン。
`accounting/api/transforms.ts` では `.slice(0, 10)`、本ファイルでは `.split("T")[0]` だが
根本原因は同じ（日付フィールドの null 非対応）。

## 優先度
**High** — シフトデータに null の日付が存在する場合、シフト管理ページが完全にクラッシュする。

## 関連ファイル
- `frontend/src/features/shifts/api/transforms.ts`
- 関連: FE-244（同種問題 accounting transforms）
