# BUG-CHECKUPS-MISSING-CRUD: 定期健診一覧に新規登録ボタン・操作カラムが欠落

## ステータス
✅ **修正済み**（コミット `76132fbb` fix(checkups): 定期健診一覧に新規登録ボタン・操作カラム追加 — 対応済み）

## 優先度
High

## 再現手順
1. `/checkups` を開く
2. ヘッダー右上・テーブル列を確認

## 症状
他の全管理ページ（カルテ・トリミング・検査・予防接種等）には存在する
「+ 新規登録」ボタンと「操作」カラムが定期健診ページのみ欠落している。

| ページ | 新規登録ボタン | 操作カラム |
|--------|--------------|-----------|
| カルテ管理 | ✅ | ✅ |
| トリミング管理 | ✅ | ✅ |
| 予防接種管理 | ✅ | ✅ |
| **定期健診** | ❌ | ❌ |

## 根本原因
`CheckupsList.tsx` の `PageLayout` に `onNew` props を渡しておらず、
テーブルの `renderRow` にも操作セル（編集/削除）が実装されていない。

```tsx
// 現状 (CheckupsList.tsx)
<PageLayout title="定期健診" resource={ResourceCheckups} ...>
  // onNew=undefined → 「+ 新規登録」ボタンが非表示
```

## 修正方針
1. `PageLayout` に `onNew` を追加（`/checkups/select-pet` または `/checkups/new` へ遷移）
2. `renderRow` に操作列（編集・削除）を追加
3. router.tsx に `/checkups/select-pet`, `/checkups/new`, `/checkups/:id/edit` ルートが
   存在するか確認し、なければ追加

## 影響範囲
- `/checkups` 一覧ページ
- 定期健診の CRUD 全般
