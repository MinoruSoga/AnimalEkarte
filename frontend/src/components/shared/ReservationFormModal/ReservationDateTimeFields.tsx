import { useCallback, useState } from "react";
import { format } from "date-fns";
import { ja } from "date-fns/locale";
import { C, ICON } from "@/lib/design-tokens";
import { Label } from "@/components/ui/label";
import { FormFieldError } from "@/components/shared/FormFieldError";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Calendar } from "@/components/ui/calendar";
import { CalendarNav, MonthGrid, YearNav } from "@/components/shared/DatePicker/DatePickerParts";
import { SINGLE_CALENDAR_CLASSES } from "@/components/shared/DatePicker/DatePickerModel";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Calendar as CalendarIcon, Clock, ArrowRight } from "lucide-react";
import { cn } from "@/lib/utils";
import { DISPLAY_TIME_FORMAT } from "@/lib/format/date";
import {
  DEFAULT_RESERVATION_DURATION_MINUTES,
  resolveEndTimeOnStartChange,
  slotVacancyLabel,
  TIME_OPTIONS,
  type SlotVacancy,
  withCurrentTimeOption,
} from "./reservation-time-utils";
import type { ReservationSlotVacancyStatus } from "@/hooks/use-reservation-types";
import type { Reservation } from "@/types";

const TRIGGER_CLASS = `h-11 text-sm bg-white ${C.borderMediumLight} ${C.text} ${C.hoverBgSubtle} transition-colors`;

// EMR-170: 空き状況バッジの色（色だけでなく記号＋テキストも併記する）
const VACANCY_BADGE_CLASS: Record<ReservationSlotVacancyStatus, string> = {
  available: C.text60,
  low: C.textStatusAmber,
  full: C.danger,
};

export interface FieldLabelProps {
  children: React.ReactNode;
  required?: boolean;
  trailing?: React.ReactNode;
}

export function FieldLabel({ children, required, trailing }: FieldLabelProps) {
  return (
    <div className="flex items-center justify-between gap-2">
      <Label className={`text-xs ${C.text60} font-medium`}>
        {children}
        {required ? (
          <span className={`ml-1 ${C.textRequired}`} aria-hidden="true">
            *
          </span>
        ) : null}
      </Label>
      {trailing ? <div>{trailing}</div> : null}
    </div>
  );
}

interface ReservationDateTimeFieldsProps {
  formData: Partial<Reservation>;
  onChange: (data: Partial<Reservation>) => void;
  validationErrors?: Record<string, string>;
  isCalendarDateDisabled: (date: Date) => boolean;
  handleMonthChange: (month: Date) => void;
  startTimeOptions: string[];
  availableTimeSlotMap: Map<string, string> | undefined;
  /** EMR-170: 開始時刻 → 空き状況。status なし応答・手動入力パスでは undefined/空。 */
  slotVacancyMap?: Map<string, SlotVacancy>;
  /** Guidance when LINE reservation settings are unset (manual time path). */
  settingsUnsetGuidance?: string | null;
  /** Non-unset available-times fetch failure message. */
  availableTimesErrorMessage?: string | null;
  /** EMR-191: 選択中予約区分の所要時間（分）。未選択/未設定時は既定15分。 */
  durationMinutes?: number;
}

export function ReservationDateTimeFields({
  formData,
  onChange,
  validationErrors,
  isCalendarDateDisabled,
  handleMonthChange,
  startTimeOptions,
  availableTimeSlotMap,
  slotVacancyMap,
  settingsUnsetGuidance = null,
  availableTimesErrorMessage = null,
  durationMinutes = DEFAULT_RESERVATION_DURATION_MINUTES,
}: ReservationDateTimeFieldsProps) {
  // EMR-191: 終了側も開始側(BUG-015)と同じ規則で現在値を options に注入し、
  // 非15分刻みの終了時刻を空白表示しない
  const endTimeOptions = withCurrentTimeOption(
    TIME_OPTIONS,
    formData.end ? format(formData.end, DISPLAY_TIME_FORMAT) : undefined,
  );

  const [calendarOpen, setCalendarOpen] = useState(false);
  const [calendarView, setCalendarView] = useState<"calendar" | "monthGrid">("calendar");
  const [displayMonth, setDisplayMonth] = useState<Date>(() => formData.start ?? new Date());

  // 表示月の変更を親へ通知し、当月の空き枠を再取得させる
  const goToMonth = useCallback(
    (month: Date) => {
      setDisplayMonth(month);
      handleMonthChange(month);
    },
    [handleMonthChange],
  );

  const handleCalendarOpenChange = useCallback(
    (nextOpen: boolean) => {
      setCalendarOpen(nextOpen);
      if (nextOpen) {
        setCalendarView("calendar");
        goToMonth(formData.start ?? new Date());
      }
    },
    [formData.start, goToMonth],
  );

  const handlePrevMonth = useCallback(() => {
    goToMonth(new Date(displayMonth.getFullYear(), displayMonth.getMonth() - 1, 1));
  }, [displayMonth, goToMonth]);

  const handleNextMonth = useCallback(() => {
    goToMonth(new Date(displayMonth.getFullYear(), displayMonth.getMonth() + 1, 1));
  }, [displayMonth, goToMonth]);

  const handleMonthSelect = useCallback(
    (month: number) => {
      goToMonth(new Date(displayMonth.getFullYear(), month, 1));
      setCalendarView("calendar");
    },
    [displayMonth, goToMonth],
  );

  const handleYearDelta = useCallback(
    (delta: number) => {
      goToMonth(new Date(displayMonth.getFullYear() + delta, displayMonth.getMonth(), 1));
    },
    [displayMonth, goToMonth],
  );

  return (
    <div className={`rounded-lg border ${C.bgSubtle} p-3 space-y-3 ${C.borderMediumLight}`}>
      <div className="space-y-1.5">
        <FieldLabel required>日付</FieldLabel>
        <Popover open={calendarOpen} onOpenChange={handleCalendarOpenChange}>
          <PopoverTrigger asChild>
            <button
              type="button"
              className={cn(
                `flex h-11 w-full items-center justify-between rounded-xs border px-3 py-1 text-sm transition-colors ${C.borderMediumLight} ${C.text} bg-white ${C.hoverBgSubtle}`,
                !formData.start && C.textMuted,
              )}
            >
              <span className="flex items-center">
                <CalendarIcon className={`mr-2 ${ICON.action}`} />
                {formData.start ? (
                  format(formData.start, "yyyy/MM/dd (E)", { locale: ja })
                ) : (
                  <span>日付を選択</span>
                )}
              </span>
            </button>
          </PopoverTrigger>
          <PopoverContent className="w-auto p-0">
            {calendarView === "calendar" ? (
              <div className="pt-2">
                <CalendarNav
                  displayMonth={displayMonth}
                  onPrev={handlePrevMonth}
                  onNext={handleNextMonth}
                  onTitleClick={() => setCalendarView("monthGrid")}
                />
              </div>
            ) : (
              <div className="pt-2">
                <YearNav
                  year={displayMonth.getFullYear()}
                  onPrevYear={() => handleYearDelta(-1)}
                  onNextYear={() => handleYearDelta(1)}
                />
              </div>
            )}

            {calendarView === "calendar" ? (
              <Calendar
                mode="single"
                month={displayMonth}
                onMonthChange={goToMonth}
                selected={formData.start}
                onSelect={(date) => {
                  if (!date) return;
                  const newStart = new Date(date);
                  const newEnd = new Date(date);

                  if (formData.start) {
                    newStart.setHours(formData.start.getHours(), formData.start.getMinutes());
                  }
                  if (formData.end) {
                    newEnd.setHours(formData.end.getHours(), formData.end.getMinutes());
                  }

                  onChange({ ...formData, start: newStart, end: newEnd });
                }}
                disabled={isCalendarDateDisabled}
                locale={ja}
                fixedWeeks
                autoFocus
                className="rounded-md pt-0"
                classNames={SINGLE_CALENDAR_CLASSES}
              />
            ) : (
              <MonthGrid currentMonth={displayMonth.getMonth()} onSelect={handleMonthSelect} />
            )}
          </PopoverContent>
        </Popover>
        {validationErrors?.date ? (
          <FormFieldError id="res-date-error" message={validationErrors.date} />
        ) : null}
      </div>

      <div className="space-y-1.5">
        <div className={`flex items-center gap-2 text-xs ${C.text60} font-medium`}>
          <Clock className={ICON.action} />
          時間
          <span className={`ml-0.5 ${C.textRequired}`} aria-hidden="true">
            *
          </span>
        </div>
        <div className="flex items-center gap-2">
          <Select
            value={formData.start ? format(formData.start, DISPLAY_TIME_FORMAT) : "10:00"}
            onValueChange={(v) => {
              if (!formData.start) return;
              const [h, m] = v.split(":").map(Number);
              const newStart = new Date(formData.start);
              newStart.setHours(h, m);
              // EMR-191: LINE 空き枠の終了時刻を優先し、無ければ durationMinutes 加算で終了時刻を自動設定
              const newEnd = resolveEndTimeOnStartChange(
                newStart,
                durationMinutes,
                availableTimeSlotMap?.get(v),
              );
              onChange({ ...formData, start: newStart, end: newEnd });
            }}
          >
            <SelectTrigger data-testid="res-start-time-trigger" className={TRIGGER_CLASS}>
              <SelectValue />
            </SelectTrigger>
            <SelectContent className="max-h-[200px]">
              {startTimeOptions.map((time) => {
                const vacancy = slotVacancyMap?.get(time);
                return (
                  <SelectItem
                    key={time}
                    value={time}
                    textValue={time}
                    disabled={vacancy?.status === "full"}
                  >
                    <span className="flex items-center gap-2">
                      <span>{time}</span>{" "}
                      {vacancy ? (
                        <span
                          data-testid={`res-start-time-vacancy-${time}`}
                          className={`text-xs ${VACANCY_BADGE_CLASS[vacancy.status]}`}
                        >
                          {slotVacancyLabel(vacancy.status)}
                        </span>
                      ) : null}
                    </span>
                  </SelectItem>
                );
              })}
            </SelectContent>
          </Select>
          <ArrowRight className={`${ICON.action} ${C.text40} flex-shrink-0`} />
          <Select
            value={formData.end ? format(formData.end, DISPLAY_TIME_FORMAT) : "11:00"}
            onValueChange={(v) => {
              if (!formData.end) return;
              const [h, m] = v.split(":").map(Number);
              const newEnd = new Date(formData.end);
              newEnd.setHours(h, m);
              onChange({ ...formData, end: newEnd });
            }}
          >
            <SelectTrigger data-testid="res-end-time-trigger" className={TRIGGER_CLASS}>
              <SelectValue />
            </SelectTrigger>
            <SelectContent className="max-h-[200px]">
              {endTimeOptions.map((time) => (
                <SelectItem key={time} value={time}>
                  {time}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        {settingsUnsetGuidance ? (
          <p data-testid="res-available-times-unset-guidance" className={`text-xs ${C.text60}`}>
            {settingsUnsetGuidance}
          </p>
        ) : null}
        {availableTimesErrorMessage ? (
          <FormFieldError id="res-available-times-error" message={availableTimesErrorMessage} />
        ) : null}
        {validationErrors?.time ? (
          <FormFieldError id="res-time-error" message={validationErrors.time} />
        ) : null}
      </div>
    </div>
  );
}
