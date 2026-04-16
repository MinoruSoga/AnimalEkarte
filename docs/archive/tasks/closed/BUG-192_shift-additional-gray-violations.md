# BUG-192: シフト管理追加の gray 系 Tailwind 違反（ShiftCell・ShiftCalendar・ShiftCalendarPage）

## 概要

BUG-171 で報告した `SHIFT_TYPE_COLORS` / 曜日カラーに加え、`ShiftCell.tsx`・`ShiftCalendar.tsx`・`ShiftCalendarPage.tsx` で `text-gray-*`・`border-gray-200`・`text-red-500`（エラーメッセージ）が大量にハードコードされている。特に `ShiftCalendar.tsx` は 5 箇所にわたる広範囲の gray 系違反がある。

## 再現手順

1. `/shifts`（シフト管理）を開く
2. カレンダービューの各セルのホバー状態を確認
   → **結果**: `text-gray-300 hover:text-gray-500 hover:bg-gray-50` でハードコード
3. API エラー状態を再現
   → **結果**: `text-red-500` のエラーテキストがデザイントークン未使用で表示

## 現状コード

### `frontend/src/features/shifts/components/ShiftCell.tsx:39`
```tsx
// ❌ セル未入力・ホバー状態に gray ハードコード
<span className="text-gray-300 hover:text-gray-500 hover:bg-gray-50 ...">
  +
</span>
```

### `frontend/src/features/shifts/components/ShiftCalendar.tsx:20,139,207,219,238`
```tsx
// ❌ 複数箇所で gray 系ハードコード
<th className="border-gray-200 ...">                   // L20
<div className="text-gray-400 ...">休日</div>           // L139
<div className="border-gray-200 ...">                  // L207
<td className="bg-gray-50 text-gray-500 ...">          // L219
<div className="text-gray-300">シフトなし</div>         // L238
```

### `frontend/src/features/shifts/routes/ShiftCalendarPage.tsx:54`
```tsx
// ❌ エラーメッセージに red ハードコード
<p className="text-red-500">{errorMessage}</p>
```

### 比較: 正しい実装
```tsx
import { C } from '@/lib/design-tokens';

// ✅ グレー系テキスト
style={{ color: C.textSecondary }}      // text-gray-500相当
style={{ color: C.text30 }}             // text-gray-300相当

// ✅ グレー系ボーダー
style={{ borderColor: C.borderLight }}  // border-gray-200相当

// ✅ グレー系背景
style={{ backgroundColor: C.bgPage }}  // bg-gray-50相当

// ✅ エラーメッセージ
import { FormFieldError } from '@/components/shared/FormFieldError';
<FormFieldError message={errorMessage} />
```

## 影響範囲

| 対象ファイル | 行番号 | 違反 | 状態 |
|---|---|---|---|
| `features/shifts/components/ShiftCell.tsx` | 39 | text-gray-300, hover:text-gray-500, hover:bg-gray-50 | 未修正 |
| `features/shifts/components/ShiftCalendar.tsx` | 20, 139, 207, 219, 238 | border-gray-200, text-gray-400, bg-gray-50, text-gray-500, text-gray-300 | 未修正 |
| `features/shifts/routes/ShiftCalendarPage.tsx` | 54 | text-red-500（エラーメッセージ） | 未修正 |

## 修正方針

### 1. `ShiftCell.tsx:39`
```tsx
import { C } from '@/lib/design-tokens';

// Before
<span className="text-gray-300 hover:text-gray-500 hover:bg-gray-50">+</span>

// After
<span
  style={{ color: C.text30 }}
  className="hover:bg-[var(--bg-page)] ..."
  onMouseEnter={e => { e.currentTarget.style.color = C.textSecondary; }}
  onMouseLeave={e => { e.currentTarget.style.color = C.text30; }}
>+</span>
// または Tailwind 4 の CSS変数インライン記法を使用
```

### 2. `ShiftCalendar.tsx` — gray 系一括置換
```tsx
import { C } from '@/lib/design-tokens';

// L20: border-gray-200 → style={{ borderColor: C.borderLight }}
// L139: text-gray-400 → style={{ color: C.textSecondary }}
// L207: border-gray-200 → style={{ borderColor: C.borderLight }}
// L219: bg-gray-50 text-gray-500 → style={{ backgroundColor: C.bgPage, color: C.textSecondary }}
// L238: text-gray-300 → style={{ color: C.text30 }}
```

### 3. `ShiftCalendarPage.tsx:54`
```tsx
import { FormFieldError } from '@/components/shared/FormFieldError';

// Before
<p className="text-red-500">{errorMessage}</p>

// After
<FormFieldError message={errorMessage} />
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリング（Tailwind 4, Inline styles）で `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。

## 優先度
**Medium** — シフト管理 UI のグレー系一貫性。BUG-171 の対応時に合わせて修正することで効率化できる。

## 関連チケット
- BUG-171: シフト管理の SHIFT_TYPE_COLORS 包括的カラー違反（第一報）
- BUG-173: エラーメッセージカラーパターンのハードコード

## 関連ファイル
- `frontend/src/features/shifts/components/ShiftCell.tsx`
- `frontend/src/features/shifts/components/ShiftCalendar.tsx`
- `frontend/src/features/shifts/routes/ShiftCalendarPage.tsx`
- `frontend/src/lib/design-tokens.ts`
