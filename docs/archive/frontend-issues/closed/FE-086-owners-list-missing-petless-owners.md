# FE-086: ペットなし飼主が飼主一覧に表示されない

**Status**: Open
**Priority**: High
**Affects**: 飼主一覧 (`/owners`)
**Date Created**: 2026-03-21
**Related**: BUG-001

## Summary

`features/owners/loaders.ts` の `ownersLoader` が `/v1/pets` エンドポイントからペット一覧を取得し、`OwnersList.tsx` でペット行ベースに一覧を表示しているため、ペットが0件の飼主が一覧に表示されない。

## 現状のコード

```typescript
// frontend/src/features/owners/loaders.ts:21-45
export const ownersLoader = async (): Promise<OwnersLoaderData> => {
  // /v1/pets を取得 → ペットなし飼主はここに含まれない
  const { data: firstPage } = await axios.get<PetsResponse>("/v1/pets", {
    params: { page: 1, limit: PER_PAGE },
  });
  // ...
  const allPets: Pet[] = [firstPage, ...remainingPages]
    .flatMap(page => page.data.map(transformBackendPetToFrontend));

  return { pets: allPets };  // ← ペットなし飼主は pets に含まれない
};
```

```typescript
// frontend/src/features/owners/routes/OwnersList.tsx:123
const { pets } = useLoaderData<OwnersLoaderData>();
// ペット行ベースで一覧表示 → ペットなし飼主は表示されない
```

## 必要な変更

### 1. loaders.ts を owners ベースに変更

```typescript
// frontend/src/features/owners/loaders.ts

import { transformBackendOwnerToFrontend } from "@/lib/transforms/owner"; // 存在確認必要
import type { Owner as BackendOwner } from "@/types/generated/models";

interface OwnersResponse {
  data: BackendOwner[];
  total: number;
  page: number;
  limit: number;
}

export interface OwnersLoaderData {
  owners: Owner[];  // pets[] から owners[] に変更
}

export const ownersLoader = async (): Promise<OwnersLoaderData> => {
  try {
    const { data: firstPage } = await axios.get<OwnersResponse>("/v1/owners", {
      params: { page: 1, limit: PER_PAGE },
    });

    const totalPages = Math.ceil(firstPage.total / PER_PAGE);

    const remainingPages = await Promise.all(
      Array.from({ length: totalPages - 1 }, (_, i) =>
        axios.get<OwnersResponse>("/v1/owners", {
          params: { page: i + 2, limit: PER_PAGE },
        }).then(r => r.data)
      )
    );

    const allOwners: Owner[] = [firstPage, ...remainingPages]
      .flatMap(page => page.data.map(transformBackendOwnerToFrontend));

    return { owners: allOwners };
  } catch {
    throw new Response("飼主一覧の取得に失敗しました", { status: 500 });
  }
};
```

### 2. OwnersList.tsx を owners ベースの表示に変更

現在は `pets` をフラットに並べているが、`owners` を行単位で表示し、各行にペット情報を展開する形式にする。

- `useLoaderData<OwnersLoaderData>()` から `owners` を取得
- ペットなし飼主は空のペット列で1行表示
- 各オーナー内の複数ペットは展開するかどうかはUIデザインに依存（現在の動作に合わせてペット行展開を維持しつつ、ペットなし飼主も1行表示する）

主な変更箇所:
- `const { pets } = useLoaderData<OwnersLoaderData>()` → `const { owners } = useLoaderData<OwnersLoaderData>()`
- フィルタリング・ソートロジックをオーナーベースに変更
- `SortKey` の `ownerNumber` / `ownerName` はオーナーレベル、`name` / `species` / `birthDate` / `lastVisit` はペットレベルで処理

### 3. `lib/transforms/owner.ts` 確認

`transformBackendOwnerToFrontend` が存在するか確認し、なければ作成する。

## UIフロー

1. ペットなし飼主を登録後、飼主一覧を開く
2. 該当飼主が1行（ペットカラム空欄）で表示される
3. 行クリックで `/owners/:id` に遷移できる

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] 型は `models.ts` から導出

## 依存関係

- バックエンド `/v1/owners` エンドポイントが Owner ネスト Pets を返すこと（確認済み: `Preload("Pets")` 実装済み）

## 完了条件

- [ ] ペットなし飼主が飼主一覧に表示される（ペットカラム空欄で1行）
- [ ] ペットあり飼主は従来通り表示される
- [ ] `pnpm build` が通る
- [ ] `pnpm lint` がエラーなし
