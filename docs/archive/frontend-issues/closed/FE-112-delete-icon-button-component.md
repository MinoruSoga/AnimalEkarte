# FE-112: DeleteIconButton 共有コンポーネント作成・10ファイル適用

**Status**: Closed
**Priority**: Medium
**Affects**: 全 feature のテーブル行削除ボタン
**Date Created**: 2026-03-25
**Related**: TASK-026, FE-113

## Summary

Trash2 アイコンのみの削除ボタンが各 feature で異なる className で実装されている。
`DeleteIconButton` 共有コンポーネントを作成し、10ファイルに適用してスタイルを統一する。

## 現状のコード

現在 10 ファイルでそれぞれ独自の className でアイコン削除ボタンが実装されている。

```typescript
// TreatmentTable.tsx:176-183
<Button
  variant="ghost"
  className="h-10 w-10 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded opacity-0 group-hover:opacity-100 transition-opacity"
  onClick={() => onDelete(row.id)}
  title="削除"
>
  <Trash2 className="size-4" />
</Button>

// CarePlanItemRow.tsx:83-86
<Button variant="ghost" size="sm" onClick={() => onDelete(plan.id)} className="h-9 w-9 p-0 bg-gray-50 hover:bg-red-50 hover:text-red-600">
  <Trash2 className="h-4 w-4 text-[#37352F]/60 hover:text-red-600" />
</Button>

// CarePlanTab.tsx:214-217
<Button variant="ghost" size="sm" className="h-7 w-7 text-red-400 hover:text-red-600 hover:bg-red-50" onClick={() => onDelete(item.id)}>
  <Trash2 className="h-3.5 w-3.5" />
</Button>

// AccountingDetail.tsx:259-263
<Button variant="ghost" size="icon" className="h-8 w-8 text-red-500 hover:text-red-700 hover:bg-red-50" onClick={() => handleRemoveItem(item.id)}>
  <Trash2 className="h-4 w-4" />
</Button>

// TreatmentRow.tsx:354-358
<Button variant="ghost" size="icon" className={`size-7 ${C.danger} ${C.hoverBgDanger5} hover:text-[#EB5757]`} onClick={handleDelete} title="削除">
  <Trash2 className="size-3.5" />
</Button>

// CheckupsTab.tsx:307-313 (native button)
<button
  onClick={() => handleDelete(checkup.id)}
  disabled={deleteMutation.isPending}
  className={`size-8 flex items-center justify-center rounded-[3px] ${C.text60} ${C.hoverTextDanger} ${C.hoverBgDanger5} transition-colors`}
  title="削除"
>
  <Trash2 className="size-3.5" />
</button>

// VitalsTab.tsx:430-436 (native button, same pattern as CheckupsTab)

// ClinicMasterSettings.tsx:354-358 (native button)
<button
  type="button"
  onClick={() => setPendingDelete(selectedItem)}
  className={`${STYLE.sidePeekToolbarBtn} cursor-pointer text-[#EB5757] hover:bg-[#EB5757]/10`}
>
  <Trash2 className="size-4" />
</button>

// HospitalizationTreatmentTable.tsx:99-102 (native button)
<button
  onClick={() => onRemove(plan.id)}
  className="text-[#37352F]/40 hover:text-[#E03E3E] transition-colors"
>
  <Trash2 className={H_STYLES.button.icon} />
</button>

// ReservationDetailModal.tsx:224-228
<Button
  variant="ghost"
  className="h-9 w-9 p-0 text-[#37352F]/30 hover:text-red-500 hover:bg-red-50 rounded-lg transition-colors"
  onClick={() => setShowDeleteConfirm(true)}
>
  <Trash2 className="size-4" />
</Button>
```

## 必要な変更

### 1. DeleteIconButton コンポーネント作成

```typescript
// frontend/src/components/shared/DeleteIconButton/DeleteIconButton.tsx
import { Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/components/ui/utils";

interface DeleteIconButtonProps {
  onClick: () => void;
  disabled?: boolean;
  className?: string;
  title?: string;
}

/**
 * テーブル行・カード等でのアイコン削除ボタン共通コンポーネント。
 * shadcn Button variant="ghost" + Trash2 アイコンを統一スタイルで提供する。
 */
export function DeleteIconButton({
  onClick,
  disabled,
  className,
  title = "削除",
}: DeleteIconButtonProps) {
  return (
    <Button
      variant="ghost"
      size="icon"
      onClick={onClick}
      disabled={disabled}
      title={title}
      className={cn(
        "size-8 text-[#37352F]/40 hover:text-red-600 hover:bg-red-50",
        className
      )}
    >
      <Trash2 className="size-4" />
    </Button>
  );
}
```

### 2. 各ファイルへの適用

**TreatmentTable.tsx** (`frontend/src/features/medical-records/components/TreatmentTable.tsx`):
```typescript
// Before
import { Circle, X, Trash2, PlusCircle } from "lucide-react";
<Button variant="ghost" className="h-10 w-10 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded opacity-0 group-hover:opacity-100 transition-opacity" onClick={() => onDelete(row.id)} title="削除">
  <Trash2 className="size-4" />
</Button>

// After
import { DeleteIconButton } from "@/components/shared/DeleteIconButton/DeleteIconButton";
<DeleteIconButton
  onClick={() => onDelete(row.id)}
  className="opacity-0 group-hover:opacity-100 transition-opacity"
/>
```

**CarePlanItemRow.tsx** (`frontend/src/features/hospitalization/components/CarePlan/CarePlanItemRow.tsx`):
```typescript
// Before
import { Edit2, Trash2, ... } from "lucide-react";
<Button variant="ghost" size="sm" onClick={() => onDelete(plan.id)} className="h-9 w-9 p-0 bg-gray-50 hover:bg-red-50 hover:text-red-600">
  <Trash2 className="h-4 w-4 text-[#37352F]/60 hover:text-red-600" />
</Button>

// After
import { DeleteIconButton } from "@/components/shared/DeleteIconButton/DeleteIconButton";
<DeleteIconButton onClick={() => onDelete(plan.id)} />
```

**CarePlanTab.tsx** (`frontend/src/features/hospitalization/components/CarePlanTab/CarePlanTab.tsx`):
```typescript
// Before
<Button variant="ghost" size="sm" className="h-7 w-7 text-red-400 hover:text-red-600 hover:bg-red-50" onClick={() => onDelete(item.id)}>
  <Trash2 className="h-3.5 w-3.5" />
</Button>

// After
<DeleteIconButton onClick={() => onDelete(item.id)} className="size-7" />
```

**AccountingDetail.tsx** (`frontend/src/features/accounting/routes/AccountingDetail.tsx`):
```typescript
// Before
<Button variant="ghost" size="icon" className="h-8 w-8 text-red-500 hover:text-red-700 hover:bg-red-50" onClick={() => handleRemoveItem(item.id)}>
  <Trash2 className="h-4 w-4" />
</Button>

// After
<DeleteIconButton onClick={() => handleRemoveItem(item.id)} />
```

**TreatmentRow.tsx** (`frontend/src/features/medical-records/components/TreatmentsTab/TreatmentRow.tsx`):
```typescript
// Before
<Button variant="ghost" size="icon" className={`size-7 ${C.danger} ${C.hoverBgDanger5} hover:text-[#EB5757]`} onClick={handleDelete} title="削除">
  <Trash2 className="size-3.5" />
</Button>

// After
<DeleteIconButton onClick={handleDelete} className="size-7" />
```

**CheckupsTab.tsx** (`frontend/src/features/medical-records/components/CheckupsTab/CheckupsTab.tsx`):
```typescript
// Before (native button → shadcn Button)
<button
  onClick={() => handleDelete(checkup.id)}
  disabled={deleteMutation.isPending}
  className={`size-8 flex items-center justify-center rounded-[3px] ${C.text60} ${C.hoverTextDanger} ${C.hoverBgDanger5} transition-colors`}
  title="削除"
>
  <Trash2 className="size-3.5" />
</button>

// After
<DeleteIconButton
  onClick={() => handleDelete(checkup.id)}
  disabled={deleteMutation.isPending}
/>
```

**VitalsTab.tsx** (`frontend/src/features/medical-records/components/VitalsTab/VitalsTab.tsx`):
CheckupsTab と同じパターンで置換。

**ClinicMasterSettings.tsx** (`frontend/src/features/hospital-settings/routes/ClinicMasterSettings.tsx`):
```typescript
// Before (native button → shadcn Button)
<button
  type="button"
  onClick={() => setPendingDelete(selectedItem)}
  className={`${STYLE.sidePeekToolbarBtn} cursor-pointer text-[#EB5757] hover:bg-[#EB5757]/10`}
>
  <Trash2 className="size-4" />
</button>

// After
<DeleteIconButton onClick={() => setPendingDelete(selectedItem)} />
```

**HospitalizationTreatmentTable.tsx** (`frontend/src/features/hospitalization/components/HospitalizationTreatmentTable.tsx`):
```typescript
// Before (native button → shadcn Button)
<button
  onClick={() => onRemove(plan.id)}
  className="text-[#37352F]/40 hover:text-[#E03E3E] transition-colors"
>
  <Trash2 className={H_STYLES.button.icon} />
</button>

// After
<DeleteIconButton onClick={() => onRemove(plan.id)} />
```

**ReservationDetailModal.tsx** (`frontend/src/features/reservations/components/ReservationDetailModal.tsx`):
```typescript
// Before
<Button variant="ghost" className="h-9 w-9 p-0 text-[#37352F]/30 hover:text-red-500 hover:bg-red-50 rounded-lg transition-colors" onClick={() => setShowDeleteConfirm(true)}>
  <Trash2 className="size-4" />
</Button>

// After
<DeleteIconButton onClick={() => setShowDeleteConfirm(true)} />
```

## UI 操作フロー

変更なし（見た目は統一されるが機能は同一）。

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし（直接ファイル import）
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `useTransition` で pending 管理（該当なし）
- [ ] 型は `models.ts` から導出（該当なし）

## 依存関係

なし（独立して着手可能）

## 完了条件

- [ ] `components/shared/DeleteIconButton/DeleteIconButton.tsx` が存在する
- [ ] 上記 10 ファイルで `DeleteIconButton` を使用している
- [ ] Trash2 import が残る場合はアイコン用途以外（DropdownMenu 等）のみ
- [ ] `pnpm lint` エラーなし
- [ ] `pnpm build` エラーなし
