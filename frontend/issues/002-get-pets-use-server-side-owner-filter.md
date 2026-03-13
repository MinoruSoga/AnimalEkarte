# GET /v1/pets?owner_id= — サーバー側フィルタをフロントエンドが使用していない

## 背景

バックエンドの `GET /v1/pets` は `?owner_id=` クエリパラメータによるサーバー側フィルタをすでに実装している。
しかしフロントエンドの `useGetPets(ownerId?)` はこれを使わず、全件取得してからクライアント側でフィルタしている。

```typescript
// frontend/src/features/pets/api/get-pets.ts（現状）
export const getPets = async (): Promise<Pet[]> => {
  const { data } = await axios.get<PetsListResponse>("/v1/pets");  // owner_id 未使用
  return data.data.map(transformBackendPetToFrontend);
};

export const useGetPets = (ownerId?: string) => {
  return useQuery({
    queryKey: ["pets"],          // ownerId がキーに入っていない
    queryFn: getPets,
    select: (pets) => {
      if (!ownerId) return pets;
      return pets.filter((pet) => pet.ownerId === ownerId);  // クライアント側フィルタ
    },
  });
};
```

## 問題

1. **不要な全件取得**: ペットが増えるほど転送量・メモリ使用量が増える
2. **queryKey の不整合**: `ownerId` が変わってもキャッシュキーが `["pets"]` のまま。異なる `ownerId` で同一キャッシュを共有し、`select` が変わっても再フェッチしない

## 修正方針

```typescript
// get-pets.ts
export const getPets = async (ownerId?: string): Promise<Pet[]> => {
  const params = ownerId ? { owner_id: ownerId } : {};
  const { data } = await axios.get<PetsListResponse>("/v1/pets", { params });
  return data.data.map(transformBackendPetToFrontend);
};

export const useGetPets = (ownerId?: string) => {
  return useQuery({
    queryKey: ownerId ? ["pets", { ownerId }] : ["pets"],
    queryFn: () => getPets(ownerId),
  });
};
```

## 前提

バックエンドの `GET /v1/pets?owner_id=` は実装済み（対応不要）。
