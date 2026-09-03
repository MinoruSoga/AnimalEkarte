import type { ChangeEvent } from "react";
import { AlertTriangle } from "lucide-react";

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { DateRangeInputs } from "@/components/shared/DateRangeInputs";
import { BADGE, C, ICON } from "@/lib/design-tokens";

import type { DeliveryTriggerSummaryResponse } from "../api/get-lstep-delivery-trigger-summary";
import { TriggerStatusLabels, TriggerTypeLabels } from "../constants/trigger-types";
import { DELIVERY_STATUS_CARDS } from "./lstep-delivery-monitor-page-model";

interface DeliveryMonitorFiltersProps {
  from: string;
  to: string;
  triggerType: string;
  statusFilter: string;
  onFromChange: (event: ChangeEvent<HTMLInputElement>) => void;
  onToChange: (event: ChangeEvent<HTMLInputElement>) => void;
  onTriggerTypeChange: (value: string) => void;
  onStatusChange: (value: string) => void;
}

export function DeliveryMonitorFilters({
  from,
  to,
  triggerType,
  statusFilter,
  onFromChange,
  onToChange,
  onTriggerTypeChange,
  onStatusChange,
}: DeliveryMonitorFiltersProps) {
  return (
    <div
      className={`${C.bgWhite} border ${C.borderLight} rounded-xs px-4 py-3 flex flex-wrap items-stretch sm:items-center gap-3`}
    >
      <label
        className={`text-sm ${C.text70} flex w-full flex-col items-stretch gap-2 sm:w-auto sm:flex-row sm:items-center`}
      >
        期間
        <DateRangeInputs
          fromValue={from}
          toValue={to}
          onFromChange={onFromChange}
          onToChange={onToChange}
          fromTestId="filter-from"
          toTestId="filter-to"
          className="w-full flex-col items-stretch sm:w-auto sm:flex-row sm:items-center"
          inputClassName={`h-11 w-full px-2 text-sm border ${C.borderMedium} rounded-xs ${C.bgWhite} ${C.text} sm:w-auto`}
        />
      </label>

      <Select value={triggerType || "all"} onValueChange={onTriggerTypeChange}>
        <SelectTrigger
          aria-label="配信トリガー種別"
          className={`h-11 w-full sm:w-[220px] ${C.borderMedium} ${C.text} ${C.bgWhite} text-sm`}
          data-testid="filter-trigger-type"
        >
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">すべての種別</SelectItem>
          {Object.entries(TriggerTypeLabels).map(([value, label]) => (
            <SelectItem key={value} value={value}>
              {label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      <Select value={statusFilter || "all"} onValueChange={onStatusChange}>
        <SelectTrigger
          aria-label="配信ステータス"
          className={`h-11 w-full sm:w-[140px] ${C.borderMedium} ${C.text} ${C.bgWhite} text-sm`}
          data-testid="filter-status"
        >
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">すべてのステータス</SelectItem>
          {Object.entries(TriggerStatusLabels).map(([value, label]) => (
            <SelectItem key={value} value={value}>
              {label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  );
}

interface DeliverySummaryCardsProps {
  summary: DeliveryTriggerSummaryResponse | undefined;
}

export function DeliverySummaryCards({ summary }: DeliverySummaryCardsProps) {
  if (summary === undefined) {
    return null;
  }

  return (
    <div className="grid grid-cols-1 gap-3 sm:grid-cols-3 xl:grid-cols-5">
      {DELIVERY_STATUS_CARDS.map(({ key, label }) => (
        <div
          key={key}
          className={`${C.bgWhite} border ${C.borderLight} rounded-xs px-4 py-4`}
          data-testid={`summary-card-${key}`}
        >
          <p className={`text-xs ${C.text50} mb-1`}>{label}</p>
          <p className={`text-heading-2 font-bold ${C.text}`}>
            {summary[key].toLocaleString("ja-JP")}
          </p>
        </div>
      ))}
    </div>
  );
}

export function DeliveryFailedWarning({ summary }: DeliverySummaryCardsProps) {
  if (summary === undefined || summary.failed === 0) {
    return null;
  }

  return (
    <div
      role="alert"
      className={`flex items-center gap-2 ${C.bgDanger8} border ${C.borderDanger20} rounded-xs px-4 py-3`}
      data-testid="failed-warning-banner"
    >
      <AlertTriangle className={`${ICON.sm} ${C.danger} shrink-0`} />
      <p className={`text-sm ${C.danger}`}>
        <strong>{summary.failed.toLocaleString("ja-JP")}件</strong>
        の配信が失敗しています。ログを確認してください。
      </p>
    </div>
  );
}

export function DeliveryExcludedReasonBreakdown({ summary }: DeliverySummaryCardsProps) {
  if (
    summary === undefined ||
    summary.excluded === 0 ||
    Object.keys(summary.excluded_reason_breakdown).length === 0
  ) {
    return null;
  }

  return (
    <div
      className={`${C.bgWhite} border ${C.borderLight} rounded-xs px-4 py-3`}
      data-testid="excluded-reason-breakdown"
    >
      <p className={`text-sm font-medium ${C.text70} mb-2`}>除外理由</p>
      <div className="flex flex-wrap gap-2">
        {Object.entries(summary.excluded_reason_breakdown).map(([reason, count]) => (
          <span
            key={reason}
            className={`inline-flex items-center gap-1 px-2 py-0.5 rounded border text-xs ${BADGE.gray}`}
          >
            {reason || "理由なし"}
            <span className="font-medium">{count.toLocaleString("ja-JP")}件</span>
          </span>
        ))}
      </div>
    </div>
  );
}
