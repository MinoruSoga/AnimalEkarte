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

interface BackendDailyReportDetail {
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

/** #247 DEC-16⑥: 部門×支払方法マトリクス（支払実額・会計 distinct） */
interface BackendCategoryPaymentMethodColumn {
  id?: number;
  name: string;
  is_active: boolean;
}

interface BackendCategoryPaymentMatrixRow {
  category: string;
  count: number;
  by_method: Record<string, number>;
  row_total: number;
}

interface BackendCategoryPaymentMatrixTotals {
  count: number;
  by_method: Record<string, number>;
  grand_total: number;
}

interface BackendCategoryPaymentMatrix {
  payment_methods: BackendCategoryPaymentMethodColumn[];
  rows: BackendCategoryPaymentMatrixRow[];
  totals: BackendCategoryPaymentMatrixTotals;
}

interface BackendMonthlyReportResponse {
  year: number;
  month: number;
  start_date: string;
  end_date: string;
  summary: BackendMonthlyReportSummary;
  daily_details: BackendDailyReportDetail[];
  category_payment_matrix?: BackendCategoryPaymentMatrix | null;
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

function transformDailyReportDetail(raw: BackendDailyReportDetail) {
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

function transformCategoryPaymentMatrix(
  raw: BackendCategoryPaymentMatrix | null | undefined,
): CategoryPaymentMatrix | null {
  if (!raw) return null;
  return {
    paymentMethods: (raw.payment_methods ?? []).map((m) => ({
      id: m.id,
      name: m.name,
      isActive: m.is_active,
    })),
    rows: (raw.rows ?? []).map((r) => ({
      category: r.category,
      count: r.count,
      byMethod: r.by_method ?? {},
      rowTotal: r.row_total,
    })),
    totals: {
      count: raw.totals?.count ?? 0,
      byMethod: raw.totals?.by_method ?? {},
      grandTotal: raw.totals?.grand_total ?? 0,
    },
  };
}

function transformMonthlyReport(raw: BackendMonthlyReportResponse) {
  return {
    year: raw.year,
    month: raw.month,
    startDate: raw.start_date,
    endDate: raw.end_date,
    summary: transformMonthlyReportSummary(raw.summary),
    dailyDetails: raw.daily_details.map(transformDailyReportDetail),
    categoryPaymentMatrix: transformCategoryPaymentMatrix(raw.category_payment_matrix),
  };
}

// ── Exported matrix types (pre-transform helper needs forward name) ──────────

export interface CategoryPaymentMatrix {
  paymentMethods: { id?: number; name: string; isActive: boolean }[];
  rows: {
    category: string;
    count: number;
    byMethod: Record<string, number>;
    rowTotal: number;
  }[];
  totals: {
    count: number;
    byMethod: Record<string, number>;
    grandTotal: number;
  };
}

// ── Derived types (FA9 compliant) ────────────────────────────────────────────

export type DailyReportDetail = ReturnType<typeof transformDailyReportDetail>;
export type MonthlyReportResponse = ReturnType<typeof transformMonthlyReport>;

export type MonthlyReportParams =
  | { mode: "month"; year: number; month: number }
  | { mode: "period"; startDate: string; endDate: string };

function toRequestParams(params: MonthlyReportParams) {
  if (params.mode === "month") {
    return { year: params.year, month: params.month };
  }
  return { start_date: params.startDate, end_date: params.endDate };
}

function toQueryKey(params: MonthlyReportParams) {
  if (params.mode === "month") {
    return ["monthly-report", "month", params.year, params.month] as const;
  }
  return ["monthly-report", "period", params.startDate, params.endDate] as const;
}

// ── API functions ────────────────────────────────────────────────────────────

const getMonthlyReport = async (params: MonthlyReportParams): Promise<MonthlyReportResponse> => {
  const { data } = await axios.get<BackendMonthlyReportResponse>("/v1/reports/monthly", {
    params: toRequestParams(params),
  });
  return transformMonthlyReport(data);
};

export const useGetMonthlyReport = (params: MonthlyReportParams) =>
  useQuery({
    queryKey: toQueryKey(params),
    queryFn: () => getMonthlyReport(params),
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
