# FE-241: ChangePasswordDialog の catch ブロックが実質空（handleApiError 未呼び出し）

## 概要

`frontend/src/features/auth/components/ChangePasswordDialog.tsx` の catch ブロックが
`handleApiError` を呼び出さず、ハードコードされた固定メッセージのみを返している。
API エラーの詳細（401/403/500 等）が完全に無視される。

## 問題コード

### `ChangePasswordDialog.tsx:54-56付近`

```ts
// Before: catch が実質的に空（エラー詳細を無視）
} catch {
  return { success: false, message: "パスワードの変更に失敗しました" };
  // handleApiError 未呼び出し → エラー種別が失われる
}

// After
} catch (error) {
  handleApiError(error, "パスワードの変更");
  return { success: false, message: "パスワードの変更に失敗しました" };
}
```

## 影響

- パスワード変更が 401（トークン期限切れ）で失敗した場合も「パスワードの変更に失敗しました」と表示される
- バックエンドからの具体的なエラーメッセージ（例: 「現在のパスワードが正しくありません」）が失われる
- 認証フローに直接影響するため、ユーザーが問題を特定できない

## 準拠すべきプロジェクト規約

### `.claude/rules/error-handling.md`
> すべての `catch` ブロックで `handleApiError(error, "コンテキスト")` を呼び出す。

## 優先度
**High** — 認証コンポーネントのエラーハンドリング。パスワード変更失敗の原因がユーザーに伝わらない。

## 関連ファイル
- `frontend/src/features/auth/components/ChangePasswordDialog.tsx`
- `frontend/src/lib/handle-api-error.ts`
