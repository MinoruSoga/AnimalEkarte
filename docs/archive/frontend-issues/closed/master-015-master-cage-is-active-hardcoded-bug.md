---
status: closed
closed_at: 2026-03-16
---

# [master] CageSettings: 新規作成時の is_active がハードコードされている（機能バグ）

## 優先度
高

## 種別
機能バグ

## 対象ファイル
`frontend/src/features/master/routes/CageSettings.tsx`

## 問題

`handleSave` の新規作成分岐（L285）で `is_active: true` とハードコードされており、
`CageSidePanel` でユーザーがステータスを「無効」に変更しても新規作成時は無視される。

```tsx
// 現状（バグ）
const req: CreateCageRequest = {
  name: data.name,
  cage_type: data.cageType,
  cage_size: data.cageSize,
  price: priceValue,
  description: data.description || undefined,
  is_active: true,  // ← ハードコード。ユーザーの選択が無視される
};

// 修正後
is_active: data.isActive,  // ← フォームの状態を反映
```

同様の修正が正しく行われている `HospitalizationSettings.tsx` の L285（`is_active: data.isActive`）と対比すると明確なバグであることがわかる。

## 修正方針
`CreateCageRequest` の `is_active` を `true` から `data.isActive` に変更する。

## 完了条件
- [x] 新規作成時に `is_active: data.isActive` を使用している
- [x] ビルドエラーなし
