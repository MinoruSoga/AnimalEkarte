import { useState, useCallback, useActionState, useMemo } from "react";
import { format, startOfMonth, endOfMonth, eachDayOfInterval, getDay, parseISO } from "date-fns";
import { ja } from "date-fns/locale";
import { ChevronLeft, ChevronRight, Plus, Calendar, List } from "lucide-react";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import { useAuth } from "@/features/auth";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
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
import { useGetLineReservations } from "../api/get-line-reservations";
import { useCreateLineReservation } from "../api/create-line-reservation";
import { useDeleteLineReservation } from "../api/delete-line-reservation";
import { useGetCourses } from "../api/get-courses";
import { useGetReservationStaffs } from "../api/get-staffs";
import type { ReservationAppointment, CreateReservationRequest } from "../api/types";

// ── Status badge ──

const STATUS_COLORS: Record<string, { bg: string; text: string }> = {
  confirmed: { bg: C.bgBrand10, text: C.textBrand },
  pending: { bg: C.bgNotice, text: C.textNotice },
  cancelled: { bg: C.bgInactive, text: C.textMuted },
};

function getStatusColor(status: string) {
  return STATUS_COLORS[status] ?? STATUS_COLORS.confirmed;
}

// ── Create Reservation Dialog ──

interface CreateReservationDialogProps {
  open: boolean;
  onClose: () => void;
  clinicId: string | null;
  initialDate?: string;
}

function CreateReservationDialog({
  open,
  onClose,
  clinicId,
  initialDate,
}: CreateReservationDialogProps) {
  const { data: courses = [] } = useGetCourses(clinicId);
  const { data: staffs = [] } = useGetReservationStaffs(clinicId);
  const createReservation = useCreateLineReservation(clinicId);

  const [selectedCourseId, setSelectedCourseId] = useState("");
  const [selectedDoctorId, setSelectedDoctorId] = useState("");

  const [formState, formAction] = useActionState(
    async (_prev: null, formData: FormData) => {
      const date = formData.get("date") as string;
      const startTime = formData.get("start_time") as string;
      const endTime = formData.get("end_time") as string;
      const payload: CreateReservationRequest = {
        start_time: `${date}T${startTime}:00`,
        end_time: `${date}T${endTime}:00`,
        service_type_id: Number(selectedCourseId),
        doctor_id: selectedDoctorId ? Number(selectedDoctorId) : undefined,
        notes: formData.get("notes") as string,
      };
      try {
        await createReservation.mutateAsync(payload);
        toast.success("予約を作成しました");
        onClose();
      } catch (err) {
        handleApiError(err, "予約作成");
      }
      return null;
    },
    null
  );

  return (
    <Dialog open={open} onOpenChange={(v) => { if (!v) onClose(); }}>
      <DialogContent className="max-w-[480px]">
        <DialogHeader>
          <DialogTitle>手動予約作成</DialogTitle>
        </DialogHeader>
        <form action={formAction} className="space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="res-date">日付</Label>
            <Input
              id="res-date"
              name="date"
              type="date"
              defaultValue={initialDate ?? format(new Date(), "yyyy-MM-dd")}
              required
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1.5">
              <Label htmlFor="start-time">開始時刻</Label>
              <Input id="start-time" name="start_time" type="time" defaultValue="09:00" required />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="end-time">終了時刻</Label>
              <Input id="end-time" name="end_time" type="time" defaultValue="09:30" required />
            </div>
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="course-select">コース</Label>
            <Select value={selectedCourseId} onValueChange={setSelectedCourseId}>
              <SelectTrigger id="course-select">
                <SelectValue placeholder="コースを選択..." />
              </SelectTrigger>
              <SelectContent>
                {courses.map((c) => (
                  <SelectItem key={c.id} value={String(c.id)}>
                    {c.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="doctor-select">担当スタッフ（任意）</Label>
            <Select value={selectedDoctorId} onValueChange={setSelectedDoctorId}>
              <SelectTrigger id="doctor-select">
                <SelectValue placeholder="未選択" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">未選択</SelectItem>
                {staffs.map((s) => (
                  <SelectItem key={s.id} value={String(s.id)}>
                    {s.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="notes">備考</Label>
            <Textarea id="notes" name="notes" rows={3} placeholder="備考..." />
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={onClose}>
              キャンセル
            </Button>
            <SubmitButton>作成</SubmitButton>
          </DialogFooter>
        </form>
        {formState !== undefined ? null : null}
      </DialogContent>
    </Dialog>
  );
}

// ── Month Calendar ──

interface MonthCalendarProps {
  date: string;
  reservations: ReservationAppointment[];
  onDayClick: (dateStr: string) => void;
}

function MonthCalendar({ date, reservations, onDayClick }: MonthCalendarProps) {
  const monthDate = useMemo(() => new Date(`${date}-01`), [date]);
  const days = useMemo(
    () =>
      eachDayOfInterval({
        start: startOfMonth(monthDate),
        end: endOfMonth(monthDate),
      }),
    [monthDate]
  );
  const firstDayOfWeek = getDay(days[0]);

  const reservationsByDay = useMemo(() => {
    const map = new Map<string, ReservationAppointment[]>();
    for (const r of reservations) {
      const d = r.start_time.slice(0, 10);
      if (!map.has(d)) map.set(d, []);
      map.get(d)!.push(r);
    }
    return map;
  }, [reservations]);

  const WEEKDAYS = ["日", "月", "火", "水", "木", "金", "土"];

  return (
    <div className={`border ${C.borderLight} rounded-lg overflow-hidden`}>
      <div className="grid grid-cols-7">
        {WEEKDAYS.map((day, i) => (
          <div
            key={day}
            className={`text-center text-xs font-medium py-2 ${
              i === 0 ? C.danger : i === 6 ? C.textBrand : C.textMuted
            } ${C.bgPage} border-b ${C.borderLight}`}
          >
            {day}
          </div>
        ))}

        {Array.from({ length: firstDayOfWeek }, (_, i) => (
          <div
            key={`blank-${i}`}
            className={`border-b border-r ${C.borderLight} ${C.bgPage} min-h-[100px]`}
          />
        ))}

        {days.map((day) => {
          const dateStr = format(day, "yyyy-MM-dd");
          const dayReservations = reservationsByDay.get(dateStr) ?? [];
          const dayOfWeek = getDay(day);

          return (
            <button
              key={dateStr}
              type="button"
              onClick={() => onDayClick(dateStr)}
              className={`text-left p-1.5 border-b border-r ${C.borderLight} min-h-[100px] ${C.hoverBgLight} transition-colors`}
            >
              <div
                className={`text-xs font-medium mb-1 ${
                  dayOfWeek === 0 ? C.danger : dayOfWeek === 6 ? C.textBrand : C.text
                }`}
              >
                {format(day, "d")}
              </div>
              {dayReservations.length > 0 ? (
                <div className={`text-xs ${C.textBrand} font-semibold`}>
                  {dayReservations.length}件
                </div>
              ) : null}
            </button>
          );
        })}
      </div>
    </div>
  );
}

// ── Day View ──

interface DayViewProps {
  date: string;
  reservations: ReservationAppointment[];
  onCancelClick: (reservation: ReservationAppointment) => void;
}

function DayView({ date, reservations, onCancelClick }: DayViewProps) {
  const sorted = useMemo(
    () => [...reservations].sort((a, b) => a.start_time.localeCompare(b.start_time)),
    [reservations]
  );

  return (
    <div className="space-y-2">
      <h3 className={`text-sm font-medium ${C.text65} mb-3`}>
        {format(parseISO(date), "M月d日（E）の予約", { locale: ja })}
      </h3>
      {sorted.length === 0 ? (
        <div className={`text-sm ${C.textMuted} py-4 text-center`}>
          この日の予約はありません。
        </div>
      ) : (
        sorted.map((r) => {
          const colors = getStatusColor(r.status);
          return (
            <div
              key={r.id}
              className={`flex items-center gap-3 px-4 py-3 rounded-lg border ${C.borderLight} ${C.bgWhite}`}
            >
              <div className={`text-sm font-mono ${C.text65}`}>
                {r.start_time.slice(11, 16)} - {r.end_time.slice(11, 16)}
              </div>
              <div className="flex-1 min-w-0">
                <div className={`text-sm font-medium ${C.text} truncate`}>
                  {r.service_type?.name ?? "コース未設定"}
                </div>
                <div className={`text-xs ${C.textMuted} truncate`}>
                  {r.line_customer?.display_name ?? "顧客情報なし"}
                  {r.doctor ? ` / ${r.doctor.name}` : ""}
                </div>
              </div>
              <span
                className={`text-xs px-2 py-0.5 rounded-full font-medium ${colors.bg} ${colors.text} shrink-0`}
              >
                {r.status}
              </span>
              <Button
                size="sm"
                variant="outline"
                onClick={() => onCancelClick(r)}
                className={`h-7 text-xs px-2 shrink-0 ${C.danger}`}
              >
                キャンセル
              </Button>
            </div>
          );
        })
      )}
    </div>
  );
}

// ── Main Page ──

export function LineReservationCalendar() {
  const { currentClinicId } = useAuth();

  const [view, setView] = useState<"month" | "day">("month");
  const [currentDate, setCurrentDate] = useState(() => format(new Date(), "yyyy-MM"));
  const [selectedDay, setSelectedDay] = useState<string | null>(null);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [cancelTarget, setCancelTarget] = useState<ReservationAppointment | null>(null);

  const apiDate = view === "month" ? currentDate : (selectedDay ?? currentDate + "-01");

  const { data: reservations = [], isLoading } = useGetLineReservations(currentClinicId, {
    view,
    date: apiDate,
  });

  const deleteReservation = useDeleteLineReservation(currentClinicId);

  const handlePrev = useCallback(() => {
    if (view === "month") {
      const [year, month] = currentDate.split("-").map(Number);
      const prev = new Date(year, month - 2, 1);
      setCurrentDate(format(prev, "yyyy-MM"));
    }
  }, [view, currentDate]);

  const handleNext = useCallback(() => {
    if (view === "month") {
      const [year, month] = currentDate.split("-").map(Number);
      const next = new Date(year, month, 1);
      setCurrentDate(format(next, "yyyy-MM"));
    }
  }, [view, currentDate]);

  const handleDayClick = useCallback((dateStr: string) => {
    setSelectedDay(dateStr);
    setView("day");
  }, []);

  const handleCancelConfirm = useCallback(() => {
    if (!cancelTarget) return;
    deleteReservation.mutate(cancelTarget.id, {
      onSuccess: () => {
        toast.success("予約をキャンセルしました");
        setCancelTarget(null);
      },
      onError: (err) => {
        handleApiError(err, "予約キャンセル");
        setCancelTarget(null);
      },
    });
  }, [cancelTarget, deleteReservation]);

  const navLabel =
    view === "month"
      ? format(new Date(`${currentDate}-01`), "yyyy年M月", { locale: ja })
      : selectedDay
      ? format(parseISO(selectedDay), "yyyy年M月d日", { locale: ja })
      : "";

  return (
    <PageLayout
      title="予約状況"
      headerAction={
        <Button size="sm" onClick={() => setCreateDialogOpen(true)} className={STYLE.confirmPrimary}>
          <Plus className={ICON.action} />
          手動予約
        </Button>
      }
    >
      <div className="space-y-4">
        {/* View Toggle + Navigation */}
        <div className="flex items-center gap-3 flex-wrap">
          <div className={`flex items-center rounded-md border ${C.borderLight} overflow-hidden`}>
            <button
              type="button"
              onClick={() => setView("month")}
              className={`flex items-center gap-1.5 px-3 h-8 text-sm transition-colors ${
                view === "month" ? `${C.bgBrand} text-white` : `${C.text65} ${C.hoverBgLight}`
              }`}
            >
              <Calendar className={ICON.action} />
              月表示
            </button>
            <button
              type="button"
              onClick={() => { setView("day"); if (!selectedDay) setSelectedDay(format(new Date(), "yyyy-MM-dd")); }}
              className={`flex items-center gap-1.5 px-3 h-8 text-sm transition-colors ${
                view === "day" ? `${C.bgBrand} text-white` : `${C.text65} ${C.hoverBgLight}`
              }`}
            >
              <List className={ICON.action} />
              日表示
            </button>
          </div>

          <div className="flex items-center gap-2">
            {view === "month" ? (
              <button
                type="button"
                aria-label="前の月"
                onClick={handlePrev}
                className={`p-1.5 rounded ${C.hoverBgMedium} transition-colors`}
              >
                <ChevronLeft className={ICON.action} />
              </button>
            ) : null}
            <span className={`text-sm font-semibold ${C.text} min-w-[120px] text-center`}>
              {navLabel}
            </span>
            {view === "month" ? (
              <button
                type="button"
                aria-label="次の月"
                onClick={handleNext}
                className={`p-1.5 rounded ${C.hoverBgMedium} transition-colors`}
              >
                <ChevronRight className={ICON.action} />
              </button>
            ) : null}
          </div>

          {view === "day" ? (
            <button
              type="button"
              onClick={() => setView("month")}
              className={`text-sm ${C.textBrand} ${C.hoverBgLight} px-2 py-1 rounded transition-colors`}
            >
              月表示に戻る
            </button>
          ) : null}
        </div>

        {isLoading ? (
          <div className={`text-sm ${C.textMuted} py-8 text-center`}>読み込み中...</div>
        ) : view === "month" ? (
          <MonthCalendar
            date={currentDate}
            reservations={reservations}
            onDayClick={handleDayClick}
          />
        ) : (
          <DayView
            date={selectedDay ?? format(new Date(), "yyyy-MM-dd")}
            reservations={reservations}
            onCancelClick={setCancelTarget}
          />
        )}
      </div>

      <CreateReservationDialog
        open={createDialogOpen}
        onClose={() => setCreateDialogOpen(false)}
        clinicId={currentClinicId}
        initialDate={selectedDay ?? undefined}
      />

      <ConfirmDialog
        open={cancelTarget !== null}
        onClose={() => setCancelTarget(null)}
        onConfirm={handleCancelConfirm}
        title="予約をキャンセルしますか？"
        description="この操作は元に戻せません。"
        confirmLabel="キャンセルする"
        cancelLabel="戻る"
        variant="destructive"
        isPending={deleteReservation.isPending}
      />
    </PageLayout>
  );
}
