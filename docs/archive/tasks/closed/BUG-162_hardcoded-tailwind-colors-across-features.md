# BUG-162: ボタン・バッジの Tailwind カラークラスハードコード（複数 feature）

## 概要

複数 feature のコンポーネントで `bg-blue-600`、`text-red-600`、`bg-red-100` 等の Tailwind ハードコードカラークラスを使用している。
プロジェクト規約では **すべての色指定は `src/lib/design-tokens.ts` の `C.*` / `STYLE.*` 定数経由** が必須であり、`blue-*` / `red-*` 等の Tailwind プリセットカラーを直接使うことは禁止。

## 再現手順

1. 各ファイルをコードエディタで開く
2. `bg-blue-` / `text-blue-` / `bg-red-` / `text-red-` を検索
3. **結果**: デザイントークンを経由しないカラークラスが複数箇所で使用されている

## 期待する動作

- すべての青 (accent) 色: `C.bgAccent`（`bg-[#2383E2]`）/ `C.textAccentBlue` 等のトークンを使用
- すべての赤 (danger) 色: `C.bgDanger`（`bg-[#C0392B]`）/ `C.textRequired`（`text-[#E03E3E]`）等を使用
- バッジ色: `STYLE.badge.blue` / `STYLE.badge.red` 等の既定バッジスタイルを使用

## 現状コード

### `frontend/src/features/accounting/routes/AccountingList.tsx:329付近`
```tsx
// ❌ ハードコード
className="h-8 w-8 text-blue-500 hover:text-blue-700 hover:bg-blue-50"

// ❌ バッジにハードコード
<span className="ml-2 text-[10px] text-blue-500 bg-blue-50 px-1.5 py-0.5 rounded">
  カルテ連携
</span>
```

### `frontend/src/features/auth/routes/LoginForm.tsx:68付近`
```tsx
// ❌ ハードコード
<span className="text-xs px-1.5 py-px rounded-[3px] text-red-600 bg-red-50">
  システム管理者
</span>
```

### `frontend/src/features/reservations/components/MonthView.tsx:79付近`
```tsx
// ❌ カレンダー日付ボタンにハードコード
className={`...${isSameDay(day, new Date()) ? "bg-blue-600 text-white shadow-sm" : "hover:bg-blue-100 hover:text-blue-700"}`}

// ❌ 初診/再診バッジ
? <span className="bg-red-100 text-red-600 text-[10px] px-1 rounded flex-shrink-0">初</span>
: <span className="bg-blue-100 text-blue-600 text-[10px] px-1 rounded flex-shrink-0">再</span>
```

### `frontend/src/features/medical-records/components/ExaminationGroup.tsx:105付近`
```tsx
// ❌ 削除ボタンにハードコード
className="h-10 px-3 text-sm bg-red-500 hover:bg-red-600"

// ❌ 青アウトラインボタンにハードコード
className="h-10 px-3 text-sm text-blue-600 border-blue-600 bg-blue-50"
```

### `frontend/src/features/hospitalization/components/CarePlanTab.tsx:62付近`
```tsx
// ❌ バッジにハードコード
<span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-700">

// ❌ info ボックスにハードコード
<div className="flex flex-col gap-2 p-3 bg-blue-50/50 rounded-lg border border-blue-100">
```

### `frontend/src/features/hospitalization/components/DischargeAlertDialog.tsx:52付近`
```tsx
// ❌ 破壊的アクションボタンにハードコード
className={`bg-red-600 hover:bg-red-700 ${H_STYLES.button.action}`}
```

### `frontend/src/features/owners/routes/OwnersList.tsx:333付近`
```tsx
// ❌ ステータスバッジにハードコード
<span className="inline-flex items-center rounded px-1.5 py-0.5 text-xs font-semibold bg-red-100 text-red-700 border border-red-300">
```

### 比較: 正しい実装（プロジェクト内参照実装）
```tsx
// ✅ STYLE.badge.blue / STYLE.badge.red を使用
import { C, STYLE } from '@/lib/design-tokens';

<span className={cn("text-xs px-1.5 py-0.5 rounded", STYLE.badge.blue)}>
  カルテ連携
</span>

// ✅ 破壊的ボタンは C.bgDanger
import { STYLE } from '@/lib/design-tokens';
className={STYLE.button.danger}

// ✅ 今日のハイライト
className={isSameDay(day, new Date()) ? `${C.bgAccent} text-white shadow-sm` : `hover:${C.bgAccentLight} ...`}
```

## 影響範囲

| ファイル | 該当箇所 | 問題の種類 |
|---------|---------|-----------|
| `frontend/src/features/accounting/routes/AccountingList.tsx` | L329, L147 | text-blue-* / bg-blue-* |
| `frontend/src/features/auth/routes/LoginForm.tsx` | L68 | text-red-* / bg-red-* |
| `frontend/src/features/reservations/components/MonthView.tsx` | L79, L105-106 | bg-blue-* / bg-red-* |
| `frontend/src/features/medical-records/components/ExaminationGroup.tsx` | L105, L113 | bg-red-* / text-blue-* |
| `frontend/src/features/hospitalization/components/CarePlanTab.tsx` | L62, L126 | bg-blue-* |
| `frontend/src/features/hospitalization/components/DailyRecordsTab.tsx` | L113 | bg-blue-* |
| `frontend/src/features/hospitalization/components/DischargeAlertDialog.tsx` | L52 | bg-red-* |
| `frontend/src/features/owners/routes/OwnersList.tsx` | L333 | bg-red-* / text-red-* |

## 修正方針

### 1. 青系 (accent) カラーの置換

```tsx
// Before
bg-blue-600 → C.bgAccent (bg-[#2383E2])
bg-blue-100 → C.bgAccentLight (bg-[#D3E5EF])
text-blue-600 / text-blue-700 → C.textAccentBlue (text-[#0B6E99])
hover:bg-blue-100 → C.bgAccentLight
```

### 2. 赤系 (danger) カラーの置換

```tsx
// Before
bg-red-500 / bg-red-600 → C.bgDanger (bg-[#C0392B])
text-red-600 → C.textRequired (text-[#E03E3E])
bg-red-100 → C.bgDanger10 (bg-[#C0392B]/10)
border-red-300 → C.borderDanger (確認の上)
```

### 3. バッジの置換

```tsx
// Before
bg-blue-100 text-blue-700 → STYLE.badge.blue
bg-red-100 text-red-600 → STYLE.badge.red
```

### 4. ボタンの置換

```tsx
// Before: bg-red-600 hover:bg-red-700
// After
className={STYLE.button.danger}
```

## 準拠すべきプロジェクト規約

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリング（Tailwind 4, Inline styles）で `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。
> **PROHIBITED**: Hexカラー（`#37352F` など）の直接指定は厳禁。

Tailwind プリセットカラー（`blue-*`, `red-*`）の直接使用も同規約違反に相当する。

### プロジェクト内参照実装
- `frontend/src/features/owners/routes/OwnerForm.tsx` — C.bgAccent / STYLE.button 正しく使用
- `frontend/src/lib/design-tokens.ts:549` — `STYLE.badge.blue`, `STYLE.badge.red` 定義済み
- `frontend/src/lib/design-tokens.ts:731` — `STYLE.button.danger` 定義済み

## 優先度
**Medium** — 機能的障害なし。デザインシステムの一貫性が損なわれ、将来のテーマ変更時に対応漏れが発生する。

## 関連チケット
- なし

## 関連ファイル
- `frontend/src/lib/design-tokens.ts` — デザイントークン定義
- `frontend/src/features/accounting/routes/AccountingList.tsx`
- `frontend/src/features/auth/routes/LoginForm.tsx`
- `frontend/src/features/reservations/components/MonthView.tsx`
- `frontend/src/features/medical-records/components/ExaminationGroup.tsx`
- `frontend/src/features/hospitalization/components/CarePlanTab.tsx`
- `frontend/src/features/hospitalization/components/DischargeAlertDialog.tsx`
- `frontend/src/features/owners/routes/OwnersList.tsx`
