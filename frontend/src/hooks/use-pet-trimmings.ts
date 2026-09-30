import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { HISTORY_FETCH_LIMIT } from "@/config/fetch-limits";
import { queryKeys } from "@/lib/query-keys";
import { transformTrimming, type TrimmingUI } from "@/lib/transforms/trimming";
import type { TrimmingListResponse } from "@/types/trimming";

const getTrimmingsByPetId = async (petId: string): Promise<TrimmingUI[]> => {
  const { data } = await axios.get<TrimmingListResponse>("/v1/trimmings", {
    params: { pet_id: petId, page: 1, limit: HISTORY_FETCH_LIMIT },
  });
  return data.data.map(transformTrimming);
};

/**
 * ペット単位のトリミング記録一覧を取得する shared query hook。
 * NO32: features/trimming/api/get-trimming.ts から昇格（cross-feature import
 * 禁止のため medical-records の統合タイムラインはここを使う）。queryKey は
 * 昇格前と同じ `trimmings.byPet` でキャッシュを共有する。
 *
 * `options.enabled=false` で取得自体を抑止する（例: trimming view 権限なし。
 * 権限なしで送信すると BE が RequireSelectedClinicGrant("trimming","view") で
 * 403 にするため、既知の権限外では呼ばない）。
 */
export const useGetTrimmingsByPetId = (petId: string, options?: { enabled?: boolean }) => {
  return useQuery({
    queryKey: queryKeys.trimmings.byPet(petId),
    queryFn: () => getTrimmingsByPetId(petId),
    enabled: Boolean(petId) && (options?.enabled ?? true),
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};
