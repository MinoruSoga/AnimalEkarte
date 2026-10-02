import { C, ICON } from "@/lib/design-tokens";
import { memo, useState, useCallback, useMemo } from "react";
import { Plus, ChevronLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import {
  Command,
  CommandInput,
  CommandItem,
  CommandList,
  CommandEmpty,
} from "@/components/ui/command";
import { FILTER_CONDITIONS } from "./types";
import { TextValueEditor } from "./TextValueEditor";
import { DateValueEditor } from "./DateValueEditor";
import type { FilterProperty, ActiveFilter, FilterCondition, FilterOption } from "./types";

// ─── Step tracking ────────────────────────────────────────

type AddStep = "property" | "condition" | "value" | "date-value" | "text-value";

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

  // Filter out already-used properties
  const activeKeys = useMemo(() => new Set(activeFilters.map((f) => f.key)), [activeFilters]);
  const availableProperties = useMemo(
    () => properties.filter((p) => !activeKeys.has(p.key)),
    [properties, activeKeys],
  );

  // ── Reset ──

  const resetState = useCallback(() => {
    setStep("property");
    setSelectedProperty(null);
    setSelectedCondition(null);
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
    // EMR-245: text は contains 固定のため条件ステップを飛ばして入力へ。
    if (prop.type === "text") {
      setSelectedCondition("contains");
      setStep("text-value");
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

  const applyTextFilter = useCallback(
    (value: string, displayValue: string) => {
      if (!selectedProperty) return;
      onAdd({
        key: selectedProperty.key,
        condition: "contains",
        value,
        displayValue,
      });
      resetState();
      setOpen(false);
    },
    [selectedProperty, onAdd, resetState],
  );

  const handleDateEditorApply = useCallback(
    (value: { from?: string; to?: string }, displayValue: string) => {
      if (!selectedProperty) return;
      onAdd({
        key: selectedProperty.key,
        condition: "is_between",
        value,
        displayValue,
      });
      resetState();
      setOpen(false);
    },
    [selectedProperty, onAdd, resetState],
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
      case "text-value":
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

  return (
    <Popover open={open} onOpenChange={handleOpenChange}>
      <PopoverTrigger asChild>
        <Button
          variant="ghost"
          size="sm"
          className={`h-11 gap-2 text-base ${C.text50} ${C.hoverText80} ${C.hoverBgLight} px-3`}
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
              className={`flex items-center gap-1 text-base ${C.text50} ${C.hoverText80} min-h-11 min-w-11 px-1`}
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
                  {prop.icon ? <prop.icon className={`mr-2 ${ICON.xs} ${C.text50}`} /> : null}
                  {prop.label}
                </CommandItem>
              ))}
            </CommandList>
          </Command>
        ) : step === "condition" ? (
          /* Step 2: Condition selection (select/multi-select only) */
          <div className="py-1">
            <p className={`text-base ${C.text40} px-3 py-1.5`}>{selectedProperty?.label} - 条件</p>
            {(
              FILTER_CONDITIONS[selectedProperty?.type ?? "select"] ?? FILTER_CONDITIONS.select
            ).map((cond) => (
              <button
                key={cond.value}
                type="button"
                onClick={() => handleSelectCondition(cond.value)}
                className={`w-full text-left px-3 min-h-11 text-base ${C.text} ${C.hoverBgLight} transition-colors`}
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
        ) : step === "text-value" ? (
          /* Step 3c: Text input (text) — contains 固定 */
          <div className="py-1">
            <p className={`text-base ${C.text40} px-3 py-1.5`}>{selectedProperty?.label} - 含む</p>
            <TextValueEditor label={selectedProperty?.label ?? ""} onApply={applyTextFilter} />
          </div>
        ) : step === "date-value" ? (
          /* Step 3b: 日付レンジ — 既存フィルタ再編集と同じ DateValueEditor を共有 */
          <DateValueEditor onApply={handleDateEditorApply} />
        ) : null}
      </PopoverContent>
    </Popover>
  );
});
