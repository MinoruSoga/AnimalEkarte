---
status: closed
closed_at: 2026-03-16
---

# [master] MedicineSettings: カテゴリ行の Fragment に key がない（React 警告）

## 優先度
中

## 種別
コード品質・React 警告

## 対象ファイル
`frontend/src/features/master/routes/MedicineSettings.tsx`

## 問題

カテゴリ行のレンダリング（L754-833）が `<>...</>` Fragment でラップされており、
Fragment 自体に `key` が付いていない。内側の `TableRow`（L756）に `key` が付いているが、
React はリスト描画で Fragment レベルの `key` を要求するため警告が発生する。

```tsx
// 現状（key なし Fragment）
{Array.from(groupedData.entries()).map(([parentId, { category, medicines }]) => (
  <>  // ← key がない
    <TableRow key={parentId} ...>  // ← ここに key を付けても Fragment レベルでは無効
      ...
    </TableRow>
    {expandedCategories.has(parentId) && medicines.map(...)}
  </>
))}

// 修正後
{Array.from(groupedData.entries()).map(([parentId, { category, medicines }]) => (
  <Fragment key={parentId}>  // ← Fragment に key を付ける
    <TableRow ...>
      ...
    </TableRow>
    {expandedCategories.has(parentId) && medicines.map(...)}
  </Fragment>
))}
```

## 修正方針
1. `import { Fragment } from "react"` を追加（または既存 import に追加）
2. `<>` を `<Fragment key={parentId}>` に変更
3. `TableRow` の `key={`h-${parentId}`}` は削除して Fragment に移す

## 完了条件
- [x] カテゴリ行の Fragment に `key={parentId}` が付いている
- [x] React の key 警告がなくなっている
- [x] ビルドエラーなし
