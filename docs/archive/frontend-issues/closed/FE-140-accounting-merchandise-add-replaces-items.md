# FE-140: 会計明細に物販追加後、既存明細がUIから消え合計が不正

**Status**: Open
**Priority**: High
**Affects**: features/accounting/
**Date Created**: 2026-03-29
**Related**: BUG-045

---

## Summary

会計精算ページで「物販・その他追加」から品目を選択すると、既存明細行が全て消えて選択アイテムのみ表示される。
POST リクエストも発生せず、リロードで元に戻る。

**根本原因（推定）**: 物販追加後のローカル state 更新で既存アイテムリストを**置き換え**ており、**マージ**していない。
かつ、バックエンドへの POST が実装されていない（フロント state のみの変更）。

---

## 実装手順

### 1. 原因調査

`features/accounting/` の物販追加ハンドラを確認：

```bash
grep -rn "addMerchandise\|handleAddItem\|物販" frontend/src/features/accounting/
```

物販選択後の state 更新箇所を特定し、以下を確認：
- `setItems(newItem)` → `setItems(prev => [...prev, newItem])` になっているか
- POST `/api/v1/accountings/:id/items` へのリクエストが実装されているか

### 2. state 更新をマージに修正

```typescript
// ❌ 現在（推測）: 置き換え
setItems([selectedMerchandiseItem]);

// ✅ 修正: マージ
setItems(prev => [...prev, selectedMerchandiseItem]);
```

### 3. POST API 呼び出しを実装

物販アイテム追加時に `POST /api/v1/accountings/:id/items` を呼び出す：

```typescript
// features/accounting/api/add-billing-item.ts
export async function addBillingItem(billingId: string, input: AddBillingItemInput) {
  const { data } = await axios.post(`/api/v1/accountings/${billingId}/items`, input);
  return data;
}
```

useTransition で pending 管理し、成功後に query を invalidate する。

### 4. 確認事項

- [ ] 物販追加後に既存明細が保持される
- [ ] 追加した物販がリスト末尾に表示される
- [ ] 合計金額が正しく再計算される
- [ ] POST リクエストが発生する（Network タブで確認）
- [ ] リロード後も追加した物販が残る

---

## 受入条件

- [ ] 既存明細（カルテ連携アイテム）が物販追加後も表示される
- [ ] 物販アイテムが既存明細の末尾に追加される
- [ ] 合計金額が全アイテムを含めて正しく計算される
- [ ] `POST /api/v1/accountings/:id/items` が呼ばれる
- [ ] リロード後もデータが保持される
- [ ] `docker compose exec frontend pnpm lint` エラー 0 件
