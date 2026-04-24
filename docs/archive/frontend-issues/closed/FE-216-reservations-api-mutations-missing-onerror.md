# FE-216: 予約 API mutation 3件に onError がない

## 概要

予約機能の `create-reservation.ts`、`update-reservation.ts`、`delete-reservation.ts` の
`useMutation` に `onError` コールバックが設定されていない。
予約の作成・更新・削除に失敗してもユーザーに通知されない。

## 問題コード

### `frontend/src/features/reservations/api/create-reservation.ts`
```ts
// Before: onError なし
export const useCreateReservation = () => {
  return useMutation({
    mutationFn: createReservation,
    onSuccess: ...,
    // onError なし → 予約作成失敗がサイレント
  });
};
```

### `frontend/src/features/reservations/api/update-reservation.ts`
```ts
// Before: onError なし
export const useUpdateReservation = () => {
  return useMutation({
    mutationFn: updateReservation,
    onSuccess: ...,
    // onError なし → 予約更新失敗がサイレント
  });
};
```

### `frontend/src/features/reservations/api/delete-reservation.ts`
```ts
// Before: onError なし
export const useDeleteReservation = () => {
  return useMutation({
    mutationFn: deleteReservation,
    onSuccess: ...,
    // onError なし → 予約削除失敗がサイレント
  });
};
```

## 修正方針

```ts
// After: 各 useMutation に onError を追加
export const useCreateReservation = () => {
  return useMutation({
    mutationFn: createReservation,
    onSuccess: ...,
    onError: (error) => handleApiError(error, "予約作成"),
  });
};

export const useUpdateReservation = () => {
  return useMutation({
    mutationFn: updateReservation,
    onSuccess: ...,
    onError: (error) => handleApiError(error, "予約更新"),
  });
};

export const useDeleteReservation = () => {
  return useMutation({
    mutationFn: deleteReservation,
    onSuccess: ...,
    onError: (error) => handleApiError(error, "予約削除"),
  });
};
```

## 影響範囲

| 対象ファイル | 問題 | 状態 |
|------------|------|------|
| `reservations/api/create-reservation.ts` | onError なし | 要修正 |
| `reservations/api/update-reservation.ts` | onError なし | 要修正 |
| `reservations/api/delete-reservation.ts` | onError なし | 要修正 |

## 準拠すべきプロジェクト規約

### `.claude/rules/error-handling.md`
> すべての `catch` ブロックおよび `onError` コールバックで `handleApiError(error, "コンテキスト")` を呼び出す。

### プロジェクト内参照実装（正しい実装）
- `frontend/src/features/accounting/api/create-accounting.ts` — `onError: (error) => handleApiError(error, "作成")` で正しく実装済み
- `frontend/src/features/accounting/api/update-accounting.ts` — 同上

## 優先度
**High** — 予約の作成・更新・削除が失敗してもユーザーに通知されない。
特に削除失敗は「消えた」と誤認させる UX 障害となる。

## 関連ファイル
- `frontend/src/features/reservations/api/create-reservation.ts`
- `frontend/src/features/reservations/api/update-reservation.ts`
- `frontend/src/features/reservations/api/delete-reservation.ts`
- `frontend/src/lib/handle-api-error.ts`
