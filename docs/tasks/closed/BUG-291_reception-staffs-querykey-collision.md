# BUG-291: reception/api/get-staffs.ts — queryKey衝突によるサイレント型破壊

## 概要

`reception/api/get-staffs.ts` と `master/api/staffs.ts` が同一の `queryKey: ["masters", "staffs"]` を使用しているため、React Queryキャッシュが共有される。両フックの戻り値の型が異なるため、サイレントな型破壊が発生する。

## 問題

```typescript
// frontend/src/features/reception/api/get-staffs.ts
export function useGetStaffs() {
  return useQuery({
    queryKey: ["masters", "staffs"],  // ← master/api/staffs.ts と同一
    queryFn: async () => {
      const { data } = await axios.get<BackendStaff[]>("/v1/masters/staffs");
      return data;  // BackendStaff[] { id, name, is_active }
    },
  });
}

// frontend/src/features/master/api/staffs.ts
export function useGetStaffs() {
  return useQuery({
    queryKey: ["masters", "staffs"],  // ← reception/api/get-staffs.ts と同一
    queryFn: async () => {
      const { data } = await axios.get<ModelStaff[]>("/v1/masters/staffs");
      return data;  // ModelStaff[] (全フィールド)
    },
  });
}
```

- `reception/api/get-staffs.ts` の `BackendStaff` は `{ id, name, is_active }` のみ
- `master/api/staffs.ts` の `ModelStaff` は全フィールド（`ModelStaff` 型）
- どちらのフックが先にフェッチするかによってキャッシュ内容が決まり、もう一方のフックが型不一致のデータを受け取る
- TypeScriptはコンパイル時にこれを検出できない（実行時のみ問題が発生）

## 影響

- サイレントな型破壊（ランタイムで未定義プロパティアクセスが発生する可能性）
- 不確定な再現性（ページのロード順序に依存する Heisenbug）

## 修正方針

**Option A（推奨）**: receptionのqueryKeyを `["reception", "staffs"]` に変更する

```typescript
// frontend/src/features/reception/api/get-staffs.ts
export function useGetStaffs() {
  return useQuery({
    queryKey: ["reception", "staffs"],  // ← 衝突解消
    queryFn: async () => {
      const { data } = await axios.get<BackendStaff[]>("/v1/masters/staffs");
      return data;
    },
    staleTime: QUERY_STALE_TIMES.STATIC,
  });
}
```

**Option B**: reception が master の hook を再利用し、props注入でデータを受け取る（アーキテクチャ上のより適切な解決）

## ステータス

- [x] ドキュメント作成
- [x] 実装完了（`["masters", "staffs"]` → `["reception", "staffs"]` に変更）
