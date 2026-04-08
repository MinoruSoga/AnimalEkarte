# BUG-189: 入院管理追加のハードコードカラー違反（DailyVitals・DailyStaffNotes・TimingSection・CarePlanItemRow・DischargeAlertDialog）

## 概要

`features/hospitalization/` 配下の複数コンポーネントで Tailwind プリセットカラー（blue・green・gray）がハードコードされている。BUG-170 で報告した `getTypeColor()`/`getTimelineColor()` 等の関数に加え、新たに 5 ファイルで追加違反が発見された。

## 再現手順

1. 入院管理の入院詳細画面を開く
2. 「バイタル」「スタッフノート」「ケアプラン」タブを確認
3. **結果**: blue・green 系カラーがハードコードで適用されている

## 現状コード

### `frontend/src/features/hospitalization/components/DailyVitalsSection.tsx:88,116`
```tsx
// ❌ バイタル値表示に blue ハードコード
<span className="text-blue-500">BPM</span>
<span className="text-blue-700 bg-blue-50 border-blue-100">正常</span>
```

### `frontend/src/features/hospitalization/components/DailyStaffNotesSection.tsx:78,103,105`
```tsx
// ❌ スタッフノートに green ハードコード
<span className="text-green-500">✓</span>
<div className="bg-green-50 text-green-700">記録済み</div>
```

### `frontend/src/features/hospitalization/components/TimingSection.tsx:46,62`
```tsx
// ❌ 処置完了ステータスに green ハードコード
<div className="bg-green-100 text-green-600">完了</div>
<div className="bg-green-50">...</div>
```

### `frontend/src/features/hospitalization/components/CarePlanItemRow.tsx:80`
```tsx
// ❌ ステータスインジケータドットにハードコード
<span className={`... ${isDone ? "bg-green-500" : "bg-gray-300"}`} />
```

### `frontend/src/features/hospitalization/components/DischargeAlertDialog.tsx:52`
```tsx
// ❌ Danger ボタンに Tailwind ハードコード
<button className="bg-red-600 hover:bg-red-700 text-white ...">
  退院処理を実行
</button>
```

## 影響範囲

| 対象ファイル | 行番号 | 違反 | 状態 |
|---|---|---|---|
| `features/hospitalization/components/DailyVitalsSection.tsx` | 88, 116 | text-blue-*, bg-blue-*, border-blue-* | 未修正 |
| `features/hospitalization/components/DailyStaffNotesSection.tsx` | 78, 103, 105 | text-green-*, bg-green-* | 未修正 |
| `features/hospitalization/components/TimingSection.tsx` | 46, 62 | bg-green-*, text-green-* | 未修正 |
| `features/hospitalization/components/CarePlanItemRow.tsx` | 80 | bg-green-500, bg-gray-300 | 未修正 |
| `features/hospitalization/components/DischargeAlertDialog.tsx` | 52 | bg-red-600 hover:bg-red-700 | 未修正 |

## 修正方針

### 1. DailyVitalsSection.tsx — blue 系をトークンに
```tsx
import { C } from '@/lib/design-tokens';

// Before
<span className="text-blue-500">BPM</span>
<span className="text-blue-700 bg-blue-50 border-blue-100">正常</span>

// After
<span style={{ color: C.bgAccent }}>BPM</span>
<span style={{ color: C.bgAccent, backgroundColor: `${C.bgAccent}15`, borderColor: `${C.bgAccent}30` }}>正常</span>
```

### 2. DailyStaffNotesSection.tsx / TimingSection.tsx — green 系をトークンに
```tsx
import { C, BADGE } from '@/lib/design-tokens';

// 完了バッジ
<span style={BADGE.green}>完了</span>
// または
<span style={{ backgroundColor: `${C.bgStatusGreenDot}20`, color: C.bgStatusGreenDot }}>記録済み</span>
```

### 3. CarePlanItemRow.tsx — ステータスドット
```tsx
import { C } from '@/lib/design-tokens';

// Before
<span className={isDone ? "bg-green-500" : "bg-gray-300"} />

// After
<span style={{ backgroundColor: isDone ? C.bgStatusGreenDot : C.borderMedium }} />
```

### 4. DischargeAlertDialog.tsx — Danger ボタン
```tsx
import { STYLE } from '@/lib/design-tokens';

// Before
<button className="bg-red-600 hover:bg-red-700 text-white ...">

// After
<button className={STYLE.button.danger}>
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリング（Tailwind 4, Inline styles）で `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。

## 優先度
**Medium** — 入院管理画面の状態表示・ボタン色のデザイントークン統一。DischargeAlertDialog の Danger ボタンは STYLE.button.danger への統一が必要。

## 関連チケット
- BUG-170: 入院管理 getTypeColor/getTimelineColor 関数の包括的カラー違反（第一報）

## 関連ファイル
- `frontend/src/features/hospitalization/components/DailyVitalsSection.tsx`
- `frontend/src/features/hospitalization/components/DailyStaffNotesSection.tsx`
- `frontend/src/features/hospitalization/components/TimingSection.tsx`
- `frontend/src/features/hospitalization/components/CarePlanItemRow.tsx`
- `frontend/src/features/hospitalization/components/DischargeAlertDialog.tsx`
- `frontend/src/lib/design-tokens.ts`
