import { useCallback, useState } from "react";
import { RefreshCw } from "lucide-react";
import { useQueryClient } from "@tanstack/react-query";

import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { Button } from "@/components/ui/button";
import { ICON, STYLE, LAYOUT } from "@/lib/design-tokens";
import { queryKeys } from "@/lib/query-keys";

import { useGetLstepDeliveryTriggerLogs } from "../api/get-lstep-delivery-trigger-logs";
import { useGetLstepDeliveryTriggerSummary } from "../api/get-lstep-delivery-trigger-summary";
import {
  DeliveryExcludedReasonBreakdown,
  DeliveryFailedWarning,
  DeliveryMonitorFilters,
  DeliverySummaryCards,
} from "./LstepDeliveryMonitorPageParts";
import { DeliveryLogsTable } from "./LstepDeliveryMonitorLogsTable";
import { getTodayDateString } from "./lstep-delivery-monitor-page-model";

export function LstepDeliveryMonitorPage() {
  const queryClient = useQueryClient();

  const [from, setFrom] = useState(getTodayDateString);
  const [to, setTo] = useState(getTodayDateString);
  const [triggerType, setTriggerType] = useState("");
  const [statusFilter, setStatusFilter] = useState("");
  const [page, setPage] = useState(1);

  const { data: summary, isLoading: summaryLoading } =
    useGetLstepDeliveryTriggerSummary(from, to, triggerType || undefined);

  const { data: logsPage, isLoading: logsLoading } =
    useGetLstepDeliveryTriggerLogs(
      from,
      to,
      triggerType || undefined,
      statusFilter || undefined,
      page
    );

  const isLoading = summaryLoading || logsLoading;
  const totalPages = logsPage ? Math.max(1, Math.ceil(logsPage.total / logsPage.per_page)) : 1;

  const handleRefresh = useCallback(() => {
    queryClient.invalidateQueries({ queryKey: queryKeys.lstepDeliveryTriggerSummary.all() });
    queryClient.invalidateQueries({ queryKey: queryKeys.lstepDeliveryTriggerLogs.all() });
  }, [queryClient]);

  const handleFromChange = useCallback((event: React.ChangeEvent<HTMLInputElement>) => {
    setFrom(event.target.value);
    setPage(1);
  }, []);

  const handleToChange = useCallback((event: React.ChangeEvent<HTMLInputElement>) => {
    setTo(event.target.value);
    setPage(1);
  }, []);

  const handleTriggerTypeChange = useCallback((value: string) => {
    setTriggerType(value === "all" ? "" : value);
    setPage(1);
  }, []);

  const handleStatusChange = useCallback((value: string) => {
    setStatusFilter(value === "all" ? "" : value);
    setPage(1);
  }, []);

  const handlePreviousPage = useCallback(() => {
    setPage((currentPage) => Math.max(1, currentPage - 1));
  }, []);

  const handleNextPage = useCallback(() => {
    setPage((currentPage) => Math.min(totalPages, currentPage + 1));
  }, [totalPages]);

  return (
    <PageLayout
      title="自動配信トリガー監視"
      maxWidth={LAYOUT.pageContentMaxWidth.full}
      headerAction={
        <Button variant="outline" className={STYLE.btnOutline} onClick={handleRefresh} disabled={isLoading}>
          <RefreshCw className={`mr-1.5 ${ICON.sm} ${isLoading ? "animate-spin" : ""}`} />
          更新
        </Button>
      }
    >
      <div className="flex flex-col gap-4 flex-1 min-h-0">
        <DeliveryMonitorFilters
          from={from}
          to={to}
          triggerType={triggerType}
          statusFilter={statusFilter}
          onFromChange={handleFromChange}
          onToChange={handleToChange}
          onTriggerTypeChange={handleTriggerTypeChange}
          onStatusChange={handleStatusChange}
        />

        <DeliverySummaryCards summary={summary} />
        <DeliveryFailedWarning summary={summary} />
        <DeliveryExcludedReasonBreakdown summary={summary} />
        <DeliveryLogsTable
          logsPage={logsPage}
          logsLoading={logsLoading}
          page={page}
          totalPages={totalPages}
          onPreviousPage={handlePreviousPage}
          onNextPage={handleNextPage}
        />
      </div>
    </PageLayout>
  );
}
