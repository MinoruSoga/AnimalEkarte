# FE-236: VitalsTab で isError が未チェック

## 概要

`frontend/src/features/medical-records/components/VitalsTab/VitalsTab.tsx` で
`useGetVitals` の `isError` を取得・処理していない。
API エラー時にユーザーには何も表示されず、空のバイタルタブが表示される。

## 問題コード

### `VitalsTab.tsx:283付近`

```tsx
// Before: isLoading のみ、isError なし
const { data: vitals, isLoading } = useGetVitals(medicalRecordId);

if (isLoading) return <LoadingFallback />;
// isError チェックなし → API エラーが空表示になる

// After
const { data: vitals, isLoading, isError } = useGetVitals(medicalRecordId);

if (isLoading) return <LoadingFallback />;
if (isError) return <ErrorFallback />;
```

## 影響

ネットワークエラーや認証エラーが発生した場合、バイタルデータが空として表示され、
ユーザーは「データがない」と誤認する。

## 準拠すべきプロジェクト規約

### `.claude/rules/error-handling.md`
> API エラー時は必ずユーザーに分かるフィードバックを表示すること。

### プロジェクト内参照実装
- `frontend/src/features/checkups/routes/CheckupsList.tsx` — isLoading + isError 正しく実装済み

### 関連チケット
- FE-209: 一覧ページの isError 未実装（同種問題）
- FE-214: VitalsTab mutation の onError 欠落（同ファイルの別問題）

## 優先度
**Medium** — 医療記録のバイタルタブでエラーが空表示になる UX 問題。

## 関連ファイル
- `frontend/src/features/medical-records/components/VitalsTab/VitalsTab.tsx`
- `frontend/src/components/shared/DataStates/` — LoadingFallback / ErrorFallback
