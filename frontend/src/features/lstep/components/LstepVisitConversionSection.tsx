import { C } from "@/lib/design-tokens";
import { TableCell, TableHead } from "@/components/ui/table";

import type { VisitConversionSummaryResponse } from "../api/get-lstep-visit-conversion";
import { TriggerTypeLabels } from "../constants/trigger-types";
import { formatPercent } from "./lstep-analytics-model";

interface VisitConversionSectionProps {
  data?: VisitConversionSummaryResponse;
  isLoading: boolean;
  isError: boolean;
}

export function VisitConversionSection({
  data,
  isLoading,
  isError,
}: VisitConversionSectionProps) {
  return (
    <section aria-labelledby="visit-conversion-heading" className="space-y-4 mt-8">
      <div className="flex items-center justify-between">
        <h2 id="visit-conversion-heading" className={`text-base font-semibold ${C.text80}`}>
          配信後来院率
        </h2>
        <p className={`text-xs ${C.text40}`}>配信日から30日以内の来院を集計</p>
      </div>

      <div className={`border ${C.borderLight} rounded-xs ${C.bgWhite} p-4 space-y-6`}>
        {isLoading ? (
          <p className={`text-sm ${C.text40} py-8 text-center`}>読み込み中...</p>
        ) : isError ? (
          <p className={`text-sm ${C.danger} py-8 text-center`}>
            来院率データの取得に失敗しました
          </p>
        ) : (
          <>
            <VisitConversionCards data={data} />
            <VisitConversionTable data={data} />
          </>
        )}
      </div>
    </section>
  );
}

function VisitConversionCards({ data }: { data?: VisitConversionSummaryResponse }) {
  return (
    <div className="grid gap-3 sm:grid-cols-3">
      <MetricCard label="送信件数" value={data?.delivered_count.toLocaleString() ?? "0"} />
      <MetricCard label="来院件数" value={data?.visited_count.toLocaleString() ?? "0"} />
      <MetricCard label="来院率" value={formatPercent(data?.visit_rate ?? 0)} />
    </div>
  );
}

function MetricCard({ label, value }: { label: string; value: string }) {
  return (
    <div className={`${C.bgWhite} rounded-xs border ${C.borderLight} p-4 flex flex-col gap-1`}>
      <span className={`text-sm ${C.text60}`}>{label}</span>
      <span className={`text-heading-3 font-semibold ${C.text80}`}>{value}</span>
    </div>
  );
}

function VisitConversionTable({ data }: { data?: VisitConversionSummaryResponse }) {
  if (!data || data.rows.length === 0) {
    return (
      <p className={`text-sm ${C.text40} py-6 text-center`}>
        この月の来院率データはありません
      </p>
    );
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm border-collapse">
        <thead>
          <tr className={`${C.bgLight} border-b ${C.borderLight}`}>
            <TableHead className={`${C.text55} min-w-[180px]`}>
              トリガー種別
            </TableHead>
            <TableHead className={`text-right ${C.text55}`}>送信件数</TableHead>
            <TableHead className={`text-right ${C.text55}`}>来院件数</TableHead>
            <TableHead className={`text-right ${C.text55}`}>来院率</TableHead>
          </tr>
        </thead>
        <tbody>
          {data.rows.map((row) => (
            <tr key={row.trigger_type} className={`border-b ${C.borderLight}`}>
              <TableCell className={C.text80}>
                {TriggerTypeLabels[row.trigger_type] ?? row.trigger_type}
              </TableCell>
              <TableCell className={`text-right ${C.text60} tabular-nums`}>
                {row.delivered_count.toLocaleString()}
              </TableCell>
              <TableCell className={`text-right ${C.text60} tabular-nums`}>
                {row.visited_count.toLocaleString()}
              </TableCell>
              <TableCell className={`text-right font-semibold ${C.text55} tabular-nums`}>
                {formatPercent(row.visit_rate)}
              </TableCell>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
