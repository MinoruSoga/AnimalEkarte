# FE-227: 受付ステータス更新 mutation に onError がない

## 概要

`frontend/src/features/reception/api/update-appointment-status.ts` の
`useUpdateAppointmentStatus` フックに `onError` コールバックが設定されていない。
受付ステータスの更新が失敗してもユーザーに通知されない。

## 問題コード

### `frontend/src/features/reception/api/update-appointment-status.ts:22-29`

```ts
// Before: onError なし
return useMutation({
  mutationFn: ({ id, status }: UpdateStatusPayload) =>
    updateAppointmentStatus(id, status),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ["reception"] });
  },
  // onError なし → ステータス更新失敗がサイレント
});

// After
return useMutation({
  mutationFn: ({ id, status }: UpdateStatusPayload) =>
    updateAppointmentStatus(id, status),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ["reception"] });
  },
  onError: (error) => handleApiError(error, "受付ステータスの更新"),
});
```

## 影響

受付カンバンでドラッグ&ドロップや「次へ進む」ボタンを押してステータス更新が失敗した場合、
UI は更新されたように見えているが（楽観的更新）、サーバー側に反映されていない可能性がある。
ユーザーは失敗に気づかず、診察フローが崩れる。

## 補足

`use-reception-kanban.ts` の `moveCard`・`advanceStatus` は
try/catch + handleApiError で正しく実装済み。
ただし useMutation レイヤーでの `onError` も追加することで二重のフォールバックになる。

## 準拠すべきプロジェクト規約

### `.claude/rules/error-handling.md`
> すべての `catch` ブロックおよび `onError` コールバックで `handleApiError(error, "コンテキスト")` を呼び出す。

## 優先度
**High** — 受付フローのステータス更新失敗が診察進行に影響する可能性がある。

## 関連ファイル
- `frontend/src/features/reception/api/update-appointment-status.ts`
- `frontend/src/lib/handle-api-error.ts`
