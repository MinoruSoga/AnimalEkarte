# FE-247: 受付カンバンの初期ローディング中にスケルトン/ローディング状態が表示されない

## 概要

`frontend/src/features/reception/routes/Reception.tsx` で、
カンバンデータの初回フェッチ中に `isLoading` 状態を UI にフィードバックしていない。
ページ表示直後に空のカンバン列が一瞬表示され、データがないと誤認されるUX問題がある。

## 問題箇所

### `Reception.tsx:366付近`

```tsx
// Before: isLoading チェックなしで columnElements をレンダリング
// → データ取得中に空のカンバン列が表示される
return (
  <div>
    {columnElements}  {/* isLoading 中も空カラムが見える */}
  </div>
);

// After: isLoading 中はローディング表示
if (isLoading) return <LoadingFallback />;
// または各カラムにスケルトン表示を追加
```

## 影響

- 初回表示時に「受付なし」状態が一瞬フラッシュする（Flashof Unstyled Content）
- ネットワークが遅い環境でユーザーが「データがない」と誤認する可能性がある

## 補足

`use-reception-kanban.ts` フックが `isLoading` を expose しているか確認が必要。
していない場合はフック側に `isLoading` を追加してから `Reception.tsx` で参照する。

## 準拠すべきプロジェクト規約

### 参照実装
- `frontend/src/features/checkups/routes/CheckupsList.tsx` — isLoading を適切に使用

## 優先度
**Low** — 機能的障害なし。UX のポリッシュ。

## 関連ファイル
- `frontend/src/features/reception/routes/Reception.tsx`
- `frontend/src/features/reception/hooks/use-reception-kanban.ts`
- `frontend/src/components/shared/DataStates/LoadingFallback.tsx`
