import { C, ICON } from "@/lib/design-tokens";
import { memo, useCallback, useMemo } from "react";
import { X } from "lucide-react";
import { FILTER_CONDITIONS } from "./types";
import { DateValueEditor } from "./DateValueEditor";
import { InlineSelector } from "./InlineSelector";
import type {
  ActiveFilter,
  FilterProperty,
  FilterCondition,
  FilterLogic,
  FilterOption,
} from "./types";

// ─── Condition label helpers ──────────────────────────────

function getConditionLabel(
  condition: FilterCondition,
  filterType: string,
): string {
  const conditions = FILTER_CONDITIONS[filterType as keyof typeof FILTER_CONDITIONS];
  return conditions?.find((c) => c.value === condition)?.label ?? condition;
}

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
                className={`w-full text-left px-2 py-1 text-base rounded-xxs transition-colors ${
                  logic === "and"
                    ? `${C.bgBrand5} ${C.textBrand}`
                    : `${C.text} ${C.hoverBgMedium}`
                }`}
              >
                AND
              </button>
              <button
                type="button"
                onClick={() => onLogicChange("or")}
                className={`w-full text-left px-2 py-1 text-base rounded-xxs transition-colors ${
                  logic === "or"
                    ? `${C.bgBrand5} ${C.textBrand}`
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
              className={`w-full text-left px-2 py-1 text-base rounded-xxs transition-colors ${
                filter.condition === opt.value
                  ? `${C.bgBrand5} ${C.textBrand}`
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
              className={`w-full text-left px-2 py-1 text-base rounded-xxs transition-colors ${
                filter.value === opt.value
                  ? `${C.bgBrand5} ${C.textBrand}`
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
        className={`ml-auto p-0.5 rounded-xxs ${C.text30} hover:${C.text60} ${C.hoverBgMedium} opacity-0 group-hover:opacity-100 transition-opacity`}
        aria-label={`${property?.label ?? filter.key} フィルタを削除`}
      >
        <X className={ICON.page} />
      </button>
    </div>
  );
});
