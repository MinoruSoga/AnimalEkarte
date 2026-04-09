import { useState, useCallback, useActionState, useMemo } from "react";
import { format, startOfMonth, endOfMonth, eachDayOfInterval, getDay } from "date-fns";
import { ja } from "date-fns/locale";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import { C, ICON } from "@/lib/design-tokens";
import { useAuth } from "@/features/auth";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { useGetReservationStaffs } from "../api/get-staffs";
import { useGetStaffSchedules } from "../api/get-schedules";
import { useUpsertStaffSchedule } from "../api/upsert-schedule";
import { useDeleteStaffSchedule } from "../api/delete-schedule";
import type { StaffSchedule, StaffScheduleShiftType, UpsertScheduleRequest } from "../api/types";

// ── Constants ──

const SHIFT_COLORS: Record<StaffScheduleShiftType, { bg: string; text: string; label: string }> = {
  full: { bg: C.bgBrand10, text: C.textBrand, label: "全日" },
  morning: { bg: C.bgStatusGreen, text: C.textStatusGreen, label: "午前" },
  afternoon: { bg: "bg-yellow-100", text: "text-yellow-700", label: "午後" },
  off: { bg: C.bgDanger10, text: C.danger, label: "休み" },
  paid_leave: { bg: C.bgInactive, text: C.textMuted, label: "有休" },
};

const WEEKDAYS = ["日", "月", "火", "水", "木", "金", "土"];

// rendering-hoist-jsx: 静的 SelectItem JSX をモジュール定数に巻き上げ
const SHIFT_TYPE_SELECT_ITEMS = Object.entries(SHIFT_COLORS).map(([value, config]) => (
  <SelectItem key={value} value={value}>
    {config.label}
  </SelectItem>
));

// ── Schedule Edit Dialog ──

interface ScheduleEditDialogProps {
  open: boolean;
  onClose: () => void;
  date: string | null;
  schedule: StaffSchedule | null;
  staffId: number;
  clinicId: string | null;
}

function ScheduleEditDialog({
  open,
  onClose,
  date,
  schedule,
  staffId,
  clinicId,
}: ScheduleEditDialogProps) {
  const upsert = useUpsertStaffSchedule(clinicId);
  const deleteSchedule = useDeleteStaffSchedule(clinicId);

  const [shiftType, setShiftType] = useState<StaffScheduleShiftType>(
    schedule?.shift_type ?? "full"
  );

  const [formState, formAction] = useActionState(
    async (_prev: null, formData: FormData) => {
      if (!date) return null;
      const payload: UpsertScheduleRequest = {
        shift_type: shiftType,
        work_start: (formData.get("work_start") as string) || undefined,
        work_end: (formData.get("work_end") as string) || undefined,
        break_start: (formData.get("break_start") as string) || undefined,
        break_end: (formData.get("break_end") as string) || undefined,
      };
      try {
        await upsert.mutateAsync({ staffId, date, payload });
        toast.success("スケジュールを保存しました");
        onClose();
      } catch (err) {
        handleApiError(err, "スケジュール保存");
      }
      return null;
    },
    null
  );

  const handleDelete = useCallback(async () => {
    if (!date) return;
    try {
      await deleteSchedule.mutateAsync({ staffId, date });
      toast.success("スケジュールを削除しました");
      onClose();
    } catch (err) {
      handleApiError(err, "スケジュール削除");
    }
  }, [date, staffId, deleteSchedule, onClose]);

  return (
    <Dialog open={open} onOpenChange={(v) => { if (!v) onClose(); }}>
      <DialogContent className="max-w-[400px]">
        <DialogHeader>
          <DialogTitle>
            {date ? format(new Date(date), "M月d日（E）", { locale: ja }) : ""}のシフト
          </DialogTitle>
        </DialogHeader>
        <form action={formAction} className="space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="shift-type">シフト種別</Label>
            <Select
              value={shiftType}
              onValueChange={(v) => setShiftType(v as StaffScheduleShiftType)}
            >
              <SelectTrigger id="shift-type">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>{SHIFT_TYPE_SELECT_ITEMS}</SelectContent>
            </Select>
          </div>

          {shiftType !== "off" && shiftType !== "paid_leave" ? (
            <>
              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-1.5">
                  <Label htmlFor="work-start">勤務開始</Label>
                  <Input
                    id="work-start"
                    name="work_start"
                    type="time"
                    defaultValue={schedule?.work_start ?? "09:00"}
                  />
                </div>
                <div className="space-y-1.5">
                  <Label htmlFor="work-end">勤務終了</Label>
                  <Input
                    id="work-end"
                    name="work_end"
                    type="time"
                    defaultValue={schedule?.work_end ?? "18:00"}
                  />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-1.5">
                  <Label htmlFor="break-start">休憩開始</Label>
                  <Input
                    id="break-start"
                    name="break_start"
                    type="time"
                    defaultValue={schedule?.break_start ?? "12:00"}
                  />
                </div>
                <div className="space-y-1.5">
                  <Label htmlFor="break-end">休憩終了</Label>
                  <Input
                    id="break-end"
                    name="break_end"
                    type="time"
                    defaultValue={schedule?.break_end ?? "13:00"}
                  />
                </div>
              </div>
            </>
          ) : null}

          <DialogFooter className="gap-2">
            {schedule ? (
              <Button
                type="button"
                variant="outline"
                onClick={handleDelete}
                disabled={deleteSchedule.isPending}
                className={C.danger}
              >
                削除
              </Button>
            ) : null}
            <Button type="button" variant="outline" onClick={onClose}>
              キャンセル
            </Button>
            <SubmitButton>保存</SubmitButton>
          </DialogFooter>
        </form>
        {formState !== undefined ? null : null}
      </DialogContent>
    </Dialog>
  );
}

// ── Main Page ──

export function LineStaffSchedule() {
  const { currentClinicId } = useAuth();
  const { data: staffs = [] } = useGetReservationStaffs(currentClinicId);

  const [selectedStaffId, setSelectedStaffId] = useState<number | null>(null);
  const [currentMonth, setCurrentMonth] = useState(() => format(new Date(), "yyyy-MM"));
  const [editDate, setEditDate] = useState<string | null>(null);
  const [dialogOpen, setDialogOpen] = useState(false);

  const { data: schedules = [], isLoading } = useGetStaffSchedules(
    currentClinicId,
    selectedStaffId,
    currentMonth
  );

  const scheduleMap = useMemo(() => {
    const map = new Map<string, StaffSchedule>();
    for (const s of schedules) {
      map.set(s.date, s);
    }
    return map;
  }, [schedules]);

  const calendarDays = useMemo(() => {
    const monthDate = new Date(`${currentMonth}-01`);
    const days = eachDayOfInterval({
      start: startOfMonth(monthDate),
      end: endOfMonth(monthDate),
    });
    const firstDayOfWeek = getDay(days[0]);
    const blanks = Array.from({ length: firstDayOfWeek }, (_, i) => i);
    return { days, blanks };
  }, [currentMonth]);

  const handlePrevMonth = useCallback(() => {
    const [year, month] = currentMonth.split("-").map(Number);
    const prev = new Date(year, month - 2, 1);
    setCurrentMonth(format(prev, "yyyy-MM"));
  }, [currentMonth]);

  const handleNextMonth = useCallback(() => {
    const [year, month] = currentMonth.split("-").map(Number);
    const next = new Date(year, month, 1);
    setCurrentMonth(format(next, "yyyy-MM"));
  }, [currentMonth]);

  const handleDayClick = useCallback((dateStr: string) => {
    if (!selectedStaffId) {
      toast.warning("スタッフを選択してください");
      return;
    }
    setEditDate(dateStr);
    setDialogOpen(true);
  }, [selectedStaffId]);

  const editSchedule = editDate ? scheduleMap.get(editDate) ?? null : null;

  return (
    <PageLayout title="スタッフ個人スケジュール">
      <div className="space-y-4 max-w-[800px]">
        {/* Staff Selector */}
        <div className="flex items-center gap-4">
          <Label className={`text-sm ${C.text65} shrink-0`} htmlFor="staff-select">
            スタッフ選択
          </Label>
          <Select
            value={selectedStaffId ? String(selectedStaffId) : ""}
            onValueChange={(v) => setSelectedStaffId(Number(v))}
          >
            <SelectTrigger id="staff-select" className="max-w-[240px]">
              <SelectValue placeholder="スタッフを選択..." />
            </SelectTrigger>
            <SelectContent>
              {staffs.map((s) => (
                <SelectItem key={s.id} value={String(s.id)}>
                  {s.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {/* Month Navigator */}
        <div className="flex items-center gap-3">
          <button
            type="button"
            aria-label="前の月"
            onClick={handlePrevMonth}
            className={`p-1.5 rounded ${C.hoverBgMedium} transition-colors`}
          >
            <ChevronLeft className={ICON.action} />
          </button>
          <span className={`text-base font-semibold ${C.text} min-w-[120px] text-center`}>
            {format(new Date(`${currentMonth}-01`), "yyyy年M月", { locale: ja })}
          </span>
          <button
            type="button"
            aria-label="次の月"
            onClick={handleNextMonth}
            className={`p-1.5 rounded ${C.hoverBgMedium} transition-colors`}
          >
            <ChevronRight className={ICON.action} />
          </button>
        </div>

        {/* Weekday headers */}
        <div className={`grid grid-cols-7 border ${C.borderLight} rounded-t-lg overflow-hidden`}>
          {WEEKDAYS.map((day) => (
            <div
              key={day}
              className={`text-center text-xs font-medium py-2 ${C.textMuted} ${C.bgPage} border-b ${C.borderLight}`}
            >
              {day}
            </div>
          ))}

          {/* Blank cells */}
          {calendarDays.blanks.map((i) => (
            <div
              key={`blank-${i}`}
              className={`border-b border-r ${C.borderLight} ${C.bgPage} min-h-[72px]`}
            />
          ))}

          {/* Day cells */}
          {calendarDays.days.map((day) => {
            const dateStr = format(day, "yyyy-MM-dd");
            const schedule = scheduleMap.get(dateStr);
            const shiftConfig = schedule ? SHIFT_COLORS[schedule.shift_type] : null;
            const dayOfWeek = getDay(day);
            const isSunday = dayOfWeek === 0;
            const isSaturday = dayOfWeek === 6;

            return (
              <button
                key={dateStr}
                type="button"
                onClick={() => handleDayClick(dateStr)}
                disabled={!selectedStaffId || isLoading}
                className={[
                  "text-left p-1.5 border-b border-r min-h-[72px] transition-colors",
                  C.borderLight,
                  selectedStaffId
                    ? `${C.hoverBgLight} cursor-pointer`
                    : "cursor-default opacity-60",
                ].join(" ")}
              >
                <div
                  className={`text-xs font-medium mb-1 ${
                    isSunday
                      ? C.danger
                      : isSaturday
                      ? C.textBrand
                      : C.text
                  }`}
                >
                  {format(day, "d")}
                </div>
                {shiftConfig ? (
                  <span
                    className={`inline-block text-[10px] px-1.5 py-0.5 rounded font-medium ${shiftConfig.bg} ${shiftConfig.text}`}
                  >
                    {shiftConfig.label}
                  </span>
                ) : null}
              </button>
            );
          })}
        </div>

        {/* Legend */}
        <div className="flex items-center gap-3 flex-wrap">
          {Object.entries(SHIFT_COLORS).map(([key, config]) => (
            <div key={key} className="flex items-center gap-1">
              <span
                className={`inline-block text-[10px] px-1.5 py-0.5 rounded font-medium ${config.bg} ${config.text}`}
              >
                {config.label}
              </span>
            </div>
          ))}
        </div>
      </div>

      <ScheduleEditDialog
        open={dialogOpen}
        onClose={() => { setDialogOpen(false); setEditDate(null); }}
        date={editDate}
        schedule={editSchedule}
        staffId={selectedStaffId ?? 0}
        clinicId={currentClinicId}
      />
    </PageLayout>
  );
}
