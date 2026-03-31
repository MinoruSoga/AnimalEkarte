---
status: closed
closed_at: 2026-03-16
---

# [master] CageSettings: price フィールドが string 型（HospitalizationSettings と不一致）

## 優先度
中

## 種別
型不一致・コード品質

## 対象ファイル
`frontend/src/features/master/routes/CageSettings.tsx`

## 問題

`CageFormData.price` が `string` 型で定義されており（L105）、`handleSave` で `Number(data.price)` に変換している。
`HospitalizationSettings.tsx` では同じ金額フィールドを `number` 型で統一して管理しており、一貫性がない。

```tsx
// CageSettings.tsx（現状）
interface CageFormData {
  price: string;  // ← string 型
}
const priceValue = data.price !== "" ? Number(data.price) : 0;  // 変換が必要

// HospitalizationSettings.tsx（正しいパターン）
interface HospitalizationFormData {
  price: number;  // ← number 型で統一
}
```

`Number("")` は `0` になるため、空文字のまま送信しても型エラーが出ず意図しない `price: 0` が送信されるリスクがある。

## 修正方針
1. `CageFormData.price` を `number` 型に変更
2. `useState` 初期値を `item?.price ?? 0` に変更
3. `input[type="number"]` の `value`/`onChange` を `number` 型として扱う（`valueAsNumber` または `Number(e.target.value)`）
4. `handleSave` の `Number(data.price)` 変換処理を削除

## 完了条件
- [x] `CageFormData.price` が `number` 型
- [x] `handleSave` に文字列→数値変換が不要
- [x] ビルドエラーなし
