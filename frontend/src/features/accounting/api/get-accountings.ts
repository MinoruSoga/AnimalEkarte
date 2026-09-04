import { useQuery, keepPreviousData } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { transformToAccounting } from "./transforms";
import type { Accounting } from "./transforms";
import type { BackendAccounting } from "./types";

interface AccountingsListResponse {
  data: BackendAccounting[];
  total: number;
  page: number;
  limit: number;
}

/** select フィルタ条件（PropertyFilter と同値）。 */
export type AccountingSelectFilterOp = "is" | "is_not" | "is_empty" | "is_not_empty";

export interface AccountingFilters {
  startDate?: string; // YYYY-MM-DD
  endDate?: string; // YYYY-MM-DD
  /**
   * 飼主 ID で会計履歴を絞り込む（飼主詳細の会計履歴セクションで使用）。
   * 数値文字列を想定。空文字は無視される。
   * Backend: GET /api/v1/accountings?owner_id=...
   */
  ownerId?: string;
  /**
   * 飼主名・ペット名の部分一致検索（サーバサイド・かな正規化）。
   * Backend: GET /api/v1/accountings?search=...
   */
  search?: string;
  /**
   * ステータス（waiting / pending / completed / cancelled）。
   * Backend: status + optional status_op (is | is_not)
   */
  status?: string;
  statusOp?: AccountingSelectFilterOp;
  /**
   * 支払方法（cash / credit_card / ...）。
   * Backend: payment_method + payment_method_op (is | is_not | is_empty | is_not_empty)
   */
  paymentMethod?: string;
  paymentMethodOp?: AccountingSelectFilterOp;
  /** 拠点横断表示 (#86 段階3): 2件以上の場合に clinic_ids クエリパラメータとして送信する。 */
  clinicIds?: string[];
}

function buildAccountingParams(filters?: AccountingFilters): Record<string, string> {
  const params: Record<string, string> = {};
  if (filters?.startDate) params.start_date = filters.startDate;
  if (filters?.endDate) params.end_date = filters.endDate;
  if (filters?.ownerId) params.owner_id = filters.ownerId;
  if (filters?.search?.trim()) params.search = filters.search.trim();
  if (filters?.status?.trim()) {
    params.status = filters.status.trim();
    if (filters.statusOp && filters.statusOp !== "is") {
      params.status_op = filters.statusOp;
    }
  }
  if (filters?.paymentMethodOp === "is_empty" || filters?.paymentMethodOp === "is_not_empty") {
    params.payment_method_op = filters.paymentMethodOp;
  } else if (filters?.paymentMethod?.trim()) {
    params.payment_method = filters.paymentMethod.trim();
    if (filters.paymentMethodOp && filters.paymentMethodOp !== "is") {
      params.payment_method_op = filters.paymentMethodOp;
    }
  }
  if (filters?.clinicIds && filters.clinicIds.length > 1) {
    params.clinic_ids = filters.clinicIds.join(",");
  }
  return params;
}

const getAccountings = async (filters?: AccountingFilters): Promise<Accounting[]> => {
  const { data } = await axios.get<AccountingsListResponse>("/v1/accountings", {
    params: buildAccountingParams(filters),
  });
  return data.data.map(transformToAccounting);
};

export const useGetAccountings = (filters?: AccountingFilters) => {
  return useQuery({
    queryKey: queryKeys.accountings.list(filters),
    queryFn: () => getAccountings(filters),
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
    // ownerId フィルタが空文字の場合はクエリを実行しない（誤って全件取得しないため）
    enabled: filters?.ownerId === undefined || filters.ownerId !== "",
  });
};

export interface AccountingsPage {
  data: Accounting[];
  total: number;
  page: number;
  limit: number;
}

export interface AccountingPageFilters extends AccountingFilters {
  page: number;
  limit: number;
}

const getAccountingsPage = async (filters: AccountingPageFilters): Promise<AccountingsPage> => {
  const params = buildAccountingParams(filters);
  params.page = String(filters.page);
  params.limit = String(filters.limit);
  const { data } = await axios.get<AccountingsListResponse>("/v1/accountings", { params });
  return {
    data: data.data.map(transformToAccounting),
    total: data.total,
    page: data.page,
    limit: data.limit,
  };
};

/**
 * BUG-411: サーバサイドページネーション版。AccountingList 専用。
 * useGetAccountings（DailyAccountingTab/OwnerAccountingHistory が使う全件系）とは
 * queryKey の filters shape が異なる（page/limit を含む）ため自然に別キャッシュエントリになる。
 */
export const useGetAccountingsPage = (filters: AccountingPageFilters) => {
  return useQuery({
    queryKey: queryKeys.accountings.list(filters),
    queryFn: () => getAccountingsPage(filters),
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
    placeholderData: keepPreviousData,
  });
};
