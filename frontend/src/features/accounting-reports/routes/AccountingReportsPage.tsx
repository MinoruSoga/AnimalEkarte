import { useCallback, useState, useTransition } from "react";
import { Link, useNavigate } from "react-router";
import { BarChart3, Download, Printer, Settings } from "lucide-react";
import { toast } from "sonner";
import { C, ICON, LAYOUT } from "@/lib/design-tokens";
import { paths } from "@/config/paths";
import { handleApiError } from "@/lib/handle-api-error";
import { usePermission } from "@/hooks/use-permission";
import { useCurrentClinicName } from "@/hooks/use-current-clinic-name";
import { useClinicTaxRates } from "@/hooks/use-clinic-tax-rates";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import {
  ResourceAccountingReports,
  ResourceCashRegisterClose,
  ResourceHospitalSettings,
} from "@/types/generated/models";
import { useGetMonthlyReport, type MonthlyReportParams } from "../api/get-monthly-report";
import { exportMonthlyCSV } from "../api/export-monthly-csv";
import { MonthlySummaryCards } from "../components/MonthlySummaryCards";
import { DailyBreakdownTable } from "../components/DailyBreakdownTable";
import { MonthlyReportPrintArea } from "../components/MonthlyReportPrintArea";
import { CategoryPaymentMatrixTable } from "../components/CategoryPaymentMatrixTable";
import { formatJSTWallDate, toJSTWallDate } from "@/lib/jst-date";

type ReportMode = "month" | "period";

function isReportMode(value: string): value is ReportMode {
  return value === "month" || value === "period";
}

export function AccountingReportsPage() {
  const now = toJSTWallDate(new Date());
  const [year, setYear] = useState<number>(now.getFullYear());
  const [month, setMonth] = useState<number>(now.getMonth() + 1);
  const defaultStartDate = formatJSTWallDate(new Date(now.getFullYear(), now.getMonth(), 1));
  const defaultEndDate = formatJSTWallDate(new Date(now.getFullYear(), now.getMonth() + 1, 0));
  const [reportMode, setReportMode] = useState<ReportMode>("month");
  const [startDate, setStartDate] = useState<string>(defaultStartDate);
  const [endDate, setEndDate] = useState<string>(defaultEndDate);
  const [isExporting, startExportTransition] = useTransition();

  const navigate = useNavigate();
  const { canView: canViewCloses } = usePermission(ResourceCashRegisterClose);
  const { canView: canViewClinicSettings } = usePermission(ResourceHospitalSettings);
  const clinicName = useCurrentClinicName();
  // #179 ②: 税率は病院マスタ設定（明細兼領収書と同一の正本）を参照する
  const { standardTaxRate, reducedTaxRate } = useClinicTaxRates();

  const reportParams: MonthlyReportParams =
    reportMode === "month" ? { mode: "month", year, month } : { mode: "period", startDate, endDate };
  const { data, isLoading, isError } = useGetMonthlyReport(reportParams);

  // 印刷面のラベルは実際に描画するデータ（data）の対象期間に一致させる。
  // 画面状態（year/month/startDate/endDate）は再取得中に data と一時的に乖離しうるため、
  // 印刷タイトルが明細データの対象期間とずれないよう data を正本とする。
  const periodLabel = data
    ? reportMode === "month"
      ? `${data.year}年${data.month}月`
      : `${data.startDate} 〜 ${data.endDate}`
    : "";

  const yearOptions = Array.from({ length: 5 }, (_, i) => now.getFullYear() - i);

  const handleDrillDown = useCallback(
    (date: string) => {
      navigate(`${paths.accounting.closeHistory.getHref()}?date=${encodeURIComponent(date)}`);
    },
    [navigate],
  );

  const handleYearChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    setYear(Number(e.target.value));
  };

  const handleMonthChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    setMonth(Number(e.target.value));
  };

  const handleModeChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    if (isReportMode(e.target.value)) {
      setReportMode(e.target.value);
    }
  };

  const handleStartDateChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setStartDate(e.target.value);
  };

  const handleEndDateChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setEndDate(e.target.value);
  };

  const handleExport = () => {
    startExportTransition(async () => {
      try {
        await exportMonthlyCSV(reportParams);
        toast.success("CSVをダウンロードしました");
      } catch (error) {
        handleApiError(error, "CSVエクスポート");
      }
    });
  };

  return (
    <PageLayout
      title="月次集計レポート"
      resource={ResourceAccountingReports}
      icon={<BarChart3 className={`${ICON.page} ${C.text}`} />}
      maxWidth={LAYOUT.pageContentMaxWidth.full}
      headerAction={
        <div className="flex flex-wrap items-center gap-2 print:hidden" data-testid="report-actions">
          {data ? (
            <button
              type="button"
              onClick={() => window.print()}
              data-testid="monthly-report-print-button"
              className={`flex min-h-11 items-center gap-2 px-4 text-base rounded-xs ${C.bgWhite} border ${C.borderMedium} ${C.text} ${C.hoverBgLight} transition-colors`}
            >
              <Printer className="size-4" />
              印刷 / PDF出力
            </button>
          ) : null}
          <button
            type="button"
            onClick={handleExport}
            disabled={isExporting}
            className={`flex min-h-11 items-center gap-2 px-4 text-base rounded-xs ${C.bgWhite} border ${C.borderMedium} ${C.text} ${C.hoverBgLight} transition-colors disabled:opacity-50`}
          >
            <Download className="size-4" />
            {isExporting ? "エクスポート中..." : "CSV出力"}
          </button>
        </div>
      }
    >
      <div className="space-y-6">
        {/* 年月選択 */}
        <div className="flex flex-wrap gap-3 items-center print:hidden">
          <div>
            <label htmlFor="report_mode" className={`text-base ${C.text70} mr-2`}>
              集計単位
            </label>
            <select
              id="report_mode"
              value={reportMode}
              onChange={handleModeChange}
              className={`min-h-11 text-base ${C.bgWhite} border ${C.borderMedium} ${C.text} rounded-xs px-3`}
            >
              <option value="month">月次</option>
              <option value="period">期間指定</option>
            </select>
          </div>
          {reportMode === "month" ? (
            <>
              <div>
                <label htmlFor="report_year" className={`text-base ${C.text70} mr-2`}>
                  年
                </label>
                <select
                  id="report_year"
                  value={year}
                  onChange={handleYearChange}
                  className={`min-h-11 text-base ${C.bgWhite} border ${C.borderMedium} ${C.text} rounded-xs px-3`}
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
                  className={`min-h-11 text-base ${C.bgWhite} border ${C.borderMedium} ${C.text} rounded-xs px-3`}
                >
                  {Array.from({ length: 12 }, (_, i) => i + 1).map((m) => (
                    <option key={m} value={m}>
                      {m}月
                    </option>
                  ))}
                </select>
              </div>
            </>
          ) : (
            <>
              <div>
                <label htmlFor="report_start_date" className={`text-base ${C.text70} mr-2`}>
                  開始日
                </label>
                <input
                  id="report_start_date"
                  type="date"
                  value={startDate}
                  onChange={handleStartDateChange}
                  className={`min-h-11 text-base ${C.bgWhite} border ${C.borderMedium} ${C.text} rounded-xs px-3`}
                />
              </div>
              <div>
                <label htmlFor="report_end_date" className={`text-base ${C.text70} mr-2`}>
                  終了日
                </label>
                <input
                  id="report_end_date"
                  type="date"
                  value={endDate}
                  onChange={handleEndDateChange}
                  className={`min-h-11 text-base ${C.bgWhite} border ${C.borderMedium} ${C.text} rounded-xs px-3`}
                />
              </div>
            </>
          )}
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
            {/* 印刷 / PDF 出力ビュー（印刷時のみ表示）*/}
            <MonthlyReportPrintArea
              periodLabel={periodLabel}
              clinicName={clinicName}
              summary={data.summary}
              dailyDetails={data.dailyDetails}
              standardTaxRate={standardTaxRate}
              reducedTaxRate={reducedTaxRate}
              categoryPaymentMatrix={data.categoryPaymentMatrix}
            />

            {/* ゾーン2: 日次明細（periodLabel を唯一のスケール強調見出しに） */}
            <section className={`${C.bgWhite} rounded-lg border ${C.borderLight} p-6`}>
              <h2 className={`text-xl font-semibold ${C.text} mb-4`}>
                {periodLabel ? `${periodLabel} の日次明細` : "日次明細"}
              </h2>
              <DailyBreakdownTable
                details={data.dailyDetails}
                onDrillDown={canViewCloses ? handleDrillDown : undefined}
              />
            </section>

            {/* #247: 部門×支払方法統合表（画面・印刷同一データ） */}
            {data.categoryPaymentMatrix ? (
              <section
                className={`${C.bgWhite} rounded-lg border ${C.borderLight} p-6`}
                data-testid="category-payment-matrix"
              >
                <h2 className={`text-xl font-semibold ${C.text} mb-4`}>部門×支払方法</h2>
                <CategoryPaymentMatrixTable matrix={data.categoryPaymentMatrix} />
              </section>
            ) : null}

            {/* ゾーン3: #179 ④-b KPI/内訳は日次明細より下。#179 ② 税率設定導線もここに集約 */}
            <MonthlySummaryCards
              summary={data.summary}
              standardTaxRate={standardTaxRate}
              reducedTaxRate={reducedTaxRate}
              taxSettingsLink={
                canViewClinicSettings ? (
                  <Link
                    to={paths.settings.clinic.getHref()}
                    className={`inline-flex min-h-11 items-center gap-1 text-sm ${C.text60} ${C.hoverText} underline-offset-2 hover:underline`}
                  >
                    <Settings className="size-3.5" />
                    税率設定を変更
                  </Link>
                ) : undefined
              }
            />
          </>
        )}
      </div>
    </PageLayout>
  );
}
