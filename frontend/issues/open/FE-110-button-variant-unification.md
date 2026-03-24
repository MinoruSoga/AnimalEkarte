# FE-110: ボタンカラー直書きを Button variant 拡充で統一

**Status**: Open
**Priority**: High
**Affects**: components/ui/button, components/shared/FormDialog, features/hospitalization
**Date Created**: 2026-03-25
**Related**: TASK-025

## Summary

`bg-[#2EAADC]`（水色プライマリ）の直書きが `FormDialog` の保存ボタンと `hospitalization` の3箇所に存在する。shadcn `Button` に `primary` variant を追加し、直書きを排除する。

## 現状のコード

```typescript
// frontend/src/components/shared/FormDialog/FormDialog.tsx:69
<Button
  onClick={onSave}
  disabled={isPending}
  className="bg-[#2EAADC] text-white h-11 px-4 text-sm font-medium"
>

// frontend/src/features/hospitalization/components/CarePlan/CarePlanSection.tsx:58
<Button className={`gap-2 bg-[#2EAADC] hover:bg-[#2EAADC]/90 text-white ${H_STYLES.button.action}`}>

// frontend/src/features/hospitalization/components/DailyRecord/TimingSection.tsx:71
<Button className={`bg-[#2EAADC] hover:bg-[#2EAADC]/90 text-white shadow-sm flex-shrink-0 ${H_STYLES.button.action}`}>

// frontend/src/features/hospitalization/components/DailyRecord/SimpleNoteForm.tsx:54
<Button className={`bg-[#2EAADC] hover:bg-[#2EAADC]/90 text-white gap-2 shadow-sm ${H_STYLES.button.action}`}>
```

## 必要な変更

### 1. `frontend/src/components/ui/button.tsx` — variant 追加

```typescript
// buttonVariants の variants.variant に追加
variant: {
  default: "bg-primary text-primary-foreground hover:bg-primary/90",
  destructive: "bg-destructive text-white hover:bg-destructive/90",
  outline: "border border-input bg-transparent text-foreground hover:bg-accent/50",
  secondary: "bg-secondary text-secondary-foreground hover:bg-secondary/80",
  ghost: "hover:bg-accent hover:text-accent-foreground",
  link: "text-primary underline-offset-4 hover:underline",
  // ★ 追加
  primary: "bg-[#2EAADC] text-white hover:bg-[#2EAADC]/90",
},
```

**注意**: `components/ui/` は通常「shadcn/ui 変更禁止」だが、プロジェクト固有 variant の追加は許容される。カラー値 `#2EAADC` はデザイントークン `C.accent` 系と同じ色。

### 2. `frontend/src/components/shared/FormDialog/FormDialog.tsx`

```typescript
// Before
<Button
  onClick={onSave}
  disabled={isPending}
  className="bg-[#2EAADC] text-white h-11 px-4 text-sm font-medium"
>

// After
<Button
  variant="primary"
  onClick={onSave}
  disabled={isPending}
>
```

### 3. `features/hospitalization/components/CarePlan/CarePlanSection.tsx`

```typescript
// Before
<Button onClick={handleOpenCreate} className={`gap-2 bg-[#2EAADC] hover:bg-[#2EAADC]/90 text-white ${H_STYLES.button.action}`}>

// After
<Button variant="primary" onClick={handleOpenCreate} className={`gap-2 ${H_STYLES.button.action}`}>
```

### 4. `features/hospitalization/components/DailyRecord/TimingSection.tsx`

```typescript
// Before
className={`bg-[#2EAADC] hover:bg-[#2EAADC]/90 text-white shadow-sm flex-shrink-0 ${H_STYLES.button.action}`}

// After
variant="primary"
className={`shadow-sm flex-shrink-0 ${H_STYLES.button.action}`}
```

### 5. `features/hospitalization/components/DailyRecord/SimpleNoteForm.tsx`

```typescript
// Before
className={`bg-[#2EAADC] hover:bg-[#2EAADC]/90 text-white gap-2 shadow-sm ${H_STYLES.button.action}`}

// After
variant="primary"
className={`gap-2 shadow-sm ${H_STYLES.button.action}`}
```

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）

## 依存関係

- BE 変更なし
- FE-109 より先に実装すること（FormDialog の保存ボタン変更が FE-109 でも参照されるため）

## 完了条件

- [ ] `grep -rn "bg-\[#2EAADC\]" frontend/src` の出力が 0 件
- [ ] `Button` の `variant="primary"` が shadcn buttonVariants に定義されている
- [ ] FormDialog の保存ボタン・hospitalization の3ボタンが水色で表示される（外観変化なし）
- [ ] `npm run build` パス
- [ ] `npm run lint` パス（エラー 0）
