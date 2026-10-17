import { memo } from "react";
import { EmptyState } from "@/components/shared/DataStates";
import { DataTableRowButton } from "@/components/shared/DataTable/DataTableRowButton";
import { TableCell, TableHead } from "@/components/ui/table";
import { C, STYLE } from "@/lib/design-tokens";
import { formatCurrency } from "@/lib/format/number";
import { formatDate } from "@/lib/format/date";
import type { DailyReportDetail } from "../api/get-monthly-report";

interface DailyBreakdownTableProps {
  details: DailyReportDetail[];
  /**
   * 締めのある日次行の日付ボタンから締め履歴へドリルダウンさせるコールバック。
   * 省略時（締め閲覧権限なし等）は日付を非インタラクティブにする。
   */
  onDrillDown?: (date: string) => void;
}

export const DailyBreakdownTable = memo(function DailyBreakdownTable({
  details,
  onDrillDown,
}: DailyBreakdownTableProps) {
  if (details.length === 0) {
    return <EmptyState message="日次データがありません" />;
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-base">
        <thead>
          <tr className={`border-b ${C.borderLight} ${C.bgPage}`}>
            <TableHead className={C.text55}>日付</TableHead>
            <TableHead className={C.text55}>曜日</TableHead>
            <TableHead className={`text-right ${C.text55}`}>会計件数</TableHead>
            <TableHead className={`text-right ${C.text55}`}>日計</TableHead>
            <TableHead className={`text-right ${C.text55}`}>返金</TableHead>
            <TableHead className={`text-center ${C.text55}`}>AM締</TableHead>
            <TableHead className={`text-center ${C.text55}`}>PM締</TableHead>
          </tr>
        </thead>
        <tbody>
          {details.map((detail) => {
            const isDrillable = !!onDrillDown && (detail.amClosed || detail.pmClosed);

            return (
              <tr
                key={detail.date}
                className={`border-b ${C.borderLight} ${detail.isHoliday ? C.bgNotice40 : STYLE.tableRow}`}
              >
                <TableCell className={`${C.text} whitespace-nowrap`}>
                  {isDrillable ? (
                    <DataTableRowButton
                      aria-label={`${detail.date ? formatDate(detail.date) : ""} の締め詳細を表示`}
                      onClick={() => onDrillDown?.(detail.date)}
                    >
                      {detail.date ? formatDate(detail.date) : "—"}
                    </DataTableRowButton>
                  ) : detail.date ? (
                    formatDate(detail.date)
                  ) : (
                    "—"
                  )}
                </TableCell>
                <TableCell className={C.text}>{detail.weekday}</TableCell>
                <TableCell className={`text-right whitespace-nowrap ${C.text60}`}>
                  {(detail.amCount + detail.pmCount).toLocaleString()}件
                </TableCell>
                <TableCell className={`text-right font-medium whitespace-nowrap ${C.text}`}>
                  {formatCurrency(detail.dayNet)}
                </TableCell>
                <TableCell
                  className={`text-right whitespace-nowrap ${detail.refund > 0 ? C.danger : C.text50}`}
                >
                  {detail.refund > 0 ? formatCurrency(-detail.refund) : "—"}
                </TableCell>
                <TableCell className={`text-center ${detail.amClosed ? C.text : C.text50}`}>
                  {detail.amClosed ? "締め済" : "未締め"}
                </TableCell>
                <TableCell className={`text-center ${detail.pmClosed ? C.text : C.text50}`}>
                  {detail.pmClosed ? "締め済" : "未締め"}
                </TableCell>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
});
