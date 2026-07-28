import { memo, useCallback, useEffect, useState, useTransition } from "react";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { C } from "@/lib/design-tokens";
import { handleApiError } from "@/lib/handle-api-error";
import { useCreateClinicHoliday, useDeleteClinicHoliday } from "../../api/clinic-holidays";
import type { ClinicHoliday } from "../../api/clinic-holidays";

interface ClinicHolidayModalProps {
  open: boolean;
  onClose: () => void;
  date: string; // YYYY-MM-DD
  existing: ClinicHoliday | undefined;
  canEdit: boolean;
}

export const ClinicHolidayModal = memo(function ClinicHolidayModal({
  open,
  onClose,
  date,
  existing,
  canEdit,
}: ClinicHolidayModalProps) {
  const [reason, setReason] = useState(() => existing?.reason ?? "");
  const [isSaving, startSaveTransition] = useTransition();
  const [isRemoving, startRemoveTransition] = useTransition();

  const setMutation = useCreateClinicHoliday();
  const deleteMutation = useDeleteClinicHoliday();

  // 日付変更・open時にフォームリセット
  useEffect(() => {
    if (open) {
      setReason(existing?.reason ?? "");
    }
  }, [open, date, existing?.reason]);

  const handleSave = useCallback(() => {
    startSaveTransition(async () => {
      try {
        await setMutation.mutateAsync({ date, reason });
        onClose();
      } catch (err) {
        handleApiError(err, "定休日の設定");
      }
    });
  }, [date, reason, setMutation, onClose]);

  const handleRemove = useCallback(() => {
    startRemoveTransition(async () => {
      try {
        await deleteMutation.mutateAsync(date);
        onClose();
      } catch (err) {
        handleApiError(err, "定休日の解除");
      }
    });
  }, [date, deleteMutation, onClose]);

  const formattedDate = date
    ? new Date(`${date}T00:00:00+09:00`).toLocaleDateString("ja-JP", {
        timeZone: "Asia/Tokyo",
        year: "numeric",
        month: "long",
        day: "numeric",
        weekday: "short",
      })
    : "";

  const isPending = isSaving || isRemoving;

  return (
    <Dialog open={open} onOpenChange={(o) => { if (!o) onClose(); }}>
      <DialogContent className="sm:max-w-sm">
        <DialogHeader>
          <DialogTitle>定休日の設定</DialogTitle>
          <DialogDescription>
            {formattedDate} の定休日設定と理由を編集します。
          </DialogDescription>
          <p className={`text-sm ${C.text50}`}>{formattedDate}</p>
        </DialogHeader>

        <div className="space-y-4 py-2">
          {existing ? (
            <p className={`text-sm font-medium ${C.danger}`}>この日は定休日に設定されています</p>
          ) : null}

          <div className="space-y-1.5">
            <Label htmlFor="holiday-reason">理由・メモ（任意）</Label>
            <Input
              id="holiday-reason"
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              placeholder="例: 院長不在、設備点検"
              disabled={!canEdit || isPending}
            />
          </div>
        </div>

        <DialogFooter className="gap-2">
          {existing && canEdit ? (
            <Button
              type="button"
              variant="destructive"
              size="sm"
              onClick={handleRemove}
              disabled={isPending}
            >
              {isRemoving ? "解除中..." : "定休日を解除"}
            </Button>
          ) : null}
          <Button type="button" variant="outline" onClick={onClose} disabled={isPending}>
            キャンセル
          </Button>
          {canEdit ? (
            <Button
              type="button"
              size="sm"
              onClick={handleSave}
              disabled={isPending}
            >
              {isSaving ? "設定中..." : "定休日に設定"}
            </Button>
          ) : null}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
});
