import { memo, useState, useEffect, useRef, useCallback, useActionState, useMemo } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { Plus, X } from "lucide-react";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import type { Shift, ShiftType, ShiftBreakInput, CreateShiftInput, UpdateShiftInput } from "../../types";
import { SHIFT_TYPE_LABELS, isShiftType } from "../../types";
import { C, ICON } from "@/lib/design-tokens";
import { createShift } from "../../api/create-shift";
import { updateShift } from "../../api/update-shift";
import { useDeleteShift } from "../../api/delete-shift";
import { useGetShiftTemplates } from "../../api/get-shift-templates";
import { handleApiError } from "@/lib/handle-api-error";
import { getFormString } from "@/lib/form-data";
import { queryKeys } from "@/lib/query-keys";
import { isShiftTemplateTimeHidden } from "../shift-template-form-utils";
import { DEFAULT_BREAK_START, DEFAULT_BREAK_END } from "../shift-template-form-model";

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

  // テンプレート選択
  const { data: templates = [] } = useGetShiftTemplates();

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

  // FE-RC-032: open 切り替え時の state リセットは useEffect ではなく、
  // 呼び出し側（ShiftCalendar.tsx）が dialog セッションごとに異なる key を
  // 付与することでコンポーネントを再マウントさせ、上記の useState 初期化子に
  // 委ねる（ShiftTemplateSidePanel と同様の設計に統一）。

  const formAction = useCallback(
    async (_prevState: FormActionState, formData: FormData): Promise<FormActionState> => {
      const rawShiftType = formData.get("shiftType");
      const resolvedShiftType = isShiftType(rawShiftType) ? rawShiftType : shiftType;
      const resolvedStartTime = getFormString(formData, "startTime");
      const resolvedEndTime = getFormString(formData, "endTime");
      // BUG-036: off/paid_leave 以外は開始・終了時刻必須（空のまま API に送らない）
      const timesRequired = !isShiftTemplateTimeHidden(resolvedShiftType);

      if (timesRequired && (!resolvedStartTime || !resolvedEndTime)) {
        return { timeError: "開始時刻と終了時刻を入力してください" };
      }

      if (resolvedStartTime && resolvedEndTime && resolvedEndTime <= resolvedStartTime) {
        return { timeError: "終了時刻は開始時刻より後に設定してください" };
      }

      try {
        if (isEdit && editShiftId) {
          const validBreaks = breaksRef.current.filter((b) => b.break_start && b.break_end);
          const input: UpdateShiftInput = {
            shift_type: resolvedShiftType,
            start_time: timesRequired ? resolvedStartTime : undefined,
            end_time: timesRequired ? resolvedEndTime : undefined,
            notes: getFormString(formData, "notes") || undefined,
            breaks: validBreaks,
          };
          await updateShift(editShiftId, input);
        } else {
          const validBreaks = breaksRef.current.filter((b) => b.break_start && b.break_end);
          const input: CreateShiftInput = {
            staff_id: staffId,
            date,
            shift_type: resolvedShiftType,
            start_time: timesRequired ? resolvedStartTime : undefined,
            end_time: timesRequired ? resolvedEndTime : undefined,
            notes: getFormString(formData, "notes") || undefined,
            breaks: validBreaks.length > 0 ? validBreaks : undefined,
          };
          await createShift(input);
        }
        await queryClient.invalidateQueries({ queryKey: queryKeys.shifts.all() });
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
  // BUG-092 / BUG-036: 休日・有休は時刻入力不要（テンプレート側と同判定）
  const isTimeFieldDisabled = isShiftTemplateTimeHidden(shiftType);

  const handleApplyTemplate = useCallback(
    (templateId: string) => {
      const tpl = templates.find((t) => t.id === templateId);
      if (!tpl) return;
      setShiftType(tpl.shift_type);
      setStartTime(tpl.start_time ?? "");
      setEndTime(tpl.end_time ?? "");
      setBreaks(tpl.breaks.map((b) => ({ break_start: b.break_start, break_end: b.break_end })));
    },
    [templates],
  );

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

  // js-cache-function-results: テンプレート選択肢を useMemo でキャッシュ
  const templateSelectItems = useMemo(
    () =>
      templates
        .filter((t) => t.is_active)
        .map((t) => (
          <SelectItem key={t.id} value={t.id}>
            {t.name}
          </SelectItem>
        )),
    [templates],
  );

  const formattedDate = date
    ? new Date(`${date}T00:00:00+09:00`).toLocaleDateString("ja-JP", {
        timeZone: "Asia/Tokyo",
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
          <DialogDescription>
            {staffName} の {formattedDate} のシフト内容を入力します。
          </DialogDescription>
          <p className={`text-sm ${C.text50}`}>
            {staffName} — {formattedDate}
          </p>
        </DialogHeader>

        <form action={dispatchFormAction} className="space-y-4">
          {/* hidden input を経由してコントロール値を FormData に渡す */}
          <input type="hidden" name="shiftType" value={shiftType} />

          {/* テンプレート選択（テンプレートがある場合のみ表示） */}
          {templateSelectItems.length > 0 ? (
            <div className="space-y-1.5">
              <Label htmlFor="shift-template" className={`text-xs ${C.text50}`}>テンプレートから入力</Label>
              <Select onValueChange={handleApplyTemplate}>
                <SelectTrigger id="shift-template" className="h-8 text-sm">
                  <SelectValue placeholder="テンプレートを選択..." />
                </SelectTrigger>
                <SelectContent>{templateSelectItems}</SelectContent>
              </Select>
            </div>
          ) : null}

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
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
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
            <Label htmlFor="notes">メモ</Label>
            <Input
              id="notes"
              name="notes"
              placeholder="メモ（任意）"
              defaultValue={editShift?.notes ?? ""}
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
                  onClick={() => setBreaks((prev) => [...prev, { break_start: DEFAULT_BREAK_START, break_end: DEFAULT_BREAK_END }])}
                >
                  <Plus className={`${ICON.xxs} mr-1`} />
                  追加
                </Button>
              </div>
              {breaks.map((b, i) => (
                <div key={i} className="flex items-center gap-2">
                  <Input
                    type="time"
                    aria-label={`休憩${i + 1} 開始時刻`}
                    value={b.break_start}
                    onChange={(e) => setBreaks((prev) => prev.map((br, j) => j === i ? { ...br, break_start: e.target.value } : br))}
                    className="flex-1"
                  />
                  <span className={`text-xs ${C.text50}`}>〜</span>
                  <Input
                    type="time"
                    aria-label={`休憩${i + 1} 終了時刻`}
                    value={b.break_end}
                    onChange={(e) => setBreaks((prev) => prev.map((br, j) => j === i ? { ...br, break_end: e.target.value } : br))}
                    className="flex-1"
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    className="h-8 w-8 p-0"
                    aria-label={`休憩${i + 1}を削除`}
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
