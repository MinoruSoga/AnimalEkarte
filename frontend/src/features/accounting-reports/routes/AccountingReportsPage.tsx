import { useCallback, useState, useTransition } from "react";
import { useNavigate } from "react-router";
import { BarChart3 } from "lucide-react";
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
import { formatJSTWallDate, toJSTWallDate } from "@/lib/jst-date";
import {
  AccountingReportsHeaderActions,
  AccountingReportsPeriodControls,
  AccountingReportsResults,
} from "./AccountingReportsPagePanels";

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
  const { standardTaxRate, reducedTaxRate } = useClinicTaxRates();

  const reportParams: MonthlyReportParams =
    reportMode === "month" ? { mode: "month", year, month } : { mode: "period", startDate, endDate };
  const { data, isLoading, isError } = useGetMonthlyReport(reportParams);

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
        <AccountingReportsHeaderActions
          hasData={!!data}
          isExporting={isExporting}
          onPrint={() => window.print()}
          onExport={handleExport}
        />
      }
    >
      <div className="space-y-6">
        <AccountingReportsPeriodControls
          reportMode={reportMode}
          year={year}
          month={month}
          startDate={startDate}
          endDate={endDate}
          yearOptions={yearOptions}
          onModeChange={(e) => {
            if (isReportMode(e.target.value)) {
              setReportMode(e.target.value);
            }
          }}
          onYearChange={(e) => setYear(Number(e.target.value))}
          onMonthChange={(e) => setMonth(Number(e.target.value))}
          onStartDateChange={(e) => setStartDate(e.target.value)}
          onEndDateChange={(e) => setEndDate(e.target.value)}
        />

        {isLoading ? (
          <div className="flex items-center justify-center py-12">
            <p className={`text-base ${C.text50}`}>読み込み中...</p>
          </div>
        ) : isError || !data ? (
          <div className="flex items-center justify-center py-12">
            <p className={`text-base ${C.danger}`}>データの取得に失敗しました</p>
          </div>
        ) : (
          <AccountingReportsResults
            data={data}
            periodLabel={periodLabel}
            clinicName={clinicName}
            standardTaxRate={standardTaxRate}
            reducedTaxRate={reducedTaxRate}
            canViewCloses={canViewCloses}
            canViewClinicSettings={canViewClinicSettings}
            onDrillDown={handleDrillDown}
          />
        )}
      </div>
    </PageLayout>
  );
}
