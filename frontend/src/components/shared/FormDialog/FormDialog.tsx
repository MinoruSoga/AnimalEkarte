import { memo, type ReactNode } from "react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";

export interface FormDialogProps {
  open: boolean;
  onClose: () => void;
  title: string;
  /** sr-only アクセシビリティ用説明文（省略可） */
  description?: string;
  children: ReactNode;
  onSave: () => void;
  /** 保存ボタンのラベル（デフォルト: "保存"） */
  saveLabel?: string;
  /** キャンセルボタンのラベル（デフォルト: "キャンセル"） */
  cancelLabel?: string;
  /** 保存中フラグ（ボタンを disabled にする） */
  isPending?: boolean;
  className?: string;
  /** 保存ボタンの追加 className（danger 色等の変更に使用） */
  saveClassName?: string;
  /** 保存ボタンの disabled 条件（isPending に加えて） */
  isSaveDisabled?: boolean;
}

/**
 * hospitalization feature の各 Dialog（LogDialog / VitalDialog / TaskCompleteDialog / CarePlanDialog）
 * に共通するヘッダー・フッターパターンを統一する共有ラッパー。
 *
 * フォームフィールド（children）は各 Dialog が持つ。
 */
export const FormDialog = memo(function FormDialog({
  open,
  onClose,
  title,
  description,
  children,
  onSave,
  saveLabel = "保存",
  cancelLabel = "キャンセル",
  isPending = false,
  className,
  saveClassName,
  isSaveDisabled,
}: FormDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className={className}>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          {description ? <DialogDescription>{description}</DialogDescription> : null}
        </DialogHeader>
        {children}
        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            onClick={onClose}
            disabled={isPending}
            className="h-11 px-4 text-sm font-medium"
          >
            {cancelLabel}
          </Button>
          <Button
            type="button"
            variant="primary"
            onClick={onSave}
            disabled={isPending || isSaveDisabled}
            className={saveClassName}
          >
            {isPending ? "保存中..." : saveLabel}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
});
