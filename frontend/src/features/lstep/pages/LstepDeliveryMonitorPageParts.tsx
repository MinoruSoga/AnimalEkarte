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
import {
  TriggerStatusLabels,
  TriggerTypeLabels,
} from "../constants/trigger-types";
import { DELIVERY_STATUS_CARDS } from "./LstepDeliveryMonitorPageModel";

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
    <div className={`bg-white border ${C.borderLight} rounded-[4px] px-4 py-3 flex flex-wrap items-center gap-3`}>
      <label className={`text-sm ${C.text70} flex items-center gap-2`}>
        期間
        <DateRangeInputs
          fromValue={from}
          toValue={to}
          onFromChange={onFromChange}
          onToChange={onToChange}
          fromTestId="filter-from"
          toTestId="filter-to"
          inputClassName={`h-9 w-auto px-2 text-sm border ${C.borderMedium} rounded-[4px] bg-white ${C.text}`}
        />
      </label>

      <Select value={triggerType || "all"} onValueChange={onTriggerTypeChange}>
        <SelectTrigger
          className={`h-9 w-[220px] ${C.borderMedium} ${C.text} bg-white text-sm`}
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
          className={`h-9 w-[140px] ${C.borderMedium} ${C.text} bg-white text-sm`}
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
    <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 xl:grid-cols-5">
      {DELIVERY_STATUS_CARDS.map(({ key, label }) => (
        <div
          key={key}
          className={`bg-white border ${C.borderLight} rounded-[4px] px-5 py-4`}
          data-testid={`summary-card-${key}`}
        >
          <p className={`text-xs ${C.text50} mb-1`}>{label}</p>
          <p className={`text-3xl font-bold ${C.text}`}>{summary[key].toLocaleString("ja-JP")}</p>
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
      className="flex items-center gap-2 bg-red-50 border border-red-200 rounded-[4px] px-4 py-3"
      data-testid="failed-warning-banner"
    >
      <AlertTriangle className={`${ICON.sm} text-red-600 shrink-0`} />
      <p className="text-sm text-red-700">
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
    <div className={`bg-white border ${C.borderLight} rounded-[4px] px-4 py-3`} data-testid="excluded-reason-breakdown">
      <p className={`text-sm font-medium ${C.text70} mb-2`}>除外理由</p>
      <div className="flex flex-wrap gap-2">
        {Object.entries(summary.excluded_reason_breakdown).map(([reason, count]) => (
          <span key={reason} className={`inline-flex items-center gap-1 px-2 py-0.5 rounded border text-xs ${BADGE.gray}`}>
            {reason || "理由なし"}
            <span className="font-medium">{count.toLocaleString("ja-JP")}件</span>
          </span>
        ))}
      </div>
    </div>
  );
}
