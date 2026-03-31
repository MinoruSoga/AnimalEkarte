# FE-002: handleApiError 未使用 — mutation の onError が全て独自実装

## 問題
`lib/handle-api-error.ts` に共通エラーハンドラが定義されているが、
100+ の `useMutation` が全て個別に `toast.error()` を直接呼び出している。
HTTPステータスコードを無視した画一的なエラーメッセージになっている。

## 影響範囲
全 feature の `api/*.ts` mutation ファイル（100件以上）

代表例:
- `features/accounting/api/create-accounting.ts`
- `features/estimates/api/create-estimate.ts`
- `features/hospitalization/api/update-hospitalization.ts`
- `features/medical-records/api/treatments.ts`（4 mutations）
- `features/medical-records/api/vitals.ts`（3 mutations）

## 現状（NG）
```ts
onError: (error: Error) => {
  toast.error(error.message || "操作に失敗しました");
}
```

## 問題点
1. HTTP 400/401/403/404/409/5xx を区別できない
2. サーバーの `response.data.message` が無視される
3. feature ごとにエラーメッセージが異なる（UX不統一）
4. `handleApiError()` が dead code 化している

## 修正方針
```ts
import { handleApiError } from "@/lib/handle-api-error";

onError: (error) => {
  handleApiError(error);
}
```

`handle-api-error.ts` の実装を確認し、必要であれば
ステータスコード別メッセージ対応を強化する。

## 優先度
MEDIUM（UX・エラーハンドリング一貫性）
