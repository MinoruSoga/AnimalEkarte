import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { PaymentMethodMaster } from "@/types/generated/models";
import type { CashRegisterPeriod } from "../lib/constants";

// ── Backend (raw) types ──────────────────────────────────────────────────────

interface BackendCashRegisterTaxBreakdownEntry {
  taxable_amount: number;
  tax_amount: number;
}

interface BackendCashRegisterTaxBreakdown {
  standard: BackendCashRegisterTaxBreakdownEntry;
  reduced: BackendCashRegisterTaxBreakdownEntry;
}

interface BackendCashRegisterAggregateSummary {
  categories: Record<string, Record<string, number>>;
  payment_methods: PaymentMethodMaster[];
  theoretical_cash: number;
  tax_breakdown: BackendCashRegisterTaxBreakdown;
  /** DEC-40: category=other 明細を持つ会計の distinct 件数 */
  unclassified_other_count: number;
  /** #247 DEC-16⑥: 部門ごとの会計 distinct 件数 */
  category_counts?: Record<string, number>;
}

interface BackendCloseBillingDetail {
  billing_id: number;
  paid_at: string;
  owner_name: string;
  pet_name: string;
  is_hospitalization: boolean;
  category: string;
  payment_method_id?: number;
  payment_method_name: string;
  billing_amount: number;
  refund_amount: number;
  net_amount: number;
}

interface BackendClosePreviewResult {
  date: string;
  period: string;
  period_start: string;
  period_end: string;
  is_already_closed: boolean;
  is_holiday: boolean;
  aggregate: BackendCashRegisterAggregateSummary;
  billing_details: BackendCloseBillingDetail[];
}

// ── Transform functions ──────────────────────────────────────────────────────

function transformTaxBreakdownEntry(raw: BackendCashRegisterTaxBreakdownEntry) {
  return {
    taxableAmount: raw.taxable_amount,
    taxAmount: raw.tax_amount,
  };
}

function transformAggregateSummary(raw: BackendCashRegisterAggregateSummary) {
  return {
    categories: raw.categories,
    paymentMethods: raw.payment_methods,
    theoreticalCash: raw.theoretical_cash,
    taxBreakdown: {
      standard: transformTaxBreakdownEntry(raw.tax_breakdown.standard),
      reduced: transformTaxBreakdownEntry(raw.tax_breakdown.reduced),
    },
    // DEC-40: 欠落時は 0（preview は常にサーバが返す。履歴 snapshot は category-breakdown 側で扱う）
    unclassifiedOtherCount:
      typeof raw.unclassified_other_count === "number" &&
      Number.isFinite(raw.unclassified_other_count)
        ? raw.unclassified_other_count
        : 0,
    // #247: 会計 distinct 件数（split 二重計上なし）。欠落時は undefined → FE は billingDetails フォールバック
    categoryCounts: raw.category_counts ?? undefined,
  };
}

function transformCloseBillingDetail(raw: BackendCloseBillingDetail) {
  return {
    billingId: raw.billing_id,
    paidAt: raw.paid_at,
    ownerName: raw.owner_name,
    petName: raw.pet_name,
    isHospitalization: raw.is_hospitalization,
    category: raw.category,
    paymentMethodId: raw.payment_method_id,
    paymentMethodName: raw.payment_method_name,
    billingAmount: raw.billing_amount,
    refundAmount: raw.refund_amount,
    netAmount: raw.net_amount,
  };
}

function transformClosePreviewResult(raw: BackendClosePreviewResult) {
  return {
    date: raw.date,
    period: raw.period,
    periodStart: raw.period_start,
    periodEnd: raw.period_end,
    isAlreadyClosed: raw.is_already_closed,
    isHoliday: raw.is_holiday,
    aggregate: transformAggregateSummary(raw.aggregate),
    billingDetails: raw.billing_details.map(transformCloseBillingDetail),
  };
}

// ── Derived types (FA9 compliant) ────────────────────────────────────────────

export type CloseBillingDetail = ReturnType<typeof transformCloseBillingDetail>;
export type ClosePreviewResult = ReturnType<typeof transformClosePreviewResult>;

// ── API functions ────────────────────────────────────────────────────────────

const getCashRegisterPreview = async (
  date: string,
  period: CashRegisterPeriod,
): Promise<ClosePreviewResult> => {
  const { data } = await axios.get<BackendClosePreviewResult>("/v1/cash-register/preview", {
    params: { date, period },
  });
  return transformClosePreviewResult(data);
};

export const useGetCashRegisterPreview = (
  date: string,
  period: CashRegisterPeriod,
  enabled: boolean,
  refreshNonce = 0,
) =>
  useQuery({
    queryKey: queryKeys.cashRegister.preview.byDatePeriodRefresh(date, period, refreshNonce),
    queryFn: () => getCashRegisterPreview(date, period),
    enabled: enabled && !!date,
    staleTime: QUERY_STALE_TIMES.REALTIME,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
