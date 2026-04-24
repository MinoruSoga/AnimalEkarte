# FE-144: ページネーションのページ番号が URL クエリパラメータに反映されない

**Status**: Open
**Priority**: Low
**Affects**: features/owners/, features/medical-records/, 全ページネーション付き一覧
**Date Created**: 2026-03-29
**Related**: BUG-049

---

## Summary

ページネーションで 2 ページ目に移動しても URL が `/owners` のまま変わらない。
リロード・URL 共有・ブラウザバック後にページ 1 に戻る。

---

## 実装手順

### 1. 現状確認

```bash
grep -rn "page\|currentPage\|setPage" frontend/src/features/owners/
grep -rn "Pagination\|usePagination" frontend/src/features/
```

### 2. URL クエリパラメータで page を管理

React Router の `useSearchParams` を使い、`page` を URL に反映する：

```typescript
import { useSearchParams } from "react-router-dom";

export function OwnersList() {
  const [searchParams, setSearchParams] = useSearchParams();
  const page = Number(searchParams.get("page") ?? "1");

  const handlePageChange = useCallback((newPage: number) => {
    setSearchParams(prev => {
      const next = new URLSearchParams(prev);
      if (newPage === 1) {
        next.delete("page");
      } else {
        next.set("page", String(newPage));
      }
      return next;
    });
  }, [setSearchParams]);

  // ...
}
```

### 3. ローダーでも page を読む（loader パターンを使っている場合）

```typescript
// loaders.ts
export async function ownersLoader({ request }: LoaderFunctionArgs) {
  const url = new URL(request.url);
  const page = Number(url.searchParams.get("page") ?? "1");
  const owners = await getOwners({ page });
  return { owners };
}
```

### 4. 適用対象

- `features/owners/routes/OwnersList.tsx`
- `features/medical-records/routes/MedicalRecordsList.tsx`
- その他ページネーション付き全一覧

---

## 受入条件

- [ ] ページ 2 クリック後 URL が `/owners?page=2` になる
- [ ] `/owners?page=2` で直接アクセスすると 2 ページ目が表示される
- [ ] ブラウザバック後に正しいページに戻る
- [ ] ページ 1 の URL は `/owners`（`?page=1` なし）
- [ ] `docker compose exec frontend pnpm lint` エラー 0 件
