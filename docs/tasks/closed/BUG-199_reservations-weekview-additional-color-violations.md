# BUG-199: 予約 WeekView・ReservationDetailModal 追加カラー違反

## 概要

BUG-183 で報告した `MonthView.tsx` の予約カラー違反に加え、`WeekView.tsx:444` の現在時刻インジケーターと `ReservationDetailModal.tsx:195-196` のメモ欄 amber カラーが未起票だった。また `AccountingList.tsx:329` のリンク hover 色も新規違反として確認された。

## 再現手順

1. `/reservations`（予約管理）の週表示を開く
2. 現在時刻を表す横線（赤い線）を確認
   → **結果**: `border-red-400` でハードコード
3. 予約詳細モーダルを開き、「備考」セクションを確認
   → **結果**: `bg-amber-50/50 border-amber-100 text-amber-700` でハードコード
4. `/accounting`（会計一覧）の詳細リンクを確認
   → **結果**: `text-blue-500 hover:text-blue-700 hover:bg-blue-50` でハードコード

## 現状コード

### `frontend/src/features/reservations/components/WeekView.tsx:444`
```tsx
// ❌ 現在時刻インジケーターに red ハードコード
<div className={`absolute w-full border-t-2 border-red-400 z-20 pointer-events-none`} />
```

### `frontend/src/features/reservations/components/ReservationDetailModal.tsx:195-196`
```tsx
// ❌ 備考欄に amber ハードコード（BUG-183 の L177 とは別箇所）
<div className="rounded-lg border border-amber-100 bg-amber-50/50 p-3">
  <div className="flex items-center gap-1.5 text-sm text-amber-700 mb-1.5">
```

### `frontend/src/features/accounting/routes/AccountingList.tsx:329`
```tsx
// ❌ リンクの hover 状態に blue ハードコード
<button className="text-blue-500 hover:text-blue-700 hover:bg-blue-50 ...">
  詳細を見る
</button>
```

### 比較: 正しい実装
```tsx
import { C } from '@/lib/design-tokens';

// ✅ 現在時刻ライン
<div
  style={{ borderColor: C.bgDanger }}
  className="absolute w-full border-t-2 z-20 pointer-events-none"
/>

// ✅ 備考欄（warning 系）
<div
  style={{
    borderColor: `${C.bgWarn}30`,
    backgroundColor: `${C.bgWarn}10`
  }}
  className="rounded-lg border p-3"
>
  <div style={{ color: C.bgWarn }} className="flex items-center gap-1.5 text-sm mb-1.5">

// ✅ リンクボタン
<button
  style={{ color: C.bgAccent }}
  className="hover:bg-[var(--bg-hover)] ..."
>
```

## 影響範囲

| 対象ファイル | 行番号 | 違反 | 状態 |
|---|---|---|---|
| `features/reservations/components/WeekView.tsx` | 444 | border-red-400（現在時刻ライン） | 未修正 |
| `features/reservations/components/ReservationDetailModal.tsx` | 195-196 | bg-amber-50/50, border-amber-100, text-amber-700 | 未修正 |
| `features/accounting/routes/AccountingList.tsx` | 329 | text-blue-500, hover:text-blue-700, hover:bg-blue-50 | 未修正 |

## 修正方針

### 1. `WeekView.tsx:444`
```tsx
// Before
<div className="... border-red-400 ...">

// After
<div className="..." style={{ borderColor: C.bgDanger }}>
```

### 2. `ReservationDetailModal.tsx:195-196`
```tsx
// Before
<div className="rounded-lg border border-amber-100 bg-amber-50/50 p-3">
  <div className="... text-amber-700 ...">

// After
<div
  style={{ borderColor: `${C.bgWarn}30`, backgroundColor: `${C.bgWarn}10` }}
  className="rounded-lg border p-3"
>
  <div style={{ color: C.bgWarn }} className="...">
```

### 3. `AccountingList.tsx:329`
```tsx
// Before
<button className="text-blue-500 hover:text-blue-700 hover:bg-blue-50 ...">

// After
<button
  style={{ color: C.bgAccent }}
  className="hover:bg-[var(--bg-hover)] ..."
>
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリング（Tailwind 4, Inline styles）で `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。

## 優先度
**Medium** — 予約管理・会計はコアユースケース。現在時刻ラインの red は特に視覚的影響あり。

## 関連チケット
- BUG-183: 予約カレンダー・モーダルのハードコードカラー（第一報）
- BUG-190: AccountingList orange バッジ（同ファイルの別違反）

## 関連ファイル
- `frontend/src/features/reservations/components/WeekView.tsx`
- `frontend/src/features/reservations/components/ReservationDetailModal.tsx`
- `frontend/src/features/accounting/routes/AccountingList.tsx`
