# FE-218: toast.error() 直接呼び出し — handleApiError に統一が必要

## 概要

複数箇所で `handleApiError()` を使わずに `toast.error()` を直接呼び出している。
プロジェクト規約「すべての catch ブロックで `handleApiError` を呼び出す」に違反。
エラー詳細のログ出力・型別メッセージ切り替えが機能しない。

## 問題箇所

### `frontend/src/features/owners/routes/OwnersList.tsx:314`
```tsx
// Before: toast.error 直接呼び出し（削除 transition 内）
startDeleteTransition(async () => {
  try {
    await deleteOwner(ownerId);
    // ...
  } catch {
    toast.error("削除に失敗しました");  // ← handleApiError でない
  }
});

// After: handleApiError を使用
startDeleteTransition(async () => {
  try {
    await deleteOwner(ownerId);
    // ...
  } catch (error) {
    handleApiError(error, "オーナー削除");
  }
});
```

### `frontend/src/features/master/routes/CompanySettings.tsx:127-135付近`
```ts
// Before: useMutation の onError で toast 直接呼び出し
updateMutation.mutate(payload, {
  onSuccess: () => { ... },
  onError: () => {
    toast.error("更新に失敗しました");  // ← handleApiError でない
  },
});

// After
updateMutation.mutate(payload, {
  onSuccess: () => { ... },
  onError: (error) => handleApiError(error, "法人情報の更新"),
});
```

## 影響範囲

| 対象ファイル | 行 | 問題 | 状態 |
|------------|---|------|------|
| `owners/routes/OwnersList.tsx` | 314 | catch ブロックで `toast.error()` 直呼び | 要修正 |
| `master/routes/CompanySettings.tsx` | 127-135 | onError で `toast.error()` 直呼び | 要修正 |

## `handleApiError` を使う理由

`handleApiError` は以下を一元処理する：
1. HTTP ステータスコード別のメッセージ分岐（401/403/404/409/500）
2. バックエンドの `{"message": "..."}` レスポンスの抽出
3. 構造化ログ出力
4. toast 表示

`toast.error("固定文字列")` では、ステータスコード/バックエンドメッセージが無視される。

## 準拠すべきプロジェクト規約

### `.claude/rules/error-handling.md`
> すべての `catch` ブロックで `handleApiError(error, "コンテキスト")` を呼び出す。

### プロジェクト内参照実装
- `frontend/src/features/medical-records/hooks/use-medical-record-form.ts:246,268,292` — 正しい実装

## 優先度
**Medium** — エラー種別によるメッセージ切り替えが機能しない。
特に 401/403 エラー時に「更新に失敗しました」という汎用メッセージが表示されてしまう。

## 関連ファイル
- `frontend/src/features/owners/routes/OwnersList.tsx`
- `frontend/src/features/master/routes/CompanySettings.tsx`
- `frontend/src/lib/handle-api-error.ts`
