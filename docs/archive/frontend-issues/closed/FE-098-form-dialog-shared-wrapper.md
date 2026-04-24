# FE-098: FormDialog 共有ラッパー作成（hospitalization Dialog 標準化）

**Status**: Closed
**Priority**: Medium
**Affects**: hospitalization/components/DailyRecord/, hospitalization/components/CarePlan/
**Date Created**: 2026-03-24
**Related**: TASK-024

## Summary

hospitalization feature 内の LogDialog / VitalDialog / TaskCompleteDialog / CarePlanDialog は、それぞれ独立実装されており、ダイアログのヘッダー・フッター（キャンセル/保存ボタン）パターンが重複している。共有 `FormDialog` ラッパーを作成し、これら 4 つの Dialog を統一する。

## 現状のコード

```typescript
// frontend/src/features/hospitalization/components/DailyRecord/LogDialog.tsx:90-115
// frontend/src/features/hospitalization/components/DailyRecord/VitalDialog.tsx:85-110
// frontend/src/features/hospitalization/components/DailyRecord/TaskCompleteDialog.tsx:60-100
// frontend/src/features/hospitalization/components/CarePlan/CarePlanDialog.tsx:200-225

// 各 Dialog が共通して持つパターン（4ファイルで重複）:
// 1. Dialog + DialogContent + DialogHeader + DialogTitle
// 2. フッター: キャンセルボタン + 保存ボタン（bg-[#2EAADC] text-white）
// 3. isPending 状態でボタン disabled
// 4. onClose / onSave コールバック

// 例: LogDialog のフッター部分
<DialogFooter>
  <Button
    variant="outline"
    onClick={onClose}
    disabled={isSavePending}
    className={H_STYLES.button.action}
  >
    キャンセル
  </Button>
  <Button
    onClick={handleSave}
    disabled={isSavePending}
    className={`bg-[#2EAADC] text-white ${H_STYLES.button.action}`}
  >
    {isSavePending ? "保存中..." : "記録"}
  </Button>
</DialogFooter>
// VitalDialog, TaskCompleteDialog, CarePlanDialog でほぼ同一パターンが繰り返される
```

## 必要な変更

### 1. 共有 FormDialog コンポーネント作成

```typescript
// frontend/src/components/shared/FormDialog/FormDialog.tsx（新規作成）

interface FormDialogProps {
  open: boolean;
  onClose: () => void;
  title: string;
  children: React.ReactNode;
  onSave: () => void;
  saveLabel?: string;          // デフォルト: "保存"
  cancelLabel?: string;        // デフォルト: "キャンセル"
  isPending?: boolean;
  className?: string;
}

export function FormDialog({
  open,
  onClose,
  title,
  children,
  onSave,
  saveLabel = "保存",
  cancelLabel = "キャンセル",
  isPending = false,
  className,
}: FormDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className={className}>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
        </DialogHeader>
        {children}
        <DialogFooter>
          <Button
            variant="outline"
            onClick={onClose}
            disabled={isPending}
            className={H_STYLES.button.action}
          >
            {cancelLabel}
          </Button>
          <Button
            onClick={onSave}
            disabled={isPending}
            className={`bg-[#2EAADC] text-white ${H_STYLES.button.action}`}
          >
            {isPending ? "保存中..." : saveLabel}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
```

```typescript
// frontend/src/components/shared/FormDialog/index.ts（新規作成）
export { FormDialog } from "./FormDialog";
export type { FormDialogProps } from "./FormDialog";
```

### 2. LogDialog を FormDialog に置き換え

```typescript
// frontend/src/features/hospitalization/components/DailyRecord/LogDialog.tsx
// Before: Dialog + DialogContent + DialogHeader + DialogFooter を手書き
// After: FormDialog ラッパーを使用

import { FormDialog } from "@/components/shared/FormDialog/FormDialog";

export function LogDialog({ open, onClose, onSave, record, date }: LogDialogProps) {
  const [formData, setFormData] = useState(...);
  const [isSavePending, startSaveTransition] = useTransition();

  const handleSave = () => {
    startSaveTransition(async () => { await onSave(formData); });
  };

  return (
    <FormDialog
      open={open}
      onClose={onClose}
      title={`${date} 日誌記録`}
      onSave={handleSave}
      saveLabel="記録"
      isPending={isSavePending}
    >
      {/* フォームフィールド（変更なし） */}
      ...
    </FormDialog>
  );
}
```

### 3. VitalDialog / TaskCompleteDialog / CarePlanDialog も同様に置き換え

各ファイルで `Dialog + DialogContent + DialogHeader + DialogFooter` を `FormDialog` ラッパーに置き換える。フォームフィールド（children 部分）は変更しない。

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし（`FormDialog/FormDialog` を直接 import）
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `useTransition` で pending 管理（各 Dialog 内は既存のまま変更なし）
- [ ] 型は明示的 Props 型（手書き interface OK）

## 依存関係

- Backend 変更なし。単独で実施可能。

## 完了条件

- [ ] `frontend/src/components/shared/FormDialog/FormDialog.tsx` が作成されている
- [ ] LogDialog / VitalDialog / TaskCompleteDialog / CarePlanDialog が `FormDialog` ラッパーを使用している
- [ ] 各 Dialog のフォームフィールド（コンテンツ部分）は変更されていない
- [ ] hospitalization の入院管理画面で全 Dialog が正常に開閉・保存できる
- [ ] `pnpm lint` エラーなし
- [ ] `pnpm build` エラーなし

## クローズ情報

- **Closed At**: 2026-03-24
- **変更ファイル**:
  - `frontend/src/components/shared/FormDialog/FormDialog.tsx` — 新規作成（title / description / children / onSave / saveLabel / isPending props）
  - `frontend/src/features/hospitalization/components/DailyRecord/LogDialog.tsx` — FormDialog ラッパーに置き換え
  - `frontend/src/features/hospitalization/components/DailyRecord/VitalDialog.tsx` — 同上
  - `frontend/src/features/hospitalization/components/DailyRecord/TaskCompleteDialog.tsx` — 同上
  - `frontend/src/features/hospitalization/components/CarePlan/CarePlanDialog.tsx` — FormDialog + TreatmentSearchDialog を Fragment に分離
