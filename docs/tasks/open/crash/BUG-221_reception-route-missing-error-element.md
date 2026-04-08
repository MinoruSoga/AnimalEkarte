# BUG-221: Reception ルート `/` に errorElement 欠如

## 概要

`app/router.tsx:55-72` で Reception ルート（`path: "/"`）に `errorElement: <RouteErrorBoundary />` が設定されていない。他の全フィーチャールートは `<RouteErrorBoundary />` を持つが、Reception のみ欠落している。エラーが発生した場合、レイアウトなしの `<RootErrorBoundary />` にフォールバックし UX が一貫しない。

## 再現手順

1. Reception コンポーネント（`/`）でランタイムエラーを発生させる
2. **結果**: `<RootErrorBoundary />` が表示される（Layout なし・全画面エラー）
3. **期待**: `<RouteErrorBoundary />` が表示される（Layout 内にエラー UI）

## 期待する動作

- Reception ルートのエラーは他のフィーチャールートと同様に `<RouteErrorBoundary />` でハンドルされること

## 現状コード

### `frontend/src/app/router.tsx:55-72`
```tsx
// ── Reception（当日の受付）────────────────────────────────────
{
  path: "/",
  element: (
    <RequirePermission resource={ResourceReception}>
      <Outlet />
    </RequirePermission>
  ),
  // ❌ errorElement がない
  children: [
    {
      index: true,
      lazy: async () => {
        const { Reception } = await import("@/features/reception");
        return { Component: Reception };
      },
    },
  ],
},
```

### 比較: 正しい実装（Owners ルート）
```tsx
// router.tsx:75-83
{
  path: "/owners",
  element: (
    <RequirePermission resource={ResourceOwners}>
      <Outlet />
    </RequirePermission>
  ),
  errorElement: <RouteErrorBoundary />,   // ✅ あり
  children: [ ... ],
},
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `app/router.tsx:55-72` | Reception ルート errorElement 欠如 | 未修正 |

## 修正方針

### `frontend/src/app/router.tsx:56-61`

```tsx
{
  path: "/",
  element: (
    <RequirePermission resource={ResourceReception}>
      <Outlet />
    </RequirePermission>
  ),
  errorElement: <RouteErrorBoundary />,   // ← 追加
  children: [ ... ],
},
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `crash/BUG-187` 修正実績（閉チケット）
前回 BUG-187 で `/checkups`, `/estimates`, `/shifts`, `/settings` に `errorElement` を追加したが、
Reception ルート (`/`) が漏れていた。

### プロジェクト内参照実装
- `frontend/src/app/router.tsx:82` — `/owners` の `errorElement: <RouteErrorBoundary />`
- `frontend/src/app/router.tsx:132` — `/reservations` の `errorElement: <RouteErrorBoundary />`

## 優先度
**Medium** — Reception エラー時に Layout が失われ UX が他ルートと一貫しない。Reception は病院スタッフが最も頻繁に使うページ。

## 関連チケット
- closed/BUG-187 (前回修正で漏れたルート)

## 関連ファイル
- `frontend/src/app/router.tsx:55-72`
- `frontend/src/components/errors/RouteErrorBoundary.tsx`
