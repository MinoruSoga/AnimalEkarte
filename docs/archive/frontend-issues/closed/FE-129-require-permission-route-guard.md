# FE-129: RequirePermission コンポーネント + ルートガード適用

**Status**: Closed
**Priority**: High
**Affects**: components/shared/RequirePermission.tsx（新規）, app/router.tsx
**Date Created**: 2026-03-26
**Related**: TASK-032, FE-128（先に完了必要）

## Summary

ページアクセス制御用の `RequirePermission` コンポーネントを作成し、`router.tsx` の各ルートに適用する。
`view` 権限がないページへのアクセスはアクセス拒否画面にリダイレクトする。

## 現状のコード

**`frontend/src/app/router.tsx`（現状: 権限チェックなし）:**
```typescript
// 現状は保護ルートが単純な AuthProvider ラップのみ
// 各ルートにアクセス制御はない
```

**既存の保護ルート構造（router.tsx）:**
```typescript
// protected layout（ログイン必須）は既に実装済み
// → この下に RequirePermission を追加する
```

## 必要な変更

### 1. RequirePermission コンポーネント（新規: `frontend/src/components/shared/RequirePermission.tsx`）

```typescript
import { usePermission } from "@/features/auth/hooks/use-permission";
import { useAuth } from "@/features/auth/hooks/use-auth";

interface RequirePermissionProps {
  resource: string;
  action?: "view" | "create" | "edit" | "delete";
  children: React.ReactNode;
  fallback?: React.ReactNode;
}

/**
 * 指定リソース×アクションの権限がない場合に fallback を表示する。
 * デフォルト action は "view"（ページアクセス制御に使用）。
 * fallback を省略すると <AccessDenied /> を表示する。
 */
export function RequirePermission({
  resource,
  action = "view",
  children,
  fallback,
}: RequirePermissionProps) {
  const { hasPermission } = useAuth();

  if (!hasPermission(resource, action)) {
    return fallback !== undefined ? <>{fallback}</> : <AccessDenied />;
  }

  return <>{children}</>;
}

function AccessDenied() {
  return (
    <div className="flex flex-col items-center justify-center h-full gap-4 text-center">
      <p className="text-lg font-medium text-foreground">アクセス権限がありません</p>
      <p className="text-sm text-muted-foreground">
        このページを表示するための権限が付与されていません。
      </p>
    </div>
  );
}
```

### 2. router.tsx へのルートガード適用

各ルートの Component を `RequirePermission` でラップする。
実装パターン: lazy import した Component を app/pages/ 内でラップするか、
ルートレベルで `element` に直接 `RequirePermission` を渡す。

**推奨実装（ルートレベルで適用）:**

```typescript
// router.tsx 内の各ルートを以下のパターンで変更

// Before（権限チェックなし）
{
  path: "accounting",
  lazy: async () => {
    const { AccountingList } = await import("@/features/accounting/routes/AccountingList");
    return { Component: AccountingList };
  },
},

// After（RequirePermission でラップ）
{
  path: "accounting",
  lazy: async () => {
    const [{ AccountingList }, { RequirePermission }] = await Promise.all([
      import("@/features/accounting/routes/AccountingList"),
      import("@/components/shared/RequirePermission"),
    ]);
    function GuardedAccountingList() {
      return (
        <RequirePermission resource="accounting" action="view">
          <AccountingList />
        </RequirePermission>
      );
    }
    return { Component: GuardedAccountingList };
  },
},
```

**ルートとリソース名の対応表:**

| ルートパス | resource |
|-----------|---------|
| `/` (dashboard) | `dashboard` |
| `/owners` | `owners` |
| `/reservations` | `reservations` |
| `/medical-records` | `medical-records` |
| `/hospitalization` | `hospitalization` |
| `/trimming` | `trimming` |
| `/examinations` | `examinations` |
| `/accounting` | `accounting` |
| `/vaccinations` | `vaccinations` |
| `/checkups` | `checkups` |
| `/inventory` | `inventory` |
| `/estimates` | `estimates` |
| `/shifts` | `shifts` |
| `/settings/*` (master) | `master` |
| `/settings/clinic` | `hospital-settings` |
| `/settings/permission-groups` | `master` |

### 3. Sidebar からアクセス不可メニューを非表示にする（オプション）

```typescript
// Sidebar.tsx や NavigationMenu.tsx でメニュー項目を権限チェックでフィルタ
import { usePermission } from "@/features/auth/hooks/use-permission";

function AccountingNavItem() {
  const { canView } = usePermission("accounting");
  return canView ? (
    <NavLink to="/accounting">会計</NavLink>
  ) : null;
}
```

サイドバーのメニュー非表示は UX 改善のため実装推奨。ルートガードとの二重実装になるが、
「アクセス試みた後に拒否」よりも「そもそも見えない」方がユーザビリティが高い。

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）

## 依存関係

- FE-128 が完了していること（`usePermission` hook が実装済み）

## 完了条件

- [ ] `RequirePermission` コンポーネントが `frontend/src/components/shared/RequirePermission.tsx` に存在する
- [ ] `view` 権限がないリソースのページにアクセスすると「アクセス権限がありません」が表示される
- [ ] 15個すべてのルートに `RequirePermission` が適用されている
- [ ] system_admin / clinic_admin は全ページにアクセスできる
- [ ] `pnpm build` で型エラーなし、`pnpm lint` パス

## クローズ情報

- **Closed At**: 2026-03-26
- **変更ファイル**:
  - `frontend/src/components/shared/RequirePermission.tsx` — 新規作成
  - `frontend/src/app/router.tsx` — 全ルートに RequirePermission（Outlet ラップ）パターン適用
