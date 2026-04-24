# FE-132: 権限リソース定数を生成 models.ts に移行・permission-resources.ts 廃止

**Status**: Open
**Priority**: High
**Affects**: features/master/types/permission-resources.ts, app/router.tsx, features/auth/hooks/use-permission.ts, features/master/components/PermissionRuleTable.tsx, features/master/routes/PermissionGroupSettings.tsx
**Date Created**: 2026-03-29
**Related**: TASK-048, BE-077（先に完了必要）
**Blocked by**: BE-077（`make codegen` で `models.ts` に `ResourceXxx` 定数が出力されることが前提）

---

## Summary

BE-077 完了後、`make codegen` で `frontend/src/types/generated/models.ts` に
`ResourceXxx` 定数が出力される。それを参照するよう各ファイルを更新し、
手書きの `permission-resources.ts` の `key` フィールドを廃止する。

---

## 現状

### `permission-resources.ts`（廃止対象部分）

```typescript
// frontend/src/features/master/types/permission-resources.ts
export const PERMISSION_RESOURCES: ResourceDefinition[] = [
  { key: "dashboard",         label: "ダッシュボード" },
  { key: "owners",            label: "オーナー管理" },
  { key: "reservations",      label: "予約" },
  { key: "medical-records",   label: "カルテ" },
  { key: "hospitalization",   label: "入院・ホテル" },
  { key: "trimming",          label: "トリミング" },
  { key: "examinations",      label: "検査" },
  { key: "accounting",        label: "会計" },
  { key: "vaccinations",      label: "ワクチン" },
  { key: "checkups",          label: "健診" },
  { key: "inventory",         label: "在庫" },
  { key: "estimates",         label: "見積" },
  { key: "shifts",            label: "シフト" },
  { key: "master",            label: "マスタ管理" },
  { key: "hospital-settings", label: "病院設定" },
];
```

`key` フィールドが文字列リテラルのハードコード。

### `router.tsx`（文字列リテラル直書き）

```tsx
// 現状: 文字列リテラル
<RequirePermission resource="medical-records">
<RequirePermission resource="accounting">
<RequirePermission resource="owners">
// ... 全15箇所
```

---

## 実装手順

### 前提確認

BE-077 完了後に `make codegen` を実行し、以下が `models.ts` に追加されていることを確認する。

```typescript
// frontend/src/types/generated/models.ts（自動生成・編集禁止）
export type Resource = string
export const ResourceDashboard       = "dashboard"
export const ResourceOwners          = "owners"
export const ResourceReservations    = "reservations"
export const ResourceMedicalRecords  = "medical-records"
export const ResourceHospitalization = "hospitalization"
export const ResourceTrimming        = "trimming"
export const ResourceExaminations    = "examinations"
export const ResourceAccounting      = "accounting"
export const ResourceVaccinations    = "vaccinations"
export const ResourceCheckups        = "checkups"
export const ResourceInventory       = "inventory"
export const ResourceEstimates       = "estimates"
export const ResourceShifts          = "shifts"
export const ResourceMaster          = "master"
export const ResourceHospitalSettings = "hospital-settings"
```

### 1. `permission-resources.ts` を label のみのマッピングに縮小

`key` を `models.ts` の定数で参照し、`label`（日本語表示名）のみ管理するファイルに変更する。

```typescript
// frontend/src/features/master/types/permission-resources.ts（変更後）
import {
  ResourceDashboard, ResourceOwners, ResourceReservations,
  ResourceMedicalRecords, ResourceHospitalization, ResourceTrimming,
  ResourceExaminations, ResourceAccounting, ResourceVaccinations,
  ResourceCheckups, ResourceInventory, ResourceEstimates,
  ResourceShifts, ResourceMaster, ResourceHospitalSettings,
  type Resource,
} from "@/types/generated/models";

export interface ResourceDefinition {
  key: Resource;
  label: string;
}

export const PERMISSION_RESOURCES: ResourceDefinition[] = [
  { key: ResourceDashboard,        label: "ダッシュボード" },
  { key: ResourceOwners,           label: "オーナー管理" },
  { key: ResourceReservations,     label: "予約" },
  { key: ResourceMedicalRecords,   label: "カルテ" },
  { key: ResourceHospitalization,  label: "入院・ホテル" },
  { key: ResourceTrimming,         label: "トリミング" },
  { key: ResourceExaminations,     label: "検査" },
  { key: ResourceAccounting,       label: "会計" },
  { key: ResourceVaccinations,     label: "ワクチン" },
  { key: ResourceCheckups,         label: "健診" },
  { key: ResourceInventory,        label: "在庫" },
  { key: ResourceEstimates,        label: "見積" },
  { key: ResourceShifts,           label: "シフト" },
  { key: ResourceMaster,           label: "マスタ管理" },
  { key: ResourceHospitalSettings, label: "病院設定" },
];
```

`label` は UI 表示専用のため、Go 側で管理する必要がなく `permission-resources.ts` に残すのが適切。

### 2. `router.tsx` の文字列リテラルを定数に置換

```tsx
// 変更前
import { /* ... */ } from "react-router";

// 変更後: models.ts から定数を import
import {
  ResourceDashboard, ResourceOwners, ResourceReservations,
  ResourceMedicalRecords, ResourceHospitalization, ResourceTrimming,
  ResourceExaminations, ResourceAccounting, ResourceVaccinations,
  ResourceCheckups, ResourceInventory, ResourceEstimates,
  ResourceShifts, ResourceMaster, ResourceHospitalSettings,
} from "@/types/generated/models";

// 変更前
<RequirePermission resource="medical-records">
<RequirePermission resource="accounting">
<RequirePermission resource="owners">

// 変更後
<RequirePermission resource={ResourceMedicalRecords}>
<RequirePermission resource={ResourceAccounting}>
<RequirePermission resource={ResourceOwners}>
```

**対象箇所**: `router.tsx` 内の全 `RequirePermission resource="..."` を定数に置換（15箇所）。

### 3. `usePermission` 呼び出し箇所の置換

`usePermission("...")` の文字列リテラルを定数に置換する。
`grep -r 'usePermission("' frontend/src/` で全箇所を特定してから修正する。

```typescript
// 変更前
const { canView, canCreate } = usePermission("medical-records");

// 変更後
import { ResourceMedicalRecords } from "@/types/generated/models";
const { canView, canCreate } = usePermission(ResourceMedicalRecords);
```

### 4. `RequirePermission` の Props 型を強化

`resource` prop の型を `string` から `Resource` に変更し、無効なキーを型レベルで弾く。

```typescript
// frontend/src/components/shared/RequirePermission.tsx
import type { Resource } from "@/types/generated/models";

interface RequirePermissionProps {
  resource: Resource;   // string → Resource 型に変更
  action?: ResourceAction;
  children: React.ReactNode;
  fallback?: React.ReactNode;
}
```

### 5. `hasPermission` の引数型を強化

```typescript
// frontend/src/features/auth/hooks/use-auth.tsx
import type { Resource } from "@/types/generated/models";

// 変更前
hasPermission: (resource: string, action: ResourceAction) => boolean;

// 変更後
hasPermission: (resource: Resource, action: ResourceAction) => boolean;
```

---

## 変更ファイル一覧

| ファイル | 変更内容 |
|---------|---------|
| `features/master/types/permission-resources.ts` | `key` を定数参照に変更（`label` は残す） |
| `app/router.tsx` | 文字列リテラル → `ResourceXxx` 定数に置換（15箇所） |
| `components/shared/RequirePermission.tsx` | `resource: string` → `resource: Resource` に型強化 |
| `features/auth/hooks/use-auth.tsx` | `hasPermission` の引数型を `Resource` に変更 |
| `features/auth/hooks/use-permission.ts` | 引数型を `Resource` に変更 |
| `features/auth/types/index.ts` | `hasPermission` シグネチャの型を更新 |
| 各 feature の `usePermission("...")` 呼び出し | 文字列リテラル → 定数に置換 |

---

## 受入条件

- [ ] `permission-resources.ts` の `key` フィールドが `models.ts` の定数を参照している
- [ ] `router.tsx` に `resource="..."` の文字列リテラルが残っていない
- [ ] `RequirePermission` の `resource` prop 型が `Resource` になっており、不正な文字列を渡すと TypeScript エラーになる
- [ ] `hasPermission` の第一引数型が `Resource` になっている
- [ ] 存在しないリソースキー（例: `"invalid-key"`）を `usePermission` に渡すと TypeScript エラーになる
- [ ] `docker compose exec frontend pnpm build` 成功
- [ ] `docker compose exec frontend pnpm lint` エラー 0 件
- [ ] `docker compose exec frontend pnpm test:run` 成功
