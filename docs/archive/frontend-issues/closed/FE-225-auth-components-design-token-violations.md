# FE-225: 認証コンポーネントのデザイントークン違反（LoginForm・ChangePasswordDialog）

## 概要

認証関連コンポーネント2ファイルで直接 Tailwind カラークラスが使用されている。

## 違反ファイル

### `frontend/src/features/auth/components/LoginForm.tsx:68`

```tsx
// Before: デモアカウント管理者バッジに直接カラークラス
<span className="text-xs px-1.5 py-px rounded-[3px] text-red-600 bg-red-50">
  admin
</span>

// After: デザイントークン使用
<span className={`text-xs px-1.5 py-px rounded-[3px] ${C.danger} ${C.bgDangerLight}`}>
  admin
</span>
```

### `frontend/src/features/auth/components/ChangePasswordDialog.tsx:111`

```tsx
// Before: エラーメッセージに直接カラークラス
<p className="text-sm text-red-600" role="alert">
  {errorMessage}
</p>

// After
<p className={`text-sm ${C.danger}`} role="alert">
  {errorMessage}
</p>
```

## 準拠すべきプロジェクト規約

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリングで `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。
> **PROHIBITED**: 直接 Tailwind カラークラスの指定は厳禁。

## 優先度
**Low** — 2ファイル・2箇所のみ。機能的障害なし。

## 関連ファイル
- `frontend/src/features/auth/components/LoginForm.tsx`
- `frontend/src/features/auth/components/ChangePasswordDialog.tsx`
- `frontend/src/lib/design-tokens.ts`
