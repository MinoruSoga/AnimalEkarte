import { memo, useActionState, useEffect, useState } from "react";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
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

type ClinicHolidayFormState = { error?: string } | null;

export const ClinicHolidayModal = memo(function ClinicHolidayModal({
  open,
  onClose,
  date,
  existing,
  canEdit,
}: ClinicHolidayModalProps) {
  const [reason, setReason] = useState(() => existing?.reason ?? "");

  const setMutation = useCreateClinicHoliday();
  const deleteMutation = useDeleteClinicHoliday();

  // 日付変更・open時にフォームリセット
  useEffect(() => {
    if (open) {
      setReason(existing?.reason ?? "");
    }
  }, [open, date, existing?.reason]);

  const [state, formAction] = useActionState<ClinicHolidayFormState, FormData>(
    async (_prev, formData) => {
      const intent = formData.get("intent");
      if (intent === "remove") {
        try {
          await deleteMutation.mutateAsync(date);
          onClose();
          return null;
        } catch (err) {
          handleApiError(err, "定休日の解除");
          return { error: "定休日の解除に失敗しました" };
        }
      }
      try {
        await setMutation.mutateAsync({ date, reason });
        onClose();
        return null;
      } catch (err) {
        handleApiError(err, "定休日の設定");
        return { error: "定休日の設定に失敗しました" };
      }
    },
    null,
  );

  const formattedDate = date
    ? new Date(`${date}T00:00:00+09:00`).toLocaleDateString("ja-JP", {
        timeZone: "Asia/Tokyo",
        year: "numeric",
        month: "long",
        day: "numeric",
        weekday: "short",
      })
    : "";

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

        <form action={formAction} noValidate>
          <div className="space-y-4 py-2">
            {existing ? (
              <p className={`text-sm font-medium ${C.danger}`}>この日は定休日に設定されています</p>
            ) : null}

            <div className="space-y-1.5">
              <Label htmlFor="holiday-reason">理由・メモ（任意）</Label>
              <Input
                id="holiday-reason"
                name="reason"
                value={reason}
                onChange={(e) => setReason(e.target.value)}
                placeholder="例: 院長不在、設備点検"
                disabled={!canEdit}
              />
            </div>

            {state?.error ? (
              <p className={`text-sm ${C.danger}`} role="alert">
                {state.error}
              </p>
            ) : null}
          </div>

          <DialogFooter className="gap-2">
            {existing && canEdit ? (
              <SubmitButton
                name="intent"
                value="remove"
                colorVariant="destructive"
                loadingText="解除中..."
              >
                定休日を解除
              </SubmitButton>
            ) : null}
            <Button type="button" variant="outline" onClick={onClose}>
              キャンセル
            </Button>
            {canEdit ? (
              <SubmitButton name="intent" value="save" loadingText="設定中...">
                定休日に設定
              </SubmitButton>
            ) : null}
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
});
