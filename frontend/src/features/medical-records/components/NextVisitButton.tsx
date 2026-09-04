import { CalendarPlus, Calendar, Pencil } from "lucide-react";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Button } from "@/components/ui/button";
import { Tooltip } from "@/components/ui/tooltip";
import { C, ICON } from "@/lib/design-tokens";
import { NextVisitDateField } from "./NextVisitDateField";

interface NextVisitButtonProps {
  value: string;
  onChange: (date: string) => void;
  onValidationChange: (valid: boolean) => void;
  hasLineIntegration?: boolean;
  disabled?: boolean;
}

export function NextVisitButton({
  value,
  onChange,
  onValidationChange,
  hasLineIntegration,
  disabled,
}: NextVisitButtonProps) {
  const hasValue = value !== "" && value != null;
  const tooltipText = hasValue ? "次回予定を変更" : "次回予定を追加";
  const ariaLabel = tooltipText;

  return (
    <div className="flex flex-col gap-0 shrink-0">
      <span className={`text-xs ${C.text50}`}>次回予定</span>
      <Popover>
        <Tooltip content={tooltipText}>
          <PopoverTrigger asChild>
            <Button
              type="button"
              variant="ghost"
              size="sm"
              disabled={disabled}
              aria-label={ariaLabel}
              className={`h-8 px-2 gap-1 text-sm font-medium ${C.text} ${C.hoverBgPage} border-none`}
            >
              {hasValue ? (
                <>
                  <Calendar className={`${ICON.sm} ${C.text60}`} aria-hidden="true" />
                  <span>{value}</span>
                  <Pencil className={`${ICON.sm} ${C.text40}`} aria-hidden="true" />
                </>
              ) : (
                <CalendarPlus className={`${ICON.sm} ${C.text40}`} aria-hidden="true" />
              )}
            </Button>
          </PopoverTrigger>
        </Tooltip>
        <PopoverContent className="w-auto p-4" align="start">
          <NextVisitDateField
            value={value}
            onChange={onChange}
            onValidationChange={onValidationChange}
            hasLineIntegration={hasLineIntegration}
            disabled={disabled}
          />
        </PopoverContent>
      </Popover>
    </div>
  );
}
