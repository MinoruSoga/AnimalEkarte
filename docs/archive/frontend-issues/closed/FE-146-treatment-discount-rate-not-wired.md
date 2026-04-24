# FE-146: カルテ治療プランの割引率フィールドが非機能（discountRate 分岐なし）

**Status**: Open
**Priority**: Medium
**Affects**: features/medical-records/components/MedicalRecordDiagnosisPlan.tsx
**Date Created**: 2026-03-29
**Related**: BUG-051

---

## Summary

治療プランテーブルの「割引(%)」フィールドに入力しても即座に 0 にリセットされる。
`handleUpdateItem` の `field === "discountRate"` 分岐が存在しないため、
`UpdateTreatmentInput` に `discount_rate` がセットされず空 PATCH が送られ、
React が古い値（0）で再レンダリングする。

また「割引(%)」ラベルと実際の計算ロジック（比率 0〜1）が不一致（入力値 10 → 10倍割引）。

---

## 実装手順

### 1. `handleUpdateItem` に `discountRate` 分岐を追加（91〜103行目付近）

```typescript
const handleUpdateItem = useCallback((id: number, field: keyof TreatmentItem, value: ...) => {
  const input: UpdateTreatmentInput = {};
  if (field === "content") input.content = String(value);
  if (field === "memo") input.memo = String(value);
  if (field === "insurance") input.insurance = Boolean(value);
  if (field === "unitPrice") input.unit_price = Number(value);
  if (field === "quantity") input.quantity = Number(value);
  if (field === "discountAmount") input.discount_amount = Number(value);
  if (field === "discountRate") input.discount_rate = Number(value);  // ← 追加
  if (field === "status") input.status = String(value);
  if (field === "selected") input.selected = Boolean(value);
  updateMutation.mutate({ treatmentId: String(id), input });
}, [updateMutation]);
```

### 2. ラベル表記の修正

現在「割引(%)」ラベルで実際は比率（0〜1）を受け付ける設計は混乱を招く。
以下のいずれかに統一する：

**案 A: ラベルを「割引率(0〜1)」に変更**（最小変更）

**案 B: UI入力を % 表記（0〜100）にして内部で 0〜1 に変換**（推奨）
```typescript
// 表示: 10（%入力）
// 保存: 0.1（API送信）
if (field === "discountRate") {
  input.discount_rate = Number(value) / 100;
}
```

案 B の場合、表示時も `discountRate * 100` に変換する。

### 3. `MedicalRecordBillCheck` / `MedicalRecordEstimate` にも同様の修正を確認

同じ `handleUpdateItem` パターンを使っている他コンポーネントにも `discountRate` 分岐漏れがないか確認する。

---

## 受入条件

- [ ] 「割引(%)」フィールドに 10 を入力すると値が保持される
- [ ] PATCH リクエストに `discount_rate` が含まれる
- [ ] 小計に割引が正しく反映される
- [ ] ラベルと入力値の単位が一致している
- [ ] `docker compose exec frontend pnpm lint` エラー 0 件
