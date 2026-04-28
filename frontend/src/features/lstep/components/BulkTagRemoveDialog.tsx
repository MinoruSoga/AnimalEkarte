import { useCallback } from "react";
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import { C, STYLE } from "@/lib/design-tokens";
import { useBulkRemoveTag } from "../api/delete-owner-tag-bulk";

interface BulkTagRemoveDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  tagName: string;
  ownerCount: number;
  ownerIds: string[];
}

export function BulkTagRemoveDialog({
  open,
  onOpenChange,
  tagName,
  ownerCount,
  ownerIds,
}: BulkTagRemoveDialogProps) {
  const { bulkRemove, progress } = useBulkRemoveTag();

  const handleConfirm = useCallback(async () => {
    await bulkRemove(tagName, ownerIds);
    onOpenChange(false);
  }, [bulkRemove, tagName, ownerIds, onOpenChange]);

  const handleCancel = useCallback(() => {
    if (!progress.isRunning) {
      onOpenChange(false);
    }
  }, [progress.isRunning, onOpenChange]);

  const progressPercent =
    progress.total > 0 ? Math.round((progress.done / progress.total) * 100) : 0;

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>タグを一括解除します</AlertDialogTitle>
          <AlertDialogDescription>
            タグ「{tagName}」を{ownerCount}名全員から解除します。この操作は取り消せません。
          </AlertDialogDescription>
        </AlertDialogHeader>

        {progress.isRunning ? (
          <div className="flex flex-col gap-2 py-2">
            <p className={`text-sm ${C.text70}`}>
              {progress.total}名中{progress.done}名完了...
            </p>
            <div className={`w-full h-2 bg-[#E3E2E0] rounded-full overflow-hidden`}>
              <div
                className={`h-full ${C.bgBrand} transition-all duration-200 rounded-full`}
                style={{ width: `${progressPercent}%` }}
              />
            </div>
            <p className={`text-xs ${C.text50} text-right`}>{progressPercent}%</p>
          </div>
        ) : null}

        <AlertDialogFooter>
          <AlertDialogCancel
            onClick={handleCancel}
            disabled={progress.isRunning}
          >
            キャンセル
          </AlertDialogCancel>
          <Button
            className={STYLE.btnDanger}
            onClick={handleConfirm}
            disabled={progress.isRunning}
          >
            {progress.isRunning ? "解除中..." : "一括解除"}
          </Button>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
