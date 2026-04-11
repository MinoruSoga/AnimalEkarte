import { memo, useState, useEffect, useRef, useCallback, useActionState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { Plus, X } from "lucide-react";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import type { Shift, ShiftType, ShiftBreakInput, CreateShiftInput, UpdateShiftInput } from "@/features/shifts/types";
import { SHIFT_TYPE_LABELS } from "@/features/shifts/types";
import { ShiftTypeOff } from "@/types/generated/models";
import { C } from "@/lib/design-tokens";
import { createShift } from "@/features/shifts/api/create-shift";
import { updateShift } from "@/features/shifts/api/update-shift";
import { useDeleteShift } from "@/features/shifts/api/delete-shift";
import { handleApiError } from "@/lib/handle-api-error";

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
  canEdit?: boolean;
  canDelete?: boolean;
}

interface FormActionState {
  timeError?: string;
}

export const ShiftFormDialog = memo(function ShiftFormDialog({
  open,
  onClose,
  staffId,
  staffName,
  date,
  editShift,
  canEdit = false,
  canDelete = false,
}: ShiftFormDialogProps) {
  const isEdit = editShift !== undefined;
  // rerender-dependencies: editShift オブジェクトの代わりに primitive id を deps に使用
  const editShiftId = editShift?.id;
  const queryClient = useQueryClient();

  // Controlled state for Select and time inputs (needed for UI feedback and FormData relay)
  const [shiftType, setShiftType] = useState<ShiftType>(() => editShift?.shift_type ?? "full");
  const [startTime, setStartTime] = useState(() => normalizeTimeToHHmm(editShift?.start_time ?? ""));
  const [endTime, setEndTime] = useState(() => normalizeTimeToHHmm(editShift?.end_time ?? ""));
  const [breaks, setBreaks] = useState<ShiftBreakInput[]>(() =>
    (editShift?.breaks ?? []).map((b) => ({ break_start: normalizeTimeToHHmm(b.break_start), break_end: normalizeTimeToHHmm(b.break_end) })),
  );
  // rerender-dependencies: breaks 配列を ref 経由で参照し formAction deps から除外
  const breaksRef = useRef(breaks);
  useEffect(() => { breaksRef.current = breaks; }, [breaks]);

  useEffect(() => {
    if (open) {
      // eslint-disable-next-line react-hooks/set-state-in-effect -- ダイアログ open 時のフォームリセットパターン
      setShiftType(editShift?.shift_type ?? "full");
      setStartTime(normalizeTimeToHHmm(editShift?.start_time ?? ""));
      setEndTime(normalizeTimeToHHmm(editShift?.end_time ?? ""));
      setBreaks((editShift?.breaks ?? []).map((b) => ({ break_start: normalizeTimeToHHmm(b.break_start), break_end: normalizeTimeToHHmm(b.break_end) })));
    }
  }, [open, editShift]);

  const formAction = useCallback(
    async (_prevState: FormActionState, formData: FormData): Promise<FormActionState> => {
      const resolvedShiftType = (formData.get("shiftType") as ShiftType) ?? shiftType;
      const resolvedStartTime = (formData.get("startTime") as string) ?? "";
      const resolvedEndTime = (formData.get("endTime") as string) ?? "";

      if (resolvedStartTime && resolvedEndTime && resolvedEndTime <= resolvedStartTime) {
        return { timeError: "終了時刻は開始時刻より後に設定してください" };
      }

      try {
        if (isEdit && editShiftId) {
          const validBreaks = breaksRef.current.filter((b) => b.break_start && b.break_end);
          const input: UpdateShiftInput = {
            shift_type: resolvedShiftType,
            start_time: resolvedStartTime || undefined,
            end_time: resolvedEndTime || undefined,
            note: (formData.get("note") as string) || undefined,
            breaks: validBreaks,
          };
          await updateShift(editShiftId, input);
        } else {
          const validBreaks = breaksRef.current.filter((b) => b.break_start && b.break_end);
          const input: CreateShiftInput = {
            staff_id: staffId,
            date,
            shift_type: resolvedShiftType,
            start_time: resolvedStartTime || undefined,
            end_time: resolvedEndTime || undefined,
            note: (formData.get("note") as string) || undefined,
            breaks: validBreaks.length > 0 ? validBreaks : undefined,
          };
          await createShift(input);
        }
        await queryClient.invalidateQueries({ queryKey: ["shifts"] });
        onClose();
        return {};
      } catch (err) {
        handleApiError(err, isEdit ? "シフト更新" : "シフト追加");
        return {};
      }
    },
    // rerender-dependencies: editShift → id（primitive）、breaks → ref 経由
    [isEdit, editShiftId, staffId, date, shiftType, queryClient, onClose],
  );

  const [state, dispatchFormAction, isPending] = useActionState<FormActionState, FormData>(formAction, {});

  const { mutate: deleteShift, isPending: isDeletePending } = useDeleteShift();
  // BUG-093: 削除確認ダイアログの表示状態
  const [isDeleteConfirmOpen, setIsDeleteConfirmOpen] = useState(false);
  // BUG-092: 休日・有休は時刻入力不要
  const isTimeFieldDisabled = shiftType === ShiftTypeOff;

  const handleShiftTypeChange = useCallback((value: string) => {
    setShiftType(value as ShiftType);
  }, []);

  const handleStartTimeChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setStartTime(e.target.value);
  }, []);

  const handleEndTimeChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setEndTime(e.target.value);
  }, []);

  // BUG-093: 削除確認後に実際の削除を実行
  const handleDeleteConfirm = useCallback(() => {
    if (!editShift) return;
    setIsDeleteConfirmOpen(false);
    deleteShift(editShift.id, { onSuccess: () => onClose() });
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
          <p className={`text-sm ${C.text50}`}>
            {staffName} — {formattedDate}
          </p>
        </DialogHeader>

        <form action={dispatchFormAction} className="space-y-4">
          {/* hidden input を経由してコントロール値を FormData に渡す */}
          <input type="hidden" name="shiftType" value={shiftType} />

          <div className="space-y-1.5">
            <Label htmlFor="shift-type">シフト種別</Label>
            <Select value={shiftType} onValueChange={handleShiftTypeChange}>
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
                  value={startTime}
                  onChange={handleStartTimeChange}
                  disabled={isTimeFieldDisabled}
                />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="end-time" className={isTimeFieldDisabled ? "opacity-40" : ""}>終了時刻</Label>
                <Input
                  id="end-time"
                  name="endTime"
                  type="time"
                  value={endTime}
                  onChange={handleEndTimeChange}
                  disabled={isTimeFieldDisabled}
                />
              </div>
            </div>
            {state.timeError ? (
              <FormFieldError message={state.timeError} />
            ) : null}
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="note">メモ</Label>
            <Input
              id="note"
              name="note"
              placeholder="メモ（任意）"
              defaultValue={editShift?.note ?? ""}
            />
          </div>

          {/* 休憩時間 */}
          {!isTimeFieldDisabled ? (
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <Label className="text-sm">休憩時間</Label>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  className="h-7 px-2 text-xs"
                  onClick={() => setBreaks((prev) => [...prev, { break_start: "12:00", break_end: "13:00" }])}
                >
                  <Plus className={`${ICON.xxs} mr-1`} />
                  追加
                </Button>
              </div>
              {breaks.map((b, i) => (
                <div key={i} className="flex items-center gap-2">
                  <Input
                    type="time"
                    value={b.break_start}
                    onChange={(e) => setBreaks((prev) => prev.map((br, j) => j === i ? { ...br, break_start: e.target.value } : br))}
                    className="flex-1"
                  />
                  <span className={`text-xs ${C.text50}`}>〜</span>
                  <Input
                    type="time"
                    value={b.break_end}
                    onChange={(e) => setBreaks((prev) => prev.map((br, j) => j === i ? { ...br, break_end: e.target.value } : br))}
                    className="flex-1"
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    className="h-8 w-8 p-0"
                    onClick={() => setBreaks((prev) => prev.filter((_, j) => j !== i))}
                  >
                    <X className={ICON.smXs} />
                  </Button>
                </div>
              ))}
            </div>
          ) : null}

          <DialogFooter className="gap-2">
            {isEdit && canDelete ? (
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
            {canEdit ? (
              <SubmitButton size="sm" disabled={isDeletePending}>
                保存
              </SubmitButton>
            ) : null}
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
});
