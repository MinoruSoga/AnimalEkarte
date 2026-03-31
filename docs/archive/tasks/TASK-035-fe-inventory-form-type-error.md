# TASK-035: FE InventoryForm.tsx の型エラー修正

**作成日**: 2026-03-27
**ステータス**: Open
**優先度**: 高
**領域**: Frontend

---

## 概要

`tsc --noUnusedLocals` で検出された `InventoryForm.tsx:411` の型エラーを修正する。
`string` 型を `"medicine" | "food" | "other" | "consumable"` という union 型に代入できない。

---

## エラー詳細

```
src/features/inventory/routes/InventoryForm.tsx(411,11):
  error TS2322: Type 'string' is not assignable to type
    '"medicine" | "food" | "other" | "consumable"'.
```

---

## 調査方針

1. `InventoryForm.tsx:411` 周辺のコードを確認
2. `category` の初期値・渡し方を確認（`defaultCategory`、`existingItem?.category` など）
3. 以下のいずれかで修正：
   - 型アサーション `as "medicine" | "food" | "other" | "consumable"` （最終手段）
   - バリデーション関数で型ガード
   - `CategoryType` 型エイリアスを定義して統一
4. `inventory/api/types.ts` または `models.ts` の型定義と整合させる

---

## 受入条件

- [ ] `docker compose exec frontend npx tsc --noEmit` でこのエラーが消えている
- [ ] `docker compose exec frontend npm run build` 成功
- [ ] 在庫登録・編集フォームの動作に変化なし（機能デグレなし）
