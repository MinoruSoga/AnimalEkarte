# BUG-191: auth feature のハードコードカラー違反（ChangePasswordDialog・LoginForm）

## 概要

`features/auth/` 配下の `ChangePasswordDialog.tsx` と `LoginForm.tsx` で Tailwind プリセットカラーがハードコードされている。認証フローの最重要 UI であるログイン画面・パスワード変更ダイアログがデザイントークン体系に準拠していない。

## 再現手順

1. `/login`（ログイン画面）を開く
2. システム管理者アカウントでログインし、管理者バッジを確認
   → **結果**: `text-red-600 bg-red-50` でハードコードされたバッジ表示
3. プロフィール → パスワード変更を開き、誤ったパスワードを入力
   → **結果**: `text-red-600` のエラーメッセージが表示されるが、デザイントークン未使用

## 現状コード

### `frontend/src/features/auth/routes/ChangePasswordDialog.tsx:111`
```tsx
// ❌ エラーメッセージに red ハードコード
<p className="text-red-600 text-sm">{errorMessage}</p>
```

### `frontend/src/features/auth/routes/LoginForm.tsx:68`
```tsx
// ❌ システム管理者バッジに red ハードコード
<span className="text-red-600 bg-red-50 ...">
  システム管理者
</span>
```

### 比較: 正しい実装
```tsx
import { C } from '@/lib/design-tokens';
import { FormFieldError } from '@/components/shared/FormFieldError';

// ✅ エラーメッセージ
<FormFieldError message={errorMessage} />
// または
<p style={{ color: C.bgDanger }} className="text-sm">{errorMessage}</p>

// ✅ 管理者バッジ
import { BADGE } from '@/lib/design-tokens';
<span style={BADGE.red}>システム管理者</span>
```

## 影響範囲

| 対象ファイル | 行番号 | 違反 | 状態 |
|---|---|---|---|
| `features/auth/routes/ChangePasswordDialog.tsx` | 111 | text-red-600（エラーメッセージ） | 未修正 |
| `features/auth/routes/LoginForm.tsx` | 68 | text-red-600 bg-red-50（管理者バッジ） | 未修正 |

## 修正方針

### 1. `ChangePasswordDialog.tsx:111` — エラーメッセージをトークンに
```tsx
import { FormFieldError } from '@/components/shared/FormFieldError';

// Before
<p className="text-red-600 text-sm">{errorMessage}</p>

// After
<FormFieldError message={errorMessage} />
```

### 2. `LoginForm.tsx:68` — 管理者バッジをトークンに
```tsx
import { C, BADGE } from '@/lib/design-tokens';

// Before
<span className="text-red-600 bg-red-50 ...">システム管理者</span>

// After
<span style={BADGE.red} className="...">システム管理者</span>
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリング（Tailwind 4, Inline styles）で `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。

### プロジェクト内参照実装
- `components/shared/FormFieldError.tsx` — エラーメッセージの標準実装

## 優先度
**Medium** — 認証画面の UX 一貫性。機能的問題はないが、エラーメッセージと管理者バッジのデザイントークン統一は全体の一貫性に関わる。

## 関連チケット
- BUG-169: required field marker の text-red-500（同パターン）
- BUG-173: エラーメッセージカラーパターンのハードコード（同パターン）

## 関連ファイル
- `frontend/src/features/auth/routes/ChangePasswordDialog.tsx`
- `frontend/src/features/auth/routes/LoginForm.tsx`
- `frontend/src/lib/design-tokens.ts`
