import { useState, useEffect, useCallback, useTransition } from "react";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { FormFieldError } from "@/components/shared/FormFieldError";
import type { Shift, ShiftType, CreateShiftInput, UpdateShiftInput } from "@/features/shifts/types";
import { SHIFT_TYPE_LABELS } from "@/features/shifts/types";
import { ShiftTypeOff } from "@/types/generated/models";
import { useCreateShift } from "@/features/shifts/api/create-shift";
import { useUpdateShift } from "@/features/shifts/api/update-shift";
import { useDeleteShift } from "@/features/shifts/api/delete-shift";

/**
 * バックエンドから "HH:MM:SS" 形式で来る時刻を "HH:mm" に正規化する。
 */
function normalizeTimeToHHmm(time: string): string {
  if (!time) return "";
  const parts = time.split(":");
  if (parts.length >= 2) {
    return `${parts[0].padStart(2, "0")}:${parts[1].padStart(2, "0")}`;
  }
  return time;
}

const SHIFT_TYPE_OPTIONS = (Object.entries(SHIFT_TYPE_LABELS) as [ShiftType, string][]).map(
  ([value, label]) => (
    <SelectItem key={value} value={value}>
      {label}
    </SelectItem>
  ),
);

interface ShiftFormDialogProps {
  open: boolean;
  onClose: () => void;
  staffId: string;
  staffName: string;
  date: string; // YYYY-MM-DD
  editShift?: Shift;
}

interface FormValues {
  shiftType: ShiftType;
  startTime: string;
  endTime: string;
  note: string;
}

export function ShiftFormDialog({
  open,
  onClose,
  staffId,
  staffName,
  date,
  editShift,
}: ShiftFormDialogProps) {
  const isEdit = editShift !== undefined;

  const [form, setForm] = useState<FormValues>(() => ({
    shiftType: editShift?.shift_type ?? "full",
    startTime: normalizeTimeToHHmm(editShift?.start_time ?? ""),
    endTime: normalizeTimeToHHmm(editShift?.end_time ?? ""),
    note: editShift?.note ?? "",
  }));
  const [timeError, setTimeError] = useState<string>("");

  useEffect(() => {
    if (open) {
      // eslint-disable-next-line react-hooks/set-state-in-effect -- ダイアログ open 時のフォームリセットパターン
      setForm({
        shiftType: editShift?.shift_type ?? "full",
        startTime: normalizeTimeToHHmm(editShift?.start_time ?? ""),
        endTime: normalizeTimeToHHmm(editShift?.end_time ?? ""),
        note: editShift?.note ?? "",
      });
      setTimeError("");
    }
  }, [open, editShift]);

  const { mutateAsync: createShift } = useCreateShift();
  const { mutateAsync: updateShift } = useUpdateShift();
  const { mutateAsync: deleteShift } = useDeleteShift();
  const [isPending, startSaveTransition] = useTransition();
  const [isDeletePending, startDeleteTransition] = useTransition();
  // BUG-093: 削除確認ダイアログの表示状態
  const [isDeleteConfirmOpen, setIsDeleteConfirmOpen] = useState(false);
  // BUG-092: 休日・有休は時刻入力不要
  const isTimeFieldDisabled = form.shiftType === ShiftTypeOff;

  const handleSubmit = useCallback(
    (e: React.FormEvent<HTMLFormElement>) => {
      e.preventDefault();
      if (form.startTime && form.endTime && form.endTime <= form.startTime) {
        setTimeError("終了時刻は開始時刻より後に設定してください");
        return;
      }
      setTimeError("");
      startSaveTransition(async () => {
        try {
          if (isEdit && editShift) {
            const input: UpdateShiftInput = {
              shift_type: form.shiftType,
              start_time: form.startTime || undefined,
              end_time: form.endTime || undefined,
              note: form.note || undefined,
            };
            await updateShift({ id: editShift.id, input });
          } else {
            const input: CreateShiftInput = {
              staff_id: staffId,
              date,
              shift_type: form.shiftType,
              start_time: form.startTime || undefined,
              end_time: form.endTime || undefined,
              note: form.note || undefined,
            };
            await createShift(input);
          }
          onClose();
        } catch {
          // エラーはReact QueryがToast等で処理
        }
      });
    },
    [form, isEdit, editShift, staffId, date, updateShift, createShift, onClose],
  );

  const handleShiftTypeChange = useCallback((value: string) => {
    setForm((prev) => ({ ...prev, shiftType: value as ShiftType }));
  }, []);

  const handleInputChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const { name, value } = e.target;
      setForm((prev) => ({ ...prev, [name]: value }));
      if (name === "startTime" || name === "endTime") {
        setTimeError("");
      }
    },
    [],
  );

  // BUG-093: 削除確認後に実際の削除を実行
  const handleDeleteConfirm = useCallback(() => {
    if (!editShift) return;
    setIsDeleteConfirmOpen(false);
    startDeleteTransition(async () => {
      await deleteShift(editShift.id);
      onClose();
    });
  }, [editShift, deleteShift, onClose]);

  const formattedDate = date
    ? new Date(date + "T00:00:00").toLocaleDateString("ja-JP", {
        month: "long",
        day: "numeric",
        weekday: "short",
      })
    : "";

  return (
    <>
    <Dialog open={open} onOpenChange={(o) => { if (!o) onClose(); }}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>
            {isEdit ? "シフト編集" : "シフト追加"}
          </DialogTitle>
          <p className="text-sm text-muted-foreground">
            {staffName} — {formattedDate}
          </p>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="shift-type">シフト種別</Label>
            <Select value={form.shiftType} onValueChange={handleShiftTypeChange}>
              <SelectTrigger id="shift-type">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>{SHIFT_TYPE_OPTIONS}</SelectContent>
            </Select>
          </div>

          {/* BUG-092: 休日選択時は時刻フィールドを非活性 */}
          <div className="space-y-1.5">
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label htmlFor="start-time" className={isTimeFieldDisabled ? "opacity-40" : ""}>開始時刻</Label>
                <Input
                  id="start-time"
                  name="startTime"
                  type="time"
                  value={form.startTime}
                  onChange={handleInputChange}
                  disabled={isTimeFieldDisabled}
                />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="end-time" className={isTimeFieldDisabled ? "opacity-40" : ""}>終了時刻</Label>
                <Input
                  id="end-time"
                  name="endTime"
                  type="time"
                  value={form.endTime}
                  onChange={handleInputChange}
                  disabled={isTimeFieldDisabled}
                />
              </div>
            </div>
            {timeError ? (
              <FormFieldError message={timeError} />
            ) : null}
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="note">メモ</Label>
            <Input
              id="note"
              name="note"
              placeholder="メモ（任意）"
              value={form.note}
              onChange={handleInputChange}
            />
          </div>

          <DialogFooter className="gap-2">
            {isEdit ? (
              <Button
                type="button"
                variant="destructive"
                size="sm"
                onClick={() => setIsDeleteConfirmOpen(true)}
                disabled={isPending || isDeletePending}
              >
                {isDeletePending ? "削除中..." : "削除"}
              </Button>
            ) : null}
            <Button type="button" variant="outline" onClick={onClose} disabled={isPending || isDeletePending}>
              キャンセル
            </Button>
            <Button type="submit" size="sm" disabled={isPending || isDeletePending}>
              {isPending ? "保存中..." : "保存"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>

    {/* BUG-093: 削除確認ダイアログ */}
    <ConfirmDialog
      open={isDeleteConfirmOpen}
      onClose={() => setIsDeleteConfirmOpen(false)}
      title="このシフトを削除しますか？"
      description="この操作は取り消せません。"
      confirmLabel="削除"
      variant="destructive"
      onConfirm={handleDeleteConfirm}
    />
    </>
  );
}
