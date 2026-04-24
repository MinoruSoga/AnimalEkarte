import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";

// ── Backend (raw) types ──────────────────────────────────────────────────────

interface BackendMonthlyTaxBreakdownEntry {
  taxable_amount: number;
  tax_amount: number;
}

interface BackendMonthlyTaxBreakdownSummary {
  standard: BackendMonthlyTaxBreakdownEntry;
  reduced: BackendMonthlyTaxBreakdownEntry;
}

interface BackendMonthlyReportSummary {
  working_days: number;
  total_billings: number;
  total_amount: number;
  total_refund: number;
  net_amount: number;
  by_payment_method: Record<string, number>;
  by_category: Record<string, number>;
  tax_breakdown: BackendMonthlyTaxBreakdownSummary;
}

export interface BackendDailyReportDetail {
  date: string;
  weekday: string;
  am_count: number;
  am_net: number;
  pm_count: number;
  pm_net: number;
  day_net: number;
  refund: number;
  am_closed: boolean;
  pm_closed: boolean;
  is_holiday: boolean;
}

interface BackendMonthlyReportResponse {
  year: number;
  month: number;
  summary: BackendMonthlyReportSummary;
  daily_details: BackendDailyReportDetail[];
}

// ── Transform functions ──────────────────────────────────────────────────────

function transformTaxBreakdownEntry(raw: BackendMonthlyTaxBreakdownEntry) {
  return {
    taxableAmount: raw.taxable_amount,
    taxAmount: raw.tax_amount,
  };
}

function transformMonthlyReportSummary(raw: BackendMonthlyReportSummary) {
  return {
    workingDays: raw.working_days,
    totalBillings: raw.total_billings,
    totalAmount: raw.total_amount,
    totalRefund: raw.total_refund,
    netAmount: raw.net_amount,
    byPaymentMethod: raw.by_payment_method,
    byCategory: raw.by_category,
    taxBreakdown: {
      standard: transformTaxBreakdownEntry(raw.tax_breakdown.standard),
      reduced: transformTaxBreakdownEntry(raw.tax_breakdown.reduced),
    },
  };
}

export function transformDailyReportDetail(raw: BackendDailyReportDetail) {
  return {
    date: raw.date,
    weekday: raw.weekday,
    amCount: raw.am_count,
    amNet: raw.am_net,
    pmCount: raw.pm_count,
    pmNet: raw.pm_net,
    dayNet: raw.day_net,
    refund: raw.refund,
    amClosed: raw.am_closed,
    pmClosed: raw.pm_closed,
    isHoliday: raw.is_holiday,
  };
}

function transformMonthlyReport(raw: BackendMonthlyReportResponse) {
  return {
    year: raw.year,
    month: raw.month,
    summary: transformMonthlyReportSummary(raw.summary),
    dailyDetails: raw.daily_details.map(transformDailyReportDetail),
  };
}

// ── Derived types (FA9 compliant) ────────────────────────────────────────────

export type DailyReportDetail = ReturnType<typeof transformDailyReportDetail>;
export type MonthlyReportResponse = ReturnType<typeof transformMonthlyReport>;

// ── API functions ────────────────────────────────────────────────────────────

export const getMonthlyReport = async (
  year: number,
  month: number,
): Promise<MonthlyReportResponse> => {
  const { data } = await axios.get<BackendMonthlyReportResponse>("/v1/reports/monthly", {
    params: { year, month },
  });
  return transformMonthlyReport(data);
};

export const useGetMonthlyReport = (year: number, month: number) =>
  useQuery({
    queryKey: ["monthly-report", year, month],
    queryFn: () => getMonthlyReport(year, month),
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
