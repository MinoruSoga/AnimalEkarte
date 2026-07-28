import type { ChangeEventHandler, ReactNode } from "react";

import { Input } from "@/components/ui/input";
import { C } from "@/lib/design-tokens";
import { cn } from "@/lib/utils";

interface DateRangeInputsProps {
  fromValue: string;
  toValue: string;
  onFromChange: ChangeEventHandler<HTMLInputElement>;
  onToChange: ChangeEventHandler<HTMLInputElement>;
  /** 呼び出し側の従来スタイルをそのまま渡す。省略時は shadcn Input の既定スタイル */
  inputClassName?: string;
  className?: string;
  fromTestId?: string;
  toTestId?: string;
  fromAriaLabel?: string;
  toAriaLabel?: string;
  separator?: ReactNode;
  separatorClassName?: string;
}

/**
 * ネイティブ日付 Input x2 + 「〜」区切りの軽量レンジフィルタ。
 * AggregationFilterPanelParts の VisitFilters と LstepDeliveryMonitorPageParts の
 * DeliveryMonitorFilters から抽出。重量級の DatePickerRange とは別物。
 */
export function DateRangeInputs({
  fromValue,
  toValue,
  onFromChange,
  onToChange,
  inputClassName,
  className,
  fromTestId,
  toTestId,
  fromAriaLabel = "開始日",
  toAriaLabel = "終了日",
  separator = "〜",
  separatorClassName = `text-sm ${C.text50}`,
}: DateRangeInputsProps) {
  return (
    <div className={cn("flex items-center gap-2", className)}>
      <Input
        type="date"
        value={fromValue}
        onChange={onFromChange}
        aria-label={fromAriaLabel}
        className={inputClassName}
        data-testid={fromTestId}
      />
      <span className={separatorClassName}>{separator}</span>
      <Input
        type="date"
        value={toValue}
        onChange={onToChange}
        aria-label={toAriaLabel}
        className={inputClassName}
        data-testid={toTestId}
      />
    </div>
  );
}
