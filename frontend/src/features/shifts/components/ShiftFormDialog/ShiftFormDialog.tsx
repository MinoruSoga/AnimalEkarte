import { useState, useEffect, useCallback } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { Shift, ShiftType, CreateShiftInput, UpdateShiftInput } from "../../types";
import { SHIFT_TYPE_LABELS } from "../../types";
import { useCreateShift } from "../../api/create-shift";
import { useUpdateShift } from "../../api/update-shift";
import { useDeleteShift } from "../../api/delete-shift";

// 静的 JSX リスト: rendering-hoist-jsx
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

interface FormState {
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

  const [form, setForm] = useState<FormState>(() => ({
    shiftType: editShift?.shift_type ?? "full",
    startTime: editShift?.start_time ?? "",
    endTime: editShift?.end_time ?? "",
    note: editShift?.note ?? "",
  }));

  useEffect(() => {
    if (open) {
      // eslint-disable-next-line react-hooks/set-state-in-effect -- ダイアログ open 時にフォームをリセット。key prop パターンの代替
      setForm({
        shiftType: editShift?.shift_type ?? "full",
        startTime: editShift?.start_time ?? "",
        endTime: editShift?.end_time ?? "",
        note: editShift?.note ?? "",
      });
    }
  }, [open, editShift]);

  const createShift = useCreateShift();
  const updateShift = useUpdateShift();
  const deleteShift = useDeleteShift();

  const isPending =
    createShift.isPending || updateShift.isPending || deleteShift.isPending;

  const handleShiftTypeChange = useCallback((value: string) => {
    setForm((prev) => ({ ...prev, shiftType: value as ShiftType }));
  }, []);

  const handleInputChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const { name, value } = e.target;
      setForm((prev) => ({ ...prev, [name]: value }));
    },
    [],
  );

  const handleSubmit = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault();
      if (isEdit && editShift) {
        const input: UpdateShiftInput = {
          shift_type: form.shiftType,
          start_time: form.startTime || undefined,
          end_time: form.endTime || undefined,
          note: form.note || undefined,
        };
        await updateShift.mutateAsync({ id: editShift.id, input });
      } else {
        const input: CreateShiftInput = {
          staff_id: staffId,
          date,
          shift_type: form.shiftType,
          start_time: form.startTime || undefined,
          end_time: form.endTime || undefined,
          note: form.note || undefined,
        };
        await createShift.mutateAsync(input);
      }
      onClose();
    },
    [isEdit, editShift, form, staffId, date, updateShift, createShift, onClose],
  );

  const handleDelete = useCallback(async () => {
    if (!editShift) return;
    await deleteShift.mutateAsync(editShift.id);
    onClose();
  }, [editShift, deleteShift, onClose]);

  const formattedDate = date
    ? new Date(date + "T00:00:00").toLocaleDateString("ja-JP", {
        month: "long",
        day: "numeric",
        weekday: "short",
      })
    : "";

  return (
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

          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1.5">
              <Label htmlFor="start-time">開始時刻</Label>
              <Input
                id="start-time"
                name="startTime"
                type="time"
                value={form.startTime}
                onChange={handleInputChange}
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="end-time">終了時刻</Label>
              <Input
                id="end-time"
                name="endTime"
                type="time"
                value={form.endTime}
                onChange={handleInputChange}
              />
            </div>
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
                onClick={handleDelete}
                disabled={isPending}
              >
                削除
              </Button>
            ) : null}
            <Button type="button" variant="outline" onClick={onClose} disabled={isPending}>
              キャンセル
            </Button>
            <Button type="submit" disabled={isPending}>
              {isPending ? "保存中..." : "保存"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
