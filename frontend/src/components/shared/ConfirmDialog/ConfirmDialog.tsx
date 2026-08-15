import { memo, useRef } from "react";
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle } from "@/components/ui/alert-dialog";
import { C } from "@/lib/design-tokens";

interface ConfirmDialogProps {
  open: boolean;
  onClose: () => void;
  onConfirm: () => void;
  title: string;
  description?: string;
  confirmLabel?: string;
  cancelLabel?: string;
  variant?: "default" | "destructive";
  isPending?: boolean;
  /** BUG-084: ref to the element that triggered the dialog; focus is restored here on close */
  triggerRef?: React.RefObject<HTMLElement | null>;
}

export const ConfirmDialog = memo(function ConfirmDialog({
  open,
  onClose,
  onConfirm,
  title,
  description,
  confirmLabel = "確認",
  cancelLabel = "キャンセル",
  variant = "default",
  isPending = false,
  triggerRef,
}: ConfirmDialogProps) {
  // BUG-084: Track the element that had focus before the dialog opened so we can
  // restore it when the dialog closes (Radix cannot do this automatically when
  // the dialog is opened programmatically without an AlertDialogTrigger).
  const previousFocusRef = useRef<Element | null>(null);

  const handleOpenChange = (nextOpen: boolean) => {
    if (nextOpen) {
      previousFocusRef.current = document.activeElement;
    } else {
      onClose();
    }
  };

  const handleCloseAutoFocus = (e: Event) => {
    // Prefer the explicitly provided triggerRef, then the previously focused element.
    const target = triggerRef?.current ?? previousFocusRef.current;
    if (target instanceof HTMLElement) {
      e.preventDefault();
      target.focus();
    }
  };

  return (
    <AlertDialog open={open} onOpenChange={handleOpenChange}>
      <AlertDialogContent onCloseAutoFocus={handleCloseAutoFocus}>
        <AlertDialogHeader>
          <AlertDialogTitle>{title}</AlertDialogTitle>
          {description ? (
            <AlertDialogDescription>{description}</AlertDialogDescription>
          ) : null}
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>{cancelLabel}</AlertDialogCancel>
          <AlertDialogAction
            onClick={onConfirm}
            disabled={isPending}
            className={
              variant === "destructive"
                ? `${C.bgDanger} ${C.textWhite} ${C.hoverBgDanger90}`
                : `${C.bgBrand} ${C.textOnBrand} ${C.hoverBgBrand} ${C.hoverTextOnBrand}`
            }
          >
            {confirmLabel}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
});
