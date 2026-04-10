# BUG-296: Feature index.ts — 外部未使用のover-export

## 概要

複数のFeatureの `index.ts` が外部から一度もインポートされていないシンボルをexportしている。Feature Indexingの意図はPublic APIの明確化だが、実際に使われていないexportは認知負荷を増やすだけである。

## 問題箇所

### auth/index.ts（5件）
```typescript
export { ME_QUERY_KEY } from "./api/get-me";         // 外部使用: 0件
export type { AuthUser } from "./types";              // 外部使用: 0件
export type { AuthContextValue } from "./types";      // 外部使用: 0件
export type { ResourcePermission } from "./types";    // 外部使用: 0件
export type { ResourcePermissions } from "./types";   // 外部使用: 0件
```

### owners/index.ts（3件）
```typescript
export { getOwner } from "./api/get-owner";    // 外部使用: 0件
export { getOwners } from "./api/get-owners";  // 外部使用: 0件
export { useGetOwners } from "./api/get-owners"; // 外部使用: 0件（app/pages経由で使用）
```

### inventory/index.ts（2件）
```typescript
export { useInventory } from "./hooks/use-inventory";     // 外部使用: 0件
export { useInventoryForm } from "./hooks/use-inventory-form"; // 外部使用: 0件
```

### estimates/index.ts（1件）
```typescript
export { useEstimateForm } from "./hooks/use-estimate-form"; // 外部使用: 0件
```

### vaccinations/index.ts（1件）
```typescript
export { useCreateVaccination } from "./api/create-vaccination"; // 外部使用: 0件
```

### line-reservation/index.ts（4件）
```typescript
export { useGetReservationSetting } from "./api/...";    // 外部使用: 0件
export { useUpdateReservationSetting } from "./api/..."; // 外部使用: 0件
export { useGetReservationCustomers } from "./api/...";  // 外部使用: 0件
export { useUpdateOwnerLink } from "./api/...";          // 外部使用: 0件
```

## 判断

- hooks（`useXxx`）はルール上「Feature内部で使うもの」であり index.ts からのexportは不要
- 生関数（`getOwner`, `getOwners`）は loader 等で直接利用される場合があるため削除前に要確認
- TypeScript型（`AuthUser`等）は router.tsx や他のコンポーネントで暗黙的に使われる場合があるため慎重に対応

## ステータス

- [x] ドキュメント作成
- [ ] 実装（各Feature index.ts の精査と不要exportの削除）
