import { memo } from "react";
import { EmptyState } from "@/components/shared/DataStates";
import { DataTableRowButton } from "@/components/shared/DataTable/DataTableRowButton";
import { TableCell, TableHead } from "@/components/ui/table";
import { C, STYLE } from "@/lib/design-tokens";
import { formatCurrency } from "@/lib/format/number";
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
            <TableHead className={`text-right ${C.text55}`}>午前件数</TableHead>
            <TableHead className={`text-right ${C.text55}`}>午前売上</TableHead>
            <TableHead className={`text-right ${C.text55}`}>午後件数</TableHead>
            <TableHead className={`text-right ${C.text55}`}>午後売上</TableHead>
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
                <TableCell className={C.text}>
                  {isDrillable ? (
                    <DataTableRowButton
                      aria-label={`${detail.date} の締め詳細を表示`}
                      onClick={() => onDrillDown?.(detail.date)}
                    >
                      {detail.date}
                    </DataTableRowButton>
                  ) : (
                    detail.date
                  )}
                </TableCell>
                <TableCell className={C.text}>{detail.weekday}</TableCell>
                <TableCell className={`text-right ${C.text60}`}>{detail.amCount}件</TableCell>
                <TableCell className={`text-right ${C.text}`}>
                  {formatCurrency(detail.amNet)}
                </TableCell>
                <TableCell className={`text-right ${C.text60}`}>{detail.pmCount}件</TableCell>
                <TableCell className={`text-right ${C.text}`}>
                  {formatCurrency(detail.pmNet)}
                </TableCell>
                <TableCell className={`text-right font-medium ${C.text}`}>
                  {formatCurrency(detail.dayNet)}
                </TableCell>
                <TableCell className={`text-right ${detail.refund > 0 ? C.danger : C.text50}`}>
                  {detail.refund > 0 ? `-${formatCurrency(detail.refund)}` : "—"}
                </TableCell>
                <TableCell className="text-center">
                  <span
                    className={`inline-block size-2 rounded-full ${detail.amClosed ? C.bgStatusGreenDot : C.bgInactive}`}
                  />
                </TableCell>
                <TableCell className="text-center">
                  <span
                    className={`inline-block size-2 rounded-full ${detail.pmClosed ? C.bgStatusGreenDot : C.bgInactive}`}
                  />
                </TableCell>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
});
