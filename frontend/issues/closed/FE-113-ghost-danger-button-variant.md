# FE-113: ghost-danger variant 追加・outline 赤ボタン整理

**Status**: Closed
**Priority**: Medium
**Affects**: HospitalizationDetailActions, EstimateDetail, TrimmingForm, MedicalRecordForm
**Date Created**: 2026-03-25
**Related**: TASK-026, FE-112

## Summary

テキスト+アイコンの赤色アウトラインボタン（削除・退院処理など）が各ファイルで `className` 直書きで実装されている。
`Button` に `ghost-danger` variant を追加し、4 ファイルの重複 className を排除する。

## 現状のコード

```typescript
// frontend/src/features/hospitalization/components/HospitalizationDetailActions.tsx:29
<Button
  variant="outline"
  className={`gap-2 text-red-600 hover:text-red-700 hover:bg-red-50 border-red-200 ${H_STYLES.button.action}`}
  onClick={onDischargeClick}
>
  <LogOut className={H_STYLES.button.icon} />
  退院処理
</Button>

// frontend/src/features/estimates/routes/EstimateDetail.tsx:65-70
<Button
  variant="outline"
  size="sm"
  onClick={() => setShowDeleteDialog(true)}
  disabled={isDeleting}
  className="h-9 gap-1.5 text-sm text-red-600 hover:text-red-700 hover:bg-red-50 border-red-200"
>
  <Trash2 className="size-4" />
  削除
</Button>

// frontend/src/features/trimming/routes/TrimmingForm.tsx:516-524
<Button
  onClick={() => setDeleteConfirmOpen(true)}
  variant="ghost"
  className="h-10 text-red-600 hover:text-red-700 hover:bg-red-50 rounded-[6px] text-sm px-4"
  disabled={isDeleting}
>
  <Trash2 className="mr-1.5 size-4" />
  削除
</Button>

// frontend/src/features/medical-records/routes/MedicalRecordForm.tsx:332-339
<Button
  variant="outline"
  onClick={() => setIsDeleteConfirmOpen(true)}
  className={`${C.borderDanger} ${C.danger} ${C.hoverBgDanger5} h-10 text-sm px-4`}
>
  <Trash2 className="h-4 w-4" />
  削除
</Button>
```

## 必要な変更

### 1. button.tsx に ghost-danger variant 追加

```typescript
// frontend/src/components/ui/button.tsx
// buttonVariants の variants.variant に追記:
"ghost-danger": "text-red-600 hover:bg-red-50 hover:text-red-700",
```

### 2. HospitalizationDetailActions.tsx の修正

```typescript
// frontend/src/features/hospitalization/components/HospitalizationDetailActions.tsx
// Before:
<Button
  variant="outline"
  className={`gap-2 text-red-600 hover:text-red-700 hover:bg-red-50 border-red-200 ${H_STYLES.button.action}`}
  onClick={onDischargeClick}
>

// After:
<Button
  variant="ghost-danger"
  className={`gap-2 border border-red-200 ${H_STYLES.button.action}`}
  onClick={onDischargeClick}
>
```

### 3. EstimateDetail.tsx の修正

```typescript
// frontend/src/features/estimates/routes/EstimateDetail.tsx
// Before:
<Button
  variant="outline"
  size="sm"
  onClick={() => setShowDeleteDialog(true)}
  disabled={isDeleting}
  className="h-9 gap-1.5 text-sm text-red-600 hover:text-red-700 hover:bg-red-50 border-red-200"
>

// After:
<Button
  variant="ghost-danger"
  size="sm"
  onClick={() => setShowDeleteDialog(true)}
  disabled={isDeleting}
  className="h-9 gap-1.5 text-sm border border-red-200"
>
```

### 4. TrimmingForm.tsx の修正

```typescript
// frontend/src/features/trimming/routes/TrimmingForm.tsx
// Before:
<Button
  onClick={() => setDeleteConfirmOpen(true)}
  variant="ghost"
  className="h-10 text-red-600 hover:text-red-700 hover:bg-red-50 rounded-[6px] text-sm px-4"
  disabled={isDeleting}
>

// After:
<Button
  variant="ghost-danger"
  onClick={() => setDeleteConfirmOpen(true)}
  disabled={isDeleting}
  className="h-10 rounded-[6px] text-sm px-4"
>
```

### 5. MedicalRecordForm.tsx の修正

```typescript
// frontend/src/features/medical-records/routes/MedicalRecordForm.tsx
// Before:
<Button
  variant="outline"
  onClick={() => setIsDeleteConfirmOpen(true)}
  className={`${C.borderDanger} ${C.danger} ${C.hoverBgDanger5} h-10 text-sm px-4`}
>

// After:
<Button
  variant="ghost-danger"
  onClick={() => setIsDeleteConfirmOpen(true)}
  className={`border ${C.borderDanger} h-10 text-sm px-4`}
>
```

## UI 操作フロー

変更なし（見た目は統一されるが機能は同一）。

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] shadcn `components/ui/button.tsx` への変更は variant 追加のみ（必要最小限）

## 依存関係

なし（独立して着手可能）。FE-112 と並行実装可。

## 完了条件

- [ ] `button.tsx` に `ghost-danger` variant が存在する
- [ ] 上記 4 ファイルで `variant="ghost-danger"` を使用している
- [ ] `text-red-600 hover:text-red-700 hover:bg-red-50` の直書き className がこれら 4 ファイルから消えている
- [ ] `npm run lint` エラーなし
- [ ] `npm run build` エラーなし
