---
name: BUG-242_trimming-includes-inside-map-on-memo
description: memo'd LeftColumn で optionIds.includes() を options.map() 内で呼び出し O(n²)
type: bug
---

# BUG-242: TrimmingForm LeftColumn — `optionIds.includes()` in `options.map()` が O(n²)

## 概要

`TrimmingForm.tsx` の `LeftColumn`（`memo()` 済み）内で、
`options.map()` のループ内に `formData.optionIds.includes(option.id)` が存在する。
これは O(n × m) の二重線形探索。`js-set-map-lookups` 違反。

`LeftColumn` は `memo()` で囲まれているため、`formData` が変わるたびに再レンダーされ、
毎回この二重探索が走る。Set を使えば O(n) に削減できる。

## 現状コード

### `features/trimming/routes/TrimmingForm.tsx:76, 139`

```tsx
// LeftColumn は memo() 済み
const LeftColumn = memo(function LeftColumn({
  formData,   // ← formData 全体がオブジェクトで渡される
  options,
  ...
}: LeftColumnProps) {
  // ...
  {options.map((option) => (       // ← O(n) ループ
    <Checkbox
      checked={formData.optionIds.includes(option.id)}  // ← O(m) 探索 = 合計 O(n×m)
```

`formData.optionIds` は配列（`string[]`）。`options` はコース依存のオプション一覧。
どちらも通常 10〜20 件程度だが、アイテム数が増えると二乗で劣化する。

## 期待する動作

`formData.optionIds` を `Set<string>` に変換し `O(1)` の `has()` を使う。

## 修正方針

### `features/trimming/routes/TrimmingForm.tsx:86付近` — LeftColumn 内に useMemo 追加

```tsx
const LeftColumn = memo(function LeftColumn({
  formData,
  options,
  ...
}: LeftColumnProps) {
  // js-set-map-lookups: array.includes() O(n) → Set.has() O(1)
  const optionIdSet = useMemo(
    () => new Set(formData.optionIds),
    [formData.optionIds]
  );

  // ...
  {options.map((option) => (
    <Checkbox
      checked={optionIdSet.has(option.id)}   // ← O(1)
      // ...
    />
  ))}
```

`useMemo` の deps を `formData.optionIds`（配列参照）にすることで、
チェックボックスが変わったときのみ Set を再生成する。

## 参照実装

同一プロジェクト内の BUG-234 では同パターンを `PetSelection` 等 3 箇所で修正予定。

`.claude/rules/code-style.md`:
> `js-set-map-lookups`: Use Set/Map for O(1) lookups

## 影響範囲

| ファイル | 行 | 内容 |
|---------|-----|------|
| `features/trimming/routes/TrimmingForm.tsx:139` | O(n²) includes | 修正必要 |

## 優先度

**Low** — options/optionIds は通常 < 20 件のため実害は軽微。ただし `memo()` 済みコンポーネント内の最適化として対応すべき。

## 関連チケット

- BUG-234: 同一パターン（PetSelection 等 3箇所）— 未修正
