# TASK-020: React Query キャッシュキー型衝突 — reception/get-staffs.ts

## 概要

`reception/api/get-staffs.ts` が `["masters", "staffs"]` というクエリキーで `BackendStaff[]`（未変換）を返しているが、同じキーを `features/master/api/staffs.ts` は変換済みの `Staff[]`、`hooks/use-staffs.ts` は `StaffItem[]` で使用している。React Query のキャッシュは型を区別しないため、フックの実行順序によってランタイム型エラーが発生する。

## 優先度

HIGH（ランタイムバグ・型安全性破綻）

## 影響ファイル

| ファイル | 行 | 問題 |
|---------|-----|------|
| `frontend/src/features/reception/api/get-staffs.ts` | L14 | `["masters", "staffs"]` キーで異なる型を返す |

## 規約違反

`.claude/rules/code-style.md`:
> Feature 間の直接 import 禁止。共有フックは `src/hooks/` を経由すること。

## 修正方針

`hooks/use-staffs.ts` の `useGetStaffs` を再利用してキャッシュを共有する。

```typescript
// reception/api/get-staffs.ts（修正後）
import { useGetStaffs } from "@/hooks/use-staffs";

// useGetStaffs は StaffItem[] を返すため buildStaffMap に十分
export function useGetReceptionStaffs() {
  return useGetStaffs();
}

export function buildStaffMap(staffs: { id: string; name: string }[]): Map<string, string> {
  return new Map(staffs.map((s) => [s.id, s.name]));
}
```

## テスト

- reception ページと master ページを同時に開いたとき、スタッフ一覧が正しい型で表示されることを確認
- React Query DevTools でキャッシュに `["masters", "staffs"]` が1エントリのみ存在することを確認
