import { memo, useState, useCallback, useMemo } from "react";
import { X, ChevronDown } from "lucide-react";
import {
  format,
  startOfDay,
  endOfDay,
  startOfWeek,
  endOfWeek,
  startOfMonth,
  endOfMonth,
  startOfYear,
  endOfYear,
  subDays,
  subWeeks,
  subMonths,
  subYears,
  addDays,
  addWeeks,
  addMonths,
  addYears,
} from "date-fns";
import { ja } from "date-fns/locale";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Calendar } from "@/components/ui/calendar";
import type { DateRange } from "react-day-picker";
import {
  FILTER_CONDITIONS,
  RELATIVE_POINT_LABELS,
  RELATIVE_UNIT_LABELS,
} from "./types";
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

function getRelativeLabel(point: RelativePoint, unit: RelativeUnit): string {
  const pointLabel = RELATIVE_POINT_LABELS.find((p) => p.value === point)?.label ?? point;
  const unitLabel = RELATIVE_UNIT_LABELS.find((u) => u.value === unit)?.label ?? unit;
  return `${pointLabel} ${unitLabel}`;
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
}

const InlineSelector = memo(function InlineSelector({
  label,
  children,
  popoverWidth = "w-[180px]",
}: InlineSelectorProps) {
  const [open, setOpen] = useState(false);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button
          type="button"
          className="flex items-center gap-1 px-2 py-1 text-sm text-[#37352F] bg-[#F1F1EF] hover:bg-[#E8E7E4] rounded-[3px] transition-colors whitespace-nowrap max-w-[180px] truncate"
        >
          <span className="truncate">{label}</span>
          <ChevronDown className="h-4 w-4 shrink-0 opacity-50" />
        </button>
      </PopoverTrigger>
      <PopoverContent className={`${popoverWidth} p-1`} align="start">
        {children}
      </PopoverContent>
    </Popover>
  );
});

// ─── Date value editor ────────────────────────────────────

interface DateValueEditorProps {
  condition: FilterCondition;
  onApply: (value: { from?: string; to?: string }, displayValue: string) => void;
}

const DateValueEditor = memo(function DateValueEditor({
  condition,
  onApply,
}: DateValueEditorProps) {
  const [relPoint, setRelPoint] = useState<RelativePoint>("this");
  const [relUnit, setRelUnit] = useState<RelativeUnit>("week");
  const [dateRange, setDateRange] = useState<DateRange | undefined>(undefined);
  const [showCalendar, setShowCalendar] = useState(false);

  const handleRelativeApply = useCallback(
    (point: RelativePoint, unit: RelativeUnit) => {
      const resolved = resolveRelativeDate(point, unit);
      const label = getRelativeLabel(point, unit);
      onApply(
        {
          from: format(resolved.from, "yyyy-MM-dd"),
          to: format(resolved.to, "yyyy-MM-dd"),
        },
        label,
      );
    },
    [onApply],
  );

  const handleCalendarSelect = useCallback(
    (range: DateRange | undefined) => {
      setDateRange(range);
      if (!range?.from) return;

      if (condition === "is_between") {
        if (range.from && range.to && range.from.getTime() !== range.to.getTime()) {
          onApply(
            {
              from: format(range.from, "yyyy-MM-dd"),
              to: format(range.to, "yyyy-MM-dd"),
            },
            `${format(range.from, "MM/dd")}~${format(range.to, "MM/dd")}`,
          );
        }
      } else {
        onApply(
          {
            from: format(range.from, "yyyy-MM-dd"),
            to: format(range.from, "yyyy-MM-dd"),
          },
          format(range.from, "yyyy/MM/dd"),
        );
      }
    },
    [condition, onApply],
  );

  if (showCalendar) {
    return (
      <div className="p-2">
        <button
          type="button"
          onClick={() => setShowCalendar(false)}
          className="text-sm text-[#37352F]/50 hover:text-[#37352F]/80 mb-2"
        >
          ← 戻る
        </button>
        {condition === "is_between" ? (
          <Calendar
            mode="range"
            selected={dateRange}
            onSelect={(range) => { if (range) handleCalendarSelect(range); }}
            locale={ja}
            numberOfMonths={2}
            className="rounded-md"
          />
        ) : (
          <Calendar
            mode="single"
            selected={dateRange?.from}
            onSelect={(date) => { if (date) handleCalendarSelect({ from: date, to: date }); }}
            locale={ja}
            numberOfMonths={1}
            className="rounded-md"
          />
        )}
      </div>
    );
  }

  return (
    <div className="p-2 space-y-2">
      <p className="text-sm text-[#37352F]/50 px-1">相対日付</p>
      <div className="flex gap-1.5">
        <select
          value={relPoint}
          onChange={(e) => {
            const p = e.target.value as RelativePoint;
            setRelPoint(p);
            handleRelativeApply(p, relUnit);
          }}
          className="flex-1 h-8 text-sm border border-[rgba(55,53,47,0.16)] rounded-[3px] px-2 bg-white text-[#37352F]"
        >
          {RELATIVE_POINT_LABELS.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {opt.label}
            </option>
          ))}
        </select>
        <select
          value={relUnit}
          onChange={(e) => {
            const u = e.target.value as RelativeUnit;
            setRelUnit(u);
            handleRelativeApply(relPoint, u);
          }}
          className="flex-1 h-8 text-sm border border-[rgba(55,53,47,0.16)] rounded-[3px] px-2 bg-white text-[#37352F]"
        >
          {RELATIVE_UNIT_LABELS.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {opt.label}
            </option>
          ))}
        </select>
      </div>
      <button
        type="button"
        onClick={() => setShowCalendar(true)}
        className="w-full text-left px-2 py-1.5 text-sm text-[#37352F]/60 hover:bg-[#F1F1EF] rounded-[3px] transition-colors"
      >
        カレンダーから選択...
      </button>
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

  // Condition options for this property type
  const conditionOptions = useMemo(
    () => FILTER_CONDITIONS[filterType] ?? FILTER_CONDITIONS.select,
    [filterType],
  );

  const currentConditionLabel = useMemo(
    () => getConditionLabel(filter.condition, filterType),
    [filter.condition, filterType],
  );

  // ── Handlers ──

  const handleConditionChange = useCallback(
    (newCondition: FilterCondition) => {
      // If switching to is_empty/is_not_empty, clear value
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

  // ── Logic label / toggle (first row only) ──

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
                className={`w-full text-left px-2 py-1 text-sm rounded-[3px] transition-colors ${
                  logic === "and"
                    ? "bg-[#2383E2]/10 text-[#2383E2]"
                    : "text-[#37352F] hover:bg-[#F1F1EF]"
                }`}
              >
                AND
              </button>
              <button
                type="button"
                onClick={() => onLogicChange("or")}
                className={`w-full text-left px-2 py-1 text-sm rounded-[3px] transition-colors ${
                  logic === "or"
                    ? "bg-[#2383E2]/10 text-[#2383E2]"
                    : "text-[#37352F] hover:bg-[#F1F1EF]"
                }`}
              >
                OR
              </button>
            </InlineSelector>
          ) : (
            <span className="text-sm text-[#37352F]/40 px-1.5">Where</span>
          )
        ) : (
          <span className="text-sm text-[#37352F]/40 px-1.5">{logicLabel}</span>
        )}
      </div>

      {/* Property column */}
      <span className="text-sm text-[#37352F]/60 px-1 shrink-0 max-w-[140px] truncate">
        {property?.label ?? filter.key}
      </span>

      {/* Condition column */}
      <InlineSelector label={currentConditionLabel || "条件"} popoverWidth="w-[140px]">
        {conditionOptions.map((opt) => (
          <button
            key={opt.value}
            type="button"
            onClick={() => handleConditionChange(opt.value)}
            className={`w-full text-left px-2 py-1 text-sm rounded-[3px] transition-colors ${
              filter.condition === opt.value
                ? "bg-[#2383E2]/10 text-[#2383E2]"
                : "text-[#37352F] hover:bg-[#F1F1EF]"
            }`}
          >
            {opt.label}
          </button>
        ))}
      </InlineSelector>

      {/* Value column */}
      {isEmptyCondition ? null : filterType === "date-range" ? (
        <InlineSelector
          label={filter.displayValue || "値を選択"}
          popoverWidth="w-auto"
        >
          <DateValueEditor
            condition={filter.condition}
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
              className={`w-full text-left px-2 py-1 text-sm rounded-[3px] transition-colors ${
                filter.value === opt.value
                  ? "bg-[#2383E2]/10 text-[#2383E2]"
                  : "text-[#37352F] hover:bg-[#F1F1EF]"
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
        className="ml-auto p-0.5 rounded-[3px] text-[#37352F]/30 hover:text-[#37352F]/60 hover:bg-[#F1F1EF] opacity-0 group-hover:opacity-100 transition-opacity"
        aria-label={`${property?.label ?? filter.key} フィルタを削除`}
      >
        <X className="h-4 w-4" />
      </button>
    </div>
  );
});
