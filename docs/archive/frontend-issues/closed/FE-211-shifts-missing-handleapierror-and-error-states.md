# FE-211: シフト管理で handleApiError 未使用・エラー状態未処理

## 概要

シフト管理（`frontend/src/features/shifts/`）の複数箇所で:
1. `ShiftFormDialog` の catch ブロックで `handleApiError` を呼んでいない
2. `ShiftCalendarPage` でスタッフ API の isLoading / isError を未処理
3. シフト API mutation に `onError` コールバックがなく、エラー処理がコンポーネント側にも存在しない

## 問題1: ShiftFormDialog の catch ブロック

### `frontend/src/features/shifts/components/ShiftFormDialog/ShiftFormDialog.tsx:127-129`（作成・更新）
```tsx
// Before: catch ブロックにコメントのみ
} catch {
  // エラーはReact QueryがToast等で処理
}
// → React Query の onError コールバックも未設定のため、エラーが完全に握り潰される

// After
} catch (error) {
  handleApiError(error, "シフト保存");
}
```

### `frontend/src/features/shifts/components/ShiftFormDialog/ShiftFormDialog.tsx:150-158`（削除）
```tsx
// Before: try/catch なし
startDeleteTransition(async () => {
  await deleteShift(editShift.id);
  onClose();
});

// After
startDeleteTransition(async () => {
  try {
    await deleteShift(editShift.id);
    onClose();
  } catch (error) {
    handleApiError(error, "シフト削除");
  }
});
```

## 問題2: ShiftCalendarPage のスタッフ API エラー・ローディング未処理

### `frontend/src/features/shifts/routes/ShiftCalendarPage.tsx:26,48-49`
```tsx
// Before: staffsQuery の isLoading/isError が取得されているが未使用
const staffsQuery = useGetStaffs({ isActive: true });
const staffs = staffsQuery.data ?? [];
// → isError 時に staffs が空配列になり、ユーザーには空のカレンダーが表示される

// After
const staffsQuery = useGetStaffs({ isActive: true });
if (staffsQuery.isLoading) return <LoadingFallback />;
if (staffsQuery.isError) return <ErrorFallback />;
const staffs = staffsQuery.data ?? [];
```

## 問題3: シフト API mutations に onError なし

### `frontend/src/features/shifts/api/create-shift.ts`
### `frontend/src/features/shifts/api/update-shift.ts`
### `frontend/src/features/shifts/api/delete-shift.ts`

```ts
// Before: onError コールバックなし
export const useCreateShift = () => {
  return useMutation({
    mutationFn: createShift,
    onSuccess: ...,
    // onError なし
  });
};

// After
export const useCreateShift = () => {
  return useMutation({
    mutationFn: createShift,
    onSuccess: ...,
    onError: (error) => handleApiError(error, "シフト作成"),
  });
};
```

## 影響範囲

| 対象 | 問題 | 状態 |
|------|------|------|
| `shifts/components/ShiftFormDialog/ShiftFormDialog.tsx:127-129` | catch で handleApiError なし | 要修正 |
| `shifts/components/ShiftFormDialog/ShiftFormDialog.tsx:150-158` | 削除に try/catch なし | 要修正 |
| `shifts/routes/ShiftCalendarPage.tsx:26,48-49` | staffsQuery エラー・ローディング未処理 | 要修正 |
| `shifts/api/create-shift.ts` | onError なし | 要修正 |
| `shifts/api/update-shift.ts` | onError なし | 要修正 |
| `shifts/api/delete-shift.ts` | onError なし | 要修正 |

## 準拠すべきプロジェクト規約

### `.claude/rules/error-handling.md`
> すべての `catch` ブロックで `handleApiError(error, "コンテキスト")` を呼び出す。

### プロジェクト内参照実装
- `frontend/src/features/vaccinations/api/create-vaccination.ts` — `onError: (error) => handleApiError(error, "...")` で正しく実装

## 優先度
**High** — シフト作成・更新・削除の API エラーが完全に握り潰される。
ユーザーは失敗したことに気づかず、データが保存されていないと勘違いする。

## 関連ファイル
- `frontend/src/features/shifts/components/ShiftFormDialog/ShiftFormDialog.tsx`
- `frontend/src/features/shifts/routes/ShiftCalendarPage.tsx`
- `frontend/src/features/shifts/api/create-shift.ts`
- `frontend/src/features/shifts/api/update-shift.ts`
- `frontend/src/features/shifts/api/delete-shift.ts`
