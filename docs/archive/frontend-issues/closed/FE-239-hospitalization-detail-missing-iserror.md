# FE-239: HospitalizationDetail で isError が処理されていない

## 概要

`frontend/src/features/hospitalization/routes/HospitalizationDetail.tsx` が
`isLoading` のみを確認しており、API エラー発生時に `isError` をチェックしていない。
クエリ失敗時にローディング状態のまま画面が固まる。

## 問題コード

### `HospitalizationDetail.tsx:23-25, 43-44`

```tsx
// Before: isLoading のみ
const { data: hospitalization, isLoading } = useHospitalizationDetail(id);

if (isLoading || !hospitalization) {
  return <LoadingFallback />;  // isError でもここに留まる
}

// After: isError も処理
const { data: hospitalization, isLoading, isError } = useHospitalizationDetail(id);

if (isLoading) return <LoadingFallback />;
if (isError) return <ErrorFallback />;
if (!hospitalization) return <div>データが見つかりません</div>;
```

## 根本原因

`use-hospitalization-detail.ts` フックが `isError` を返していない可能性がある。
フック側で `isError` を expose するよう修正も必要。

```ts
// use-hospitalization-detail.ts
const { data, isLoading, isError } = useQuery(...);
return { hospitalization: data, isLoading, isError };  // isError を追加
```

## 準拠すべきプロジェクト規約

### `.claude/rules/error-handling.md`
> API エラー時は必ずユーザーに分かるフィードバックを表示すること。

### 関連チケット
- FE-209: 一覧ページの isError 未処理（同種問題）
- FE-236: VitalsTab の isError 未処理（同種問題）

## 優先度
**Medium** — 入院詳細ページで API エラー時にローディング画面から抜け出せない。

## 関連ファイル
- `frontend/src/features/hospitalization/routes/HospitalizationDetail.tsx`
- `frontend/src/features/hospitalization/hooks/use-hospitalization-detail.ts`（isError を expose）
- `frontend/src/components/shared/DataStates/` — ErrorFallback
