# BUG-166: ReservationManagement — 遅延ロードコンポーネントに ErrorBoundary がない

## 概要

`frontend/src/features/reservations/routes/ReservationManagement.tsx` で、
`MonthView`・`WeekView`・`ReservationDetailModal` を `lazy()` で遅延ロードしているが、
`<Suspense>` には fallback のみ指定されており `ErrorBoundary` が存在しない。
チャンクの読み込みに失敗した場合（ネットワークエラー、デプロイ後のキャッシュ不整合等）、
Suspense fallback が永続表示されるか、`React` がトップレベルにエラーを伝播させてページ全体がクラッシュする。

## 再現手順

1. 予約管理ページを開く（`/reservations`）
2. 遅延ロード対象のチャンクファイルを DevTools → Network タブで「Block request URL」で遮断
3. ページをリロード
4. **結果**: `読み込み中...` が永続表示されるか、ページ全体がクラッシュする

## 期待する動作

- チャンクロード失敗時は `<ErrorFallback />` またはリトライ可能なエラーメッセージを表示する
- 予約ページ全体がクラッシュすることなく、部分的なエラー表示で留まる

## 現状コード

### `frontend/src/features/reservations/routes/ReservationManagement.tsx:28-50付近`
```tsx
// Before: lazy + Suspense のみ。ErrorBoundary なし
const MonthView = lazy(() => import("../components/MonthView"));
const WeekView = lazy(() => import("../components/WeekView"));
const ReservationDetailModal = lazy(() => import("../components/ReservationDetailModal"));

// ...

<Suspense fallback={<div>読み込み中...</div>}>
  {view === "month" ? <MonthView ... /> : <WeekView ... />}
</Suspense>
// ↑ チャンクロード失敗時のエラーハンドリングが存在しない
```

### 比較: 正しい実装（プロジェクト内参照実装）
```tsx
// ErrorBoundary を Suspense の外側に配置するパターン
import { RouteErrorBoundary } from "@/components/errors/RouteErrorBoundary";

<RouteErrorBoundary>
  <Suspense fallback={<LoadingFallback />}>
    {view === "month" ? <MonthView ... /> : <WeekView ... />}
  </Suspense>
</RouteErrorBoundary>
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `ReservationManagement.tsx:240-250` | Suspense に ErrorBoundary なし | 未修正 |
| `MonthView` チャンク | ロード失敗時エラー未処理 | 未修正 |
| `WeekView` チャンク | ロード失敗時エラー未処理 | 未修正 |
| `ReservationDetailModal` チャンク | ロード失敗時エラー未処理 | 未修正 |

## 修正方針

### 1. ErrorBoundary を Suspense の外側に追加 — `ReservationManagement.tsx:240付近`

```tsx
import { RouteErrorBoundary } from "@/components/errors/RouteErrorBoundary";
import { LoadingFallback } from "@/components/shared/DataStates/LoadingFallback";

// After
<RouteErrorBoundary>
  <Suspense fallback={<LoadingFallback />}>
    {view === "month" ? (
      <MonthView ... />
    ) : (
      <WeekView ... />
    )}
  </Suspense>
</RouteErrorBoundary>
```

### 2. `ReservationDetailModal` の Suspense も同様に対処

```tsx
<RouteErrorBoundary>
  <Suspense fallback={null}>
    {selectedReservationId ? (
      <ReservationDetailModal reservationId={selectedReservationId} onClose={...} />
    ) : null}
  </Suspense>
</RouteErrorBoundary>
```

## 準拠すべきプロジェクト規約

### `.claude/rules/code-style.md` — bundle-dynamic-imports
> 重いモーダル・ダイアログは `lazy()` + `Suspense` で遅延ロード

ErrorBoundary の追加は React の公式推奨事項（[React docs: Error Boundaries](https://react.dev/reference/react/Component#catching-rendering-errors-with-an-error-boundary)）であり、
`Suspense` と組み合わせる場合は必須。

### プロジェクト内参照実装
- `frontend/src/components/errors/RouteErrorBoundary.tsx` — ErrorBoundary 実装済み

## 優先度
**High** — デプロイ後のキャッシュ不整合（古いチャンク URL へのアクセス）はプロダクションで頻繁に発生する既知のケース。予約ページが完全にクラッシュするリスクがある。

## 関連チケット
- なし

## 関連ファイル
- `frontend/src/features/reservations/routes/ReservationManagement.tsx:28-38,240-250`
- `frontend/src/components/errors/RouteErrorBoundary.tsx`
- `frontend/src/components/shared/DataStates/LoadingFallback.tsx`
