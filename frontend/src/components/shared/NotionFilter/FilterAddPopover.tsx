import { memo, useState, useCallback, useMemo } from "react";
import { Plus, ChevronLeft } from "lucide-react";
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
import { Button } from "@/components/ui/button";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import {
  Command,
  CommandInput,
  CommandItem,
  CommandList,
  CommandEmpty,
} from "@/components/ui/command";
import { Calendar } from "@/components/ui/calendar";
import type { DateRange } from "react-day-picker";
import {
  FILTER_CONDITIONS,
  RELATIVE_POINT_LABELS,
  RELATIVE_UNIT_LABELS,
} from "./types";
import type {
  FilterProperty,
  ActiveFilter,
  FilterCondition,
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

// ─── Step tracking ────────────────────────────────────────

type AddStep = "property" | "condition" | "value" | "date-value";

// ─── Component ────────────────────────────────────────────

interface FilterAddPopoverProps {
  properties: FilterProperty[];
  activeFilters: ActiveFilter[];
  onAdd: (filter: ActiveFilter) => void;
}

export const FilterAddPopover = memo(function FilterAddPopover({
  properties,
  activeFilters,
  onAdd,
}: FilterAddPopoverProps) {
  const [open, setOpen] = useState(false);
  const [step, setStep] = useState<AddStep>("property");
  const [selectedProperty, setSelectedProperty] = useState<FilterProperty | null>(null);
  const [selectedCondition, setSelectedCondition] = useState<FilterCondition | null>(null);
  const [dateRange, setDateRange] = useState<DateRange | undefined>(undefined);
  const [relPoint, setRelPoint] = useState<RelativePoint>("this");
  const [relUnit, setRelUnit] = useState<RelativeUnit>("week");
  const [showCalendar, setShowCalendar] = useState(false);

  // Filter out already-used properties
  const activeKeys = useMemo(
    () => new Set(activeFilters.map((f) => f.key)),
    [activeFilters],
  );
  const availableProperties = useMemo(
    () => properties.filter((p) => !activeKeys.has(p.key)),
    [properties, activeKeys],
  );

  // ── Reset ──

  const resetState = useCallback(() => {
    setStep("property");
    setSelectedProperty(null);
    setSelectedCondition(null);
    setDateRange(undefined);
    setShowCalendar(false);
    setRelPoint("this");
    setRelUnit("week");
  }, []);

  // ── Handlers ──

  const handleSelectProperty = useCallback((prop: FilterProperty) => {
    setSelectedProperty(prop);
    setStep("condition");
  }, []);

  const handleSelectCondition = useCallback(
    (condition: FilterCondition) => {
      if (!selectedProperty) return;
      setSelectedCondition(condition);

      // is_empty / is_not_empty -> apply immediately
      if (condition === "is_empty" || condition === "is_not_empty") {
        onAdd({
          key: selectedProperty.key,
          condition,
          value: "",
          displayValue: condition === "is_empty" ? "空" : "空でない",
        });
        resetState();
        setOpen(false);
        return;
      }

      // Date type -> date value step
      if (selectedProperty.type === "date-range") {
        setStep("date-value");
        return;
      }

      // Select / multi-select -> value step
      setStep("value");
    },
    [selectedProperty, onAdd, resetState],
  );

  const handleSelectValue = useCallback(
    (option: FilterOption) => {
      if (!selectedProperty || !selectedCondition) return;
      onAdd({
        key: selectedProperty.key,
        condition: selectedCondition,
        value: option.value,
        displayValue: option.label,
      });
      resetState();
      setOpen(false);
    },
    [selectedProperty, selectedCondition, onAdd, resetState],
  );

  const applyDateFilter = useCallback(
    (from: Date, to: Date, label: string) => {
      if (!selectedProperty || !selectedCondition) return;
      onAdd({
        key: selectedProperty.key,
        condition: selectedCondition,
        value: { from: format(from, "yyyy-MM-dd"), to: format(to, "yyyy-MM-dd") },
        displayValue: label,
      });
      resetState();
      setOpen(false);
    },
    [selectedProperty, selectedCondition, onAdd, resetState],
  );

  const handleRelativeApply = useCallback(
    (point: RelativePoint, unit: RelativeUnit) => {
      const resolved = resolveRelativeDate(point, unit);
      applyDateFilter(resolved.from, resolved.to, getRelativeLabel(point, unit));
    },
    [applyDateFilter],
  );

  const handleCalendarSelect = useCallback(
    (range: DateRange | undefined) => {
      setDateRange(range);
      if (!range?.from) return;

      const isBetween = selectedCondition === "is_between";
      if (isBetween) {
        if (range.from && range.to && range.from.getTime() !== range.to.getTime()) {
          applyDateFilter(
            range.from,
            range.to,
            `${format(range.from, "MM/dd")}~${format(range.to, "MM/dd")}`,
          );
        }
      } else {
        applyDateFilter(
          range.from,
          range.from,
          format(range.from, "yyyy/MM/dd"),
        );
      }
    },
    [selectedCondition, applyDateFilter],
  );

  const handleOpenChange = useCallback(
    (v: boolean) => {
      setOpen(v);
      if (!v) resetState();
    },
    [resetState],
  );

  const handleBack = useCallback(() => {
    if (showCalendar) {
      setShowCalendar(false);
      return;
    }
    switch (step) {
      case "condition":
        setStep("property");
        setSelectedProperty(null);
        break;
      case "value":
      case "date-value":
        setStep("condition");
        setSelectedCondition(null);
        break;
      default:
        break;
    }
  }, [step, showCalendar]);

  // Hide if all properties are in use
  if (availableProperties.length === 0) return null;

  // ── Render helpers ──

  const showBackButton = step !== "property" || showCalendar;
  const isBetweenCalendar = selectedCondition === "is_between";

  return (
    <Popover open={open} onOpenChange={handleOpenChange}>
      <PopoverTrigger asChild>
        <Button
          variant="ghost"
          size="sm"
          className="h-8 gap-1.5 text-sm text-[#37352F]/50 hover:text-[#37352F]/80 hover:bg-[#F1F1EF] px-3"
        >
          <Plus className="h-4 w-4" />
          フィルタを追加
        </Button>
      </PopoverTrigger>
      <PopoverContent
        className={
          step === "date-value" && showCalendar
            ? "w-auto p-0"
            : "w-[220px] p-0"
        }
        align="start"
      >
        {/* Back button */}
        {showBackButton ? (
          <div className="px-2 pt-2">
            <button
              type="button"
              onClick={handleBack}
              className="flex items-center gap-1 text-sm text-[#37352F]/50 hover:text-[#37352F]/80 px-1"
            >
              <ChevronLeft className="h-4 w-4" />
              戻る
            </button>
          </div>
        ) : null}

        {/* Step 1: Property selection */}
        {step === "property" ? (
          <Command>
            <CommandInput placeholder="フィルタを検索..." />
            <CommandList>
              <CommandEmpty>フィルタが見つかりません</CommandEmpty>
              {availableProperties.map((prop) => (
                <CommandItem
                  key={prop.key}
                  onSelect={() => handleSelectProperty(prop)}
                  className="text-sm"
                >
                  {prop.icon ? (
                    <prop.icon className="mr-2 h-5 w-5 text-[#37352F]/50" />
                  ) : null}
                  {prop.label}
                </CommandItem>
              ))}
            </CommandList>
          </Command>
        ) : step === "condition" ? (
          /* Step 2: Condition selection */
          <div className="py-1">
            <p className="text-sm text-[#37352F]/40 px-3 py-1.5">
              {selectedProperty?.label} - 条件
            </p>
            {(
              FILTER_CONDITIONS[selectedProperty?.type ?? "select"] ??
              FILTER_CONDITIONS.select
            ).map((cond) => (
              <button
                key={cond.value}
                type="button"
                onClick={() => handleSelectCondition(cond.value)}
                className="w-full text-left px-3 py-1.5 text-sm text-[#37352F] hover:bg-[#F1F1EF] transition-colors"
              >
                {cond.label}
              </button>
            ))}
          </div>
        ) : step === "value" ? (
          /* Step 3a: Value selection (select/multi-select) */
          <Command>
            <CommandInput placeholder={`${selectedProperty?.label ?? ""}を選択...`} />
            <CommandList>
              <CommandEmpty>選択肢がありません</CommandEmpty>
              {(selectedProperty?.options ?? []).map((opt) => (
                <CommandItem
                  key={opt.value}
                  onSelect={() => handleSelectValue(opt)}
                  className="text-sm"
                >
                  {opt.label}
                </CommandItem>
              ))}
            </CommandList>
          </Command>
        ) : step === "date-value" && showCalendar ? (
          /* Step 3b-calendar: Calendar picker */
          <div className="p-2">
            {isBetweenCalendar ? (
              <Calendar
                mode="range"
                selected={dateRange}
                onSelect={(range) => { if (range) handleCalendarSelect(range); }}
                numberOfMonths={2}
                locale={ja}
                className="rounded-md"
              />
            ) : (
              <Calendar
                mode="single"
                selected={dateRange?.from}
                onSelect={(date) => { if (date) handleCalendarSelect({ from: date, to: date }); }}
                numberOfMonths={1}
                locale={ja}
                className="rounded-md"
              />
            )}
          </div>
        ) : step === "date-value" ? (
          /* Step 3b: Date value (relative selector) */
          <div className="py-2 px-3 space-y-2">
            <p className="text-sm text-[#37352F]/40">相対日付</p>
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
        ) : null}
      </PopoverContent>
    </Popover>
  );
});
