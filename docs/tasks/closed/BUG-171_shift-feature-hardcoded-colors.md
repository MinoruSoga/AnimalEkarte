# BUG-171: シフト管理 feature のハードコードカラー違反（SHIFT_TYPE_COLORS・カレンダー曜日色）

## 概要

`features/shifts/` 配下で SHIFT_TYPE_COLORS 定数（型定義ファイル内）とカレンダーコンポーネントの曜日色が Tailwind プリセットカラーをハードコードしている。特に `shifts/types/index.ts` 内の `SHIFT_TYPE_COLORS` マップは Tailwind クラス文字列を直接持っており、`design-tokens.ts` のトークンを一切参照していない。

## 再現手順

1. シフト管理画面を開く
2. シフトカレンダーの曜日ヘッダー（日曜日 = 赤、土曜日 = 青）と各シフト種別のバッジ色を確認
3. **結果**: `text-red-500`（日曜）、`text-blue-500`（土曜）、`bg-blue-100`、`bg-green-100` 等のハードコード色が使われている

## 期待する動作

- シフト種別の色マッピングは `BADGE.*` または `C.*` トークンを使用した CSSProperties を返す
- 曜日色（日・土）もトークン定数を使用する

## 現状コード

### `frontend/src/features/shifts/types/index.ts:63-67`
```tsx
// ❌ 型定義ファイル内に Tailwind クラスのハードコードマップ
export const SHIFT_TYPE_COLORS: Record<string, string> = {
  '通常':     'bg-blue-100 text-blue-800',
  '午前のみ': 'bg-green-100 text-green-800',
  '午後のみ': 'bg-teal-100 text-teal-800',
  '休日':     'bg-gray-100 text-gray-800',
  '特別':     'bg-purple-100 text-purple-800',
};
```

### `frontend/src/features/shifts/components/ShiftCalendar/ShiftCalendar.tsx`
```tsx
// ❌ 曜日色ハードコード
className={cn(
  dayOfWeek === 0 && "text-red-500",    // 日曜
  dayOfWeek === 6 && "text-blue-500",   // 土曜
)}
// ❌ その他 text-gray-* ハードコード
className="text-gray-400"
className="text-gray-500"
className="text-gray-600"
```

### 比較: 正しい実装
```tsx
import { BADGE, C } from '@/lib/design-tokens';
import type React from 'react';

// ✅ CSSProperties マップ（BADGE トークン使用）
export const SHIFT_TYPE_STYLES: Record<string, React.CSSProperties> = {
  '通常':     BADGE.blue,
  '午前のみ': BADGE.green,
  '午後のみ': BADGE.teal,   // もしくは近似トークン
  '休日':     BADGE.gray,
  '特別':     BADGE.purple,
};

// ✅ 曜日色
const sundayStyle = { color: C.bgDanger };    // 赤 (日曜)
const saturdayStyle = { color: C.bgAccent };  // 青 (土曜)
```

## 影響範囲

| 対象ファイル | 違反箇所 | 状態 |
|---|---|---|
| `features/shifts/types/index.ts` | SHIFT_TYPE_COLORS マップ (L63-67) | 未修正 |
| `features/shifts/components/ShiftCalendar/ShiftCalendar.tsx` | 曜日色 text-red-500/text-blue-500 + text-gray-* | 未修正 |
| `features/shifts/components/ShiftCell/ShiftCell.tsx` | 未調査（SHIFT_TYPE_COLORS 参照箇所） | 未修正 |

## 修正方針

### 1. `shifts/types/index.ts` の SHIFT_TYPE_COLORS を CSSProperties に変更
```tsx
import { BADGE } from '@/lib/design-tokens';
import type React from 'react';

// Before
export const SHIFT_TYPE_COLORS: Record<string, string> = { ... };

// After
export const SHIFT_TYPE_STYLES: Record<string, React.CSSProperties> = {
  '通常':     BADGE.blue,
  '午前のみ': BADGE.green,
  '午後のみ': { ...BADGE.green, filter: 'hue-rotate(180deg)' },  // teal近似 or 専用トークン追加
  '休日':     BADGE.gray,
  '特別':     BADGE.purple,
};
```

### 2. `ShiftCalendar.tsx` の曜日色をトークンに置換
```tsx
import { C } from '@/lib/design-tokens';

// 曜日色
const getDayStyle = (dayOfWeek: number): React.CSSProperties => {
  if (dayOfWeek === 0) return { color: C.bgDanger };   // 日曜
  if (dayOfWeek === 6) return { color: C.bgAccent };   // 土曜
  return {};
};

// text-gray-* → C.textSecondary など
style={{ color: C.textSecondary }}
```

### 3. SHIFT_TYPE_COLORS 参照箇所を SHIFT_TYPE_STYLES に更新
`ShiftCell.tsx` 等で `SHIFT_TYPE_COLORS[type]` を参照している箇所を `SHIFT_TYPE_STYLES[type]` に変更し、`className` → `style` に変更。

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリング（Tailwind 4, Inline styles）で `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。

型定義ファイル（`types/index.ts`）内であっても、スタイル値を定義する場合はトークンを使用すること。

### プロジェクト内参照実装
- `utils/status-helpers.ts` — `BADGE.*` トークンを使ったステータス別スタイルマッピングの参照実装

## 優先度
**Medium** — シフト管理画面のデザイン整合性の問題。型定義ファイルにスタイル値が混在している設計上の問題も含む。

## 関連チケット
- BUG-162: 複数 feature のハードコードカラー違反
- BUG-164: シフトカレンダーのローディング状態欠如

## 関連ファイル
- `frontend/src/features/shifts/types/index.ts`
- `frontend/src/features/shifts/components/ShiftCalendar/ShiftCalendar.tsx`
- `frontend/src/features/shifts/components/ShiftCell/ShiftCell.tsx`（要調査）
- `frontend/src/lib/design-tokens.ts` — BADGE, C トークン定義
