# BUG-200: 入院管理 追加コンポーネント群のハードコードカラー違反（CarePlanDialog・DailyRecordTimeline・DailyCareNoteForm・DailyRecordSection・HospitalizationDetailActions）

## 概要

BUG-170・BUG-189 で報告した入院管理カラー違反の第三報。`CarePlanDialog.tsx`（purple）・`DailyRecordTimeline.tsx`（rose/orange）・`DailyCareNoteForm.tsx`（gray）・`DailyRecordSection.tsx`（orange/yellow/indigo colorClass）・`HospitalizationDetailActions.tsx`（red border）が新たに発見された。入院管理モジュール全体でほぼすべてのコンポーネントがデザイントークン非準拠の状態。

## 再現手順

1. 入院管理の入院詳細画面を開く
2. 「ケアプラン」タブでケアプラン追加ダイアログを開く
   → **結果**: `bg-purple-50 text-purple-700 border-purple-200` のタグ表示
3. 「日次記録」タブのタイムラインを確認
   → **結果**: `bg-rose-50 text-rose-600` / `bg-orange-50 text-orange-600` のステータス表示
4. 「日次ケアノート」入力フォームを確認
   → **結果**: `border-t border-gray-100 text-gray-500` のグレー系ハードコード

## 現状コード

### `frontend/src/features/hospitalization/components/CarePlan/CarePlanDialog.tsx:168`
```tsx
// ❌ タグに purple ハードコード
<span className="bg-purple-50 text-purple-700 border-purple-200 ...">
  {tag}
</span>
```

### `frontend/src/features/hospitalization/components/DailyRecord/DailyRecordTimeline.tsx:34,38-42`
```tsx
// ❌ ステータス別カラーに rose/orange ハードコード
// L34
<span className="bg-rose-50 text-rose-600 border-rose-200">緊急</span>
// L38-42（getTimelineColor 関数の返り値ではなく直接記述）
<span className="bg-orange-50 text-orange-600 border-orange-200">警告</span>
<span className="bg-yellow-50 text-yellow-700 border-yellow-200">注意</span>
<span className="bg-blue-50 text-blue-600 border-blue-200">情報</span>
<span className="bg-gray-50 text-gray-500 border-gray-200">完了</span>
```

### `frontend/src/features/hospitalization/components/DailyRecord/DailyCareNoteForm.tsx:49,51,52`
```tsx
// ❌ gray 系ハードコード
<div className="border-t border-gray-100 ...">  // L51
<textarea placeholder="..." className="placeholder:text-gray-400 ...">  // L49
<span className="text-gray-500">文字</span>  // L52
```

### `frontend/src/features/hospitalization/components/DailyRecord/DailyRecordSection.tsx:74,81,88`
```tsx
// ❌ colorClass として orange/yellow/indigo ハードコード文字列
const colorClass = type === "vital" ? "text-orange-600"
  : type === "note" ? "text-yellow-600"
  : "text-indigo-600";
<span className={colorClass}>...</span>
```

### `frontend/src/features/hospitalization/components/HospitalizationDetailActions.tsx:31`
```tsx
// ❌ 危険操作ボタンのボーダーに red ハードコード
<button className="... border border-red-200 ...">
  退院
</button>
```

## 影響範囲

| 対象ファイル | 行番号 | 違反 | 状態 |
|---|---|---|---|
| `features/hospitalization/components/CarePlan/CarePlanDialog.tsx` | 168 | bg-purple-50, text-purple-700, border-purple-200 | 未修正 |
| `features/hospitalization/components/DailyRecord/DailyRecordTimeline.tsx` | 34, 38-42 | bg-rose-*, text-rose-*, bg-orange-*, bg-yellow-*, bg-blue-*, bg-gray-* | 未修正 |
| `features/hospitalization/components/DailyRecord/DailyCareNoteForm.tsx` | 49, 51, 52 | placeholder:text-gray-400, border-gray-100, text-gray-500 | 未修正 |
| `features/hospitalization/components/DailyRecord/DailyRecordSection.tsx` | 74, 81, 88 | text-orange-600, text-yellow-600, text-indigo-600（colorClass） | 未修正 |
| `features/hospitalization/components/HospitalizationDetailActions.tsx` | 31 | border-red-200（退院ボタン） | 未修正 |

## 修正方針

### 1. `CarePlanDialog.tsx:168` — タグカラー
```tsx
import { BADGE } from '@/lib/design-tokens';

// Before
<span className="bg-purple-50 text-purple-700 border-purple-200">

// After（BADGE.purple がなければ追加、または直接指定）
<span style={{ backgroundColor: '#EDE9FE', color: '#6D28D9', borderColor: '#DDD6FE' }}>
// → design-tokens に BADGE.purple を追加することを推奨
```

### 2. `DailyRecordTimeline.tsx:34,38-42` — ステータス別バッジをトークンに
```tsx
import { BADGE, C } from '@/lib/design-tokens';

// ステータスマッピング関数
const getTimelineStyle = (status: string) => {
  switch (status) {
    case 'urgent':  return BADGE.red;
    case 'warning': return BADGE.orange;  // or { backgroundColor: `${C.bgWarn}20`, color: C.bgWarn }
    case 'info':    return BADGE.blue;
    case 'done':    return { backgroundColor: C.bgHover, color: C.textSecondary, borderColor: C.borderLight };
    default:        return {};
  }
};
```

### 3. `DailyCareNoteForm.tsx` — gray 系をトークンに
```tsx
import { C } from '@/lib/design-tokens';

// L51: border-gray-100 → style={{ borderColor: C.borderLight }}
// L49: placeholder:text-gray-400 → style={{ "--placeholder-color": C.textSecondary }}（カスタムプロパティ）
// L52: text-gray-500 → style={{ color: C.textSecondary }}
```

### 4. `DailyRecordSection.tsx:74,81,88` — colorClass をトークンに
```tsx
import { C } from '@/lib/design-tokens';

// Before
const colorClass = type === "vital" ? "text-orange-600" : ...;

// After: className ではなく style で
const getTypeStyle = (type: string): React.CSSProperties => {
  if (type === 'vital') return { color: C.bgWarn };
  if (type === 'note')  return { color: C.bgStatusGreenDot };  // yellow 相当
  return { color: C.bgAccent };  // indigo 相当
};
<span style={getTypeStyle(type)}>
```

### 5. `HospitalizationDetailActions.tsx:31`
```tsx
import { C } from '@/lib/design-tokens';

// Before
<button className="... border border-red-200 ...">

// After
<button
  style={{ borderColor: `${C.bgDanger}40` }}
  className="..."
>
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリング（Tailwind 4, Inline styles）で `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。

## 優先度
**Medium** — 入院管理は臨床上重要な機能。BUG-170/BUG-189 の対応時に合わせて入院管理モジュール全体を一括修正することを推奨。

## 関連チケット
- BUG-170: 入院管理 getTypeColor 系関数の包括的カラー違反（第一報）
- BUG-189: 入院管理 DailyVitals/DailyStaffNotes/TimingSection/CarePlanItemRow/DischargeAlertDialog（第二報）

## 関連ファイル
- `frontend/src/features/hospitalization/components/CarePlan/CarePlanDialog.tsx`
- `frontend/src/features/hospitalization/components/DailyRecord/DailyRecordTimeline.tsx`
- `frontend/src/features/hospitalization/components/DailyRecord/DailyCareNoteForm.tsx`
- `frontend/src/features/hospitalization/components/DailyRecord/DailyRecordSection.tsx`
- `frontend/src/features/hospitalization/components/HospitalizationDetailActions.tsx`
- `frontend/src/lib/design-tokens.ts`
