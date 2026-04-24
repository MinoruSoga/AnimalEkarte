# FE-237: 本番コードの unsafe type cast（as unknown as）3件

## 概要

本番コード3箇所で `as unknown as X` の二重キャストが使用されている。
TypeScript の型安全性を迂回するアンチパターンであり、プロジェクト規約（`any` 禁止）の精神に反する。

## 違反箇所

### `frontend/src/features/estimates/routes/EstimateList.tsx:154-155`

```ts
// ソートロジックでの動的プロパティアクセス
const aVal = String((a as unknown as Record<string, unknown>)[sort.key] ?? "");
const bVal = String((b as unknown as Record<string, unknown>)[sort.key] ?? "");
```

**問題**: ソートキー `sort.key` が `string` 型で、Estimate 型のプロパティとして型安全にアクセスできないため二重キャストで回避している。

**修正方針**:
```ts
// 型安全なアクセスに変更
type EstimateSortKey = keyof Estimate;
const aVal = String(a[sort.key as EstimateSortKey] ?? "");
const bVal = String(b[sort.key as EstimateSortKey] ?? "");
```

### `frontend/src/components/shared/NotionDatePicker/NotionDatePicker.tsx:607`

```tsx
// キーボードイベントをマウスイベントにキャスト
onClick(e as unknown as React.MouseEvent<HTMLSpanElement>);
```

**問題**: `<span role="button">` の `onKeyDown` ハンドラで `KeyboardEvent` を `MouseEvent` にキャスト。
そもそも `<span role="button">` ではなく `<button>` を使うべき（FE-205 で指摘済みの別問題）。

**修正方針**:
```tsx
// <button> に変更すれば onClick は自然に動作し、キャストが不要になる
// または onClick の型シグネチャを合わせる
```

## 影響

`as unknown as` キャストはランタイムでは `undefined` や予期しない型の値をすり抜けさせる可能性がある。
`EstimateList` のソート処理でアクセス対象プロパティが存在しない場合、`undefined` が `String()` で `"undefined"` になり、ソート順が壊れる可能性がある。

## 準拠すべきプロジェクト規約

### `.claude/CLAUDE.md` — 型安全性最優先
> TypeScript で `any` を禁止し、厳格な型定義を行う。
> `as unknown as X` は `any` と同等の型安全性破壊であるため同様に禁止。

## 優先度
**Medium** — 型安全性の破壊。`EstimateList` のソートは実際の動作に影響する可能性がある。

## 関連ファイル
- `frontend/src/features/estimates/routes/EstimateList.tsx`
- `frontend/src/components/shared/NotionDatePicker/NotionDatePicker.tsx`
