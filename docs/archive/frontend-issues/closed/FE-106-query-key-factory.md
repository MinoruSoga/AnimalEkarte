# FE-106: queryKey ファクトリーパターン導入

**Status**: Closed
**Priority**: Medium
**Affects**: 全 feature の api/get-xxx.ts
**Date Created**: 2026-03-24
**Related**: TASK-024

## Summary

React Query の queryKey が feature ごとに命名規則がばらばらで、`["owners", id]` vs `["accounting-detail", id]` のような不統一が存在する。`queryClient.invalidateQueries` の効果範囲が曖昧になり、キャッシュ無効化バグの温床になっている。`queryKeys` ファクトリーオブジェクトを導入して統一する。

## 現状のコード

```typescript
// frontend/src/features/owners/api/get-owner.ts
queryKey: ["owners", id]              // 詳細: "owners" + id

// frontend/src/features/owners/api/get-owners.ts
queryKey: ["owners"]                  // リスト: "owners" のみ

// frontend/src/features/accounting/api/get-accounting.ts
queryKey: ["accounting-detail", id]   // ← 詳細に "-detail" サフィックス（命名不統一）

// frontend/src/features/accounting/api/get-accountings.ts
queryKey: ["accountings", { status }] // ← 複数形 "accountings"（命名不統一）

// frontend/src/features/medical-records/api/get-medical-records.ts
queryKey: ["medical-records"]

// frontend/src/features/master/api/get-master-items.ts
queryKey: ["masterItems", category]   // ← camelCase（他は kebab-case）

// frontend/src/features/dashboard/api/get-staffs.ts
queryKey: ["masters", "staffs"]       // ← masters 配下

// frontend/src/features/owners/api/get-animal-species.ts
queryKey: ["masters", "animal-species"] // ← masters 配下（masterItems と不統一）
```

**問題例**: `accounting` の詳細を削除後に `invalidateQueries({ queryKey: ["accountings"] })` を呼んでも、詳細キー `["accounting-detail", id]` は無効化されない（異なるキー体系のため）。

## 必要な変更

### 1. queryKeys ファクトリー作成

```typescript
// frontend/src/lib/query-keys.ts（新規作成）

export const queryKeys = {
  owners: {
    all: () => ["owners"] as const,
    detail: (id: string) => ["owners", id] as const,
  },
  pets: {
    byOwner: (ownerId: string) => ["pets", { ownerId }] as const,
    detail: (id: string) => ["pets", id] as const,
  },
  accountings: {
    all: (params?: { status?: string }) => ["accountings", ...(params ? [params] : [])] as const,
    detail: (id: string) => ["accountings", id] as const,  // "accounting-detail" → "accountings", id に統一
  },
  medicalRecords: {
    all: () => ["medical-records"] as const,
    detail: (id: string) => ["medical-records", id] as const,
  },
  vaccinations: {
    all: () => ["vaccinations"] as const,
    detail: (id: string) => ["vaccinations", id] as const,
  },
  inventory: {
    all: () => ["inventory"] as const,
    detail: (id: string) => ["inventory", id] as const,
  },
  trimming: {
    all: () => ["trimming"] as const,
    detail: (id: string) => ["trimming", id] as const,
  },
  estimates: {
    all: () => ["estimates"] as const,
    detail: (id: string) => ["estimates", id] as const,
  },
  reservations: {
    all: () => ["reservations"] as const,
    detail: (id: string) => ["reservations", id] as const,
  },
  hospitalization: {
    all: () => ["hospitalization"] as const,
    detail: (id: string) => ["hospitalization", id] as const,
  },
  masters: {
    all: (category: string) => ["masters", category] as const,
    // masterItems → masters に統一
  },
  staffs: {
    all: () => ["masters", "staffs"] as const,
  },
  checkups: {
    all: () => ["checkups"] as const,
    detail: (id: string) => ["checkups", id] as const,
  },
} as const;
```

### 2. 各 api ファイルの queryKey を置き換え

```typescript
// frontend/src/features/accounting/api/get-accounting.ts
// Before:
queryKey: ["accounting-detail", id]

// After:
import { queryKeys } from "@/lib/query-keys";
queryKey: queryKeys.accountings.detail(id)
```

```typescript
// frontend/src/features/master/api/get-master-items.ts
// Before:
queryKey: ["masterItems", category]

// After:
queryKey: queryKeys.masters.all(category)
```

同様の置き換えを全 feature の `api/get-*.ts` および `api/create-*.ts` / `api/update-*.ts` / `api/delete-*.ts` の `invalidateQueries` 呼び出しに実施。

## 影響

- `queryClient.invalidateQueries({ queryKey: queryKeys.accountings.all() })` により、詳細も含めた全 accountings キャッシュが確実に無効化される（TanStack Query の階層一致による）

## プロジェクトルール遵守チェック

- [ ] `any` 型なし（`as const` で型安全）
- [ ] barrel index 経由 import なし（`lib/query-keys` を直接 import）

## 依存関係

- Backend 変更なし。他の FE イシューとも独立。

## 完了条件

- [ ] `frontend/src/lib/query-keys.ts` が作成されている
- [ ] `["accounting-detail", id]` のような非標準キーが削除されている
- [ ] `["masterItems", category]` が `queryKeys.masters.all(category)` に統一されている
- [ ] 全 feature の `invalidateQueries` が `queryKeys.*` を使用している
- [ ] キャッシュ無効化の動作が変化なし（削除後にリストが再取得される）
- [ ] `pnpm lint` エラーなし
- [ ] `pnpm build` エラーなし
