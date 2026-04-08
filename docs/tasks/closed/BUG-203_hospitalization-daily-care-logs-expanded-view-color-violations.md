# BUG-203: 入院管理 DailyCareLogsSection・HospitalizationExpandedView のハードコードカラー違反

## 概要

`DailyCareLogsSection.tsx` で orange・teal・purple・red・gray 系の Tailwind カラーが 8 箇所、`HospitalizationExpandedView.tsx` で `bg-gray-50/50` が 1 箇所ハードコードされている。前者はケアログの種別バッジ（投薬・排泄・食事・異常等）の色定義を含み、デザイントークン体系外で視覚表現が決まっている。

## 再現手順

1. 入院管理の入院詳細画面を開く
2. 「日次記録」タブのケアログセクションを確認
   → **結果**: 種別バッジが orange/teal/purple/red/gray でハードコード表示
3. 入院一覧のホバー時拡張ビューを確認
   → **結果**: `bg-gray-50/50` でハードコード背景

## 現状コード

### `frontend/src/features/hospitalization/components/DailyRecordsTab/DailyCareLogsSection.tsx`
```tsx
// ❌ L68-76: 種別バッジハードコード
<span className="bg-orange-50 border-orange-100 text-orange-700">投薬</span>   // L68
<span className="bg-teal-50 border-teal-100 text-teal-700">排泄</span>         // L70
<span className="bg-purple-50 border-purple-100 text-purple-700">観察</span>   // L72
<span className="bg-red-50 border-red-100 text-red-700">異常</span>            // L74
<span className="bg-gray-50 border-gray-100 text-gray-700">その他</span>       // L76

// ❌ L128: 警告アイコン色ハードコード
<AlertTriangle className="text-orange-500" />

// ❌ L145: 背景ハードコード
<div className="bg-gray-50 ...">

// ❌ L189: 必須マーカーハードコード
<span className="text-red-500">*</span>
```

### `frontend/src/features/hospitalization/components/HospitalizationExpandedView.tsx:54`
```tsx
// ❌ 拡張ビュー背景ハードコード
<div className="bg-gray-50/50 ...">
```

## 影響範囲

| 対象ファイル | 行番号 | 違反 | 状態 |
|---|---|---|---|
| `DailyCareLogsSection.tsx` | 68 | bg-orange-50, border-orange-100, text-orange-700 | 未修正 |
| `DailyCareLogsSection.tsx` | 70 | bg-teal-50, border-teal-100, text-teal-700 | 未修正 |
| `DailyCareLogsSection.tsx` | 72 | bg-purple-50, border-purple-100, text-purple-700 | 未修正 |
| `DailyCareLogsSection.tsx` | 74 | bg-red-50, border-red-100, text-red-700 | 未修正 |
| `DailyCareLogsSection.tsx` | 76 | bg-gray-50, border-gray-100, text-gray-700 | 未修正 |
| `DailyCareLogsSection.tsx` | 128 | text-orange-500 | 未修正 |
| `DailyCareLogsSection.tsx` | 145 | bg-gray-50 | 未修正 |
| `DailyCareLogsSection.tsx` | 189 | text-red-500（必須マーカー） | 未修正 |
| `HospitalizationExpandedView.tsx` | 54 | bg-gray-50/50 | 未修正 |

## 修正方針

### `DailyCareLogsSection.tsx` — 種別バッジをトークンに
```tsx
import { C, BADGE } from '@/lib/design-tokens';

// getCareLogTypeStyle 関数を定義
const getCareLogTypeStyle = (type: string): React.CSSProperties => {
  switch (type) {
    case 'medication': return BADGE.orange;  // 投薬
    case 'excretion':  return BADGE.green;   // 排泄（teal → green 近似）
    case 'observation': return BADGE.purple; // 観察
    case 'abnormal':   return BADGE.red;     // 異常
    default:           return BADGE.gray;    // その他
  }
};

// L68-76: Before
<span className="bg-orange-50 border-orange-100 text-orange-700">投薬</span>

// After
<span style={getCareLogTypeStyle('medication')} className="border text-xs px-1.5 py-0.5 rounded">
  投薬
</span>

// L128: 警告アイコン
<AlertTriangle style={{ color: C.bgWarn }} />

// L145: 背景
<div style={{ backgroundColor: C.bgPage }}>

// L189: 必須マーカー
<span style={{ color: C.textRequired }}>*</span>
```

### `HospitalizationExpandedView.tsx:54`
```tsx
import { C } from '@/lib/design-tokens';

// Before
<div className="bg-gray-50/50 ...">

// After
<div style={{ backgroundColor: `${C.bgPage}80` }} className="...">
// または opacity で表現
<div style={{ backgroundColor: C.bgPage }} className="... opacity-80">
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリング（Tailwind 4, Inline styles）で `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。

### プロジェクト内参照実装
- `utils/status-helpers.ts` — ステータス別スタイルマッピング関数の参照パターン

## 優先度
**Medium** — 入院管理ケアログの種別色は臨床業務で頻繁に確認する情報。BUG-202 と合わせて入院管理モジュール全体を一括修正することを推奨。

## 関連チケット
- BUG-170: 入院管理 getTypeColor/getTimelineColor 系（第一報）
- BUG-202: CarePlanTab.tsx（同モジュール前チケット）

## 関連ファイル
- `frontend/src/features/hospitalization/components/DailyRecordsTab/DailyCareLogsSection.tsx`
- `frontend/src/features/hospitalization/components/HospitalizationExpandedView.tsx`
- `frontend/src/lib/design-tokens.ts`
