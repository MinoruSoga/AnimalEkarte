import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";

interface BackendOwnerUnpaidBalance {
  unpaid_total: number;
  unpaid_count: number;
}

export interface OwnerUnpaidBalance {
  /** 未入金 billing（status=waiting）の合計金額。 */
  unpaidTotal: number;
  /** 未入金 billing の件数。 */
  unpaidCount: number;
}

/**
 * 飼主の未納残高を取得する（#182）。
 * Backend: GET /v1/accountings/unpaid-balance?owner_id=...
 * 残高定義は未納者一覧(#120)/月次繰越(#114)と同一（waiting 全額 + クレジット訂正差額 — BUG-007）。
 *
 * P2-11 (PR #186 review): 拠点横断の会計詳細から開いた場合、未納残高は
 * グローバル選択クリニックではなく対象会計（accounting.clinicId）のクリニックで
 * 解決しなければならない。X-Clinic-ID ヘッダーを明示指定し、axios interceptor の
 * グローバル値フォールバックを上書きする。
 */
const getOwnerUnpaidBalance = async (
  clinicId: string,
  ownerId: string,
): Promise<OwnerUnpaidBalance> => {
  const { data } = await axios.get<BackendOwnerUnpaidBalance>("/v1/accountings/unpaid-balance", {
    params: { owner_id: ownerId },
    headers: { "X-Clinic-ID": clinicId },
  });
  return {
    unpaidTotal: data.unpaid_total ?? 0,
    unpaidCount: data.unpaid_count ?? 0,
  };
};

export const useGetOwnerUnpaidBalance = (clinicId: string, ownerId: string | undefined) =>
  useQuery({
    queryKey: queryKeys.ownerUnpaidBalance(clinicId, ownerId),
    queryFn: () => getOwnerUnpaidBalance(clinicId, ownerId as string),
    // 未確定（空文字 / "0" / 未指定）の飼主では実行しない
    enabled: ownerId !== undefined && ownerId !== "" && ownerId !== "0",
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
