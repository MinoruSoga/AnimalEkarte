import { useState, useEffect, useCallback, useTransition, useRef } from "react";
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
import { FormFieldError } from "@/components/shared/FormFieldError";
import type { Shift, ShiftType, CreateShiftInput, UpdateShiftInput } from "../../types";
import { SHIFT_TYPE_LABELS } from "../../types";
import { useCreateShift } from "../../api/create-shift";
import { useUpdateShift } from "../../api/update-shift";
import { useDeleteShift } from "../../api/delete-shift";

/**
 * バックエンドから "HH:MM:SS" 形式で来る時刻を "HH:mm" に正規化する。
 * BUG-029: input[type="time"] は "HH:mm" 形式のみ受け付けるため必要。
 */
function normalizeTimeToHHmm(time: string): string {
  if (!time) return "";
  // "HH:MM:SS" → "HH:MM"
  const parts = time.split(":");
  if (parts.length >= 2) {
    return `${parts[0].padStart(2, "0")}:${parts[1].padStart(2, "0")}`;
  }
  return time;
}

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
    startTime: normalizeTimeToHHmm(editShift?.start_time ?? ""),
    endTime: normalizeTimeToHHmm(editShift?.end_time ?? ""),
    note: editShift?.note ?? "",
  }));
  const [timeError, setTimeError] = useState<string>("");

  useEffect(() => {
    if (open) {
      // eslint-disable-next-line react-hooks/set-state-in-effect -- ダイアログ open 時にフォームをリセット。key prop パターンの代替
      setForm({
        shiftType: editShift?.shift_type ?? "full",
        startTime: normalizeTimeToHHmm(editShift?.start_time ?? ""),
        endTime: normalizeTimeToHHmm(editShift?.end_time ?? ""),
        note: editShift?.note ?? "",
      });
      setTimeError("");
    }
  }, [open, editShift]);

  const createShift = useCreateShift();
  const updateShift = useUpdateShift();
  const deleteShift = useDeleteShift();
  const [isSavePending, startSaveTransition] = useTransition();

  // rerender-dependencies: editShift/form はオブジェクト。useRef で保持し deps から除外
  const editShiftRef = useRef(editShift);
  useEffect(() => { editShiftRef.current = editShift; }, [editShift]);
  const formRef = useRef(form);
  useEffect(() => { formRef.current = form; }, [form]);

  const handleShiftTypeChange = useCallback((value: string) => {
    setForm((prev) => ({ ...prev, shiftType: value as ShiftType }));
  }, []);

  const handleInputChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const { name, value } = e.target;
      setForm((prev) => ({ ...prev, [name]: value }));
      // 時刻フィールド変更時にエラーをクリア
      if (name === "startTime" || name === "endTime") {
        setTimeError("");
      }
    },
    [],
  );

  const handleSubmit = useCallback(
    (e: React.FormEvent) => {
      e.preventDefault();
      const currentForm = formRef.current;
      const currentEditShift = editShiftRef.current;

      // BUG-028: end_time > start_time バリデーション
      if (currentForm.startTime && currentForm.endTime) {
        if (currentForm.endTime <= currentForm.startTime) {
          setTimeError("終了時刻は開始時刻より後に設定してください");
          return;
        }
      }
      setTimeError("");

      startSaveTransition(async () => {
        if (isEdit && currentEditShift) {
          const input: UpdateShiftInput = {
            shift_type: currentForm.shiftType,
            start_time: currentForm.startTime || undefined,
            end_time: currentForm.endTime || undefined,
            note: currentForm.note || undefined,
          };
          await updateShift.mutateAsync({ id: currentEditShift.id, input });
        } else {
          const input: CreateShiftInput = {
            staff_id: staffId,
            date,
            shift_type: currentForm.shiftType,
            start_time: currentForm.startTime || undefined,
            end_time: currentForm.endTime || undefined,
            note: currentForm.note || undefined,
          };
          await createShift.mutateAsync(input);
        }
        onClose();
      });
    },
    [isEdit, staffId, date, updateShift, createShift, onClose, startSaveTransition],
  );

  const handleDelete = useCallback(() => {
    const currentEditShift = editShiftRef.current;
    if (!currentEditShift) return;
    startSaveTransition(async () => {
      await deleteShift.mutateAsync(currentEditShift.id);
      onClose();
    });
  }, [deleteShift, onClose, startSaveTransition]);

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

          <div className="space-y-1.5">
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
                onClick={handleDelete}
                disabled={isSavePending}
              >
                削除
              </Button>
            ) : null}
            <Button type="button" variant="outline" onClick={onClose} disabled={isSavePending}>
              キャンセル
            </Button>
            <Button type="submit" disabled={isSavePending}>
              {isSavePending ? "保存中..." : "保存"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
