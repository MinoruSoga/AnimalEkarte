# FE-136: 権限グループ割当後に /me を再取得して権限キャッシュを即時更新

**Status**: Open
**Priority**: Medium
**Affects**: features/master/routes/UserAccountSettings.tsx（または useSetUserPermissionGroups hook）
**Date Created**: 2026-03-29
**Related**: FE-134, BUG-057

---

## Summary

`PUT /v1/users/:id/permission-groups` の成功後、フロントエンドは
ユーザー一覧・詳細のキャッシュを無効化するが、**`AuthContext` が保持する
`user.permissions` は更新しない**。

結果として、**権限変更を受けたスタッフが変更直後も旧権限でページ操作を継続できる**。
`RequirePermission` コンポーネントや `usePermission` hook の判定が古い値のままになる。

---

## 現状の問題

```typescript
// features/master/api/set-user-permission-groups.ts（現状）
export function useSetUserPermissionGroups() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (params) => setUserPermissionGroups(params),
    onSuccess: () => {
      // ✅ ユーザー一覧・詳細のキャッシュは無効化
      queryClient.invalidateQueries({ queryKey: USER_QUERY_KEYS.lists() });
      queryClient.invalidateQueries({ queryKey: USER_QUERY_KEYS.detail(userId) });
      // ❌ user.permissions は更新されない
      //    → 次ページナビゲーションまで旧権限で動作
    },
  });
}
```

### 具体的な不具合シナリオ

1. スタッフ A が `canDelete = false` の権限グループに所属している
2. clinic_admin が UserAccountSettings で スタッフ A に `canDelete = true` のグループを割当・保存
3. スタッフ A のタブで即座に「削除」ボタンが表示されない
4. スタッフ A がページをリロードして初めて新権限が反映される

---

## 修正方針

### パターン 1（推奨）: `/me` を再取得して AuthContext を更新

`useSetUserPermissionGroups` の `onSuccess` で、変更対象ユーザーが **自分自身** の場合に
`/me` を再取得して `AuthContext` を更新する。

```typescript
// features/master/api/set-user-permission-groups.ts
import { useAuth } from "@/features/auth/hooks/use-auth";

export function useSetUserPermissionGroups(userId: string) {
  const queryClient = useQueryClient();
  const { user, refreshPermissions } = useAuth();  // ← refreshPermissions を追加

  return useMutation({
    mutationFn: (params: SetUserPermissionGroupsParams) =>
      setUserPermissionGroups(params),
    onSuccess: async () => {
      queryClient.invalidateQueries({ queryKey: USER_QUERY_KEYS.lists() });
      queryClient.invalidateQueries({ queryKey: USER_QUERY_KEYS.detail(userId) });

      // 変更対象が自分自身の場合: 権限を即時更新
      if (user?.id === userId) {
        await refreshPermissions();
      }
    },
  });
}
```

### `refreshPermissions` を `AuthContext` に追加

```typescript
// features/auth/hooks/use-auth.tsx

interface AuthContextValue {
  // ... 既存フィールド ...
  refreshPermissions: () => Promise<void>;  // ← 追加
}

// AuthProvider 内
const refreshPermissions = useCallback(async () => {
  const me = await fetchMe();  // GET /v1/me を再呼び出し
  setUser(transformMeResponse(me));
}, []);
```

### パターン 2（シンプル）: `/me` の React Query キャッシュを無効化

より単純な方法として、`/me` のキャッシュを直接無効化する。

```typescript
onSuccess: async () => {
  queryClient.invalidateQueries({ queryKey: USER_QUERY_KEYS.lists() });
  queryClient.invalidateQueries({ queryKey: USER_QUERY_KEYS.detail(userId) });

  // /me キャッシュを無効化 → AuthProvider が自動的に再フェッチ
  if (user?.id === userId) {
    queryClient.invalidateQueries({ queryKey: ["me"] });
  }
},
```

**注意**: `/me` のキャッシュキーが `["me"]` として管理されている場合のみ有効。
現在の実装がどう管理しているか確認してから適用すること。

---

## 変更ファイル一覧

| ファイル | 変更内容 |
|---------|---------|
| `features/auth/hooks/use-auth.tsx` | `refreshPermissions()` メソッドを AuthContext に追加 |
| `features/master/api/set-user-permission-groups.ts` | `onSuccess` で `refreshPermissions()` を呼ぶ（自分自身の変更時のみ） |

---

## 補足: 他ユーザーへの権限変更は非同期で問題ない

clinic_admin がスタッフ B の権限を変更した場合、スタッフ B のセッションへの即時反映は
本チケットのスコープ外（BUG-057 で対応）。**自分自身の権限を変更する UI を経由した際に
即時反映されることだけを保証する** のが本チケットのスコープ。

---

## 受入条件

- [ ] 自分自身の権限グループを変更・保存後、即座に新権限でページが動作する
- [ ] 保存後 `usePermission()` が更新された権限を返す
- [ ] `clinic_admin` / `system_admin` には変更なし（全権限バイパスのため影響なし）
- [ ] `docker compose exec frontend pnpm build` 成功
- [ ] `docker compose exec frontend pnpm lint` エラー 0 件
- [ ] `docker compose exec frontend pnpm test:run` 成功
