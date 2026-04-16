# BUG-173: エラーメッセージの text-red-600 ハードコード（estimates/inventory/shifts/accounting）

## 概要

フォームのエラーメッセージ表示で `text-red-600` / `text-red-500` / `border-red-500` を Tailwind クラスとして直接使用しているケースが複数 feature に存在する。プロジェクトでは `C.bgDanger` または `C.textRequired` をトークンとして使用するべきであり、`text-red-*` のプリセットカラー直接指定は禁止。

## 再現手順

1. 見積フォーム / 在庫フォーム / シフト管理 / 会計詳細でバリデーションエラーを発生させる
2. エラーメッセージのテキスト色を DevTools で確認
3. **結果**: `text-red-600` / `text-red-500` クラスが適用されており、`C.bgDanger` / `C.textRequired` トークンを経由していない

## 期待する動作

- エラーメッセージテキスト: `style={{ color: C.bgDanger }}` または `style={{ color: C.textRequired }}` を使用
- エラー状態のインプット枠線: `style={{ borderColor: C.bgDanger }}` を使用
- 共通エラー表示コンポーネント（`FormFieldError`）が既に存在する場合はそれを使用する

## 現状コード

### `frontend/src/features/estimates/routes/EstimateForm.tsx`
```tsx
// ❌ バリデーションエラーテキスト
<p className="text-red-600 text-sm mt-1">{errors.amount}</p>
// ❌ エラー枠線
<Input className="border-red-500" />
```

### `frontend/src/features/inventory/routes/InventoryForm.tsx`
```tsx
// ❌ クロスフィールドバリデーションエラー
<p className="text-red-600 text-sm">{crossFieldError}</p>
```

### `frontend/src/features/shifts/routes/ShiftManagement.tsx` 等
```tsx
// ❌ シフト関連エラー表示
<span className="text-red-600">{errorMessage}</span>
```

### `frontend/src/features/accounting/routes/AccountingDetail.tsx`
```tsx
// ❌ 金額マイナス時の色 (エラー系)
<span className={amount < 0 ? "text-red-500" : "text-gray-900"}>
```

### 比較: 正しい実装（共通コンポーネント使用）
```tsx
// ✅ FormFieldError コンポーネントを使用（既存）
import { FormFieldError } from '@/components/shared/FormFieldError/FormFieldError';

{errors.amount ? <FormFieldError message={errors.amount} /> : null}

// ✅ FormFieldError の実装例（内部でトークン使用）
import { C } from '@/lib/design-tokens';
export function FormFieldError({ message }: { message: string }) {
  return <p style={{ color: C.bgDanger }} className="text-sm mt-1">{message}</p>;
}

// ✅ エラー枠線
style={{ borderColor: errors.amount ? C.bgDanger : undefined }}
```

## 影響範囲

| 対象ファイル | 違反箇所 | 状態 |
|---|---|---|
| `features/estimates/routes/EstimateForm.tsx` | text-red-600 エラーメッセージ + border-red-500 | 未修正 |
| `features/inventory/routes/InventoryForm.tsx` | text-red-600 クロスフィールドエラー | 未修正 |
| `features/shifts/routes/` 配下 | text-red-600 エラー表示 | 未修正 |
| `features/accounting/routes/AccountingDetail.tsx` | text-red-500 マイナス金額色 (BUG-172 と重複) | 未修正 |

## 修正方針

### 1. `FormFieldError` コンポーネントの確認・使用
既存の `components/shared/FormFieldError/` コンポーネントが `C.bgDanger` を使って実装されているか確認し、使用する:

```tsx
// ❌ Before
<p className="text-red-600 text-sm mt-1">{error}</p>

// ✅ After
import { FormFieldError } from '@/components/shared/FormFieldError/FormFieldError';
{error ? <FormFieldError message={error} /> : null}
```

### 2. `FormFieldError` がトークンを使っていない場合は修正
```tsx
// components/shared/FormFieldError/FormFieldError.tsx
import { C } from '@/lib/design-tokens';

interface Props { message: string }

export function FormFieldError({ message }: Props) {
  return (
    <p
      role="alert"
      style={{ color: C.bgDanger }}
      className="text-sm mt-1"
    >
      {message}
    </p>
  );
}
```

### 3. エラー枠線を直接使用している箇所
```tsx
import { C } from '@/lib/design-tokens';

// border-red-500 → style prop
<Input
  style={hasError ? { borderColor: C.bgDanger } : undefined}
/>
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリング（Tailwind 4, Inline styles）で `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。

### `.claude/rules/accessibility-rules.md` — aria-live
エラーメッセージには `role="alert"` または `aria-live="polite"` を付与すること（スクリーンリーダー対応）。`FormFieldError` コンポーネントにこれを組み込む。

### プロジェクト内参照実装
- `components/shared/FormFieldError/` — 既存の共通エラー表示コンポーネント（内部実装要確認）

## 優先度
**Medium** — デザイン整合性の問題。既存の `FormFieldError` コンポーネントへの集約で解決できる標準的なリファクタ。

## 関連チケット
- BUG-162: 複数 feature のハードコードカラー違反
- BUG-169: 必須フィールドマーカーのハードコード
- BUG-172: AccountingDetail.tsx の包括的色違反

## 関連ファイル
- `frontend/src/components/shared/FormFieldError/` — 共通エラーコンポーネント
- `frontend/src/lib/design-tokens.ts` — C.bgDanger / C.textRequired
- `frontend/src/features/estimates/routes/EstimateForm.tsx`
- `frontend/src/features/inventory/routes/InventoryForm.tsx`
