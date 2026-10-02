/**
 * BugReportsPage — バグ報告の一覧・対応管理ページ（管理者向け）
 *
 * /settings/bug-reports（hospital-settings 権限、settings-routes でゲート）。
 * 件名セルの詳細ボタンで詳細ダイアログ（スクリーンショット・画面文脈・ステータス切替）。
 */
import { useState } from "react";
import { Bug, ExternalLink, Trash2 } from "lucide-react";
import type { UseMutationResult } from "@tanstack/react-query";

import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { EmptyState, ErrorFallback, LoadingFallback } from "@/components/shared/DataStates";
import { DataTableRowButton } from "@/components/shared/DataTable/DataTableRowButton";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { BADGE, C, ICON, STYLE } from "@/lib/design-tokens";
import { formatJSTDate, formatJSTTime } from "@/lib/jst-date";

import { usePermission } from "@/hooks/use-permission";
import { useGetBugReports } from "../api/get-bug-reports";
import { useCreatePlaneTicket } from "../api/create-plane-ticket";
import { useDeleteBugReport } from "../api/delete-bug-report";
import { useUpdateBugReportStatus } from "../api/update-bug-report-status";
import type { BugReport, BugReportStatus } from "../types";

const STATUS_LABEL: Record<BugReportStatus, string> = {
  open: "未対応",
  resolved: "対応済み",
};

const STATUS_BADGE: Record<BugReportStatus, string> = {
  open: BADGE.yellow,
  resolved: BADGE.green,
};

function StatusBadge({ status }: { status: BugReportStatus }) {
  return (
    <span
      className={`inline-flex items-center rounded-xxs border px-1.5 py-0.5 text-2xs ${STATUS_BADGE[status]}`}
    >
      {STATUS_LABEL[status]}
    </span>
  );
}

type PlaneTicketMutation = UseMutationResult<BugReport, unknown, number, unknown>;

interface PlaneTicketCellProps {
  report: BugReport;
  canEdit: boolean;
  mutation: PlaneTicketMutation;
}

/**
 * Plane 連携状態。
 * 起票済み → チケットへの外部リンク / 直近失敗 → 失敗表示 + 再送ボタン（edit 権限）/ 未連携 → ―
 */
function PlaneTicketCell({ report, canEdit, mutation }: PlaneTicketCellProps) {
  if (report.plane_issue_url) {
    return (
      <a
        href={report.plane_issue_url}
        target="_blank"
        rel="noopener noreferrer"
        className={`inline-flex items-center gap-1 underline ${C.textActionPrimary} hover:opacity-70`}
      >
        <ExternalLink className={ICON.xs} />
        チケット
      </a>
    );
  }
  if (report.plane_sync_error) {
    return (
      <span className="inline-flex items-center gap-1.5">
        <span className={`text-2xs ${C.danger}`} title={report.plane_sync_error}>
          起票失敗
        </span>
        {canEdit ? (
          <Button
            type="button"
            variant="outline"
            size="sm"
            className="h-6 px-2 text-2xs"
            onClick={() => mutation.mutate(report.id)}
            disabled={mutation.isPending}
          >
            再送
          </Button>
        ) : null}
      </span>
    );
  }
  return <span className={C.text50}>―</span>;
}

interface BugReportDetailDialogProps {
  report: BugReport | null;
  canEdit: boolean;
  onClose: () => void;
}

function BugReportDetailDialog({ report, canEdit, onClose }: BugReportDetailDialogProps) {
  const updateStatus = useUpdateBugReportStatus();
  const createTicket = useCreatePlaneTicket();
  const nextStatus: BugReportStatus | null =
    report === null ? null : report.status === "open" ? "resolved" : "open";

  const handleToggle = () => {
    if (report === null || nextStatus === null) return;
    updateStatus.mutate({ id: report.id, status: nextStatus }, { onSuccess: onClose });
  };

  return (
    <Dialog open={report !== null} onOpenChange={(open) => (open ? undefined : onClose())}>
      {report !== null ? (
        <DialogContent className="sm:max-w-2xl max-h-[85vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle className={`flex items-center gap-2 text-base ${C.text}`}>
              {report.title}
              <StatusBadge status={report.status} />
            </DialogTitle>
            <DialogDescription className="text-left">
              {report.reporter_name || `スタッフ#${report.reporter_staff_id}`} ・{" "}
              {formatJSTDate(report.created_at)} {formatJSTTime(report.created_at)}
            </DialogDescription>
          </DialogHeader>

          <dl className="grid grid-cols-[96px_1fr] gap-x-3 gap-y-1.5 text-sm">
            <dt className={C.text50}>詳細</dt>
            <dd className={`whitespace-pre-wrap break-words ${C.text}`}>
              {report.detail || "（記載なし）"}
            </dd>
            <dt className={C.text50}>画面</dt>
            <dd className={`break-all ${C.text70}`}>
              {report.route_path || report.page_url || "―"}
            </dd>
            <dt className={C.text50}>ページURL</dt>
            <dd className={`break-all ${C.text70}`}>{report.page_url || "―"}</dd>
            <dt className={C.text50}>表示領域</dt>
            <dd className={C.text70}>{report.viewport || "―"}</dd>
            <dt className={C.text50}>UA</dt>
            <dd className={`break-all text-2xs ${C.text60}`}>{report.user_agent || "―"}</dd>
            <dt className={C.text50}>Plane</dt>
            <dd className={C.text70}>
              <PlaneTicketCell report={report} canEdit={canEdit} mutation={createTicket} />
            </dd>
          </dl>

          {report.screenshot_url ? (
            <figure className={`rounded-xxs border ${C.borderLight} overflow-hidden`}>
              <img
                src={report.screenshot_url}
                alt={`「${report.title}」の添付スクリーンショット`}
                className={`w-full object-contain max-h-[320px] ${C.bgPrimary5}`}
              />
            </figure>
          ) : (
            <p className={`text-sm ${C.text60}`}>スクリーンショットは添付されていません</p>
          )}

          <div className="flex justify-end gap-2 pt-1">
            <Button variant="outline" size="sm" onClick={onClose}>
              閉じる
            </Button>
            <Button
              size="sm"
              variant={nextStatus === "open" ? "outline" : "default"}
              onClick={handleToggle}
              disabled={updateStatus.isPending}
            >
              {nextStatus === "resolved" ? "対応済みにする" : "未対応に戻す"}
            </Button>
          </div>
        </DialogContent>
      ) : (
        <DialogContent className="sm:max-w-2xl" aria-label="バグ報告の詳細" />
      )}
    </Dialog>
  );
}

export function BugReportsPage() {
  const { data: reports, isLoading, isError } = useGetBugReports();
  const { canEdit, canDelete } = usePermission("hospital-settings");
  const createTicket = useCreatePlaneTicket();
  const deleteReport = useDeleteBugReport();
  const [selected, setSelected] = useState<BugReport | null>(null);
  const [pendingDelete, setPendingDelete] = useState<BugReport | null>(null);

  return (
    <PageLayout
      title="バグ報告"
      description="サポートウィジェットから送信されたバグ報告の一覧です"
      icon={<Bug className={`${ICON.page} ${C.text}`} />}
      resource="hospital-settings"
      maxWidth="max-w-5xl"
    >
      {isLoading ? (
        <LoadingFallback />
      ) : isError ? (
        <ErrorFallback message="バグ報告の取得に失敗しました。時間をおいて再度お試しください。" />
      ) : reports === undefined || reports.length === 0 ? (
        <EmptyState message="バグ報告はまだありません" />
      ) : (
        <div className={STYLE.tableContainer}>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[130px]">日時</TableHead>
                <TableHead className="w-[110px]">報告者</TableHead>
                <TableHead>件名</TableHead>
                <TableHead className="w-[160px]">画面</TableHead>
                <TableHead className="w-[90px]">スクショ</TableHead>
                <TableHead className="w-[90px]">状態</TableHead>
                <TableHead className="w-[120px]">Plane</TableHead>
                {canDelete ? <TableHead className="w-[64px]" /> : null}
              </TableRow>
            </TableHeader>
            <TableBody>
              {reports.map((report) => (
                <TableRow key={report.id} className={`${C.hoverBgPageHalf} h-14`}>
                  <TableCell className={C.text70}>
                    {formatJSTDate(report.created_at)} {formatJSTTime(report.created_at)}
                  </TableCell>
                  <TableCell className={C.text70}>
                    {report.reporter_name || `#${report.reporter_staff_id}`}
                  </TableCell>
                  <TableCell>
                    <DataTableRowButton
                      aria-label={`詳細: ${report.title}`}
                      onClick={() => setSelected(report)}
                    >
                      <span className="line-clamp-2">{report.title}</span>
                    </DataTableRowButton>
                  </TableCell>
                  <TableCell className={`${C.text50} break-all`}>
                    {report.route_path || "―"}
                  </TableCell>
                  <TableCell className={C.text50}>{report.screenshot_url ? "あり" : "―"}</TableCell>
                  <TableCell>
                    <StatusBadge status={report.status} />
                  </TableCell>
                  <TableCell>
                    <PlaneTicketCell report={report} canEdit={canEdit} mutation={createTicket} />
                  </TableCell>
                  {canDelete ? (
                    <TableCell>
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        className={`h-7 px-2 ${C.danger}`}
                        aria-label={`削除: ${report.title}`}
                        onClick={() => setPendingDelete(report)}
                      >
                        <Trash2 className={ICON.xs} />
                      </Button>
                    </TableCell>
                  ) : null}
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}

      <BugReportDetailDialog
        report={selected}
        canEdit={canEdit}
        onClose={() => setSelected(null)}
      />

      <ConfirmDialog
        open={pendingDelete !== null}
        onClose={() => setPendingDelete(null)}
        onConfirm={() => {
          const target = pendingDelete;
          if (target !== null) {
            deleteReport.mutate(target.id, {
              onSuccess: () => setPendingDelete(null),
            });
          }
        }}
        title={
          pendingDelete !== null ? `「${pendingDelete.title}」を削除しますか？` : "削除しますか？"
        }
        description="報告と添付スクリーンショットが削除されます。Plane に作成済みのチケットは残ります。この操作は取り消せません。"
        confirmLabel="削除する"
        variant="destructive"
        isPending={deleteReport.isPending}
      />
    </PageLayout>
  );
}
