# BUG-170: 入院管理 feature 全体のハードコードカラー違反（40+箇所）

## 概要

`features/hospitalization/` 配下で Tailwind プリセットカラーのハードコードが 40 箇所以上存在する。特に `CarePlanTab`、`CarePlanItemRow`、`DailyCareLogsSection`、`DailyRecordTimeline` でそれぞれ `getTypeColor()` / `getTimelineColor()` / `getCareLogColor()` などの内部関数が Tailwind クラス文字列を直接返しており、設計トークンを一切使っていない。

## 再現手順

1. 入院管理画面（入院詳細 → ケアプランタブ / 日次記録タブ）を開く
2. ケアプランの種別バッジ、タイムラインのカラー、ケアログの種別アイコン色を確認
3. **結果**: `bg-orange-100 text-orange-800`、`bg-blue-100`、`text-green-500` 等のハードコード色が使われている

## 期待する動作

- 種別・状態別の色分けはすべて `C.*` / `STYLE.*` / `BADGE.*` トークンを使用する
- 動的な色マッピングは `getXxxColor()` 関数内でもトークン値を返す

## 現状コード

### `frontend/src/features/hospitalization/components/CarePlan/CarePlanItemRow.tsx:46-49`
```tsx
// ❌ getTypeColor() が Tailwind クラスをハードコード返却
function getTypeColor(type: string) {
  switch (type) {
    case '投薬': return 'bg-orange-100 text-orange-800';
    case '処置': return 'bg-blue-100 text-blue-800';
    case '検査': return 'bg-purple-100 text-purple-800';
    default:     return 'bg-gray-100 text-gray-800';
  }
}
```

### `frontend/src/features/hospitalization/components/DailyRecord/DailyRecordTimeline.tsx:38-42`
```tsx
// ❌ getTimelineColor() が Tailwind クラスをハードコード返却
function getTimelineColor(type: string) {
  switch (type) {
    case '投薬': return 'text-orange-500 bg-orange-50';
    case '処置': return 'text-blue-500 bg-blue-50';
    case '検査': return 'text-purple-500 bg-purple-50';
    default:     return 'text-gray-500 bg-gray-50';
  }
}
```

### `frontend/src/features/hospitalization/components/DailyRecordsTab/DailyCareLogsSection.tsx:68-76`
```tsx
// ❌ getCareLogColor() が Tailwind クラスをハードコード返却
function getCareLogColor(category: string) {
  switch (category) {
    case '食事':  return 'text-orange-500';
    case '排泄':  return 'text-teal-500';
    case '運動':  return 'text-purple-500';
    case '異常':  return 'text-red-500';
    default:      return 'text-gray-500';
  }
}
```

### `frontend/src/features/hospitalization/components/CarePlanTab/CarePlanTab.tsx`
```tsx
// ❌ アイコン色ハードコード
className="text-orange-500"
className="text-blue-500"
className="text-purple-500"
className="text-green-500"
// ❌ バッジ色ハードコード
className="bg-blue-100 text-blue-800"
className="bg-green-100 text-green-800"
className="bg-gray-100 text-gray-800"
```

### 比較: 正しい実装
```tsx
import { C, BADGE } from '@/lib/design-tokens';

// ✅ BADGE トークンを使ったカラーマッピング
function getTypeStyle(type: string): React.CSSProperties {
  switch (type) {
    case '投薬': return BADGE.orange;
    case '処置': return BADGE.blue;
    case '検査': return BADGE.purple;
    default:     return BADGE.gray;
  }
}

// ✅ アイコン色はトークン
style={{ color: C.bgStatusOrangeDot }}  // オレンジ
style={{ color: C.bgAccent }}           // ブルー
```

## 影響範囲

| 対象ファイル | 違反箇所数 | 状態 |
|---|---|---|
| `features/hospitalization/components/CarePlan/CarePlanItemRow.tsx` | ~4箇所 (L46-49) | 未修正 |
| `features/hospitalization/components/DailyRecord/DailyRecordTimeline.tsx` | ~4箇所 (L38-42) | 未修正 |
| `features/hospitalization/components/DailyRecordsTab/DailyCareLogsSection.tsx` | ~5箇所 (L68-76) | 未修正 |
| `features/hospitalization/components/CarePlanTab/CarePlanTab.tsx` | ~15箇所以上 | 未修正 |
| その他 hospitalization コンポーネント | 未調査 | 未修正 |

## 修正方針

### 1. `getTypeColor()` / `getTimelineColor()` / `getCareLogColor()` 関数の置換

各関数を `string` (Tailwindクラス) 返却から `React.CSSProperties` 返却に変更し、BADGE トークンを使用:

```tsx
import { BADGE, C } from '@/lib/design-tokens';

// Before: Tailwindクラス文字列を返す
function getTypeColor(type: string): string { ... }

// After: CSSProperties を返す
function getTypeStyle(type: string): React.CSSProperties {
  switch (type) {
    case '投薬': return BADGE.orange;
    case '処置': return BADGE.blue;
    case '検査': return BADGE.purple;
    default:     return BADGE.gray;
  }
}

// 使用側
<span style={getTypeStyle(item.type)}>...</span>
```

### 2. `CarePlanTab.tsx` のアイコン・バッジ色置換

```tsx
import { C, BADGE } from '@/lib/design-tokens';

// アイコン色
style={{ color: C.bgStatusOrangeDot }}  // text-orange-500 の代替
style={{ color: C.bgAccent }}           // text-blue-500 の代替
style={{ color: C.bgStatusPurpleDot }}  // text-purple-500 の代替
style={{ color: C.bgStatusGreenDot }}   // text-green-500 の代替

// バッジ
style={BADGE.blue}
style={BADGE.green}
style={BADGE.gray}
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリング（Tailwind 4, Inline styles）で `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。
> **PROHIBITED**: Hexカラー（`#37352F` など）の直接指定は厳禁。

動的な色マッピング関数（`getTypeColor()` 等）もこのルールの対象。Tailwind クラス文字列ではなく `React.CSSProperties` + トークンを返すよう統一する。

### プロジェクト内参照実装
- `utils/status-helpers.ts` — ステータス別色マッピングをトークンで実装している参照実装

## 優先度
**Medium** — 機能に影響はないが、入院管理画面全体で設計トークンが無視されており、将来のテーマ変更時に手動修正が必要になる。

## 関連チケット
- BUG-162: 複数 feature のハードコードカラー違反
- BUG-167: PALETTE 直接使用

## 関連ファイル
- `frontend/src/features/hospitalization/components/CarePlan/CarePlanItemRow.tsx`
- `frontend/src/features/hospitalization/components/DailyRecord/DailyRecordTimeline.tsx`
- `frontend/src/features/hospitalization/components/DailyRecordsTab/DailyCareLogsSection.tsx`
- `frontend/src/features/hospitalization/components/CarePlanTab/CarePlanTab.tsx`
- `frontend/src/lib/design-tokens.ts` — BADGE, C トークン定義
