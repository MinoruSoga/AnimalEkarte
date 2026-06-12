import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
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

export interface AccountingFilters {
  startDate?: string; // YYYY-MM-DD
  endDate?: string;   // YYYY-MM-DD
  /**
   * 飼主 ID で会計履歴を絞り込む（飼主詳細の会計履歴セクションで使用）。
   * 数値文字列を想定。空文字は無視される。
   * Backend: GET /api/v1/accountings?owner_id=...
   */
  ownerId?: string;
  /** 拠点横断表示 (#86 段階3): 2件以上の場合に clinic_ids クエリパラメータとして送信する。 */
  clinicIds?: string[];
}

export const getAccountings = async (
  filters?: AccountingFilters,
): Promise<Accounting[]> => {
  const params: Record<string, string> = {};
  if (filters?.startDate) params.start_date = filters.startDate;
  if (filters?.endDate) params.end_date = filters.endDate;
  if (filters?.ownerId) params.owner_id = filters.ownerId;
  if (filters?.clinicIds && filters.clinicIds.length > 1) params.clinic_ids = filters.clinicIds.join(",");
  const { data } = await axios.get<AccountingsListResponse>("/v1/accountings", { params });
  return data.data.map(transformToAccounting);
};

export const useGetAccountings = (filters?: AccountingFilters) => {
  return useQuery({
    queryKey: ["accountings", filters],
    queryFn: () => getAccountings(filters),
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
    // ownerId フィルタが空文字の場合はクエリを実行しない（誤って全件取得しないため）
    enabled: filters?.ownerId === undefined || filters.ownerId !== "",
  });
};
