import { C, ICON } from "@/lib/design-tokens";
import { memo, useState, useCallback, useMemo } from "react";
import { X, ChevronDown } from "lucide-react";
import { cn } from "@/components/ui/utils";
import { format, startOfDay, endOfDay, startOfWeek, endOfWeek, startOfMonth, endOfMonth, startOfYear, endOfYear, subDays, subWeeks, subMonths, subYears, addDays, addWeeks, addMonths, addYears } from "date-fns";
import { ja } from "date-fns/locale";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Calendar } from "@/components/ui/calendar";
import type { DateRange } from "react-day-picker";
import { FILTER_CONDITIONS } from "./types";
import type {
  ActiveFilter,
  FilterProperty,
  FilterCondition,
  FilterLogic,
  FilterOption,
  RelativePoint,
  RelativeUnit,
} from "./types";

// ─── Relative date resolution ─────────────────────────────

function resolveRelativeDate(
  point: RelativePoint,
  unit: RelativeUnit,
): { from: Date; to: Date } {
  const now = new Date();
  const today = startOfDay(now);

  switch (unit) {
    case "day": {
      const target =
        point === "this" ? today :
        point === "last" ? subDays(today, 1) :
        addDays(today, 1);
      return { from: target, to: endOfDay(target) };
    }
    case "week": {
      const base =
        point === "this" ? today :
        point === "last" ? subWeeks(today, 1) :
        addWeeks(today, 1);
      return {
        from: startOfWeek(base, { locale: ja }),
        to: endOfWeek(base, { locale: ja }),
      };
    }
    case "month": {
      const base =
        point === "this" ? today :
        point === "last" ? subMonths(today, 1) :
        addMonths(today, 1);
      return { from: startOfMonth(base), to: endOfMonth(base) };
    }
    case "year": {
      const base =
        point === "this" ? today :
        point === "last" ? subYears(today, 1) :
        addYears(today, 1);
      return { from: startOfYear(base), to: endOfYear(base) };
    }
  }
}

// ─── Date presets (module-level constant) ─────────────────

type DatePreset =
  | { label: string; type: "relative"; point: RelativePoint; unit: RelativeUnit }
  | { label: string; type: "last_n_days"; n: number };

const DATE_PRESETS: DatePreset[] = [
  { label: "今日",     type: "relative",    point: "this", unit: "day" },
  { label: "昨日",     type: "relative",    point: "last", unit: "day" },
  { label: "今週",     type: "relative",    point: "this", unit: "week" },
  { label: "先週",     type: "relative",    point: "last", unit: "week" },
  { label: "今月",     type: "relative",    point: "this", unit: "month" },
  { label: "先月",     type: "relative",    point: "last", unit: "month" },
  { label: "直近7日",  type: "last_n_days", n: 7 },
  { label: "直近30日", type: "last_n_days", n: 30 },
];

function resolvePreset(preset: DatePreset): { from: Date; to: Date } {
  if (preset.type === "last_n_days") {
    const today = startOfDay(new Date());
    return { from: subDays(today, preset.n - 1), to: endOfDay(today) };
  }
  return resolveRelativeDate(preset.point, preset.unit);
}

// ─── Condition label helpers ──────────────────────────────

function getConditionLabel(
  condition: FilterCondition,
  filterType: string,
): string {
  const conditions = FILTER_CONDITIONS[filterType as keyof typeof FILTER_CONDITIONS];
  return conditions?.find((c) => c.value === condition)?.label ?? condition;
}

// ─── Inline selector sub-component ───────────────────────

interface InlineSelectorProps {
  label: string;
  children: React.ReactNode;
  popoverWidth?: string;
  noPadding?: boolean;
}

const InlineSelector = memo(function InlineSelector({
  label,
  children,
  popoverWidth = "w-[180px]",
  noPadding = false,
}: InlineSelectorProps) {
  const [open, setOpen] = useState(false);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button
          type="button"
          className={`flex items-center gap-1 px-2 py-1 text-base ${C.text} ${C.bgMutedBadge} ${C.hoverBgMutedBadge} rounded-[3px] transition-colors whitespace-nowrap max-w-[200px] truncate`}
        >
          <span className="truncate">{label}</span>
          <ChevronDown className={`${ICON.page} shrink-0 opacity-50`} />
        </button>
      </PopoverTrigger>
      <PopoverContent
        className={`${popoverWidth} ${noPadding ? "p-0" : "p-1"}`}
        align="start"
      >
        {children}
      </PopoverContent>
    </Popover>
  );
});

// ─── Date value editor ────────────────────────────────────

interface DateValueEditorProps {
  currentValue?: { from?: string; to?: string };
  onApply: (value: { from?: string; to?: string }, displayValue: string) => void;
}

const DateValueEditor = memo(function DateValueEditor({
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

  // FROM → TO display
  const hasFrom = !!dateRange?.from;
  const hasTo = !!(dateRange?.to && dateRange.from && dateRange.to.getTime() !== dateRange.from.getTime());
  const fromDisplay = hasFrom ? format(dateRange!.from!, "M月d日") : "開始日";
  const toDisplay = hasTo ? format(dateRange!.to!, "M月d日") : "終了日";

  return (
    <div className={`flex divide-x ${C.divideDivider}`}>
      {/* Presets column */}
      <div className="w-[108px] py-1 shrink-0">
        {DATE_PRESETS.map((preset) => (
          <button
            key={preset.label}
            type="button"
            onClick={() => {
              const { from, to } = resolvePreset(preset);
              handlePresetClick(from, to, preset.label);
            }}
            className={cn(`w-full text-left px-3 py-1.5 text-sm ${C.bgMutedBadge} ${C.hoverBgMutedBadge} transition-colors`, C.text)}
          >
            {preset.label}
          </button>
        ))}
      </div>

      {/* Calendar column */}
      <div className="p-3">
        {/* FROM → TO header */}
        <div className={`flex items-center justify-center gap-3 mb-3 px-3 py-2 ${C.bgPage} rounded-[4px]`}>
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
          onSelect={(range) => { if (range) handleCalendarSelect(range); }}
          locale={ja}
          numberOfMonths={1}
          className="rounded-md"
          captionLayout="dropdown"
          fromYear={2020}
          toYear={new Date().getFullYear() + 2}
          classNames={{
            months: "relative flex flex-col",
            month_caption: "flex justify-center items-center h-9 w-full",
            caption_label: "sr-only",
            nav: "absolute top-1 left-0 right-0 flex justify-between items-center px-1 pointer-events-none",
            button_previous: `size-8 p-0 rounded-sm ${C.bgMutedBadge} opacity-50 hover:opacity-100 inline-flex items-center justify-center pointer-events-auto`,
            button_next: `size-8 p-0 rounded-sm ${C.bgMutedBadge} opacity-50 hover:opacity-100 inline-flex items-center justify-center pointer-events-auto`,
            dropdowns: "flex items-center gap-1",
            dropdown: `text-sm font-medium bg-transparent border-none cursor-pointer focus:outline-none hover:opacity-70 py-0.5 px-1 rounded ${C.bgMutedBadge}`,
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

// ─── Main Component ───────────────────────────────────────

interface FilterRuleRowProps {
  filter: ActiveFilter;
  property: FilterProperty | undefined;
  isFirst: boolean;
  logic: FilterLogic;
  onLogicChange?: (logic: FilterLogic) => void;
  onUpdate: (updated: ActiveFilter) => void;
  onRemove: () => void;
}

export const FilterRuleRow = memo(function FilterRuleRow({
  filter,
  property,
  isFirst,
  logic,
  onLogicChange,
  onUpdate,
  onRemove,
}: FilterRuleRowProps) {
  const filterType = property?.type ?? "select";
  const isDateRange = filterType === "date-range";

  // rerender-dependencies: optional chaining を deps から排除し stable な変数に抽出
  const propertyConditions = property?.conditions;
  // Condition options: property.conditions が指定されていればそれを優先
  const conditionOptions = useMemo(() => {
    if (propertyConditions && propertyConditions.length > 0) {
      const base = FILTER_CONDITIONS[filterType] ?? FILTER_CONDITIONS.select;
      return base.filter((c) => propertyConditions.includes(c.value));
    }
    return FILTER_CONDITIONS[filterType] ?? FILTER_CONDITIONS.select;
  }, [filterType, propertyConditions]);

  const currentConditionLabel = useMemo(
    () => getConditionLabel(filter.condition, filterType),
    [filter.condition, filterType],
  );

  // ── Handlers ──

  const handleConditionChange = useCallback(
    (newCondition: FilterCondition) => {
      if (newCondition === "is_empty" || newCondition === "is_not_empty") {
        onUpdate({
          ...filter,
          condition: newCondition,
          value: "",
          displayValue: newCondition === "is_empty" ? "空" : "空でない",
        });
      } else {
        onUpdate({ ...filter, condition: newCondition });
      }
    },
    [filter, onUpdate],
  );

  const handleValueChange = useCallback(
    (option: FilterOption) => {
      onUpdate({
        ...filter,
        value: option.value,
        displayValue: option.label,
      });
    },
    [filter, onUpdate],
  );

  const handleDateValueApply = useCallback(
    (value: { from?: string; to?: string }, displayValue: string) => {
      onUpdate({ ...filter, value, displayValue });
    },
    [filter, onUpdate],
  );

  // Extract current date value for initializing calendar
  const currentDateValue = useMemo(
    () =>
      typeof filter.value === "object" && !Array.isArray(filter.value)
        ? (filter.value as { from?: string; to?: string })
        : undefined,
    [filter.value],
  );

  // ── Logic label ──

  const logicLabel = logic === "and" ? "AND" : "OR";
  const isEmptyCondition =
    filter.condition === "is_empty" || filter.condition === "is_not_empty";

  return (
    <div className="flex items-center gap-1.5 py-0.5 group">
      {/* Logic column */}
      <div className="w-[56px] shrink-0">
        {isFirst ? (
          onLogicChange ? (
            <InlineSelector label={logicLabel} popoverWidth="w-[100px]">
              <button
                type="button"
                onClick={() => onLogicChange("and")}
                className={`w-full text-left px-2 py-1 text-base rounded-[3px] transition-colors ${
                  logic === "and"
                    ? `${C.bgAccent5} ${C.accent}`
                    : `${C.text} ${C.hoverBgMedium}`
                }`}
              >
                AND
              </button>
              <button
                type="button"
                onClick={() => onLogicChange("or")}
                className={`w-full text-left px-2 py-1 text-base rounded-[3px] transition-colors ${
                  logic === "or"
                    ? `${C.bgAccent5} ${C.accent}`
                    : `${C.text} ${C.hoverBgMedium}`
                }`}
              >
                OR
              </button>
            </InlineSelector>
          ) : (
            <span className={`text-base ${C.text40} px-1.5`}>絞込</span>
          )
        ) : (
          <span className={`text-base ${C.text40} px-1.5`}>{logicLabel}</span>
        )}
      </div>

      {/* Property column */}
      <span className={`text-base ${C.text60} px-1 shrink-0 max-w-[140px] truncate`}>
        {property?.label ?? filter.key}
      </span>

      {/* Condition column — hidden for date-range (always "期間内") */}
      {isDateRange ? null : (
        <InlineSelector label={currentConditionLabel || "条件"} popoverWidth="w-[140px]">
          {conditionOptions.map((opt) => (
            <button
              key={opt.value}
              type="button"
              onClick={() => handleConditionChange(opt.value)}
              className={`w-full text-left px-2 py-1 text-base rounded-[3px] transition-colors ${
                filter.condition === opt.value
                  ? `${C.bgAccent5} ${C.accent}`
                  : `${C.text} ${C.hoverBgMedium}`
              }`}
            >
              {opt.label}
            </button>
          ))}
        </InlineSelector>
      )}

      {/* Value column */}
      {isEmptyCondition ? null : isDateRange ? (
        <InlineSelector
          label={filter.displayValue || "期間を選択"}
          popoverWidth="w-auto"
          noPadding
        >
          <DateValueEditor
            currentValue={currentDateValue}
            onApply={handleDateValueApply}
          />
        </InlineSelector>
      ) : (
        <InlineSelector
          label={filter.displayValue || "値を選択"}
          popoverWidth="w-[180px]"
        >
          {(property?.options ?? []).map((opt) => (
            <button
              key={opt.value}
              type="button"
              onClick={() => handleValueChange(opt)}
              className={`w-full text-left px-2 py-1 text-base rounded-[3px] transition-colors ${
                filter.value === opt.value
                  ? `${C.bgAccent5} ${C.accent}`
                  : `${C.text} ${C.hoverBgMedium}`
              }`}
            >
              {opt.label}
            </button>
          ))}
        </InlineSelector>
      )}

      {/* Remove button */}
      <button
        type="button"
        onClick={onRemove}
        className={`ml-auto p-0.5 rounded-[3px] ${C.text30} hover:${C.text60} ${C.hoverBgMedium} opacity-0 group-hover:opacity-100 transition-opacity`}
        aria-label={`${property?.label ?? filter.key} フィルタを削除`}
      >
        <X className={ICON.page} />
      </button>
    </div>
  );
});
