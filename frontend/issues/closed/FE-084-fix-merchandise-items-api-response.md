# FE-084: 物販マスタ API hook レスポンスパース修正

**Status**: Open
**Priority**: High
**Affects**: master feature（MerchandiseItemSettings）、accounting feature（会計物販モーダル）
**Date Created**: 2026-03-19
**Related**: TASK-023, BE-048

## Summary

merchandise-items API のフロントエンド API hook が `axios.get<MerchandiseItem[]>()` で直接配列を期待しているが、バックエンドは `{ data: [...], total, page, limit }` 形式（PaginatedResponse）を返す。BE-048 で配列形式に修正するが、FE 側でも PaginatedResponse の場合のフォールバック処理を入れて堅牢にする。

## 現状のコード

### master feature API（merchandise-items.ts:60-63）

```typescript
// frontend/src/features/master/api/merchandise-items.ts:60-63
export const getAllMerchandiseItems = async (): Promise<FrontendMerchandiseItem[]> => {
  const { data } = await axios.get<MerchandiseItem[]>("/v1/masters/merchandise-items");
  return data.map(transformMerchandiseItem);  // ← data がオブジェクトの場合 TypeError
};
```

### accounting feature API（get-merchandise-items.ts:29-31）

```typescript
// frontend/src/features/accounting/api/get-merchandise-items.ts:29-31
const { data } = await axios.get<MerchandiseItem[]>("/v1/masters/merchandise-items");
return data.map(transformMerchandiseItem);  // ← 同上
```

## 必要な変更

### 1. master feature API 修正

```typescript
// frontend/src/features/master/api/merchandise-items.ts
export const getAllMerchandiseItems = async (): Promise<FrontendMerchandiseItem[]> => {
  const { data } = await axios.get<MerchandiseItem[] | { data: MerchandiseItem[] }>("/v1/masters/merchandise-items");
  // BE-048 修正前は PaginatedResponse、修正後は直接配列
  const items = Array.isArray(data) ? data : data.data;
  return items.map(transformMerchandiseItem);
};
```

### 2. accounting feature API 修正

```typescript
// frontend/src/features/accounting/api/get-merchandise-items.ts
const { data } = await axios.get<MerchandiseItem[] | { data: MerchandiseItem[] }>("/v1/masters/merchandise-items");
const items = Array.isArray(data) ? data : data.data;
return items.map(transformMerchandiseItem);
```

## プロジェクトルール遵守チェック

- [x] `any` 型なし
- [x] barrel index 経由 import なし

## 依存関係

- BE-048 で配列形式に修正されれば、`Array.isArray` フォールバックは不要になるが、防御的に残しておく

## 完了条件

- [ ] `/settings/merchandise-items` にアクセスすると品目一覧が表示される
- [ ] 会計詳細の「物販・その他追加」モーダルで品目一覧が表示される
- [ ] `npm run build` パス
