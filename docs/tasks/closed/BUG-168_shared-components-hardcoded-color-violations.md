# BUG-168: 共有コンポーネントのハードコードカラー違反（cross-feature 高影響）

## 概要

`components/shared/` 配下の複数コンポーネントで Tailwind プリセットカラー（`bg-blue-50`、`text-blue-600`、`text-red-600`、`hover:bg-red-50` 等）を直接使用している。共有コンポーネントは全 feature で使用されるため、影響範囲が最大級。プロジェクト規約では **すべての色指定は `C.*` / `STYLE.*` トークン経由が必須**。

## 再現手順

1. `components/shared/` 配下のファイルを開く
2. `bg-blue-` / `text-blue-` / `text-red-` / `hover:bg-red-` を検索
3. **結果**: デザイントークンを経由しないカラークラスが複数共有コンポーネントで使用されている

## 期待する動作

- 青（accent）色: `C.bgAccent` / `C.textAccentBlue` 等のトークンを使用
- 赤（danger）色: `C.bgDanger` / `C.textRequired` 等のトークンを使用
- hover/focus 状態も同様にトークンを使用

## 現状コード

### `frontend/src/components/shared/PetSelection/PetSelection.tsx`
```tsx
// ❌ ハードコード
className="bg-blue-50/50"
className="text-blue-600"
className="text-blue-700"
className="text-blue-800"
```

### `frontend/src/components/shared/ReservationFormModal/ReservationFormModal.tsx`
```tsx
// ❌ ハードコード
className="text-blue-600"
className="bg-blue-600"
className="hover:bg-red-50"
```

### `frontend/src/components/shared/ReservationFormModal/ReservationFormFields.tsx`
```tsx
// ❌ ハードコード
className="border-red-600 bg-red-50"
className="border-blue-600 bg-blue-50"
```

### `frontend/src/components/shared/RowActionDropdown/RowActionDropdown.tsx`
```tsx
// ❌ ハードコード
className="text-red-600 focus:bg-red-50"
```

### `frontend/src/components/shared/DeleteIconButton/DeleteIconButton.tsx`
```tsx
// ❌ ハードコード
className="hover:text-red-600 hover:bg-red-50"
```

### `frontend/src/components/shared/CharCountTextarea/CharCountTextarea.tsx`
```tsx
// ❌ ハードコード
className="text-red-500"
```

### 比較: 正しい実装（プロジェクト内参照実装）
```tsx
// ✅ features/owners/components/OwnerCard.tsx 等
import { C, STYLE } from '@/lib/design-tokens';

className={cn(STYLE.button.danger)}
style={{ color: C.textRequired }}
style={{ backgroundColor: C.bgAccent }}
```

## 影響範囲

| 対象ファイル | 問題箇所 | 状態 |
|---|---|---|
| `components/shared/PetSelection/PetSelection.tsx` | bg-blue-50/50, text-blue-600/700/800 | 未修正 |
| `components/shared/ReservationFormModal/ReservationFormModal.tsx` | text-blue-600, bg-blue-600, hover:bg-red-50 | 未修正 |
| `components/shared/ReservationFormModal/ReservationFormFields.tsx` | border-red-600 bg-red-50, border-blue-600 bg-blue-50 | 未修正 |
| `components/shared/RowActionDropdown/RowActionDropdown.tsx` | text-red-600 focus:bg-red-50 | 未修正 |
| `components/shared/DeleteIconButton/DeleteIconButton.tsx` | hover:text-red-600 hover:bg-red-50 | 未修正 |
| `components/shared/CharCountTextarea/CharCountTextarea.tsx` | text-red-500 | 未修正 |

## 修正方針

### 1. `PetSelection.tsx` — blue トークンに置換
```tsx
import { C } from '@/lib/design-tokens';

// bg-blue-50/50 → inline style
style={{ backgroundColor: `${C.bgAccentLight}80` }}
// text-blue-600/700/800 → トークン
style={{ color: C.textAccentBlue }}
```

### 2. `ReservationFormModal.tsx` / `ReservationFormFields.tsx` — accent/danger トークンに置換
```tsx
import { C, STYLE } from '@/lib/design-tokens';

// bg-blue-600 ボタン → STYLE.button.primary
className={cn(STYLE.button.primary)}
// border-red-600 エラー枠 → C.bgDanger
style={{ borderColor: C.bgDanger, backgroundColor: `${C.bgDanger}20` }}
```

### 3. `RowActionDropdown.tsx` / `DeleteIconButton.tsx` — danger hover をトークンに
```tsx
import { C } from '@/lib/design-tokens';

// text-red-600 hover:bg-red-50 → CSS変数 or inline style
style={{ color: C.bgDanger }}
// hover は onMouseEnter/Leave or Tailwind arbitrary value
className="hover:bg-[#FFE2DD]"  // BADGE.red の背景値を使用
```

### 4. `CharCountTextarea.tsx` — text-red-500 → C.textRequired
```tsx
import { C } from '@/lib/design-tokens';

// text-red-500 → C.textRequired
style={{ color: C.textRequired }}
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリング（Tailwind 4, Inline styles）で `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。
> **PROHIBITED**: Hexカラー（`#37352F` など）の直接指定は厳禁。

共有コンポーネントはすべての feature から参照されるため、ここでのトークン違反は全 feature に影響する。最優先で修正すること。

### プロジェクト内参照実装
- `features/owners/routes/OwnersList.tsx` — `C.bgAccent` / `STYLE.button.primary` の正しい使用例

## 優先度
**High** — 共有コンポーネントへの違反のため全 feature (16機能) に影響が及ぶ。次回リリースまでに対応が必要。

## 関連チケット
- BUG-162: 複数 feature のハードコードカラー違反（feature 固有）
- BUG-167: PALETTE 直接使用

## 関連ファイル
- `frontend/src/lib/design-tokens.ts` — C, STYLE, BADGE トークン定義
- `frontend/src/components/shared/PetSelection/PetSelection.tsx`
- `frontend/src/components/shared/ReservationFormModal/ReservationFormModal.tsx`
- `frontend/src/components/shared/ReservationFormModal/ReservationFormFields.tsx`
- `frontend/src/components/shared/RowActionDropdown/RowActionDropdown.tsx`
- `frontend/src/components/shared/DeleteIconButton/DeleteIconButton.tsx`
- `frontend/src/components/shared/CharCountTextarea/CharCountTextarea.tsx`
