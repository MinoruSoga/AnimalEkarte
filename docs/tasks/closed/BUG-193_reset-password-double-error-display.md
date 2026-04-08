# BUG-193: ResetPasswordPage のパスワード不一致時に二重エラー表示

## 概要

`features/auth/routes/ResetPasswordPage.tsx` のパスワード不一致バリデーションで、`toast.error()` と `return { error: "..." }` の両方が同時に実行される二重エラー表示バグがある。ユーザーには Toast 通知とインラインエラーの両方が表示され、UX が混乱する。

## 再現手順

1. `/reset-password`（パスワードリセット画面）にアクセス
2. 「新しいパスワード」と「確認用パスワード」に異なる文字列を入力して送信
3. **結果**: 
   - Toast 通知「パスワードが一致しません」が表示される
   - かつ、フォームインラインエラー「パスワードが一致しません」も同時に表示される
   - ユーザーは同じエラーを 2 回見ることになる

## 期待する動作

- エラーはどちらか一方にのみ表示する（インラインエラーが推奨）
- Toast はサーバーサイドエラー・成功通知用として予約する

## 現状コード

### `frontend/src/features/auth/routes/ResetPasswordPage.tsx:50-51`
```tsx
// ❌ パスワード不一致時に toast と return error の両方を実行
if (newPassword !== confirmPassword) {
  toast.error("パスワードが一致しません");         // Toast 表示
  return { error: "パスワードが一致しません" };    // インライン表示
  // 両方実行される → 二重エラー
}
```

### 比較: 正しい実装
```tsx
// ✅ インラインエラーのみ（フォームバリデーション）
if (newPassword !== confirmPassword) {
  return { error: "パスワードが一致しません" };
  // toast は使わない（クライアントバリデーション）
}

// または ✅ Toast のみ（早期リターン）
if (newPassword !== confirmPassword) {
  toast.error("パスワードが一致しません");
  return;  // state は更新しない
}
```

## 影響範囲

| 対象ファイル | 行番号 | 問題 | 状態 |
|---|---|---|---|
| `features/auth/routes/ResetPasswordPage.tsx` | 50-51 | toast + return error の二重エラー | 未修正 |

## 修正方針

### `ResetPasswordPage.tsx:50-51`

クライアントサイドバリデーション（パスワード不一致）はインラインエラーのみで表示する。Toast はサーバーサイドエラー・成功時に使用する方針に統一する。

```tsx
// Before
if (newPassword !== confirmPassword) {
  toast.error("パスワードが一致しません");
  return { error: "パスワードが一致しません" };
}

// After
if (newPassword !== confirmPassword) {
  return { error: "パスワードが一致しません" };
}
```

エラー表示コンポーネント側（`state.error` を表示している箇所）で `FormFieldError` を使い、`C.bgDanger` に統一すること（BUG-191 / BUG-173 参照）。

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — Frontend ベストプラクティス
`useActionState` でフォームアクション管理。`return { error: "..." }` と `toast.error()` を同時に使用することは UX 上の二重通知であり、どちらか一方に統一すべきである。

### プロジェクト内参照実装
- `features/owners/routes/OwnerForm.tsx` — `useActionState` + インラインエラーのみの正しい実装

## 優先度
**Medium** — 機能的には動作するがユーザーが同じエラーを 2 回見る UX 問題。次回リリースで修正すべき。

## 関連チケット
- BUG-173: エラーメッセージカラーパターンのハードコード
- BUG-191: auth feature のハードコードカラー違反

## 関連ファイル
- `frontend/src/features/auth/routes/ResetPasswordPage.tsx`
