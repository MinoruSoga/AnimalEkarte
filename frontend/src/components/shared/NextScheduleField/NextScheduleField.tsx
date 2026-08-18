import { DatePicker } from "@/components/shared/DatePicker/DatePicker";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { C } from "@/lib/design-tokens";

import {
  NEXT_SCHEDULE_OPTIONS,
  type NextScheduleOption,
} from "./next-schedule";

interface NextScheduleFieldProps {
  typeId: string;
  dateId: string;
  scheduleType: string;
  nextDate: string;
  onScheduleTypeChange: (value: string) => void;
  onNextDateChange: (value: string) => void;
  dateAriaLabel?: string;
  options?: readonly NextScheduleOption[];
  error?: string;
  className?: string;
}

export function NextScheduleField({
  typeId,
  dateId,
  scheduleType,
  nextDate,
  onScheduleTypeChange,
  onNextDateChange,
  dateAriaLabel = "次回予定日",
  options = NEXT_SCHEDULE_OPTIONS,
  error,
  className,
}: NextScheduleFieldProps) {
  return (
    <div className={`space-y-2 ${className ?? ""}`}>
      <Label htmlFor={typeId}>次回の予定</Label>
      <div className="flex flex-wrap items-center gap-3">
        <Select value={scheduleType} onValueChange={onScheduleTypeChange}>
          <SelectTrigger id={typeId} className="w-[130px]">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {options.map((option) => (
              <SelectItem key={option.value} value={option.value}>
                {option.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Label htmlFor={dateId} className="sr-only">
          {dateAriaLabel}
        </Label>
        <DatePicker
          id={dateId}
          value={nextDate}
          onChange={onNextDateChange}
          className="min-w-[160px] flex-1"
        />
      </div>
      {error ? (
        <p role="alert" className={`text-sm ${C.danger}`}>
          {error}
        </p>
      ) : null}
    </div>
  );
}
