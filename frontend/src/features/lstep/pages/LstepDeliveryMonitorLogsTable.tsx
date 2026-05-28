import { Button } from "@/components/ui/button";
import { BADGE, C, STYLE } from "@/lib/design-tokens";

import type { DeliveryTriggerLogsPageResponse } from "../api/get-lstep-delivery-trigger-logs";
import {
  TriggerStatusBadge,
  TriggerStatusLabels,
  TriggerTypeLabels,
} from "../constants/trigger-types";
import { formatDeliveryMonitorDatetime } from "./LstepDeliveryMonitorPageModel";

interface DeliveryLogsTableProps {
  logsPage: DeliveryTriggerLogsPageResponse | undefined;
  logsLoading: boolean;
  page: number;
  totalPages: number;
  onPreviousPage: () => void;
  onNextPage: () => void;
}

export function DeliveryLogsTable({
  logsPage,
  logsLoading,
  page,
  totalPages,
  onPreviousPage,
  onNextPage,
}: DeliveryLogsTableProps) {
  return (
    <div className={`bg-white border ${C.borderLight} rounded-[4px] flex flex-col flex-1 min-h-0`}>
      <div className={`${STYLE.tableContainer} flex-1 overflow-auto`}>
        <table className="w-full text-sm border-collapse">
          <thead>
            <tr className={STYLE.tableHeaderRow}>
              <th className={`${STYLE.tableHeaderCell} text-left w-48`}>種別</th>
              <th className={`${STYLE.tableHeaderCell} text-left w-36`}>飼い主</th>
              <th className={`${STYLE.tableHeaderCell} text-left w-40`}>予定日時</th>
              <th className={`${STYLE.tableHeaderCell} text-left w-24`}>ステータス</th>
              <th className={`${STYLE.tableHeaderCell} text-left w-40`}>送信日時</th>
              <th className={`${STYLE.tableHeaderCell} text-left`}>除外理由</th>
            </tr>
          </thead>
          <tbody>
            <DeliveryLogsTableBody logsPage={logsPage} logsLoading={logsLoading} />
          </tbody>
        </table>
      </div>

      {logsPage !== undefined && logsPage.total > 0 ? (
        <DeliveryLogsPagination
          logsPage={logsPage}
          page={page}
          totalPages={totalPages}
          onPreviousPage={onPreviousPage}
          onNextPage={onNextPage}
        />
      ) : null}
    </div>
  );
}

function DeliveryLogsTableBody({
  logsPage,
  logsLoading,
}: Pick<DeliveryLogsTableProps, "logsPage" | "logsLoading">) {
  if (logsLoading) {
    return (
      <tr>
        <td colSpan={6} className={STYLE.tableEmpty} data-testid="logs-loading">
          読み込み中...
        </td>
      </tr>
    );
  }

  if (logsPage?.items.length === 0) {
    return (
      <tr>
        <td colSpan={6} className={STYLE.tableEmpty} data-testid="logs-empty">
          ログがありません
        </td>
      </tr>
    );
  }

  return logsPage?.items.map((item) => {
    const badgeColor = TriggerStatusBadge[item.status] ?? "gray";
    return (
      <tr key={item.id} className={STYLE.tableRow} data-testid="log-row">
        <td className={STYLE.tableCell}>{TriggerTypeLabels[item.trigger_type] ?? item.trigger_type}</td>
        <td className={STYLE.tableCell}>{item.owner_name}</td>
        <td className={`${STYLE.tableCell} font-mono`}>{formatDeliveryMonitorDatetime(item.scheduled_at)}</td>
        <td className={STYLE.tableCell}>
          <span className={`inline-flex items-center px-2 py-0.5 rounded border text-xs ${BADGE[badgeColor]}`}>
            {TriggerStatusLabels[item.status] ?? item.status}
          </span>
        </td>
        <td className={`${STYLE.tableCell} font-mono`}>
          {item.fired_at ? formatDeliveryMonitorDatetime(item.fired_at) : "—"}
        </td>
        <td className={`${STYLE.tableCell} ${C.text50}`}>{item.excluded_reason ?? "—"}</td>
      </tr>
    );
  });
}

interface DeliveryLogsPaginationProps {
  logsPage: DeliveryTriggerLogsPageResponse;
  page: number;
  totalPages: number;
  onPreviousPage: () => void;
  onNextPage: () => void;
}

function DeliveryLogsPagination({
  logsPage,
  page,
  totalPages,
  onPreviousPage,
  onNextPage,
}: DeliveryLogsPaginationProps) {
  return (
    <div className={`flex items-center justify-between px-4 py-2 border-t ${C.borderLight}`}>
      <p className={`text-sm ${C.text50}`}>全 {logsPage.total.toLocaleString("ja-JP")} 件</p>
      <div className="flex items-center gap-2">
        <Button
          variant="outline"
          className="h-8 px-3 text-sm"
          disabled={page <= 1}
          onClick={onPreviousPage}
          data-testid="pagination-prev"
        >
          前へ
        </Button>
        <span className={`text-sm ${C.text70}`}>
          {page} / {totalPages}
        </span>
        <Button
          variant="outline"
          className="h-8 px-3 text-sm"
          disabled={page >= totalPages}
          onClick={onNextPage}
          data-testid="pagination-next"
        >
          次へ
        </Button>
      </div>
    </div>
  );
}
