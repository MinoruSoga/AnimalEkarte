import { useState, useTransition } from "react";
import { Download } from "lucide-react";
import { toast } from "sonner";
import { C } from "@/lib/design-tokens";
import { handleApiError } from "@/lib/handle-api-error";
import { useGetMonthlyReport } from "../api/get-monthly-report";
import { exportMonthlyCSV } from "../api/export-monthly-csv";
import { MonthlySummaryCards } from "../components/MonthlySummaryCards";
import { DailyBreakdownTable } from "../components/DailyBreakdownTable";

export function AccountingReportsPage() {
  const now = new Date();
  const [year, setYear] = useState<number>(now.getFullYear());
  const [month, setMonth] = useState<number>(now.getMonth() + 1);
  const [isExporting, startExportTransition] = useTransition();

  const { data, isLoading, isError } = useGetMonthlyReport(year, month);

  const yearOptions = Array.from({ length: 5 }, (_, i) => now.getFullYear() - i);

  const handleYearChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    setYear(Number(e.target.value));
  };

  const handleMonthChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    setMonth(Number(e.target.value));
  };

  const handleExport = () => {
    startExportTransition(async () => {
      try {
        await exportMonthlyCSV(year, month);
        toast.success("CSVをダウンロードしました");
      } catch (error) {
        handleApiError(error, "CSVエクスポート");
      }
    });
  };

  return (
    <div className="flex-1 overflow-y-auto px-6 py-6 space-y-6">
      <div className="flex items-center justify-between flex-wrap gap-3">
        <h1 className={`text-xl font-bold ${C.text}`}>月次集計レポート</h1>
        <button
          type="button"
          onClick={handleExport}
          disabled={isExporting}
          className={`flex items-center gap-2 px-4 h-10 text-base rounded-[4px] ${C.bgWhite} border ${C.borderMedium} ${C.text} ${C.hoverBgLight} transition-colors disabled:opacity-50`}
        >
          <Download className="size-4" />
          {isExporting ? "エクスポート中..." : "CSV出力"}
        </button>
      </div>

      {/* 年月選択 */}
      <div className="flex flex-wrap gap-3 items-center">
        <div>
          <label htmlFor="report_year" className={`text-base ${C.text70} mr-2`}>
            年
          </label>
          <select
            id="report_year"
            value={year}
            onChange={handleYearChange}
            className={`h-10 text-base ${C.bgWhite} border ${C.borderMedium} ${C.text} rounded-[4px] px-3`}
          >
            {yearOptions.map((y) => (
              <option key={y} value={y}>
                {y}年
              </option>
            ))}
          </select>
        </div>
        <div>
          <label htmlFor="report_month" className={`text-base ${C.text70} mr-2`}>
            月
          </label>
          <select
            id="report_month"
            value={month}
            onChange={handleMonthChange}
            className={`h-10 text-base ${C.bgWhite} border ${C.borderMedium} ${C.text} rounded-[4px] px-3`}
          >
            {Array.from({ length: 12 }, (_, i) => i + 1).map((m) => (
              <option key={m} value={m}>
                {m}月
              </option>
            ))}
          </select>
        </div>
      </div>

      {isLoading ? (
        <div className="flex items-center justify-center py-12">
          <p className={`text-base ${C.text50}`}>読み込み中...</p>
        </div>
      ) : isError || !data ? (
        <div className="flex items-center justify-center py-12">
          <p className={`text-base ${C.danger}`}>データの取得に失敗しました</p>
        </div>
      ) : (
        <>
          <MonthlySummaryCards summary={data.summary} />

          {/* 支払方法別集計 */}
          {Object.keys(data.summary.byPaymentMethod).length > 0 ? (
            <section className={`${C.bgWhite} rounded-lg border ${C.borderLight} p-6`}>
              <h2 className={`text-base font-semibold ${C.text} mb-4`}>支払方法別集計</h2>
              <div className="flex flex-wrap gap-4">
                {Object.entries(data.summary.byPaymentMethod).map(([method, amount]) => (
                  <div key={method} className={`flex flex-col gap-0.5`}>
                    <span className={`text-base ${C.text60}`}>{method}</span>
                    <span className={`text-base font-semibold ${C.text}`}>
                      ¥{amount.toLocaleString()}
                    </span>
                  </div>
                ))}
              </div>
            </section>
          ) : null}

          {/* 消費税内訳 */}
          <section className={`${C.bgWhite} rounded-lg border ${C.borderLight} p-6`}>
            <h2 className={`text-base font-semibold ${C.text} mb-4`}>消費税内訳</h2>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <p className={`text-base ${C.text60} mb-1`}>標準税率 (10%)</p>
                <p className={`text-base ${C.text}`}>
                  課税額: ¥{data.summary.taxBreakdown.standard.taxableAmount.toLocaleString()}
                </p>
                <p className={`text-base ${C.text}`}>
                  税額: ¥{data.summary.taxBreakdown.standard.taxAmount.toLocaleString()}
                </p>
              </div>
              <div>
                <p className={`text-base ${C.text60} mb-1`}>軽減税率 (8%)</p>
                <p className={`text-base ${C.text}`}>
                  課税額: ¥{data.summary.taxBreakdown.reduced.taxableAmount.toLocaleString()}
                </p>
                <p className={`text-base ${C.text}`}>
                  税額: ¥{data.summary.taxBreakdown.reduced.taxAmount.toLocaleString()}
                </p>
              </div>
            </div>
          </section>

          {/* 日次明細 */}
          <section className={`${C.bgWhite} rounded-lg border ${C.borderLight} p-6`}>
            <h2 className={`text-base font-semibold ${C.text} mb-4`}>日次明細</h2>
            <DailyBreakdownTable details={data.dailyDetails} />
          </section>
        </>
      )}
    </div>
  );
}
