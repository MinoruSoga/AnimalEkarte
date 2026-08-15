import { Pagination } from "@/components/shared/Pagination";
import { BADGE, C, STYLE } from "@/lib/design-tokens";

import type { DeliveryTriggerLogsPageResponse } from "../api/get-lstep-delivery-trigger-logs";
import {
  TriggerStatusBadge,
  TriggerStatusLabels,
  TriggerTypeLabels,
} from "../constants/trigger-types";
import { formatDeliveryMonitorDatetime } from "./lstep-delivery-monitor-page-model";

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
    <div className={`${C.bgWhite} border ${C.borderLight} rounded-xs flex flex-col flex-1 min-h-0`}>
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
        <DeliveryLogsPaginationBar
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

interface DeliveryLogsPaginationBarProps {
  logsPage: DeliveryTriggerLogsPageResponse;
  page: number;
  totalPages: number;
  onPreviousPage: () => void;
  onNextPage: () => void;
}

function DeliveryLogsPaginationBar({
  logsPage,
  page,
  totalPages,
  onPreviousPage,
  onNextPage,
}: DeliveryLogsPaginationBarProps) {
  const totalCount = logsPage.total;
  const perPage = logsPage.per_page;
  const startIndex = totalCount === 0 ? 0 : (page - 1) * perPage + 1;
  const endIndex = Math.min(page * perPage, totalCount);

  // FE5-15: サーバ側ページネーション（logsPage は現在ページの items のみ保持）のため
  // usePagination（クライアント側の配列全件を必要とする）は使えない。<Pagination> の
  // onPageChange（ページ番号ジャンプ・先頭/末尾ボタン）は既存の onPreviousPage/onNextPage
  // （関数更新型 setPage を親で保持）を目標ページまで多重呼び出しすることで実現し、
  // 親（LstepDeliveryMonitorPage.tsx）の状態管理・API呼び出しロジックには一切触れない。
  const onPageChange = (target: number) => {
    const diff = target - page;
    if (diff > 0) {
      for (let i = 0; i < diff; i++) onNextPage();
    } else if (diff < 0) {
      for (let i = 0; i < -diff; i++) onPreviousPage();
    }
  };

  return (
    <div className={`border-t ${C.borderLight}`}>
      <Pagination
        currentPage={page}
        totalPages={totalPages}
        totalCount={totalCount}
        startIndex={startIndex}
        endIndex={endIndex}
        onPageChange={onPageChange}
        onPrev={onPreviousPage}
        onNext={onNextPage}
      />
    </div>
  );
}
