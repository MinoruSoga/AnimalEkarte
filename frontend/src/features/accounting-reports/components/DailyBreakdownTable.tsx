import { memo } from "react";
import { C, STYLE } from "@/lib/design-tokens";
import type { DailyReportDetail } from "@/types/generated/models";

interface DailyBreakdownTableProps {
  details: DailyReportDetail[];
}

export const DailyBreakdownTable = memo(function DailyBreakdownTable({
  details,
}: DailyBreakdownTableProps) {
  if (details.length === 0) {
    return <p className={`text-base ${C.text50} py-4 text-center`}>日次データがありません</p>;
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-base">
        <thead>
          <tr className={`border-b ${C.borderLight} ${C.bgPage}`}>
            <th className={`text-left px-3 py-2 font-medium ${C.text70}`}>日付</th>
            <th className={`text-left px-3 py-2 font-medium ${C.text70}`}>曜日</th>
            <th className={`text-right px-3 py-2 font-medium ${C.text70}`}>午前件数</th>
            <th className={`text-right px-3 py-2 font-medium ${C.text70}`}>午前売上</th>
            <th className={`text-right px-3 py-2 font-medium ${C.text70}`}>午後件数</th>
            <th className={`text-right px-3 py-2 font-medium ${C.text70}`}>午後売上</th>
            <th className={`text-right px-3 py-2 font-medium ${C.text70}`}>日計</th>
            <th className={`text-right px-3 py-2 font-medium ${C.text70}`}>返金</th>
            <th className={`text-center px-3 py-2 font-medium ${C.text70}`}>AM締</th>
            <th className={`text-center px-3 py-2 font-medium ${C.text70}`}>PM締</th>
          </tr>
        </thead>
        <tbody>
          {details.map((detail) => (
            <tr
              key={detail.date}
              className={`border-b ${C.borderLight} ${detail.is_holiday ? C.bgNotice40 : STYLE.tableRow}`}
            >
              <td className={`px-3 py-2 ${C.text}`}>{detail.date}</td>
              <td className={`px-3 py-2 ${C.text}`}>{detail.weekday}</td>
              <td className={`px-3 py-2 text-right ${C.text60}`}>{detail.am_count}件</td>
              <td className={`px-3 py-2 text-right ${C.text}`}>
                ¥{detail.am_net.toLocaleString()}
              </td>
              <td className={`px-3 py-2 text-right ${C.text60}`}>{detail.pm_count}件</td>
              <td className={`px-3 py-2 text-right ${C.text}`}>
                ¥{detail.pm_net.toLocaleString()}
              </td>
              <td className={`px-3 py-2 text-right font-medium ${C.text}`}>
                ¥{detail.day_net.toLocaleString()}
              </td>
              <td
                className={`px-3 py-2 text-right ${detail.refund > 0 ? C.danger : C.text50}`}
              >
                {detail.refund > 0 ? `-¥${detail.refund.toLocaleString()}` : "—"}
              </td>
              <td className="px-3 py-2 text-center">
                <span
                  className={`inline-block size-2 rounded-full ${detail.am_closed ? C.bgStatusGreenDot : C.bgInactive}`}
                />
              </td>
              <td className="px-3 py-2 text-center">
                <span
                  className={`inline-block size-2 rounded-full ${detail.pm_closed ? C.bgStatusGreenDot : C.bgInactive}`}
                />
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
});
