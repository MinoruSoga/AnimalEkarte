---
status: closed
closed_at: 2026-03-13
commit: ef82ffa
---

# GET /v1/pets?owner_id= — サーバー側フィルタをフロントエンドが使用していない

## 背景

バックエンドの `GET /v1/pets` は `?owner_id=` クエリパラメータによるサーバー側フィルタをすでに実装している。
しかしフロントエンドの `useGetPets(ownerId?)` はこれを使わず、全件取得してからクライアント側でフィルタしていた。

## 対応内容

```typescript
// After
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

## 完了条件

- [x] `getPets` が `?owner_id=` パラメータをサーバーに送信している
- [x] `queryKey` に `ownerId` が含まれ、キャッシュが正しく分離されている
- [x] クライアント側 `select` フィルタを削除
