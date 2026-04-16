# BUG-292: accounting/api/get-merchandise-items.ts — queryKey衝突

## 概要

`accounting/api/get-merchandise-items.ts` と `master/api/merchandise-items.ts` が同一の `queryKey: ["masters", "merchandise-items"]` を使用している。BUG-291と同種の問題。

## 問題

```typescript
// frontend/src/features/accounting/api/get-merchandise-items.ts
export const useGetAllMerchandiseItems = () => {
  return useQuery({
    queryKey: ["masters", "merchandise-items"],  // ← master と同一
    queryFn: async () => {
      const { data } = await axios.get<MerchandiseItem[]>("/v1/masters/merchandise-items");
      return data;
    },
  });
};
```

- accountingとmasterで同一queryKeyを共有しているため、型不一致のキャッシュ汚染が発生し得る

## 修正

```typescript
// queryKeyをfeature固有に変更
queryKey: ["accounting", "merchandise-items"],
```

## ステータス

- [x] ドキュメント作成
- [x] 実装完了（`["masters", "merchandise-items"]` → `["accounting", "merchandise-items"]` に変更）
