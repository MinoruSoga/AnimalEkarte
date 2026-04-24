# FE-137: 権限変更をリアルタイム反映 — useMe の refetchInterval + refetchOnWindowFocus 有効化

**Status**: Open
**Priority**: Low-Medium
**Affects**: features/auth/api/（useMe hook）
**Date Created**: 2026-03-29
**Related**: BUG-057, FE-136（先行: 自セッション権限更新）

---

## Summary

権限グループの変更が他セッション（別タブ・別ブラウザ）にリアルタイム反映されない。
`GET /me` のポーリング間隔が長すぎるか、`refetchOnWindowFocus` が無効になっているため。

BUG-055（JWT を 15 分短縮）が完了すると影響範囲が「最大 24時間 → 最大 20 分」に縮小されるため、
BUG-055 完了後に本チケットを実装することで最大遅延を 5 分程度に抑えられる。

---

## 実装手順

### 1. `useMe` hook の確認

```bash
grep -rn "useMe\|getMe\|/me" frontend/src/features/auth/api/
```

### 2. refetchInterval と refetchOnWindowFocus を追加

```typescript
// features/auth/api/get-me.ts（または useMe hook）
export function useGetMe() {
  return useQuery({
    queryKey: ["me"],
    queryFn: getMe,
    staleTime: 4 * 60 * 1000,          // 4分（キャッシュ有効期間）
    refetchInterval: 5 * 60 * 1000,    // 5分ごとにポーリング
    refetchOnWindowFocus: true,         // タブフォーカス時に即再取得
    refetchIntervalInBackground: false, // バックグラウンドタブはポーリング停止
    retry: 1,
    retryDelay: 3000,
  });
}
```

### 3. 401 時にポーリングを停止

権限変更でログアウト状態になった場合、無限ループを防ぐ：

```typescript
export function useGetMe() {
  return useQuery({
    queryKey: ["me"],
    queryFn: getMe,
    staleTime: 4 * 60 * 1000,
    refetchInterval: (query) => {
      // 401 (Unauthorized) の場合はポーリング停止
      if (query.state.error instanceof AxiosError) {
        if (query.state.error.response?.status === 401) {
          return false;
        }
      }
      return 5 * 60 * 1000;
    },
    refetchOnWindowFocus: true,
    refetchIntervalInBackground: false,
  });
}
```

### 4. 確認事項

- [ ] ポーリング間隔: 5 分（過剰なリクエストを避けつつ許容範囲内）
- [ ] タブフォーカス時に即再取得される
- [ ] バックグラウンドタブではポーリングしない（サーバー負荷軽減）
- [ ] 401 時にポーリングが停止しログインページにリダイレクトされる

---

## 受入条件

- [ ] 権限変更後、最大 5 分以内に他セッションに反映される
- [ ] タブがフォーカスされたタイミングで `/me` が再取得される
- [ ] 401 時にポーリングが停止する
- [ ] `docker compose exec frontend pnpm lint` エラー 0 件
