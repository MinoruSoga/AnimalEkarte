# BUG-051: カルテ治療プランの割引率(%)フィールドが非機能

## 概要
カルテ編集画面の「診察/治療プランタブ」にある治療プランテーブルの「割引(%)」フィールドが、入力しても即座に 0 にリセットされ、値が保持されない。

## 再現手順
1. `/medical-records/:id`（作成中ステータスのカルテ）を開く
2. 「診察/治療プラン」タブをクリック
3. 「治療プラン」テーブルの任意の行の「割引(%)」フィールドに数値を入力する（例: 10）
4. Tabキーまたは他のフィールドをクリックしてフォーカスを外す

## 期待動作
- 入力した割引率(%)が保持される
- 小計に `Math.floor(subtotal × discountRate)` が適用される

## 実際の動作
- 入力値が即座に 0 にリセットされる
- 割引率が一切適用されない

## 原因（コード）
`frontend/src/features/medical-records/components/MedicalRecordDiagnosisPlan.tsx` の `handleUpdateItem`（91〜103行目）:

```typescript
const handleUpdateItem = useCallback((id: number, field: keyof TreatmentItem, value: ...) => {
  const input: UpdateTreatmentInput = {};
  if (field === "content") input.content = String(value);
  if (field === "memo") input.memo = String(value);
  if (field === "insurance") input.insurance = Boolean(value);
  if (field === "unitPrice") input.unit_price = Number(value);
  if (field === "quantity") input.quantity = Number(value);
  if (field === "discountAmount") input.discount_amount = Number(value);
  if (field === "status") input.status = String(value);
  if (field === "selected") input.selected = Boolean(value);
  // ← "discountRate" の分岐が存在しない
  updateMutation.mutate({ treatmentId: String(id), input });
}, [updateMutation]);
```

`field === "discountRate"` の分岐がなく、`UpdateTreatmentInput` に `discount_rate` が設定されないため、React が古い状態値（0）で再レンダリングする。

## 追加: ラベルとロジックの不一致
`TreatmentTable.tsx` のヘッダは「割引(%)」と表記しているが、計算式 `calcLineItemAmount` は `Math.floor(subtotal × discountRate)` であり、入力値を比率（0〜1）として扱う。
- 10% 割引を意図するなら入力値は「0.1」が正しい
- 「割引(%)」ラベルは誤解を招く（10と入力すると10倍の割引になる）

## 影響範囲
- カルテ治療プランの割引率フィールド（診察/治療プランタブ）
- `MedicalRecordBillCheck` / `MedicalRecordEstimate` も同様に `discountRate` 未処理の可能性あり

## 優先度
中（割引率機能が完全に未実装）

## 発見日
2026-03-29

## 派生イシュー

| イシュー | 領域 | 内容 |
|---------|------|------|
| FE-146 | Frontend | `handleUpdateItem` に `discountRate` 分岐追加・ラベル「割引(%)」と 0〜100 値の 0〜1 変換修正 |
