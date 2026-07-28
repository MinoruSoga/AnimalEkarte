import { C, ICON } from "@/lib/design-tokens";
import { memo, useState, useCallback, useMemo } from "react";
import { Plus, ChevronLeft } from "lucide-react";
import { format } from "date-fns";
import { ja } from "date-fns/locale";
import { Button } from "@/components/ui/button";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Command, CommandInput, CommandItem, CommandList, CommandEmpty } from "@/components/ui/command";
import { Calendar } from "@/components/ui/calendar";
import { toJSTWallDate } from "@/lib/jst-date";
import type { DateRange } from "react-day-picker";
import { FILTER_CONDITIONS } from "./types";
import { DATE_PRESETS, resolvePreset } from "./date-preset-utils";
import type { DatePreset } from "./date-preset-utils";
import type {
  FilterProperty,
  ActiveFilter,
  FilterCondition,
  FilterOption,
} from "./types";

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
  }, []);

  // ── Handlers ──

  const handleSelectProperty = useCallback((prop: FilterProperty) => {
    setSelectedProperty(prop);
    // For date-range, skip condition step — always range mode
    if (prop.type === "date-range") {
      setSelectedCondition("is_between");
      setStep("date-value");
      return;
    }
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
      if (!selectedProperty) return;
      onAdd({
        key: selectedProperty.key,
        condition: "is_between",
        value: { from: format(from, "yyyy-MM-dd"), to: format(to, "yyyy-MM-dd") },
        displayValue: label,
      });
      resetState();
      setOpen(false);
    },
    [selectedProperty, onAdd, resetState],
  );

  const handlePresetClick = useCallback(
    (preset: DatePreset) => {
      const { from, to } = resolvePreset(preset);
      applyDateFilter(from, to, preset.label);
    },
    [applyDateFilter],
  );

  const handleCalendarSelect = useCallback(
    (range: DateRange | undefined) => {
      setDateRange(range);
      if (!range?.from) return;
      if (range.to && range.from.getTime() !== range.to.getTime()) {
        applyDateFilter(
          range.from,
          range.to,
          `${format(range.from, "M/d")}〜${format(range.to, "M/d")}`,
        );
      }
    },
    [applyDateFilter],
  );

  const handleOpenChange = useCallback(
    (v: boolean) => {
      setOpen(v);
      if (!v) resetState();
    },
    [resetState],
  );

  const handleBack = useCallback(() => {
    switch (step) {
      case "condition":
        setStep("property");
        setSelectedProperty(null);
        break;
      case "value":
        setStep("condition");
        setSelectedCondition(null);
        break;
      case "date-value":
        // Skip condition step — go back to property
        setStep("property");
        setSelectedProperty(null);
        setSelectedCondition(null);
        break;
      default:
        break;
    }
  }, [step]);

  // Hide if all properties are in use
  if (availableProperties.length === 0) return null;

  const showBackButton = step !== "property";

  // FROM → TO display for date-value step
  const hasFrom = !!dateRange?.from;
  const hasTo = !!(dateRange?.to && dateRange.from && dateRange.to.getTime() !== dateRange.from.getTime());
  const fromDisplay = hasFrom ? format(dateRange!.from!, "M月d日") : "開始日";
  const toDisplay = hasTo ? format(dateRange!.to!, "M月d日") : "終了日";

  return (
    <Popover open={open} onOpenChange={handleOpenChange}>
      <PopoverTrigger asChild>
        <Button
          variant="ghost"
          size="sm"
          className={`h-9 gap-2 text-base ${C.text50} hover:${C.text80} ${C.hoverBgLight} px-3`}
        >
          <Plus className={ICON.page} />
          フィルタを追加
        </Button>
      </PopoverTrigger>
      <PopoverContent
        className={step === "date-value" ? "w-auto p-0" : "w-[220px] p-0"}
        align="start"
      >
        {/* Back button */}
        {showBackButton ? (
          <div className={step === "date-value" ? "px-2 pt-2 pb-0" : "px-2 pt-2"}>
            <button
              type="button"
              onClick={handleBack}
              className={`flex items-center gap-1 text-base ${C.text50} hover:${C.text80} px-1`}
            >
              <ChevronLeft className={ICON.action} />
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
                  className="text-base"
                >
                  {prop.icon ? (
                    <prop.icon className={`mr-2 ${ICON.xs} ${C.text50}`} />
                  ) : null}
                  {prop.label}
                </CommandItem>
              ))}
            </CommandList>
          </Command>
        ) : step === "condition" ? (
          /* Step 2: Condition selection (select/multi-select only) */
          <div className="py-1">
            <p className={`text-base ${C.text40} px-3 py-1.5`}>
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
                className={`w-full text-left px-3 py-1.5 text-base ${C.text} ${C.hoverBgLight} transition-colors`}
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
                  className="text-base"
                >
                  {opt.label}
                </CommandItem>
              ))}
            </CommandList>
          </Command>
        ) : step === "date-value" ? (
          /* Step 3b: Date range picker — presets + calendar */
          <div className={`flex divide-x ${C.divideDivider}`}>
            {/* Presets column */}
            <div className="w-[108px] py-1 shrink-0">
              {DATE_PRESETS.map((preset) => (
                <button
                  key={preset.label}
                  type="button"
                  onClick={() => handlePresetClick(preset)}
                  className={`w-full text-left px-3 py-1.5 text-sm ${C.text} ${C.hoverBgLight} transition-colors`}
                >
                  {preset.label}
                </button>
              ))}
            </div>

            {/* Calendar column */}
            <div className="p-3">
              {/* FROM → TO header */}
              <div className={`flex items-center justify-center gap-3 mb-3 px-3 py-2 ${C.bgPage} rounded-xs`}>
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
                numberOfMonths={1}
                locale={ja}
                className="rounded-md"
                captionLayout="dropdown"
                startMonth={new Date(2020, 0)}
                endMonth={new Date(toJSTWallDate(new Date()).getFullYear() + 2, 11)}
                classNames={{
                  months: "relative flex flex-col",
                  month_caption: "flex justify-center items-center h-9 w-full",
                  caption_label: "sr-only",
                  nav: "absolute top-1 left-0 right-0 flex justify-between items-center px-1 pointer-events-none",
                  button_previous: `size-8 p-0 rounded-sm ${C.hoverBgLight} opacity-50 hover:opacity-100 inline-flex items-center justify-center pointer-events-auto`,
                  button_next: `size-8 p-0 rounded-sm ${C.hoverBgLight} opacity-50 hover:opacity-100 inline-flex items-center justify-center pointer-events-auto`,
                  dropdowns: "flex items-center gap-1",
                  dropdown: `text-sm font-medium bg-transparent border-none cursor-pointer focus:outline-none hover:opacity-70 py-0.5 px-1 rounded ${C.hoverBgLight}`,
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
        ) : null}
      </PopoverContent>
    </Popover>
  );
});
