---
status: closed
closed_at: 2026-03-16
---

# [master] Settings.tsx: useEffect setState・useTransition 未使用・renderRow useCallback なし

## 優先度
高

## 種別
プロジェクトルール違反・パフォーマンス

## 対象ファイル
`frontend/src/features/master/routes/Settings.tsx`

## 問題

### 1. useEffect による setState（プロジェクトルール明示禁止）

L106-113 に `useEffect` で `setIsEditing`・`setSelectedItem`・`setSearchTerm`・`setFormData` を呼ぶコードが存在する。
コード自体に `// eslint-disable-next-line react-hooks/exhaustive-deps` が付いており、自身が問題を認識した上で回避している状態。

```tsx
// 現状（禁止パターン）
useEffect(() => {
  setIsEditing(false);
  setSelectedItem(null);
  setSearchTerm("");
  setFormData({});
  // eslint-disable-next-line react-hooks/exhaustive-deps
}, [category]);
```

`category` 変更時に状態をリセットする目的であれば、`key` prop を使ったコンポーネントリマウントで代替できる。

```tsx
// 修正案: 呼び出し側で key={category} を付ける、または
// Settings コンポーネント内で category を key として管理する
```

### 2. useTransition 未使用

`handleSave`（L133）の `update`・`add` 呼び出しに `useTransition` が適用されていない。
他のすべてのマスタページ（StaffSettings, HospitalizationSettings, CageSettings 等）は
`startSaveTransition` で API 書き込みを非緊急マークしており、Settings.tsx だけが不一致。

### 3. renderRow が useCallback でラップされていない

L211-238 の `renderRow` が `useMemo`/`useCallback` なしでインライン定義されており、
`DataTable` に毎レンダリングで新しい関数参照が渡される。
`DataTable` が `memo()` でラップされている場合、再レンダリング抑制の効果がない。

## 修正方針

1. **useEffect 削除**: `category` 変更時のリセットを `key` prop（`<Settings key={category} ...>`）または `useReducer` で代替する
2. **useTransition 追加**: `handleSave` を `startSaveTransition` でラップする
3. **renderRow を useCallback でラップ**: deps に `[filteredItems, handleEdit, showCode, showCategory, showPrice]` 等を指定する

## 完了条件
- [ ] `useEffect` による `setState` が削除されている
- [ ] `eslint-disable` コメントが残っていない
- [ ] `handleSave` が `startSaveTransition` でラップされている
- [ ] `renderRow` が `useCallback` でラップされている
- [ ] ビルドエラーなし
