# FE-212: chief-complaint-categories のクエリキーが2箇所で不一致（キャッシュ不整合）

## 概要

主訴カテゴリ（chief-complaint-categories）の React Query クエリキーが
`medical-records/api/` と `master/api/` の2箇所で異なる値を使用しており、
一方を invalidate しても他方のキャッシュが更新されない不整合が発生している。

## 問題コード

### `frontend/src/features/medical-records/api/get-chief-complaint-categories.ts:21`
```ts
queryKey: ["masters", "chief-complaints"],  // ← "chief-complaints"（略語）
```

### `frontend/src/features/master/api/chief-complaint-categories.ts:47`
```ts
queryKey: ["masters", "chief-complaint-categories"],  // ← "chief-complaint-categories"（正式名）
```

**2つの異なるキーが同じリソースを参照している。**

`master` 機能でカテゴリを追加・更新・削除すると `["masters", "chief-complaint-categories"]` が invalidate されるが、
`medical-records` 側の `["masters", "chief-complaints"]` は invalidate されず、古いキャッシュが残る。

## 影響

1. マスタ設定でカテゴリを追加しても、カルテ側のカテゴリ選択に即時反映されない
2. マスタ設定でカテゴリを削除しても、カルテ側で削除済みカテゴリが表示され続ける

## 影響範囲

| 対象 | 行番号 | キー | 状態 |
|------|--------|------|------|
| `frontend/src/features/medical-records/api/get-chief-complaint-categories.ts` | 21 | `["masters", "chief-complaints"]` | 要修正 |
| `frontend/src/features/master/api/chief-complaint-categories.ts` | 47 | `["masters", "chief-complaint-categories"]` | 正しい（変更不要） |

## 修正方針

`medical-records` 側のクエリキーを `master` 側に統一する。

### `get-chief-complaint-categories.ts:21`
```ts
// Before
queryKey: ["masters", "chief-complaints"],

// After
queryKey: ["masters", "chief-complaint-categories"],
```

**さらに理想的な対応**: 両ファイルで同じ `queryKeys` 定数を参照する。
`src/lib/query-keys.ts` 等で一元管理すべき。

```ts
// lib/query-keys.ts
export const masterQueryKeys = {
  chiefComplaintCategories: () => ["masters", "chief-complaint-categories"] as const,
};
```

## 準拠すべきプロジェクト規約

### `.claude/CLAUDE.md` — 型安全性最優先
> 重複コードの排除。同一リソースのクエリキーは一元管理すること。

## 優先度
**Medium** — カルテ入力時に削除済みカテゴリが表示されたり、新しいカテゴリが反映されないデータ不整合が発生する。

## 関連ファイル
- `frontend/src/features/medical-records/api/get-chief-complaint-categories.ts:21`
- `frontend/src/features/master/api/chief-complaint-categories.ts:47`
