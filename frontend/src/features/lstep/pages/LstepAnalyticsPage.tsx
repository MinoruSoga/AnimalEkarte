// React/Framework
import { useActionState, useRef, useState } from "react";

// External
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from "recharts";
import { UploadCloud } from "lucide-react";

// Internal
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { C, ICON, PALETTE } from "@/lib/design-tokens";
import { handleApiError } from "@/lib/handle-api-error";
import { axios } from "@/lib/axios";
import { useQueryClient } from "@tanstack/react-query";

// Relative
import { useGetLstepDeliveryStats } from "../api/get-lstep-delivery-stats";
import { useGetLstepCsvImports } from "../api/get-lstep-csv-imports";
import { TriggerTypeLabels } from "../constants/trigger-types";

// ── Helpers ──────────────────────────────────────────────────────────

function currentYearMonth(): string {
  const d = new Date();
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}`;
}

function generateMonthOptions(count = 12): { value: string; label: string }[] {
  const options: { value: string; label: string }[] = [];
  const now = new Date();
  for (let i = 0; i < count; i++) {
    const d = new Date(now.getFullYear(), now.getMonth() - i, 1);
    const value = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}`;
    const label = `${d.getFullYear()}年${d.getMonth() + 1}月`;
    options.push({ value, label });
  }
  return options;
}

const STATS_STATUSES = ["fired", "excluded", "failed", "scheduled"] as const;
type StatsStatus = (typeof STATS_STATUSES)[number];

const STATUS_LABELS: Record<StatsStatus, string> = {
  fired: "送信済",
  excluded: "除外",
  failed: "失敗",
  scheduled: "予定",
};

const STATUS_COLORS: Record<StatsStatus, string> = {
  fired: PALETTE.successGreen,
  excluded: PALETTE.chartAxisText,
  failed: PALETTE.danger,
  scheduled: PALETTE.accent,
};

const CSV_STATUS_LABELS: Record<string, string> = {
  pending: "処理待ち",
  processing: "処理中",
  completed: "完了",
  failed: "失敗",
};

interface CrossRow {
  trigger_type: string;
  fired: number;
  excluded: number;
  failed: number;
  scheduled: number;
  total: number;
}

function buildCrossRows(
  rows: { trigger_type: string; status: string; count: number }[]
): CrossRow[] {
  const map = new Map<string, CrossRow>();
  for (const row of rows) {
    if (!map.has(row.trigger_type)) {
      map.set(row.trigger_type, {
        trigger_type: row.trigger_type,
        fired: 0,
        excluded: 0,
        failed: 0,
        scheduled: 0,
        total: 0,
      });
    }
    const entry = map.get(row.trigger_type)!;
    const status = row.status as StatsStatus;
    if (status in entry) {
      (entry[status] as number) += Number(row.count);
    }
    entry.total += Number(row.count);
  }
  return Array.from(map.values()).sort((a, b) => b.total - a.total);
}

// ── Sub-components ────────────────────────────────────────────────────

interface StatsTableProps {
  rows: CrossRow[];
}

function DeliveryStatsTable({ rows }: StatsTableProps) {
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
            <th className={`text-left px-3 py-2 font-medium ${C.text80} min-w-[180px]`}>
              トリガー種別
            </th>
            {STATS_STATUSES.map((s) => (
              <th key={s} className={`text-right px-3 py-2 font-medium ${C.text80}`}>
                {STATUS_LABELS[s]}
              </th>
            ))}
            <th className={`text-right px-3 py-2 font-medium ${C.text80}`}>合計</th>
          </tr>
        </thead>
        <tbody>
          {rows.map((row) => (
            <tr
              key={row.trigger_type}
              className={`border-b ${C.borderLight} hover:${C.bgLight} transition-colors`}
            >
              <td className={`px-3 py-2 ${C.text80}`}>
                {TriggerTypeLabels[row.trigger_type] ?? row.trigger_type}
              </td>
              {STATS_STATUSES.map((s) => (
                <td key={s} className={`text-right px-3 py-2 ${C.text60} tabular-nums`}>
                  {row[s].toLocaleString()}
                </td>
              ))}
              <td className={`text-right px-3 py-2 font-medium ${C.text80} tabular-nums`}>
                {row.total.toLocaleString()}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function DeliveryStatsChart({ rows }: StatsTableProps) {
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
      <p className={`text-sm font-medium ${C.text80} mb-3`}>
        トリガー別配信内訳
      </p>
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
          <Legend
            iconType="circle"
            iconSize={8}
            wrapperStyle={{ fontSize: 12, paddingTop: 8 }}
          />
          <Bar dataKey="送信済" stackId="a" fill={STATUS_COLORS.fired} />
          <Bar dataKey="除外" stackId="a" fill={STATUS_COLORS.excluded} />
          <Bar dataKey="失敗" stackId="a" fill={STATUS_COLORS.failed} />
          <Bar dataKey="予定" stackId="a" fill={STATUS_COLORS.scheduled} />
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
}

// ── CSV Upload Section ────────────────────────────────────────────────

type UploadState = { success: true } | { error: string } | null;

function CsvUploadSection() {
  const queryClient = useQueryClient();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [state, formAction, isPending] = useActionState<UploadState, FormData>(
    async (_prev, _formData) => {
      const file = fileInputRef.current?.files?.[0];
      if (!file || file.size === 0) {
        return { error: "CSVファイルを選択してください" };
      }
      try {
        const clinicId = localStorage.getItem("auth_current_clinic:v1");
        const fd = new FormData();
        fd.append("file", file);
        await axios.post(
          `/v1/clinics/${clinicId}/lstep/csv-imports/friend-attributes`,
          fd,
          { headers: { "Content-Type": "multipart/form-data" } }
        );
        queryClient.invalidateQueries({ queryKey: ["lstep-csv-imports"] });
        return { success: true };
      } catch (err) {
        handleApiError(err, "CSVアップロード");
        return { error: "アップロードに失敗しました" };
      }
    },
    null
  );

  return (
    <form action={formAction} className="space-y-3">
      <div className={`flex items-center gap-3 p-4 border ${C.borderLight} rounded-[4px] ${C.bgLight}`}>
        <UploadCloud className={`${ICON.action} ${C.text40} shrink-0`} />
        <div className="flex-1 min-w-0">
          <p className={`text-sm font-medium ${C.text80}`}>友だち属性 CSV</p>
          <p className={`text-xs ${C.text40}`}>
            Lステップ管理画面からエクスポートした CSV をアップロード
          </p>
        </div>
        <input
          ref={fileInputRef}
          type="file"
          name="file"
          accept=".csv"
          className={`text-sm ${C.text60} file:mr-3 file:px-3 file:py-1.5 file:rounded file:border-0 file:text-xs file:font-medium file:${C.bgWhite} file:border file:${C.borderLight}`}
          aria-label="友だち属性CSVファイルを選択"
        />
      </div>
      {state && "error" in state ? (
        <p className={`text-sm text-[${PALETTE.danger}]`}>{state.error}</p>
      ) : null}
      {state && "success" in state ? (
        <p className={`text-sm text-[${PALETTE.successGreen}]`}>
          アップロードが完了しました
        </p>
      ) : null}
      <div className="flex justify-end">
        <SubmitButton disabled={isPending}>
          {isPending ? "アップロード中..." : "アップロード"}
        </SubmitButton>
      </div>
    </form>
  );
}

// ── CSV History Table ─────────────────────────────────────────────────

function CsvImportHistoryTable() {
  const { data, isLoading } = useGetLstepCsvImports(20);

  if (isLoading) {
    return <p className={`text-sm ${C.text40} py-4`}>読み込み中...</p>;
  }
  if (!data || data.length === 0) {
    return (
      <p className={`text-sm ${C.text40} py-8 text-center`}>
        インポート履歴はありません
      </p>
    );
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm border-collapse">
        <thead>
          <tr className={`${C.bgLight} border-b ${C.borderLight}`}>
            <th className={`text-left px-3 py-2 font-medium ${C.text80}`}>ファイル名</th>
            <th className={`text-right px-3 py-2 font-medium ${C.text80}`}>総行数</th>
            <th className={`text-right px-3 py-2 font-medium ${C.text80}`}>成功</th>
            <th className={`text-right px-3 py-2 font-medium ${C.text80}`}>失敗</th>
            <th className={`text-center px-3 py-2 font-medium ${C.text80}`}>ステータス</th>
            <th className={`text-left px-3 py-2 font-medium ${C.text80}`}>アップロード日時</th>
          </tr>
        </thead>
        <tbody>
          {data.map((item) => (
            <tr key={item.id} className={`border-b ${C.borderLight}`}>
              <td className={`px-3 py-2 ${C.text80} max-w-[240px] truncate`} title={item.file_name}>
                {item.file_name}
              </td>
              <td className={`text-right px-3 py-2 ${C.text60} tabular-nums`}>
                {item.row_count.toLocaleString()}
              </td>
              <td className={`text-right px-3 py-2 tabular-nums`} style={{ color: PALETTE.successGreen }}>
                {item.success_count.toLocaleString()}
              </td>
              <td className={`text-right px-3 py-2 tabular-nums`} style={{ color: item.error_count > 0 ? PALETTE.danger : undefined }}>
                {item.error_count.toLocaleString()}
              </td>
              <td className="text-center px-3 py-2">
                <span className={`text-xs px-2 py-0.5 rounded-full ${
                  item.status === "completed"
                    ? `bg-[${PALETTE.successGreen}]/10 text-[${PALETTE.successGreen}]`
                    : item.status === "failed"
                    ? `bg-[${PALETTE.danger}]/10 text-[${PALETTE.danger}]`
                    : `${C.bgLight} ${C.text60}`
                }`}>
                  {CSV_STATUS_LABELS[item.status] ?? item.status}
                </span>
              </td>
              <td className={`px-3 py-2 ${C.text60} text-xs`}>
                {new Date(item.created_at).toLocaleString("ja-JP", {
                  year: "numeric", month: "2-digit", day: "2-digit",
                  hour: "2-digit", minute: "2-digit",
                })}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

// ── Main Page ─────────────────────────────────────────────────────────

const MONTH_OPTIONS = generateMonthOptions(12);

export function LstepAnalyticsPage() {
  const [yearMonth, setYearMonth] = useState(currentYearMonth());
  const { data, isLoading, isError } = useGetLstepDeliveryStats(yearMonth);

  const crossRows = data ? buildCrossRows(data.rows) : [];

  return (
    <PageLayout
      title="Lステップ分析レポート"
      description="月次配信統計と友だち属性 CSV インポート管理"
    >
      {/* 月次配信統計セクション */}
      <section aria-labelledby="delivery-stats-heading" className="space-y-4">
        <div className="flex items-center justify-between">
          <h2
            id="delivery-stats-heading"
            className={`text-base font-semibold ${C.text80}`}
          >
            月次配信統計
          </h2>
          <select
            value={yearMonth}
            onChange={(e) => setYearMonth(e.target.value)}
            className={`text-sm border ${C.borderLight} rounded-[4px] px-3 py-1.5 ${C.bgWhite} ${C.text80}`}
            aria-label="集計対象年月"
          >
            {MONTH_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
        </div>

        <div className={`border ${C.borderLight} rounded-[4px] ${C.bgWhite} p-4`}>
          {isLoading ? (
            <p className={`text-sm ${C.text40} py-8 text-center`}>読み込み中...</p>
          ) : isError ? (
            <p className={`text-sm text-[${PALETTE.danger}] py-8 text-center`}>
              データの取得に失敗しました
            </p>
          ) : (
            <>
              <DeliveryStatsTable rows={crossRows} />
              <DeliveryStatsChart rows={crossRows} />
            </>
          )}
        </div>
      </section>

      {/* CSV インポート管理セクション */}
      <section
        aria-labelledby="csv-import-heading"
        className="space-y-4 mt-8"
      >
        <h2 id="csv-import-heading" className={`text-base font-semibold ${C.text80}`}>
          友だち属性 CSV インポート
        </h2>

        <div className={`border ${C.borderLight} rounded-[4px] ${C.bgWhite} p-4 space-y-5`}>
          <div>
            <p className={`text-sm font-medium ${C.text80} mb-3`}>新規アップロード</p>
            <CsvUploadSection />
          </div>
          <hr className={C.borderLight} />
          <div>
            <p className={`text-sm font-medium ${C.text80} mb-3`}>インポート履歴</p>
            <CsvImportHistoryTable />
          </div>
        </div>
      </section>
    </PageLayout>
  );
}
