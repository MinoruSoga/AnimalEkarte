import type { ChangeEvent, ReactNode } from "react";
import { Download, Printer, Settings } from "lucide-react";
import { Link } from "react-router";
import { C } from "@/lib/design-tokens";
import { paths } from "@/config/paths";
import { MonthlySummaryCards } from "../components/MonthlySummaryCards";
import { DailyBreakdownTable } from "../components/DailyBreakdownTable";
import { MonthlyReportPrintArea } from "../components/MonthlyReportPrintArea";
import { CategoryPaymentMatrixTable } from "../components/CategoryPaymentMatrixTable";
import type { MonthlyReportResponse } from "../api/get-monthly-report";

type ReportMode = "month" | "period";

interface AccountingReportsHeaderActionsProps {
  hasData: boolean;
  isExporting: boolean;
  onPrint: () => void;
  onExport: () => void;
}

export function AccountingReportsHeaderActions({
  hasData,
  isExporting,
  onPrint,
  onExport,
}: AccountingReportsHeaderActionsProps) {
  return (
    <div className="flex flex-wrap items-center gap-2 print:hidden" data-testid="report-actions">
      {hasData ? (
        <button
          type="button"
          onClick={onPrint}
          data-testid="monthly-report-print-button"
          className={`flex min-h-11 items-center gap-2 px-4 text-base rounded-xs ${C.bgWhite} border ${C.borderMedium} ${C.text} ${C.hoverBgLight} transition-colors`}
        >
          <Printer className="size-4" />
          印刷 / PDF出力
        </button>
      ) : null}
      <button
        type="button"
        onClick={onExport}
        disabled={isExporting}
        className={`flex min-h-11 items-center gap-2 px-4 text-base rounded-xs ${C.bgWhite} border ${C.borderMedium} ${C.text} ${C.hoverBgLight} transition-colors disabled:opacity-50`}
      >
        <Download className="size-4" />
        {isExporting ? "エクスポート中..." : "CSV出力"}
      </button>
    </div>
  );
}

interface AccountingReportsPeriodControlsProps {
  reportMode: ReportMode;
  year: number;
  month: number;
  startDate: string;
  endDate: string;
  yearOptions: number[];
  onModeChange: (e: ChangeEvent<HTMLSelectElement>) => void;
  onYearChange: (e: ChangeEvent<HTMLSelectElement>) => void;
  onMonthChange: (e: ChangeEvent<HTMLSelectElement>) => void;
  onStartDateChange: (e: ChangeEvent<HTMLInputElement>) => void;
  onEndDateChange: (e: ChangeEvent<HTMLInputElement>) => void;
}

export function AccountingReportsPeriodControls({
  reportMode,
  year,
  month,
  startDate,
  endDate,
  yearOptions,
  onModeChange,
  onYearChange,
  onMonthChange,
  onStartDateChange,
  onEndDateChange,
}: AccountingReportsPeriodControlsProps) {
  return (
    <div className="flex flex-wrap gap-3 items-center print:hidden">
      <div>
        <label htmlFor="report_mode" className={`text-base ${C.text70} mr-2`}>
          集計単位
        </label>
        <select
          id="report_mode"
          value={reportMode}
          onChange={onModeChange}
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
              onChange={onYearChange}
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
              onChange={onMonthChange}
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
              onChange={onStartDateChange}
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
              onChange={onEndDateChange}
              className={`min-h-11 text-base ${C.bgWhite} border ${C.borderMedium} ${C.text} rounded-xs px-3`}
            />
          </div>
        </>
      )}
    </div>
  );
}

interface AccountingReportsResultsProps {
  data: MonthlyReportResponse;
  periodLabel: string;
  clinicName: string;
  standardTaxRate: number;
  reducedTaxRate: number;
  canViewCloses: boolean;
  canViewClinicSettings: boolean;
  onDrillDown: (date: string) => void;
}

export function AccountingReportsResults({
  data,
  periodLabel,
  clinicName,
  standardTaxRate,
  reducedTaxRate,
  canViewCloses,
  canViewClinicSettings,
  onDrillDown,
}: AccountingReportsResultsProps) {
  const taxSettingsLink: ReactNode = canViewClinicSettings ? (
    <Link
      to={paths.settings.clinic.getHref()}
      className={`inline-flex min-h-11 items-center gap-1 text-sm ${C.text60} ${C.hoverText} underline-offset-2 hover:underline`}
    >
      <Settings className="size-3.5" />
      税率設定を変更
    </Link>
  ) : undefined;

  return (
    <>
      <MonthlyReportPrintArea
        periodLabel={periodLabel}
        clinicName={clinicName}
        summary={data.summary}
        dailyDetails={data.dailyDetails}
        standardTaxRate={standardTaxRate}
        reducedTaxRate={reducedTaxRate}
        categoryPaymentMatrix={data.categoryPaymentMatrix}
      />

      <section className={`${C.bgWhite} rounded-lg border ${C.borderLight} p-6`}>
        <h2 className={`text-xl font-semibold ${C.text} mb-4`}>
          {periodLabel ? `${periodLabel} の日次明細` : "日次明細"}
        </h2>
        <DailyBreakdownTable
          details={data.dailyDetails}
          onDrillDown={canViewCloses ? onDrillDown : undefined}
        />
      </section>

      {data.categoryPaymentMatrix ? (
        <section
          className={`${C.bgWhite} rounded-lg border ${C.borderLight} p-6`}
          data-testid="category-payment-matrix"
        >
          <h2 className={`text-xl font-semibold ${C.text} mb-4`}>部門×支払方法</h2>
          <CategoryPaymentMatrixTable matrix={data.categoryPaymentMatrix} />
        </section>
      ) : null}

      <MonthlySummaryCards
        summary={data.summary}
        standardTaxRate={standardTaxRate}
        reducedTaxRate={reducedTaxRate}
        taxSettingsLink={taxSettingsLink}
      />
    </>
  );
}
