# BUG-187: router.tsx の 5 ルートに errorElement 未設定（エラー時に全画面クラッシュ）

## 概要

`frontend/src/app/router.tsx` において `/checkups`・`/estimates`・`/shifts`・`/settings/*`・`/settings/clinic` の 5 ルートに `errorElement` が設定されていない。これらのルートでランタイムエラーが発生した場合、`<RouteErrorBoundary />` がなければ React Router がエラーをバブルアップさせ、上位の `<RootErrorBoundary />` まで到達して**アプリ全体がクラッシュ画面になる**。BUG-166（予約の lazy コンポーネント）と同カテゴリの問題。

## 再現手順

1. `/checkups`（定期健診）にアクセスし、API が 500 エラーを返すように再現する
2. **結果**: checkups ルート固有のエラーページではなく、アプリ全体がクラッシュ画面になる（Root ErrorBoundary まで伝播）

## 期待する動作

```tsx
// ✅ 全ルートに errorElement が設定されるべき
{
  path: "/checkups",
  errorElement: <RouteErrorBoundary />,  // ← 必須
  children: [...]
}
```

## 現状コード

### `frontend/src/app/router.tsx`（対象5ルート）

```tsx
// ❌ /checkups — errorElement なし
{
  path: "/checkups",
  element: <RequirePermission resource={ResourceCheckups}>...<RequirePermission />,
  // errorElement: <RouteErrorBoundary /> が欠落
  children: [...]
}

// ❌ /estimates — errorElement なし
{
  path: "/estimates",
  element: <RequirePermission resource={ResourceEstimates}>...</RequirePermission>,
  // errorElement: <RouteErrorBoundary /> が欠落
  children: [...]
}

// ❌ /shifts — errorElement なし
{
  path: "/shifts",
  element: <RequirePermission resource={ResourceShifts}>...</RequirePermission>,
  // errorElement: <RouteErrorBoundary /> が欠落
  children: [...]
}

// ❌ /settings/* — errorElement なし
{
  path: "/settings",
  element: <MasterSettingsIndex />,
  // errorElement: <RouteErrorBoundary /> が欠落
  children: [...]
}

// ❌ /settings/clinic — errorElement なし
{
  path: "/settings/clinic",
  element: <CompanySettings />,
  // errorElement: <RouteErrorBoundary /> が欠落
}
```

### 比較: 正しい実装（参照実装）
```tsx
// ✅ /owners — errorElement 設定済み
{
  path: "/owners",
  errorElement: <RouteErrorBoundary />,
  element: <RequirePermission resource={ResourceOwners}>...</RequirePermission>,
  children: [...]
}

// ✅ /medical-records — errorElement 設定済み
{
  path: "/medical-records",
  errorElement: <RouteErrorBoundary />,
  children: [...]
}
```

## 影響範囲

| ルート | errorElement | 状態 |
|---|---|---|
| `/checkups` | ❌ 欠如 | 未修正 |
| `/estimates` | ❌ 欠如 | 未修正 |
| `/shifts` | ❌ 欠如 | 未修正 |
| `/settings/*` | ❌ 欠如 | 未修正 |
| `/settings/clinic` | ❌ 欠如 | 未修正 |
| `/owners`, `/reservations`, `/medical-records`, `/hospitalization`, etc. | ✅ 設定済み | 対応不要 |

## 修正方針

### `router.tsx` — 5ルートに `errorElement` を追加

```tsx
import { RouteErrorBoundary } from "@/components/errors/RouteErrorBoundary";

// /checkups
{
  path: "/checkups",
  errorElement: <RouteErrorBoundary />,
  element: <RequirePermission resource={ResourceCheckups}>...</RequirePermission>,
  children: [...]
}

// /estimates
{
  path: "/estimates",
  errorElement: <RouteErrorBoundary />,
  element: <RequirePermission resource={ResourceEstimates}>...</RequirePermission>,
  children: [...]
}

// /shifts
{
  path: "/shifts",
  errorElement: <RouteErrorBoundary />,
  element: <RequirePermission resource={ResourceShifts}>...</RequirePermission>,
  children: [...]
}

// /settings
{
  path: "/settings",
  errorElement: <RouteErrorBoundary />,
  element: <MasterSettingsIndex />,
  children: [...]
}

// /settings/clinic
{
  path: "/settings/clinic",
  errorElement: <RouteErrorBoundary />,
  element: <CompanySettings />,
}
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/error-handling.md` — React Error Boundary
> React Error Boundary を実装すること。

### プロジェクト内参照実装
- `app/router.tsx` — `/owners`, `/medical-records` 等の `errorElement: <RouteErrorBoundary />` 設定済みパターン

## 優先度
**High** — `/estimates`（見積）・`/shifts`（シフト管理）・`/settings`（マスタ設定）はすべて業務上重要な画面であり、エラー時に全画面クラッシュが発生する。1ファイルの修正で5ルートが保護される。

## 関連チケット
- BUG-166: ReservationManagement lazy コンポーネントの ErrorBoundary 欠如（同カテゴリ）

## 関連ファイル
- `frontend/src/app/router.tsx`
- `frontend/src/components/errors/RouteErrorBoundary.tsx`
