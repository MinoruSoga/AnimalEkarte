# FE-068: 共有コンポーネントのフォントサイズ最低 text-base 化

**Status**: Open
**Priority**: High
**Affects**: StatusBadge, NotionStatusPill, Pagination, NotionFilter（全一覧ページで使用）
**Date Created**: 2026-03-18
**Related**: TASK-017, FE-067

## Summary

共有コンポーネント内でハードコードされた `text-xs` / `text-sm` を `text-base` に置換する。これらのコンポーネントは全一覧ページで使用されるため影響範囲が大きい。

## 現状のコード

### StatusBadge

```typescript
// frontend/src/components/shared/StatusBadge/StatusBadge.tsx:13
className={`text-sm px-2 h-7 font-normal border ${colorClass} ${className}`}
```

### NotionStatusPill

```typescript
// frontend/src/components/shared/StatusPill/NotionStatusPill.tsx:24
className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded-sm text-xs ${cfg.bg} ${cfg.text}`}
```

### Pagination

```typescript
// frontend/src/components/shared/Pagination/Pagination.tsx:65
<div className="text-sm text-[#37352F]/60">

// :97
className="px-1 text-sm text-[#37352F]/40"

// :108-109
? "h-8 w-8 bg-[#37352F] text-white hover:bg-[#37352F]/90 text-sm"
: "h-8 w-8 text-[#37352F]/60 hover:bg-[#F7F6F3]/50 text-sm"
```

### NotionFilter/FilterRuleRow.tsx（主要箇所）

```typescript
// frontend/src/components/shared/NotionFilter/FilterRuleRow.tsx
// 行 125, 203, 232, 241, 256, 268, 364, 375, 385, 388, 393, 404, 436
// 全箇所 text-sm → text-base
```

### NotionFilter/SortPopover.tsx

```typescript
// frontend/src/components/shared/NotionFilter/SortPopover.tsx
// 行 46, 66, 83, 182, 199, 216, 231
// 全箇所 text-sm → text-base
```

### NotionFilter/FilterAddPopover.tsx

```typescript
// frontend/src/components/shared/NotionFilter/FilterAddPopover.tsx
// 行 312, 330, 343, 354, 370, 403, 412, 427, 439
// 全箇所 text-sm → text-base
```

### NotionFilter/SortPill.tsx

```typescript
// frontend/src/components/shared/NotionFilter/SortPill.tsx
// 行 62(text-sm), 77(text-sm), 91(text-xs→text-base), 102(text-sm), 120(text-sm)
```

## 必要な変更

### 1. StatusBadge.tsx

`text-sm` → `text-base`

### 2. NotionStatusPill.tsx

`text-xs` → `text-base`。padding (`px-2 py-0.5`) の調整が必要になる可能性あり。

### 3. Pagination.tsx

全 `text-sm` → `text-base`

### 4. NotionFilter 全ファイル

`FilterRuleRow.tsx`, `SortPopover.tsx`, `FilterAddPopover.tsx`, `SortPill.tsx` 内の全 `text-xs` / `text-sm` → `text-base`

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] 型は `models.ts` から導出

## 依存関係

- FE-067 が先に完了していることが望ましい（STYLE トークン経由の変更が先に反映されるため）

## 完了条件

- [ ] 上記共有コンポーネント内に `text-xs` / `text-sm`（テキストサイズ）が残っていない
- [ ] `pnpm build` パス
- [ ] `pnpm lint` パス
- [ ] NotionStatusPill のバッジが大きくなりすぎていない（padding 調整済み）
- [ ] Pagination のレイアウトが崩れていない
