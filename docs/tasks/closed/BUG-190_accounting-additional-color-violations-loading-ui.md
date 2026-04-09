# BUG-190: 会計追加のハードコードカラー違反・ローディング UI 問題（AccountingList・AccountingDetail・AccountingDocument）

## 概要

BUG-172・BUG-174 で報告した会計機能の違反に加え、`AccountingList.tsx` の返金バッジ（orange）・`AccountingDetail.tsx` の保険金額（green）・生の `<div>読み込み中...</div>` が新たに発見された。特に `AccountingDetail.tsx:1064-1065` のローディング UI は BUG-181/BUG-188 と同一パターンの白画面/粗雑 UI 問題。

## 再現手順

1. `/accounting`（会計一覧）を開き、返金ステータスのレコードを確認
   → **結果**: `bg-orange-50 text-orange-600 border-orange-200` でバッジが表示される
2. `/accounting/:id`（会計詳細）を開き、データ取得中を確認
   → **結果**: 「読み込み中...」の生テキスト表示（`LoadingFallback` 未使用）
3. 保険金額フィールドを確認
   → **結果**: `text-green-700 bg-green-50` ハードコード

## 現状コード

### `frontend/src/features/accounting/routes/AccountingList.tsx:317`
```tsx
// ❌ 返金バッジ — orange ハードコード
<span className="bg-orange-50 text-orange-600 border-orange-200 ...">
  返金
</span>
```

### `frontend/src/features/accounting/routes/AccountingDetail.tsx:440`
```tsx
// ❌ 保険金額 — green ハードコード
<span className="text-green-700 bg-green-50">
  ¥{insuranceAmount}
</span>
```

### `frontend/src/features/accounting/routes/AccountingDetail.tsx:1064-1065`
```tsx
// ❌ 素のテキストでローディング・空データ表示
if (isLoading) return <div>読み込み中...</div>;
if (!data) return <div>データが見つかりません</div>;
```

### `frontend/src/features/accounting/components/AccountingDocument.tsx:77,177`
```tsx
// ❌ 保険金額列に green ハードコード（BUG-174 の gray 違反と別箇所）
<td className="text-green-700">保険適用額</td>
<span className="text-green-700">¥{insurance}</span>
```

## 影響範囲

| 対象ファイル | 行番号 | 違反 | 状態 |
|---|---|---|---|
| `features/accounting/routes/AccountingList.tsx` | 317 | bg-orange-50, text-orange-600, border-orange-200 | 未修正 |
| `features/accounting/routes/AccountingDetail.tsx` | 440 | text-green-700, bg-green-50 | 未修正 |
| `features/accounting/routes/AccountingDetail.tsx` | 1064-1065 | 素 div ローディング・空データ表示 | 未修正 |
| `features/accounting/components/AccountingDocument.tsx` | 77, 177 | text-green-700 | 未修正 |

## 修正方針

### 1. `AccountingList.tsx:317` — 返金バッジ
```tsx
import { C, BADGE } from '@/lib/design-tokens';

// Before
<span className="bg-orange-50 text-orange-600 border-orange-200 ...">返金</span>

// After — BADGE.orange がない場合は直接スタイル指定
<span
  style={{ backgroundColor: `${C.bgWarn}20`, color: C.bgWarn, borderColor: `${C.bgWarn}40` }}
  className="..."
>
  返金
</span>
```

### 2. `AccountingDetail.tsx:440` — 保険金額
```tsx
import { C } from '@/lib/design-tokens';

// Before
<span className="text-green-700 bg-green-50">¥{insuranceAmount}</span>

// After
<span style={{ color: C.bgStatusGreenDot, backgroundColor: `${C.bgStatusGreenDot}15` }}>
  ¥{insuranceAmount}
</span>
```

### 3. `AccountingDetail.tsx:1064-1065` — ローディング・空データ表示
```tsx
import { LoadingFallback } from '@/components/shared/LoadingFallback';
import { ErrorFallback } from '@/components/shared/ErrorFallback';

// Before
if (isLoading) return <div>読み込み中...</div>;
if (!data) return <div>データが見つかりません</div>;

// After
if (isLoading) return <LoadingFallback />;
if (!data) return <ErrorFallback message="データが見つかりません" />;
```

### 4. `AccountingDocument.tsx:77,177` — 保険金額列
```tsx
import { C } from '@/lib/design-tokens';

// Before
<td className="text-green-700">保険適用額</td>

// After
<td style={{ color: C.bgStatusGreenDot }}>保険適用額</td>
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリング（Tailwind 4, Inline styles）で `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。

### `.claude/CLAUDE.md` — Frontend ベストプラクティス
ローディング中・エラー時は適切なフォールバック UI を表示すること。

## 優先度
**Medium** — カラー違反は機能的問題なし。`AccountingDetail` のローディング UI 問題（1064-1065）は BUG-181/188 と同パターンで修正優先度を High にすべき。

## 関連チケット
- BUG-172: AccountingDetail 包括的カラー違反（第一報）
- BUG-174: AccountingDocument 印刷ビューの gray 系違反
- BUG-181: VaccinationForm/TrimmingForm の return null（同パターン）
- BUG-188: HospitalizationDetail の素 div ローディング（同パターン）

## 関連ファイル
- `frontend/src/features/accounting/routes/AccountingList.tsx`
- `frontend/src/features/accounting/routes/AccountingDetail.tsx`
- `frontend/src/features/accounting/components/AccountingDocument.tsx`
- `frontend/src/lib/design-tokens.ts`
