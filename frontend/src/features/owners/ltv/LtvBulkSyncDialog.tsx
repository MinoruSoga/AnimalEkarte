import { useState, useCallback } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { C, STYLE } from "@/lib/design-tokens";
import { getForbiddenPrefix } from "@/constants/lstep-auto-tag-prefixes";
import { useCreateLstepBulkTag } from "../api/create-lstep-bulk-tag";

interface LtvBulkSyncDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  selectedOwnerIds: string[];
  selectedCount: number;
}

export function LtvBulkSyncDialog({
  open,
  onOpenChange,
  selectedOwnerIds,
  selectedCount,
}: LtvBulkSyncDialogProps) {
  const [tagName, setTagName] = useState("");
  const [confirmOpen, setConfirmOpen] = useState(false);
  const { mutate, isPending } = useCreateLstepBulkTag();

  const forbiddenPrefix = getForbiddenPrefix(tagName);
  const isDisabled = tagName.trim() === "" || forbiddenPrefix !== null || isPending;

  const handleTagNameChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      setTagName(e.target.value);
    },
    []
  );

  const handleSyncClick = useCallback(() => {
    setConfirmOpen(true);
  }, []);

  const handleConfirm = useCallback(() => {
    mutate(
      { owner_ids: selectedOwnerIds, tag_name: tagName.trim() },
      {
        onSuccess: () => {
          setConfirmOpen(false);
          setTagName("");
          onOpenChange(false);
        },
      }
    );
  }, [mutate, selectedOwnerIds, tagName, onOpenChange]);

  const handleClose = useCallback(() => {
    if (!isPending) {
      setTagName("");
      onOpenChange(false);
    }
  }, [isPending, onOpenChange]);

  return (
    <>
      <Dialog open={open} onOpenChange={handleClose}>
        <DialogContent className="sm:max-w-[480px]">
          <DialogHeader>
            <DialogTitle>Lステップに連携</DialogTitle>
          </DialogHeader>

          <div className="flex flex-col gap-4 py-2">
            <p className={`text-sm ${C.text70}`}>
              選択中の{selectedCount}名にタグを付与してLステップと連携します。
            </p>

            <div className="flex flex-col gap-1.5">
              <label className={`text-sm ${C.text65}`} htmlFor="ltv-tag-name">
                タグ名
              </label>
              <Input
                id="ltv-tag-name"
                className={`${STYLE.formInput} ${forbiddenPrefix !== null ? STYLE.formInputError : ""}`}
                placeholder="タグ名を入力..."
                value={tagName}
                onChange={handleTagNameChange}
                disabled={isPending}
              />
              {forbiddenPrefix !== null ? (
                <p className={`text-sm ${C.danger}`}>
                  「{forbiddenPrefix}」で始まるタグ名は自動管理タグに予約されており使用できません。
                </p>
              ) : null}
            </div>
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              className={STYLE.btnOutline}
              onClick={handleClose}
              disabled={isPending}
            >
              キャンセル
            </Button>
            <Button
              className={STYLE.btnPrimary}
              onClick={handleSyncClick}
              disabled={isDisabled}
            >
              Lステップに連携
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <ConfirmDialog
        open={confirmOpen}
        onClose={() => setConfirmOpen(false)}
        onConfirm={handleConfirm}
        title={`${selectedCount}件にタグを付与します`}
        description={`タグ「${tagName}」を選択中の${selectedCount}名全員に付与します。よろしいですか？`}
        confirmLabel={isPending ? "処理中..." : "付与する"}
        cancelLabel="キャンセル"
        isPending={isPending}
      />
    </>
  );
}
