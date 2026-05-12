import { useState, useCallback } from "react";
import { AlertTriangle, RefreshCw } from "lucide-react";
import { useQueryClient } from "@tanstack/react-query";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { BADGE, C, ICON, STYLE } from "@/lib/design-tokens";
import { useGetLstepDeliveryTriggerSummary } from "../api/get-lstep-delivery-trigger-summary";
import { useGetLstepDeliveryTriggerLogs } from "../api/get-lstep-delivery-trigger-logs";
import {
  TriggerTypeLabels,
  TriggerStatusLabels,
  TriggerStatusBadge,
} from "../constants/trigger-types";

function todayStr(): string {
  return new Date().toISOString().slice(0, 10);
}

function formatDatetime(iso: string): string {
  const d = new Date(iso);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")} ${String(d.getHours()).padStart(2, "0")}:${String(d.getMinutes()).padStart(2, "0")}`;
}

export function LstepDeliveryMonitorPage() {
  const queryClient = useQueryClient();

  const [from, setFrom] = useState(todayStr());
  const [to, setTo] = useState(todayStr());
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

  const handleRefresh = useCallback(() => {
    queryClient.invalidateQueries({
      queryKey: ["lstep-delivery-trigger-summary"],
    });
    queryClient.invalidateQueries({
      queryKey: ["lstep-delivery-trigger-logs"],
    });
  }, [queryClient]);

  const handleFromChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      setFrom(e.target.value);
      setPage(1);
    },
    []
  );

  const handleToChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      setTo(e.target.value);
      setPage(1);
    },
    []
  );

  const handleTriggerTypeChange = useCallback((value: string) => {
    setTriggerType(value === "all" ? "" : value);
    setPage(1);
  }, []);

  const handleStatusChange = useCallback((value: string) => {
    setStatusFilter(value === "all" ? "" : value);
    setPage(1);
  }, []);

  const isLoading = summaryLoading || logsLoading;
  const totalPages = logsPage
    ? Math.max(1, Math.ceil(logsPage.total / logsPage.per_page))
    : 1;

  const statusCards: Array<{
    key:
      | "scheduled"
      | "fired"
      | "excluded"
      | "failed"
      | "suppressed_by_priority";
    label: string;
  }> = [
    { key: "scheduled", label: "予定" },
    { key: "fired", label: "送信済" },
    { key: "excluded", label: "除外" },
    { key: "failed", label: "失敗" },
    { key: "suppressed_by_priority", label: "優先度抑制" },
  ];

  return (
    <PageLayout
      title="自動配信トリガー監視"
      maxWidth="max-w-full"
      headerAction={
        <Button
          variant="outline"
          className={STYLE.btnOutline}
          onClick={handleRefresh}
          disabled={isLoading}
        >
          <RefreshCw
            className={`mr-1.5 ${ICON.sm} ${isLoading ? "animate-spin" : ""}`}
          />
          更新
        </Button>
      }
    >
      <div className="flex flex-col gap-4 flex-1 min-h-0">
        {/* フィルター */}
        <div
          className={`bg-white border ${C.borderLight} rounded-[4px] px-4 py-3 flex flex-wrap items-center gap-3`}
        >
          <label className={`text-sm ${C.text70} flex items-center gap-2`}>
            期間
            <input
              type="date"
              value={from}
              onChange={handleFromChange}
              className={`h-9 px-2 text-sm border ${C.borderMedium} rounded-[4px] bg-white ${C.text}`}
              data-testid="filter-from"
            />
            <span className={C.text50}>〜</span>
            <input
              type="date"
              value={to}
              onChange={handleToChange}
              className={`h-9 px-2 text-sm border ${C.borderMedium} rounded-[4px] bg-white ${C.text}`}
              data-testid="filter-to"
            />
          </label>

          <Select
            value={triggerType || "all"}
            onValueChange={handleTriggerTypeChange}
          >
            <SelectTrigger
              className={`h-9 w-[220px] ${C.borderMedium} ${C.text} bg-white text-sm`}
              data-testid="filter-trigger-type"
            >
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">すべての種別</SelectItem>
              {Object.entries(TriggerTypeLabels).map(([val, label]) => (
                <SelectItem key={val} value={val}>
                  {label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          <Select
            value={statusFilter || "all"}
            onValueChange={handleStatusChange}
          >
            <SelectTrigger
              className={`h-9 w-[140px] ${C.borderMedium} ${C.text} bg-white text-sm`}
              data-testid="filter-status"
            >
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">すべてのステータス</SelectItem>
              {Object.entries(TriggerStatusLabels).map(([val, label]) => (
                <SelectItem key={val} value={val}>
                  {label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {/* サマリーカード */}
        {summary !== undefined ? (
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 xl:grid-cols-5">
            {statusCards.map(({ key, label }) => (
              <div
                key={key}
                className={`bg-white border ${C.borderLight} rounded-[4px] px-5 py-4`}
                data-testid={`summary-card-${key}`}
              >
                <p className={`text-xs ${C.text50} mb-1`}>{label}</p>
                <p className={`text-3xl font-bold ${C.text}`}>
                  {summary[key].toLocaleString("ja-JP")}
                </p>
              </div>
            ))}
          </div>
        ) : null}

        {/* 失敗件数警告バナー */}
        {summary !== undefined && summary.failed > 0 ? (
          <div
            className={`flex items-center gap-2 bg-red-50 border border-red-200 rounded-[4px] px-4 py-3`}
            data-testid="failed-warning-banner"
          >
            <AlertTriangle className={`${ICON.sm} text-red-600 shrink-0`} />
            <p className="text-sm text-red-700">
              <strong>{summary.failed.toLocaleString("ja-JP")}件</strong>
              の配信が失敗しています。ログを確認してください。
            </p>
          </div>
        ) : null}

        {/* 除外理由内訳 */}
        {summary !== undefined &&
        summary.excluded > 0 &&
        Object.keys(summary.excluded_reason_breakdown).length > 0 ? (
          <div
            className={`bg-white border ${C.borderLight} rounded-[4px] px-4 py-3`}
            data-testid="excluded-reason-breakdown"
          >
            <p className={`text-sm font-medium ${C.text70} mb-2`}>除外理由</p>
            <div className="flex flex-wrap gap-2">
              {Object.entries(summary.excluded_reason_breakdown).map(
                ([reason, count]) => (
                  <span
                    key={reason}
                    className={`inline-flex items-center gap-1 px-2 py-0.5 rounded border text-xs ${BADGE.gray}`}
                  >
                    {reason || "理由なし"}
                    <span className="font-medium">
                      {count.toLocaleString("ja-JP")}件
                    </span>
                  </span>
                )
              )}
            </div>
          </div>
        ) : null}

        {/* ログテーブル */}
        <div
          className={`bg-white border ${C.borderLight} rounded-[4px] flex flex-col flex-1 min-h-0`}
        >
          <div className={`${STYLE.tableContainer} flex-1 overflow-auto`}>
            <table className="w-full text-sm border-collapse">
              <thead>
                <tr className={STYLE.tableHeaderRow}>
                  <th className={`${STYLE.tableHeaderCell} text-left w-48`}>
                    種別
                  </th>
                  <th className={`${STYLE.tableHeaderCell} text-left w-36`}>
                    飼い主
                  </th>
                  <th className={`${STYLE.tableHeaderCell} text-left w-40`}>
                    予定日時
                  </th>
                  <th className={`${STYLE.tableHeaderCell} text-left w-24`}>
                    ステータス
                  </th>
                  <th className={`${STYLE.tableHeaderCell} text-left w-40`}>
                    送信日時
                  </th>
                  <th className={`${STYLE.tableHeaderCell} text-left`}>
                    除外理由
                  </th>
                </tr>
              </thead>
              <tbody>
                {logsLoading ? (
                  <tr>
                    <td
                      colSpan={6}
                      className={STYLE.tableEmpty}
                      data-testid="logs-loading"
                    >
                      読み込み中...
                    </td>
                  </tr>
                ) : logsPage?.items.length === 0 ? (
                  <tr>
                    <td
                      colSpan={6}
                      className={STYLE.tableEmpty}
                      data-testid="logs-empty"
                    >
                      ログがありません
                    </td>
                  </tr>
                ) : (
                  logsPage?.items.map((item) => {
                    const badgeColor =
                      TriggerStatusBadge[item.status] ?? "gray";
                    return (
                      <tr
                        key={item.id}
                        className={STYLE.tableRow}
                        data-testid="log-row"
                      >
                        <td className={STYLE.tableCell}>
                          {TriggerTypeLabels[item.trigger_type] ??
                            item.trigger_type}
                        </td>
                        <td className={STYLE.tableCell}>{item.owner_name}</td>
                        <td className={`${STYLE.tableCell} font-mono`}>
                          {formatDatetime(item.scheduled_at)}
                        </td>
                        <td className={STYLE.tableCell}>
                          <span
                            className={`inline-flex items-center px-2 py-0.5 rounded border text-xs ${BADGE[badgeColor]}`}
                          >
                            {TriggerStatusLabels[item.status] ?? item.status}
                          </span>
                        </td>
                        <td className={`${STYLE.tableCell} font-mono`}>
                          {item.fired_at ? formatDatetime(item.fired_at) : "—"}
                        </td>
                        <td className={`${STYLE.tableCell} ${C.text50}`}>
                          {item.excluded_reason ?? "—"}
                        </td>
                      </tr>
                    );
                  })
                )}
              </tbody>
            </table>
          </div>

          {/* ページネーション */}
          {logsPage !== undefined && logsPage.total > 0 ? (
            <div
              className={`flex items-center justify-between px-4 py-2 border-t ${C.borderLight}`}
            >
              <p className={`text-sm ${C.text50}`}>
                全 {logsPage.total.toLocaleString("ja-JP")} 件
              </p>
              <div className="flex items-center gap-2">
                <Button
                  variant="outline"
                  className="h-8 px-3 text-sm"
                  disabled={page <= 1}
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
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
                  onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                  data-testid="pagination-next"
                >
                  次へ
                </Button>
              </div>
            </div>
          ) : null}
        </div>
      </div>
    </PageLayout>
  );
}
