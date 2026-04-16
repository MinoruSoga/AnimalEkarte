# BUG-202: 入院管理 CarePlanTab のハードコードカラー違反（13箇所）

## 概要

`features/hospitalization/components/CarePlanTab/CarePlanTab.tsx` で orange・blue・purple・green・gray・indigo 系の Tailwind プリセットカラーが 13 箇所にわたって使用されている。BUG-170/189/200 で入院管理の他コンポーネントを対象としてきたが、`CarePlanTab.tsx` 自体は未起票だった。

## 再現手順

1. 入院管理の入院詳細画面を開く
2. 「ケアプラン」タブを選択
3. ケアプランのタイプ別アイコン色・ステータスバッジを確認
   → **結果**: orange/blue/purple/green/gray/indigo の Tailwind プリセットが使用されている

## 現状コード

### `frontend/src/features/hospitalization/components/CarePlanTab/CarePlanTab.tsx`
```tsx
// ❌ L52-56: タイプ別アイコン色ハードコード
text-orange-500   // 投薬
text-blue-500     // 処置
text-purple-500   // 観察
text-green-500    // 食事
text-gray-400     // その他

// ❌ L62: 処置バッジ
<span className="bg-blue-100 text-blue-700">処置</span>

// ❌ L69: 食事バッジ
<span className="bg-green-100 text-green-700">食事</span>

// ❌ L75: その他バッジ
<span className="bg-gray-100 text-gray-500">その他</span>

// ❌ L89: 観察バッジ
<span className="bg-indigo-100 text-indigo-700">観察</span>

// ❌ L90: 無効状態
<span className="bg-gray-50 text-gray-300">無効</span>

// ❌ L126: 編集中背景
<div className="bg-blue-50/50 border-blue-100 ...">

// ❌ L142, L270: 補助テキスト
<span className="text-gray-500">...</span>
```

## 影響範囲

| 行番号 | 違反 | 状態 |
|---|---|---|
| L52 | text-orange-500 | 未修正 |
| L53 | text-blue-500 | 未修正 |
| L54 | text-purple-500 | 未修正 |
| L55 | text-green-500 | 未修正 |
| L56 | text-gray-400 | 未修正 |
| L62 | bg-blue-100 text-blue-700 | 未修正 |
| L69 | bg-green-100 text-green-700 | 未修正 |
| L75 | bg-gray-100 text-gray-500 | 未修正 |
| L89 | bg-indigo-100 text-indigo-700 | 未修正 |
| L90 | bg-gray-50 text-gray-300 | 未修正 |
| L126 | bg-blue-50/50 border-blue-100 | 未修正 |
| L142 | text-gray-500 | 未修正 |
| L270 | text-gray-500 | 未修正 |

## 修正方針

```tsx
import { C, BADGE } from '@/lib/design-tokens';

// L52-56: アイコン色をトークンに
const getTypeIconStyle = (type: string): React.CSSProperties => ({
  color: {
    medication: C.bgWarn,          // orange
    treatment:  C.bgAccent,        // blue
    observation: C.bgStatusPurple, // purple (C.textStatusPurple)
    meal:       C.bgStatusGreenDot,// green
    other:      C.textSecondary,   // gray
  }[type] ?? C.textSecondary,
});

// L62,69,75,89,90: バッジをトークンに
// bg-blue-100 text-blue-700   → style={BADGE.blue}
// bg-green-100 text-green-700 → style={BADGE.green}
// bg-gray-100 text-gray-500   → style={BADGE.gray}
// bg-indigo-100 text-indigo-700 → style={BADGE.purple}
// bg-gray-50 text-gray-300    → style={{ backgroundColor: C.bgPage, color: C.text30 }}

// L126: 編集中背景
style={{ backgroundColor: `${C.bgAccent}08`, borderColor: `${C.bgAccent}20` }}

// L142, L270: 補助テキスト
style={{ color: C.textSecondary }}
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリング（Tailwind 4, Inline styles）で `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。

## 優先度
**Medium** — 入院管理のケアプランはコア機能。BUG-170/189/200 の対応時に合わせて入院管理モジュール全体を一括修正することを推奨。

## 関連チケット
- BUG-170: 入院管理 getTypeColor 系関数（第一報）
- BUG-189: DailyVitals/DailyStaffNotes 等（第二報）
- BUG-200: CarePlanDialog/DailyRecordTimeline 等（第三報）

## 関連ファイル
- `frontend/src/features/hospitalization/components/CarePlanTab/CarePlanTab.tsx`
- `frontend/src/lib/design-tokens.ts`
