import { memo, useCallback, useState } from "react";
import { format } from "date-fns";
import { ja } from "date-fns/locale";
import type { DateRange } from "react-day-picker";

import { Calendar } from "@/components/ui/calendar";
import { cn } from "@/components/ui/utils";
import { C } from "@/lib/design-tokens";
import { toJSTWallDate } from "@/lib/jst-date";

import { DATE_PRESETS, resolvePreset } from "./date-preset-utils";

interface DateValueEditorProps {
  currentValue?: { from?: string; to?: string };
  onApply: (value: { from?: string; to?: string }, displayValue: string) => void;
}

export const DateValueEditor = memo(function DateValueEditor({
  currentValue,
  onApply,
}: DateValueEditorProps) {
  const [dateRange, setDateRange] = useState<DateRange | undefined>(() => {
    if (!currentValue?.from) return undefined;
    return {
      from: new Date(currentValue.from),
      to: currentValue.to ? new Date(currentValue.to) : undefined,
    };
  });

  const handlePresetClick = useCallback(
    (from: Date, to: Date, label: string) => {
      setDateRange({ from, to });
      onApply(
        { from: format(from, "yyyy-MM-dd"), to: format(to, "yyyy-MM-dd") },
        label,
      );
    },
    [onApply],
  );

  const handleCalendarSelect = useCallback(
    (range: DateRange | undefined) => {
      setDateRange(range);
      if (!range?.from) return;
      if (range.to && range.from.getTime() !== range.to.getTime()) {
        onApply(
          { from: format(range.from, "yyyy-MM-dd"), to: format(range.to, "yyyy-MM-dd") },
          `${format(range.from, "M/d")}〜${format(range.to, "M/d")}`,
        );
      }
    },
    [onApply],
  );

  const hasFrom = !!dateRange?.from;
  const hasTo = !!(
    dateRange?.to &&
    dateRange.from &&
    dateRange.to.getTime() !== dateRange.from.getTime()
  );
  const fromDisplay = hasFrom ? format(dateRange.from!, "M月d日") : "開始日";
  const toDisplay = hasTo ? format(dateRange.to!, "M月d日") : "終了日";

  return (
    <div className={`flex divide-x ${C.divideDivider}`}>
      <div className="w-[108px] py-1 shrink-0">
        {DATE_PRESETS.map((preset) => (
          <button
            key={preset.label}
            type="button"
            onClick={() => {
              const { from, to } = resolvePreset(preset);
              handlePresetClick(from, to, preset.label);
            }}
            className={cn(
              `w-full text-left px-3 py-1.5 text-sm ${C.bgMutedBadge} ${C.hoverBgMutedBadge} transition-colors`,
              C.text,
            )}
          >
            {preset.label}
          </button>
        ))}
      </div>

      <div className="p-3">
        <div
          className={`flex items-center justify-center gap-3 mb-3 px-3 py-2 ${C.bgPage} rounded-xs`}
        >
          <span className={`text-sm font-mono tabular-nums ${hasFrom ? `${C.text} font-medium` : C.text30}`}>
            {fromDisplay}
          </span>
          <span className={`${C.text30} text-xs`}>→</span>
          <span className={`text-sm font-mono tabular-nums ${hasTo ? `${C.text} font-medium` : C.text30}`}>
            {toDisplay}
          </span>
        </div>
        <Calendar
          mode="range"
          selected={dateRange}
          onSelect={(range) => {
            if (range) handleCalendarSelect(range);
          }}
          locale={ja}
          numberOfMonths={1}
          className="rounded-md"
          captionLayout="dropdown"
          startMonth={new Date(2020, 0)}
          endMonth={new Date(toJSTWallDate(new Date()).getFullYear() + 2, 11)}
          classNames={{
            months: "relative flex flex-col",
            month_caption: "flex justify-center items-center h-9 w-full",
            caption_label: "sr-only",
            nav: "absolute top-1 left-0 right-0 flex justify-between items-center px-1 pointer-events-none",
            button_previous: `size-8 p-0 rounded-sm ${C.bgMutedBadge} opacity-50 hover:opacity-100 inline-flex items-center justify-center pointer-events-auto`,
            button_next: `size-8 p-0 rounded-sm ${C.bgMutedBadge} opacity-50 hover:opacity-100 inline-flex items-center justify-center pointer-events-auto`,
            dropdowns: "flex items-center gap-1",
            dropdown: `text-sm font-medium bg-transparent border-none cursor-pointer focus:outline-none hover:opacity-70 py-0.5 px-1 rounded ${C.bgMutedBadge} focus-visible:ring-2 ${C.focusRingAccent40}`,
          }}
          formatters={{
            formatMonthDropdown: (month) => {
              const monthNumber = month instanceof Date ? month.getMonth() + 1 : Number(month) + 1;
              return `${monthNumber}月`;
            },
            formatYearDropdown: (year) => {
              const yearNumber = year instanceof Date ? year.getFullYear() : Number(year);
              return `${yearNumber}年`;
            },
          }}
        />
      </div>
    </div>
  );
});
