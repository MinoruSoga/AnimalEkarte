import {
  Bar,
  BarChart,
  CartesianGrid,
  Legend,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";

import { TableCell, TableHead } from "@/components/ui/table";
import { C, PALETTE } from "@/lib/design-tokens";

import { TriggerTypeLabels } from "../constants/trigger-types";
import {
  STATUS_COLORS,
  STATUS_LABELS,
  STATS_STATUSES,
  type CrossRow,
} from "./lstep-analytics-model";

interface DeliveryStatsSectionProps {
  yearMonth: string;
  monthOptions: { value: string; label: string }[];
  rows: CrossRow[];
  isLoading: boolean;
  isError: boolean;
  onYearMonthChange: (value: string) => void;
}

export function DeliveryStatsSection({
  yearMonth,
  monthOptions,
  rows,
  isLoading,
  isError,
  onYearMonthChange,
}: DeliveryStatsSectionProps) {
  return (
    <section aria-labelledby="delivery-stats-heading" className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 id="delivery-stats-heading" className={`text-base font-semibold ${C.text80}`}>
          月次配信統計
        </h2>
        <select
          value={yearMonth}
          onChange={(e) => onYearMonthChange(e.target.value)}
          className={`min-h-11 text-sm border ${C.borderLight} rounded-xs px-3 py-1.5 ${C.bgWhite} ${C.text80}`}
          aria-label="集計対象年月"
        >
          {monthOptions.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {opt.label}
            </option>
          ))}
        </select>
      </div>

      <div className={`border ${C.borderLight} rounded-xs ${C.bgWhite} p-4`}>
        {isLoading ? (
          <p className={`text-sm ${C.text40} py-8 text-center`}>読み込み中...</p>
        ) : isError ? (
          <p className={`text-sm ${C.danger} py-8 text-center`}>
            データの取得に失敗しました
          </p>
        ) : (
          <>
            <DeliveryStatsTable rows={rows} />
            <DeliveryStatsChart rows={rows} />
          </>
        )}
      </div>
    </section>
  );
}

function DeliveryStatsTable({ rows }: { rows: CrossRow[] }) {
  if (rows.length === 0) {
    return (
      <p className={`text-sm ${C.text40} py-8 text-center`}>
        この月のデータはありません
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
            {STATS_STATUSES.map((status) => (
              <TableHead key={status} className={`text-right ${C.text55}`}>
                {STATUS_LABELS[status]}
              </TableHead>
            ))}
            <TableHead className={`text-right ${C.text55}`}>合計</TableHead>
          </tr>
        </thead>
        <tbody>
          {rows.map((row) => (
            <tr
              key={row.trigger_type}
              className={`border-b ${C.borderLight} hover:${C.bgLight} transition-colors`}
            >
              <TableCell className={C.text80}>
                {TriggerTypeLabels[row.trigger_type] ?? row.trigger_type}
              </TableCell>
              {STATS_STATUSES.map((status) => (
                <TableCell key={status} className={`text-right ${C.text60} tabular-nums`}>
                  {row[status].toLocaleString()}
                </TableCell>
              ))}
              <TableCell className={`text-right font-semibold ${C.text55} tabular-nums`}>
                {row.total.toLocaleString()}
              </TableCell>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function DeliveryStatsChart({ rows }: { rows: CrossRow[] }) {
  if (rows.length === 0) return null;

  const chartData = rows.map((row) => ({
    name: TriggerTypeLabels[row.trigger_type] ?? row.trigger_type,
    送信済: row.fired,
    除外: row.excluded,
    失敗: row.failed,
    予定: row.scheduled,
  }));

  return (
    <div className="mt-6">
      <p className={`text-sm font-medium ${C.text80} mb-3`}>トリガー別配信内訳</p>
      <ResponsiveContainer width="100%" height={300}>
        <BarChart
          data={chartData}
          margin={{ top: 4, right: 16, left: 0, bottom: 80 }}
          aria-label="トリガー別配信内訳グラフ"
        >
          <title>トリガー別配信内訳</title>
          <CartesianGrid strokeDasharray="3 3" stroke={PALETTE.chartGrid} />
          <XAxis
            dataKey="name"
            tick={{ fontSize: 10, fill: PALETTE.chartAxisText }}
            tickLine={false}
            axisLine={{ stroke: PALETTE.chartGrid }}
            angle={-45}
            textAnchor="end"
            interval={0}
          />
          <YAxis
            tick={{ fontSize: 11, fill: PALETTE.chartAxisText }}
            tickLine={false}
            axisLine={{ stroke: PALETTE.chartGrid }}
            width={40}
          />
          <Tooltip
            contentStyle={{
              fontSize: 12,
              border: `1px solid ${PALETTE.chartGrid}`,
              borderRadius: 4,
            }}
          />
          <Legend iconType="circle" iconSize={8} wrapperStyle={{ fontSize: 12, paddingTop: 8 }} />
          <Bar dataKey="送信済" stackId="a" fill={STATUS_COLORS.fired} />
          <Bar dataKey="除外" stackId="a" fill={STATUS_COLORS.excluded} />
          <Bar dataKey="失敗" stackId="a" fill={STATUS_COLORS.failed} />
          <Bar dataKey="予定" stackId="a" fill={STATUS_COLORS.scheduled} />
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
}
