import { format } from "date-fns";
import { ja } from "date-fns/locale";
import { C, ICON } from "@/lib/design-tokens";
import { Label } from "@/components/ui/label";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Calendar } from "@/components/ui/calendar";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Calendar as CalendarIcon, Clock, ArrowRight } from "lucide-react";
import { cn } from "@/lib/utils";
import { DISPLAY_TIME_FORMAT } from "@/lib/format/date";
import { TIME_OPTIONS } from "./reservation-time-utils";
import type { Reservation } from "@/types";

const TRIGGER_CLASS =
  `h-9 text-sm bg-white ${C.borderMediumLight} ${C.text} ${C.hoverBgSubtle} transition-colors`;

export interface FieldLabelProps {
  children: React.ReactNode;
  required?: boolean;
  trailing?: React.ReactNode;
}

export function FieldLabel({ children, required, trailing }: FieldLabelProps) {
  return (
    <div className="flex items-center justify-between gap-2">
      <Label className={`text-xs ${C.text40} font-medium`}>
        {children}
        {required ? (
          <span style={{ color: C.danger }} className="ml-1" aria-hidden="true">*</span>
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
}

export function ReservationDateTimeFields({
  formData,
  onChange,
  validationErrors,
  isCalendarDateDisabled,
  handleMonthChange,
  startTimeOptions,
  availableTimeSlotMap,
}: ReservationDateTimeFieldsProps) {
  return (
    <div className={`rounded-lg border ${C.bgSubtle} p-3 space-y-3 ${C.borderMediumLight}`}>
      <div className="space-y-1.5">
        <FieldLabel required>日付</FieldLabel>
        <Popover>
          <PopoverTrigger asChild>
            <button
              type="button"
              className={cn(
                `flex h-9 w-full items-center justify-between rounded border px-3 py-1 text-sm transition-colors ${C.borderMediumLight} ${C.text} bg-white ${C.hoverBgSubtle}`,
                !formData.start && C.text40
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
            <Calendar
              mode="single"
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
              onMonthChange={handleMonthChange}
              autoFocus
            />
          </PopoverContent>
        </Popover>
        {validationErrors?.date ? (
          <FormFieldError id="res-date-error" message={validationErrors.date} />
        ) : null}
      </div>

      <div className="space-y-1.5">
        <div className={`flex items-center gap-2 text-xs ${C.text40} font-medium`}>
          <Clock className={ICON.action} />
          時間
          <span style={{ color: C.danger }} className="ml-0.5" aria-hidden="true">*</span>
        </div>
        <div className="flex items-center gap-2">
          <Select
            value={formData.start ? format(formData.start, DISPLAY_TIME_FORMAT) : "10:00"}
            onValueChange={(v) => {
              if (!formData.start) return;
              const [h, m] = v.split(":").map(Number);
              const newStart = new Date(formData.start);
              newStart.setHours(h, m);
              const nextData: Partial<Reservation> = { ...formData, start: newStart };
              const slotEnd = availableTimeSlotMap?.get(v);
              if (slotEnd && formData.end) {
                const [endHour, endMinute] = slotEnd.split(":").map(Number);
                const newEnd = new Date(formData.end);
                newEnd.setHours(endHour, endMinute);
                nextData.end = newEnd;
              }
              onChange(nextData);
            }}
          >
            <SelectTrigger data-testid="res-start-time-trigger" className={TRIGGER_CLASS}>
              <SelectValue />
            </SelectTrigger>
            <SelectContent className="max-h-[200px]">
              {startTimeOptions.map((time) => (
                <SelectItem key={time} value={time}>
                  {time}
                </SelectItem>
              ))}
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
              {TIME_OPTIONS.map((time) => (
                <SelectItem key={time} value={time}>
                  {time}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        {validationErrors?.time ? (
          <FormFieldError id="res-time-error" message={validationErrors.time} />
        ) : null}
      </div>
    </div>
  );
}
