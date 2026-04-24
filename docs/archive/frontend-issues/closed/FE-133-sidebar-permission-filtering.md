# FE-133: サイドバーメニューを権限でフィルタリング

**Status**: Open
**Priority**: Medium
**Affects**: components/shared/Sidebar.tsx（または類似のサイドバーコンポーネント）
**Date Created**: 2026-03-29
**Related**: BUG-056, TASK-048, FE-134

---

## Summary

サイドバーのメニュー項目が、ログインユーザーの権限に関わらず全件表示されている。
`canView = false` のリソースはメニューから非表示にすべきだ。
現状は **アクセス権のないページへのリンクが表示され、遷移すると空白またはエラーになる** という UX 問題がある。

セキュリティ上の主防衛線はバックエンドの認可ミドルウェア（BUG-056）だが、
フロントエンドのメニュー表示はユーザー体験に直結するため独立して修正する。

---

## 現状

### サイドバーの問題

```tsx
// 現状（推定）: 全メニュー項目を無条件に表示
const SIDEBAR_MENU_ITEMS = [
  { label: "ダッシュボード",  path: "/dashboard",          resource: "dashboard" },
  { label: "オーナー管理",   path: "/owners",              resource: "owners" },
  { label: "予約",          path: "/reservations",         resource: "reservations" },
  { label: "カルテ",        path: "/medical-records",      resource: "medical-records" },
  { label: "入院・ホテル",   path: "/hospitalization",     resource: "hospitalization" },
  { label: "トリミング",     path: "/trimming",            resource: "trimming" },
  { label: "検査",          path: "/examinations",         resource: "examinations" },
  { label: "会計",          path: "/accounting",           resource: "accounting" },
  { label: "ワクチン",      path: "/vaccinations",         resource: "vaccinations" },
  { label: "健診",          path: "/checkups",             resource: "checkups" },
  { label: "在庫",          path: "/inventory",            resource: "inventory" },
  { label: "見積",          path: "/estimates",            resource: "estimates" },
  { label: "シフト",        path: "/shifts",               resource: "shifts" },
  { label: "マスタ管理",    path: "/master",               resource: "master" },
  { label: "病院設定",      path: "/hospital-settings",   resource: "hospital-settings" },
];

export function Sidebar() {
  return (
    <nav>
      {SIDEBAR_MENU_ITEMS.map(item => (
        <SidebarLink key={item.path} to={item.path}>{item.label}</SidebarLink>
      ))}
    </nav>
  );
}
```

`clinic_admin`・`system_admin` は全リソースに常にアクセス可能なため問題ないが、
`staff` ユーザーは自分に付与されていないリソースのリンクも表示される。

---

## 実装方針

### `usePermission` を使ってメニューをフィルタリング

```tsx
// components/shared/Sidebar.tsx（変更後）
import { usePermission } from "@/features/auth/hooks/use-permission";
import type { Resource } from "@/types/generated/models";

interface SidebarMenuItem {
  label: string;
  path: string;
  resource: Resource;
  icon?: React.ReactNode;
}

const SIDEBAR_MENU_ITEMS: SidebarMenuItem[] = [
  { label: "ダッシュボード",  path: "/dashboard",         resource: "dashboard" },
  { label: "オーナー管理",   path: "/owners",             resource: "owners" },
  { label: "予約",          path: "/reservations",        resource: "reservations" },
  { label: "カルテ",        path: "/medical-records",     resource: "medical-records" },
  { label: "入院・ホテル",   path: "/hospitalization",    resource: "hospitalization" },
  { label: "トリミング",     path: "/trimming",           resource: "trimming" },
  { label: "検査",          path: "/examinations",        resource: "examinations" },
  { label: "会計",          path: "/accounting",          resource: "accounting" },
  { label: "ワクチン",      path: "/vaccinations",        resource: "vaccinations" },
  { label: "健診",          path: "/checkups",            resource: "checkups" },
  { label: "在庫",          path: "/inventory",           resource: "inventory" },
  { label: "見積",          path: "/estimates",           resource: "estimates" },
  { label: "シフト",        path: "/shifts",              resource: "shifts" },
  { label: "マスタ管理",    path: "/master",              resource: "master" },
  { label: "病院設定",      path: "/hospital-settings",  resource: "hospital-settings" },
];
```

**重要**: メニュー項目を `usePermission` でフィルタリングする場合、hooks のルールによりループ内で
hooks を呼べないため、各メニュー項目をサブコンポーネントに切り出す。

```tsx
// コンポーネント分離パターン（hooks をループ内で呼ぶ問題を回避）
const SidebarMenuItemComponent = memo(function SidebarMenuItemComponent({
  label,
  path,
  resource,
}: SidebarMenuItem) {
  const { canView } = usePermission(resource);
  if (!canView) return null;
  return <SidebarLink to={path}>{label}</SidebarLink>;
});

export function Sidebar() {
  return (
    <nav>
      {SIDEBAR_MENU_ITEMS.map(item => (
        <SidebarMenuItemComponent key={item.path} {...item} />
      ))}
    </nav>
  );
}
```

### FE-132 完了後の型強化

BE-077 / FE-132 完了後、`resource` の型を文字列リテラルから `Resource` 型定数に変更する。

```tsx
// FE-132 完了後
import {
  ResourceDashboard, ResourceOwners, ResourceReservations,
  // ... 全定数
} from "@/types/generated/models";

const SIDEBAR_MENU_ITEMS: SidebarMenuItem[] = [
  { label: "ダッシュボード", path: "/dashboard",    resource: ResourceDashboard },
  { label: "オーナー管理",  path: "/owners",        resource: ResourceOwners },
  // ...
];
```

---

## 注意点

### clinic_admin / system_admin は全メニュー表示

`usePermission` の内部実装で `userType === "clinic_admin"` または `"system_admin"` の場合は
`canView = true` を返すため、管理者は従来通り全メニューが表示される。フィルタリングは `staff` のみに効く。

### ローディング中の表示

`useAuth()` の初期化中（`user` が未定の状態）は `canView` が `false` になるケースがある。
ローディングスケルトンを表示するなど UX を考慮する。

---

## 変更ファイル一覧

| ファイル | 変更内容 |
|---------|---------|
| `components/shared/Sidebar.tsx` | `SidebarMenuItemComponent` に分割し `usePermission` でフィルタ |
| （FE-132 完了後）`components/shared/Sidebar.tsx` | `resource` 文字列を `ResourceXxx` 定数に置換 |

---

## 受入条件

- [ ] `canView = false` のリソースのメニュー項目が `staff` ユーザーに表示されない
- [ ] `clinic_admin` / `system_admin` は全メニューが表示される
- [ ] 権限が変更された後（ページリロードまたは `/me` 再取得後）にメニューが正しく更新される
- [ ] `docker compose exec frontend pnpm build` 成功
- [ ] `docker compose exec frontend pnpm lint` エラー 0 件
- [ ] `docker compose exec frontend pnpm test:run` 成功
