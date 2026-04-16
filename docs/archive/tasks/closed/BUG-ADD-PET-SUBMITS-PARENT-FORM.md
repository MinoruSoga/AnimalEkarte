# BUG: ペット追加ボタンが親フォームを送信する

**Status**: FIXED  
**Priority**: HIGH  
**Discovery**: 機能テスト Section 1.4 (2026-04-12)

## 問題

`/owners/:id` の編集ページで「ペット追加」ボタンをクリックすると、ペット追加モーダルが開かずに親フォーム（飼主更新）が送信され、`/owners` 一覧へリダイレクトされる。

### 再現手順
1. `/owners/1` を開く
2. 「ペット追加」ボタンをクリック
3. 期待: ペット新規登録モーダルが開く
4. 実際: 飼主更新 PATCH が送信され `/owners` へ遷移・「飼主情報を更新しました」トーストが表示される

## 根本原因

`OwnerForm.tsx` の「ペット追加」ボタンに `type="button"` が欠落していた。

HTML の仕様では、`<form>` 内の `<button>` の `type` 属性はデフォルトで `"submit"` となる。
このボタンは `<form action={formAction}>` の子孫要素であるため、クリック時にフォームが送信されていた。

```tsx
// 修正前（バグあり）
<Button
  size="sm"
  onClick={handleAddPet}
  ...
>

// 修正後
<Button
  type="button"     ← 追加
  size="sm"
  onClick={handleAddPet}
  ...
>
```

## 修正箇所

`frontend/src/features/owners/routes/OwnerForm.tsx:626`

## 影響範囲

- `/owners/new`（新規登録ページ）: `isEdit=false` のため PATCH は発行されないが、フォームバリデーションが走り UX が壊れていた
- `/owners/:id`（編集ページ）: フォームが即座に送信されるため ペット追加が不可能

## テスト確認事項

- [ ] ペット追加ボタンクリック → モーダルが開く（フォーム送信されない）
- [ ] キャンセルボタンクリック → モーダルが閉じる（/owners への遷移なし）
- [ ] モーダルから登録 → ペットが一覧に追加される
- [ ] 新規登録ページでも同様に動作する
