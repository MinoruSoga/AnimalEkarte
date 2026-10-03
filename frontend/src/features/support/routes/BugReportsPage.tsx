/**
 * BugReportsPage — バグ報告の一覧・対応管理ページ（全スタッフ・全医院に公開）
 *
 * /settings/bug-reports（権限ゲートなし — バグ報告は全医院共有の製品フィードバック
 * 基盤として意図的に開放。backend も同じ方針）。一覧は全医院の報告を新しい順で返し、
 * 医院列で provenance を識別できる。
 * 件名セルの詳細ボタンで詳細ダイアログ（スクリーンショット・画面文脈・ステータス切替）。
 * status 切替・Plane 再送・削除は自医院の報告のみ — 他医院の報告は閲覧専用
 * （backend も報告元 clinic_id スコープで 404 を返す。2026-10 セキュリティレビュー変更）。
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
import { useAuth } from "@/hooks/use-auth";

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
  mutation: PlaneTicketMutation;
  /** 自医院の報告のみ再送可能（他医院の報告は閲覧のみ — backend は 404） */
  canMutate: boolean;
}

/**
 * Plane 連携状態。
 * 起票済み → チケットへの外部リンク / 直近失敗 → 失敗表示 + 再送ボタン / 未連携 → ―
 */
function PlaneTicketCell({ report, mutation, canMutate }: PlaneTicketCellProps) {
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
        {canMutate ? (
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
  onClose: () => void;
  /** 自医院の報告のみ status 切替・再送を表示（他医院の報告は閲覧専用） */
  canMutate: boolean;
}

function BugReportDetailDialog({ report, onClose, canMutate }: BugReportDetailDialogProps) {
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
            <dt className={C.text50}>医院</dt>
            <dd className={C.text70}>{report.clinic_name || "―"}</dd>
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
              <PlaneTicketCell report={report} mutation={createTicket} canMutate={canMutate} />
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
            {canMutate ? (
              <Button
                size="sm"
                variant={nextStatus === "open" ? "outline" : "default"}
                onClick={handleToggle}
                disabled={updateStatus.isPending}
              >
                {nextStatus === "resolved" ? "対応済みにする" : "未対応に戻す"}
              </Button>
            ) : null}
          </div>
        </DialogContent>
      ) : (
        <DialogContent className="sm:max-w-2xl" aria-label="バグ報告の詳細" />
      )}
    </Dialog>
  );
}

export function BugReportsPage() {
  const { currentClinicId } = useAuth();
  const { data: reports, isLoading, isError } = useGetBugReports();
  const createTicket = useCreatePlaneTicket();
  const deleteReport = useDeleteBugReport();
  const [selected, setSelected] = useState<BugReport | null>(null);
  const [pendingDelete, setPendingDelete] = useState<BugReport | null>(null);

  // 一覧は全医院公開だが、操作は報告元医院スコープ（backend は他医院指定を 404 で拒否）。
  const isOwnClinic = (report: BugReport): boolean => String(report.clinic_id) === currentClinicId;

  return (
    <PageLayout
      title="バグ報告"
      description="サポートウィジェットから送信されたバグ報告の一覧です（全医院の報告を閲覧可。操作は自医院の報告のみ）"
      icon={<Bug className={`${ICON.page} ${C.text}`} />}
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
                <TableHead className="w-[110px]">医院</TableHead>
                <TableHead className="w-[110px]">報告者</TableHead>
                <TableHead>件名</TableHead>
                <TableHead className="w-[160px]">画面</TableHead>
                <TableHead className="w-[90px]">スクショ</TableHead>
                <TableHead className="w-[90px]">状態</TableHead>
                <TableHead className="w-[120px]">Plane</TableHead>
                <TableHead className="w-[64px]" />
              </TableRow>
            </TableHeader>
            <TableBody>
              {reports.map((report) => (
                <TableRow key={report.id} className={`${C.hoverBgPageHalf} h-14`}>
                  <TableCell className={C.text70}>
                    {formatJSTDate(report.created_at)} {formatJSTTime(report.created_at)}
                  </TableCell>
                  <TableCell className={C.text70}>{report.clinic_name || "―"}</TableCell>
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
                    <PlaneTicketCell
                      report={report}
                      mutation={createTicket}
                      canMutate={isOwnClinic(report)}
                    />
                  </TableCell>
                  <TableCell>
                    {isOwnClinic(report) ? (
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
                    ) : null}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}

      <BugReportDetailDialog
        report={selected}
        onClose={() => setSelected(null)}
        canMutate={selected !== null ? isOwnClinic(selected) : false}
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
