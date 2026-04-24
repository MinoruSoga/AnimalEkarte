# FE-217: AccountingList のデザイントークン違反

## 概要

`frontend/src/features/accounting/routes/AccountingList.tsx` で
直接 Tailwind カラークラスを使用している。プロジェクト規約に違反。

## 問題コード

### `frontend/src/features/accounting/routes/AccountingList.tsx`

#### 返金バッジ（Line 317付近）
```tsx
// Before: 直接 Tailwind カラー
<span className="inline-flex items-center gap-0.5 text-[10px] font-medium px-1.5 py-0.5 rounded bg-orange-50 text-orange-600 border border-orange-200">
  返金
</span>

// After: デザイントークン使用
<span className={cn(
  "inline-flex items-center gap-0.5 text-[10px] font-medium px-1.5 py-0.5 rounded border",
  C.bgWarningLight, C.textWarning, C.borderWarningLight
)}>
  返金
</span>
```

#### カルテリンクボタン（Line 329付近）
```tsx
// Before: 直接 Tailwind カラー
className="h-8 w-8 text-blue-500 hover:text-blue-700 hover:bg-blue-50"

// After: デザイントークン使用
className={`h-8 w-8 ${C.textAccent} ${C.hoverTextAccentDark} ${C.hoverBgAccentLight}`}
```

## 影響範囲

| 行 | 違反内容 | 状態 |
|----|---------|------|
| 317付近 | `bg-orange-50 text-orange-600 border-orange-200`（返金バッジ） | 要修正 |
| 329付近 | `text-blue-500 hover:text-blue-700 hover:bg-blue-50`（カルテリンクボタン） | 要修正 |

## 準拠すべきプロジェクト規約

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリングで `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。
> **PROHIBITED**: 直接 Tailwind カラークラスの指定は厳禁。

## 優先度
**Low** — 機能的障害はないが、デザイン一貫性のため修正が必要。

## 関連ファイル
- `frontend/src/features/accounting/routes/AccountingList.tsx`
- `frontend/src/lib/design-tokens.ts`
