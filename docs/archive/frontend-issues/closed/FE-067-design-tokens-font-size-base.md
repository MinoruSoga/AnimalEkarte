# FE-067: design-tokens.ts STYLE プリセットのフォントサイズ最低 text-base 化

**Status**: Open
**Priority**: High
**Affects**: 全一覧・設定ページ（STYLE トークン経由のすべての画面）
**Date Created**: 2026-03-18
**Related**: TASK-017, FE-068, FE-069, FE-070

## Summary

`frontend/src/lib/design-tokens.ts` の STYLE プリセット内で `text-xs` / `text-sm` を使用している箇所を `text-base` に変更する。STYLE 経由でフォントサイズを指定しているすべてのページに波及するため、最もインパクトが大きい変更。

## 現状のコード

```typescript
// frontend/src/lib/design-tokens.ts:623-634
tableHeaderCell:
  `text-xs font-medium ${C.text70} h-11`,
tableCell:
  `text-base ${C.text} py-2.5`,           // ← 既に text-base（変更不要）
tableCellMono:
  `font-mono text-base ${C.text} py-2.5`, // ← 既に text-base（変更不要）
tableCellMuted:
  `text-base ${C.text70} py-2.5`,         // ← 既に text-base（変更不要）
tableEmpty:
  `text-center py-12 ${C.text70} text-sm`,

// frontend/src/lib/design-tokens.ts:603-604
formHeaderTitle: `text-sm ${C.text} leading-tight`,
formHeaderDesc:  `text-xs ${C.text50} mt-0.5`,

// frontend/src/lib/design-tokens.ts:639-640
searchInput:
  `pl-8 h-11 w-full text-sm ${C.text} ...`,

// frontend/src/lib/design-tokens.ts:647-652
paginationBtn:
  `h-8 w-8 ${C.text60} ${C.hoverBgPageHalf} rounded-[4px]`,
paginationBtnActive:
  `h-8 w-8 ${C.bgPrimary} text-white ${C.hoverBgPrimaryDark} text-sm rounded-[4px]`,
paginationInfo:
  `text-xs ${C.text50}`,

// frontend/src/lib/design-tokens.ts:670-672
propertyLabel:
  `w-[140px] shrink-0 text-sm ${C.text65} select-none truncate`,
propertyInput:
  `w-full bg-transparent text-sm ${C.text} outline-none ...`,

// frontend/src/lib/design-tokens.ts:607-616
btnPrimary:
  `... text-sm shadow-none ...`,
btnAccent:
  `... text-sm rounded-[4px] ...`,
btnDanger:
  `... text-sm rounded-[4px] ...`,
btnOutline:
  `... text-sm rounded-[4px] ...`,

// frontend/src/lib/design-tokens.ts:700
selectCompact:
  `h-[30px] text-sm bg-transparent ...`,

// frontend/src/lib/design-tokens.ts:703-704
sectionLabel:
  `text-xs ${C.text55} uppercase tracking-wide select-none`,

// frontend/src/lib/design-tokens.ts:708
badge:
  "text-sm px-2 h-9 font-normal border",

// frontend/src/lib/design-tokens.ts:686-688
sidePeekCancelBtn:
  `px-4 py-[7px] text-sm ${C.text65} ...`,
sidePeekSaveBtn:
  `px-5 py-[7px] text-sm text-white ...`,

// frontend/src/lib/design-tokens.ts:729-735
formLabel:
  `text-sm ${C.text70}`,
formInput:
  `h-11 text-sm bg-white ${C.borderMedium} ${C.text}`,
formInputLight:
  `h-11 text-sm bg-white ${C.borderMediumLight} ...`,

// frontend/src/lib/design-tokens.ts:726
inlineAddBtn:
  `w-full flex items-center gap-2 px-4 py-2.5 text-sm ${C.text40} ...`,

// frontend/src/lib/design-tokens.ts:716
confirmPrimary:
  `... text-sm rounded-[4px] ...`,
```

## 必要な変更

すべての `text-xs` → `text-base`、`text-sm` → `text-base` に置換する。

### 変更一覧（design-tokens.ts 内）

| プリセット | 変更前 | 変更後 |
|-----------|--------|--------|
| `tableHeaderCell` | `text-xs` | `text-base` |
| `tableEmpty` | `text-sm` | `text-base` |
| `formHeaderTitle` | `text-sm` | `text-base` |
| `formHeaderDesc` | `text-xs` | `text-base` |
| `searchInput` | `text-sm` | `text-base` |
| `paginationBtnActive` | `text-sm` | `text-base` |
| `paginationInfo` | `text-xs` | `text-base` |
| `propertyLabel` | `text-sm` | `text-base` |
| `propertyInput` | `text-sm` | `text-base` |
| `btnPrimary` | `text-sm` | `text-base` |
| `btnAccent` | `text-sm` | `text-base` |
| `btnDanger` | `text-sm` | `text-base` |
| `btnOutline` | `text-sm` | `text-base` |
| `selectCompact` | `text-sm` | `text-base` |
| `sectionLabel` | `text-xs` | `text-base` |
| `badge` | `text-sm` | `text-base` |
| `sidePeekCancelBtn` | `text-sm` | `text-base` |
| `sidePeekSaveBtn` | `text-sm` | `text-base` |
| `formLabel` | `text-sm` | `text-base` |
| `formInput` | `text-sm` | `text-base` |
| `formInputLight` | `text-sm` | `text-base` |
| `inlineAddBtn` | `text-sm` | `text-base` |
| `confirmPrimary` | `text-sm` | `text-base` |

## フロントエンド影響

- STYLE トークン経由でフォントサイズを指定している全ページに波及
- `OwnersList`（参照実装）は既に `STYLE.tableCell`（text-base）のため影響なし
- マスタ設定の SidePeek パネルのフォントサイズも変わる

## 完了条件

- [ ] design-tokens.ts 内に `text-xs` / `text-sm` が残っていない（テキストサイズに関するもの）
- [ ] `pnpm build` パス
- [ ] `pnpm lint` パス
- [ ] OwnersList（参照実装）の表示に変化がない（既に text-base のため）
